package ktfunc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

// VerificationResult holds the outcome of replaying a winner calculation.
type VerificationResult struct {
	CalculatedWinner common.Address
	Match            bool
}

// VerifyWinnerCalculation replays the winner-selection algorithm using the
// provided stake data, epoch range, and block hash. Returns the calculated
// winner so it can be compared against the on-chain result.
//
// The previous `useLinear bool` parameter was removed in Phase 6a — the
// node now always log-normalizes, eliminating the silent operator-config
// divergence that caused multiple voters to disagree in the field.
func VerifyWinnerCalculation(
	stakeDataMap map[common.Address]map[uint64]*UserStakeData,
	epochStart, epochEnd uint64,
	blockHash common.Hash,
) (*VerificationResult, error) {
	if stakeDataMap == nil {
		return nil, fmt.Errorf("stake data map is nil")
	}

	// Calculate minimum stakes over the epoch range (same as VoteAndReward)
	totalMin, stakeDataMinsMap, err := findMinOverBlockRange(epochStart, epochEnd, stakeDataMap)
	if err != nil {
		return nil, fmt.Errorf("failed to find minimum stakes: %w", err)
	}

	// If no valid stakes, return zero address
	if len(stakeDataMinsMap) == 0 || totalMin.Cmp(big.NewInt(0)) == 0 {
		return &VerificationResult{
			CalculatedWinner: common.Address{},
			Match:            false,
		}, nil
	}

	// Calculate probabilities for each wallet
	calculateProbsForEachWallet(stakeDataMinsMap, totalMin)

	// Select winner using the same deterministic algorithm
	winner, err := defaultCalculateWinningWallet(stakeDataMinsMap, blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate winning wallet: %w", err)
	}

	return &VerificationResult{
		CalculatedWinner: winner,
		Match:            winner != (common.Address{}),
	}, nil
}

// VerifyLastWinner fetches on-chain Rwd and Voted events, then replays the winner
// calculation to verify the last rewarded winner was correctly selected.
func VerifyLastWinner(cProps *ConnectionProps) error {
	LogOperationStart("Verifying last winner")

	currentBlock, err := cProps.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get current block: %w", err)
	}

	callOpts := &bind.CallOpts{
		Context: context.Background(),
		Pending: false,
		From:    cProps.MyPubKey,
	}
	interval, err := cProps.Kt.EpochInterval(callOpts)
	if err != nil {
		return fmt.Errorf("failed to get epoch interval: %w", err)
	}

	// Search several epochs back for the most recent Rwd event
	searchRange := uint64(interval) * 5
	if searchRange < 1000 {
		searchRange = 1000
	}
	searchStart := uint64(0)
	if currentBlock > searchRange {
		searchStart = currentBlock - searchRange
	}

	log.Infof("Searching for last Rwd event from block %d to %d", searchStart, currentBlock)

	rwdIter, err := cProps.Kt.FilterRwd(&bind.FilterOpts{
		Start:   searchStart,
		End:     &currentBlock,
		Context: context.Background(),
	})
	if err != nil {
		return fmt.Errorf("failed to filter Rwd events: %w", err)
	}
	defer rwdIter.Close()

	// Find the last (most recent) Rwd event
	var lastRwdAddr common.Address
	var lastRwdAmount *big.Int
	var lastRwdBlock uint64
	found := false
	for rwdIter.Next() {
		evt := rwdIter.Event()
		if evt == nil {
			continue
		}
		lastRwdAddr = evt.Arg0
		lastRwdAmount = evt.Arg1
		lastRwdBlock = evt.Raw.BlockNumber
		found = true
	}
	if err := rwdIter.Error(); err != nil {
		return fmt.Errorf("rwd iterator error: %w", err)
	}

	if !found {
		log.Warnf("No Rwd events found in blocks %d-%d. Nothing to verify. "+
			"If you expect a recent reward, increase the search range or pick a more recent block.",
			searchStart, currentBlock)
		return nil
	}

	rwdEth := new(big.Float).Quo(new(big.Float).SetInt(lastRwdAmount), big.NewFloat(1e18))
	log.Infof("Last Rwd event at block %d", lastRwdBlock)
	log.Infof("  Winner: %s", lastRwdAddr.Hex())
	log.Infof("  Amount: %s wei (%.6f ETH)", lastRwdAmount.String(), rwdEth)

	// Find matching Voted event to get epoch start block and block hash
	votedIter, err := cProps.Kt.FilterVoted(&bind.FilterOpts{
		Start:   searchStart,
		End:     &currentBlock,
		Context: context.Background(),
	})
	if err != nil {
		return fmt.Errorf("failed to filter Voted events: %w", err)
	}
	defer votedIter.Close()

	// Iterate to find the Voted event that triggered this Rwd. Voted events
	// are returned in ascending block order, so the *last* match for the
	// winner address with BlockNumber ≤ lastRwdBlock is the most recent vote
	// preceding the reward — i.e., from the same epoch. If the same wallet
	// won earlier epochs in the search range, those earlier Voted events get
	// overwritten and we end up with the correct (latest) one.
	var votedEpochStart *big.Int
	var votedBlockHash string
	var votedFound bool
	for votedIter.Next() {
		evt := votedIter.Event()
		if evt == nil {
			continue
		}
		if evt.Arg1 == lastRwdAddr && evt.Raw.BlockNumber <= lastRwdBlock {
			votedEpochStart = evt.Arg0
			votedBlockHash = evt.Arg2
			votedFound = true
		}
	}
	if err := votedIter.Error(); err != nil {
		return fmt.Errorf("voted iterator error: %w", err)
	}

	if !votedFound {
		return fmt.Errorf("no matching Voted event found for winner %s", lastRwdAddr.Hex())
	}

	log.Infof("Matching Voted event found")
	log.Infof("  Epoch start block: %d", votedEpochStart.Uint64())
	log.Infof("  Block hash used: %s", votedBlockHash)

	endBlock := new(big.Int).Add(votedEpochStart, big.NewInt(int64(interval)))
	log.Infof("  Epoch end block: %d", endBlock.Uint64())

	// Get contract creation block
	creationBlockUint64, err := GetContractCreationBlock(cProps)
	if err != nil {
		return fmt.Errorf("failed to get contract creation block: %w", err)
	}
	creationBlock := new(big.Int).SetUint64(creationBlockUint64)

	// Gather stakes from creation to end of that epoch
	log.Infof("Gathering stakes from block %d to %d", creationBlock.Uint64(), endBlock.Uint64())
	stakeDataMap, err := GatherStakesAndWithdraws(cProps, cProps.Kt, creationBlock, endBlock)
	if err != nil {
		return fmt.Errorf("failed to gather stakes: %w", err)
	}

	// Replay the winner calculation
	blockHash := common.HexToHash(votedBlockHash)
	result, err := VerifyWinnerCalculation(
		stakeDataMap,
		votedEpochStart.Uint64(),
		endBlock.Uint64(),
		blockHash,
	)
	if err != nil {
		return fmt.Errorf("failed to verify winner calculation: %w", err)
	}

	// Compare
	log.Info("")
	log.Info("========== VERIFICATION RESULT ==========")
	log.Infof("  On-chain winner:    %s", lastRwdAddr.Hex())
	log.Infof("  Calculated winner:  %s", result.CalculatedWinner.Hex())

	if result.CalculatedWinner == lastRwdAddr {
		log.Info("  VERIFIED: The last winner was correctly selected.")
	} else {
		log.Warn("  MISMATCH: The calculated winner does not match the on-chain winner!")
		log.Warn("  Possible causes:")
		log.Warn("    - the epoch was rewarded before the min-stake or withdraw-erasure")
		log.Warn("      fix shipped, so the on-chain winner was selected by a pre-fix algorithm;")
		log.Warn("    - the voting node was running with a different code version (check banner);")
		log.Warn("    - contract state (e.g., declines) has changed since the epoch was rewarded.")
	}
	log.Info("==========================================")

	return nil
}
