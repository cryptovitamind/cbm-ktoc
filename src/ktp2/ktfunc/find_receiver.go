package ktfunc

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math"
	"math/big"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

// StakeEvent represents a stake event
type StakeEvent struct {
	Addr   common.Address
	Amount *big.Int
	Block  uint64
}

// WithdrawEvent represents a withdraw event
type WithdrawEvent struct {
	Addr   common.Address
	Amount *big.Int
	Block  uint64
}

// ChunkEvents holds events for block chunk
type ChunkEvents struct {
	StakeEvents    []StakeEvent
	WithdrawEvents []WithdrawEvent
}

// Define as a variable holding a function
var calcWinningWallet = defaultCalculateWinningWallet

var GatherStakesAndWithdraws = realGatherStakesAndWithdraws

// Exported for testing purposes if needed
func SetCalculateWinningWallet(f func(map[common.Address]*UserStakeData, common.Hash) (common.Address, error)) {
	calcWinningWallet = f
}

// VoteAndReward determines the winning wallet, votes for it, and rewards it if conditions are met.
// Returns the winning wallet address or a zero address if no winner is determined.
func VoteAndReward(cProps *ConnectionProps) error {
	log.Debugf("Voting and rewarding")

	// Get current block
	currentBlockHeader, err := GetCurrentBlock(cProps)
	if err != nil {
		log.Errorf("Failed to get current block: %v", err)
		return fmt.Errorf("failed to get current block: %w", err)
	}
	currentNum := currentBlockHeader.Number

	// Get current start block and epoch interval from contract (cached
	// with a short TTL via state_cache.go to avoid re-querying every
	// epoch).
	startBlock, err := cachedStartBlock(cProps)
	if err != nil {
		log.Errorf("Failed to get start block: %v", err)
		return fmt.Errorf("failed to get start block: %w", err)
	}
	interval, err := cachedEpochInterval(cProps)
	if err != nil {
		log.Errorf("Failed to get epoch interval: %v", err)
		return fmt.Errorf("failed to get epoch interval: %w", err)
	}
	if interval <= 0 {
		log.Errorf("Invalid epoch interval: %d", interval)
		return fmt.Errorf("invalid epoch interval")
	}

	// Calculate end block
	endBlock := new(big.Int).Add(startBlock, big.NewInt(int64(interval)))

	// Check if it's time to vote
	if !IsTimeToVote(endBlock, currentBlockHeader) {
		status := fmt.Sprintf("Not time to vote yet - Current block: %d, End block: %d", currentNum.Uint64(), endBlock.Uint64())
		fmt.Fprintf(os.Stdout, "\r\033[1;36m%s\033[0m", status) // Cyan text, reset color
		os.Stdout.Sync()                                        // Ensure it flushes immediately
		return nil                                              // Not an error, just not time yet
	}

	// Log the end-of-epoch ETH balance
	if err := printEndEpochKtEthBalance(cProps, endBlock); err != nil {
		log.Warnf("Failed to print end epoch balance: %v", err)
	}

	// Get contract creation block
	creationBlockUint64, err := GetContractCreationBlock(cProps)
	if err != nil {
		log.Errorf("Failed to get contract creation block: %v", err)
		return fmt.Errorf("failed to get contract creation block: %w", err)
	}
	creationBlock := new(big.Int).SetUint64(creationBlockUint64)
	log.Infof("Gathering stakes from creation block %d to end block %d", creationBlock.Uint64(), endBlock.Uint64())

	// Gather stake and withdrawal events
	stakeDataMap, err := GatherStakesAndWithdraws(cProps, cProps.Kt, creationBlock, endBlock)
	if err != nil {
		log.Errorf("Failed to gather stakes and withdraws: %v", err)
		return fmt.Errorf("failed to gather stakes: %w", err)
	}

	// Calculate minimum stakes over the block range
	totalMin, stakeDataMinsMap, err := findMinOverBlockRange(startBlock.Uint64(), endBlock.Uint64(), stakeDataMap)
	if err != nil {
		log.Errorf("Failed to find minimum stakes: %v", err)
		return fmt.Errorf("failed to find minimum stakes: %w", err)
	}
	if totalMin.Cmp(big.NewInt(0)) == 0 {
		log.Warn("No valid stakes found after minimum calculation.")
	}

	// Print minimum stakes
	printAllStakes(stakeDataMinsMap)

	// Filter out declined stakers
	if err := filterDeclinedStakers(stakeDataMinsMap, cProps); err != nil {
		log.Errorf("Failed to filter declined stakers: %v", err)
		return fmt.Errorf("failed to filter declined stakers: %w", err)
	}

	// Recalculate totalMin after filtering
	totalMin = big.NewInt(0)
	for _, data := range stakeDataMinsMap {
		totalMin.Add(totalMin, data.StakeAmount)
	}

	// Calculate probabilities for each wallet
	calculateProbsForEachWallet(stakeDataMinsMap, totalMin)
	if totalMin.Cmp(big.NewInt(0)) == 0 {
		log.Warn("No valid stakes detected - will vote for dead address.")
	}

	// Vote and potentially reward the winner
	winner, err := calculateVoteAndReward(stakeDataMinsMap, startBlock, endBlock, cProps, totalMin)
	if err != nil {
		log.Errorf("Failed to vote and reward: %v", err)
		return fmt.Errorf("failed to vote and reward: %w", err)
	}

	if winner != (common.Address{}) {
		log.Debugf("Winner determined: %s", winner.Hex())
	} else {
		log.Warn("No winner determined")
	}
	return nil
}

func calculateVoteAndReward(
	stakeDataMinsMap map[common.Address]*UserStakeData,
	epochStartBlock, endEpochBlockNumber *big.Int,
	cProps *ConnectionProps,
	totalMin *big.Int) (common.Address, error) {

	log.Debugf("Calculating vote and reward")

	// Validate inputs
	if cProps == nil || cProps.Client == nil || cProps.Kt == nil {
		log.Errorf("Invalid ConnectionProps - Client: %v, KT: %v", cProps.Client, cProps.Kt)
		return common.Address{}, fmt.Errorf("invalid ConnectionProps: client or KT instance is nil")
	}
	if stakeDataMinsMap == nil {
		log.Warn("Stake data map is nil")
		return common.Address{}, fmt.Errorf("stake data map is nil")
	}
	if epochStartBlock == nil || endEpochBlockNumber == nil {
		log.Errorf("Invalid block numbers - Start: %v, End: %v", epochStartBlock, endEpochBlockNumber)
		return common.Address{}, fmt.Errorf("epoch start or end block is nil")
	}

	// Seed the lottery from a block N positions past the epoch end. A
	// single-block lookahead (endBlock+1) is vulnerable to short reorgs:
	// two nodes voting milliseconds apart could see different hashes for
	// the same block number if one observed a re-org tip. Voting at
	// endBlock+ConfirmationDepth trades ~12s/depth of latency for much
	// stronger cross-node agreement on the seed.
	confirmationDepth := cProps.ConfirmationDepth
	if confirmationDepth == 0 {
		confirmationDepth = DefaultConfirmationDepth
	}
	nextBlockNumber := new(big.Int).Add(endEpochBlockNumber, new(big.Int).SetUint64(confirmationDepth))
	log.Printf("Epoch start block: %d, Seed block (endBlock+%d): %d",
		epochStartBlock.Uint64(), confirmationDepth, nextBlockNumber.Uint64())

	// Wait for the next block to be available
	var nextBlock *types.Header
	var err error
	for {
		nextBlock, err = cProps.Client.HeaderByNumber(context.Background(), nextBlockNumber)
		if err == nil && nextBlock != nil {
			break
		}
		if err != nil && !strings.Contains(err.Error(), "not found") {
			log.Errorf("Failed to get next block %d: %v", nextBlockNumber.Uint64(), err)
			return common.Address{}, fmt.Errorf("failed to get next block: %w", err)
		}
		log.Infof("Next block %d not available yet, waiting...", nextBlockNumber.Uint64())
		time.Sleep(1 * time.Second) // Adjust sleep duration as needed, e.g., based on chain block time
	}
	log.Infof("Next block: %d", nextBlock.Number)

	// Calculate winning wallet
	if totalMin.Cmp(big.NewInt(0)) == 0 {
		log.Debug("Total minimum stake is zero - Will select dead address as winner.")
	}

	winner, err := calcWinningWallet(stakeDataMinsMap, nextBlock.Hash())
	if err != nil {
		log.Errorf("Failed to calculate winning wallet: %v", err)
		return common.Address{}, fmt.Errorf("failed to calculate winning wallet: %w", err)
	}
	if winner == (common.Address{}) {
		winner = common.Address{}
		log.Warn("No winner determined - Falling back to dead address")
	}
	if totalMin.Cmp(big.NewInt(0)) == 0 {
		log.Infof("Winner selected: %s (no stakes fallback)", winner.Hex())
	} else {
		log.Infof("Winner selected: %s", winner.Hex())
	}

	// Vote for the winner
	if err := vote(cProps, winner, nextBlock.Hash().String()); err != nil {
		log.Warnf("Failed to vote for %s: %v", winner.Hex(), err)
		// Continue despite voting failure
	}

	// Get vote count and required votes
	voteCount, voteRequired, err := getVoteCountAndRequired(cProps, epochStartBlock, winner)
	if err != nil {
		log.Errorf("Failed to get vote count and required votes: %v", err)
		return winner, fmt.Errorf("failed to get vote info: %w", err)
	}
	log.Infof("Vote status - Count: %d, Required: %d", voteCount, voteRequired)

	// Reward if enough votes
	if voteCount >= voteRequired {
		if err := rewardWinningWallet(cProps, winner, totalMin); err != nil {
			log.Errorf("Failed to reward %s: %v", winner.Hex(), err)
			return winner, fmt.Errorf("failed to reward winner: %w", err)
		}
		log.Infof("Winner %s rewarded successfully", winner.Hex())
	}

	return winner, nil
}

// printEndEpochKtEthBalance logs the KT contract's ETH balance at the specified end epoch block.
func printEndEpochKtEthBalance(cProps *ConnectionProps, endBlock *big.Int) error {
	log.Debugf("Printing end epoch balance")

	balance, err := cProps.Client.BalanceAt(context.Background(), cProps.KtAddr, endBlock)
	if err != nil {
		log.Errorf("Failed to get epoch end balance: %v", err)
		return fmt.Errorf("failed to get epoch end balance: %w", err)
	}

	log.Infof("Epoch end balance: %s ETH", new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18)).String())
	return nil
}

// getStartAndEndEpochBlocks retrieves the start and end block numbers for the current epoch.
// Returns both blocks or nil values if an error occurs.
func getStartAndEndEpochBlocks(cProps *ConnectionProps) (*big.Int, *big.Int, error) {
	log.Debugf("Fetching epoch blocks")

	// Validate KT instance
	if cProps.Kt == nil {
		log.Errorf("KT instance is nil")
		return nil, nil, fmt.Errorf("KT instance is nil")
	}

	// Prepare call options
	callOpts := &bind.CallOpts{
		Context: context.Background(),
		Pending: false,
		From:    cProps.MyPubKey,
	}

	// Get start block
	startBlock, err := cProps.Kt.StartBlock(callOpts)
	if err != nil {
		log.Errorf("Failed to get start block: %v", err)
		return nil, nil, fmt.Errorf("failed to get start block: %w", err)
	}

	// Get epoch interval
	epochInterval, err := cProps.Kt.EpochInterval(callOpts)
	if err != nil {
		log.Errorf("Failed to get epoch interval: %v", err)
		return nil, nil, fmt.Errorf("failed to get epoch interval: %w", err)
	}

	// Calculate end block
	endBlock := new(big.Int).Add(startBlock, big.NewInt(int64(epochInterval)))

	// Log results
	log.Debugf("Start block: %d", startBlock.Uint64())
	log.Debugf("End block: %d", endBlock.Uint64())

	return startBlock, endBlock, nil
}

// getCurrentBlock retrieves the latest block from the blockchain.
// Returns the block or nil if an error occurs.
func GetCurrentBlock(cProps *ConnectionProps) (*types.Header, error) {
	log.Debug("Fetching current block")

	// Fetch the latest block (nil block number means latest)
	block, err := cProps.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Errorf("Failed to retrieve current block: %v", err)
		return nil, fmt.Errorf("failed to get current block: %w", err)
	}

	// Log block details with pretty formatting
	log.Debugf("Block number: %d", block.Number)
	log.Debugf("Block hash: %s", block.Hash().Hex())

	return block, nil
}

// defaultCalculateWinningWallet selects the winning wallet based on stake probabilities and a random seed.
// Returns the winning address or a zero address if no winner can be determined.
func defaultCalculateWinningWallet(
	stakeDataMinsMap map[common.Address]*UserStakeData,
	randomNumber common.Hash) (common.Address, error) {

	LogOperationStart("Calculating winning wallet")

	// Validate inputs
	if stakeDataMinsMap == nil {
		log.Warn("Stake data map is nil - No stakes to process")
		return common.Address{}, fmt.Errorf("stake data map is nil")
	}
	if len(stakeDataMinsMap) == 0 {
		log.Info("No stakes found in the map. Defaulting to dead address.")
		return common.Address{}, nil // Return zero as dead
	}

	// Convert block hash to a float between 0 and 1
	randInt := new(big.Int).SetBytes(randomNumber[:])
	denominator := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
	randFloat := new(big.Float).Quo(new(big.Float).SetInt(randInt), new(big.Float).SetInt(denominator))
	randValue, _ := randFloat.Float64()
	log.Infof("Random value from hash: %.6f", randValue)

	// Sort addresses for deterministic selection
	var addresses []common.Address
	for addr := range stakeDataMinsMap {
		addresses = append(addresses, addr)
	}
	sort.Slice(addresses, func(i, j int) bool { return addresses[i].Hex() < addresses[j].Hex() })
	log.Infof("Total addresses: %d", len(addresses))

	// Calculate cumulative probability and find winner
	cumulativeProb := new(big.Float)
	for i, addr := range addresses {
		stakeData := stakeDataMinsMap[addr]
		if stakeData == nil || stakeData.Prob == nil {
			log.Warnf("Invalid stake data for %s - Skipping", addr.Hex())
			continue
		}

		cumulativeProb.Add(cumulativeProb, stakeData.Prob)
		cumProb, _ := cumulativeProb.Float64()
		probFloat, _ := stakeData.Prob.Float64()
		log.Debugf("Address %d - %s, Stake: %s, Prob: %.6f, CumProb: %.6f",
			i+1, addr.Hex(), stakeData.StakeAmount.String(), probFloat, cumProb)

		if randFloat.Cmp(cumulativeProb) < 0 {
			return addr, nil
		}
	}

	// Verify probability sum
	totalProb, _ := cumulativeProb.Float64()
	if totalProb < 1.0-0.000001 || totalProb > 1.0+0.000001 { // Allow small float precision error
		log.Warnf("Cumulative probability %f does not sum to approximately 1", totalProb)
	}

	// Fallback to last address if no winner found (shouldn't happen with valid probs)
	if len(addresses) > 0 {
		lastAddr := addresses[len(addresses)-1]
		log.Infof("No winner found within range - Fallback to last address: %s", lastAddr.Hex())
		return lastAddr, nil
	}

	log.Info("No valid addresses available - Returning dead address")
	return common.Address{}, nil
}

func filterDeclinedStakers(stakeDataMinsMap map[common.Address]*UserStakeData, cProps *ConnectionProps) error {
	if cProps.DeclinesCache == nil {
		cProps.DeclinesCache = make(map[common.Address]bool)
	}
	for addr := range stakeDataMinsMap {
		declined, cached := cProps.DeclinesCache[addr]
		if !cached {
			var err error
			declined, err = cProps.Kt.Declines(&bind.CallOpts{}, addr)
			if err != nil {
				return fmt.Errorf("failed to check declines for %s: %w", addr.Hex(), err)
			}
			cProps.DeclinesCache[addr] = declined
		}
		if declined {
			delete(stakeDataMinsMap, addr)
			log.Debugf("Filtered out declined staker: %s", addr.Hex())
		}
	}
	return nil
}

// logNormalizeProbabilities applies log-scale normalization to the
// probabilities in stakeDataMinsMap: prob_i = log(1+stake_i) / sum(log(1+stake_j)).
//
// Larger stakes still win more often, but the log compresses disparities so
// a wallet with 1000x more stake is only modestly more likely to win — not
// 1000x — preventing whales from drowning out smaller stakers.
//
// We use log1p (log(1+x)) instead of log(x) so a 1-wei stake gets a small
// but non-zero probability instead of being silently excluded (log(1)=0).
func logNormalizeProbabilities(stakeDataMinsMap map[common.Address]*UserStakeData) error {
	if stakeDataMinsMap == nil || len(stakeDataMinsMap) == 0 {
		return nil
	}

	sumLog := new(big.Float)
	validCount := 0

	// First pass: compute sum of log(1+stake) for valid stakes
	for _, stakeData := range stakeDataMinsMap {
		if stakeData.StakeAmount == nil || stakeData.StakeAmount.Cmp(big.NewInt(0)) <= 0 {
			stakeData.Prob = new(big.Float).SetFloat64(0)
			continue
		}
		stakeFloat := new(big.Float).SetInt(stakeData.StakeAmount)
		stakeF64, _ := stakeFloat.Float64()
		logStake := math.Log1p(stakeF64)
		logStakeBig := new(big.Float).SetFloat64(logStake)
		sumLog.Add(sumLog, logStakeBig)
		validCount++
	}

	if validCount == 0 || sumLog.Cmp(new(big.Float).SetFloat64(0)) == 0 {
		return nil
	}

	// Second pass: set probabilities
	for _, stakeData := range stakeDataMinsMap {
		if stakeData.StakeAmount == nil || stakeData.StakeAmount.Cmp(big.NewInt(0)) <= 0 {
			continue
		}
		stakeFloat := new(big.Float).SetInt(stakeData.StakeAmount)
		stakeF64, _ := stakeFloat.Float64()
		logStake := math.Log1p(stakeF64)
		logStakeBig := new(big.Float).SetFloat64(logStake)
		stakeData.Prob = new(big.Float).Quo(logStakeBig, sumLog)
	}

	return nil
}

// calculateProbsForEachWallet assigns a log-normalized probability to each
// wallet in stakeDataMinsMap. The linear-mode option (and the `-linearProbs`
// CLI flag that drove it) was removed in Phase 6a — it was a per-operator
// switch with no on-chain record, so different operators silently computed
// different winners. Log normalization is now the only supported mode:
// larger stakers still win more often, but with strongly diminishing
// returns so whales don't drown out smaller wallets.
func calculateProbsForEachWallet(stakeDataMinsMap map[common.Address]*UserStakeData, totalMin *big.Int) bool {
	_ = totalMin // kept in signature for call-site symmetry; log path doesn't use it.

	if stakeDataMinsMap == nil || len(stakeDataMinsMap) == 0 {
		log.Warn("Stake data map is nil or empty - Cannot calculate probabilities")
		return false
	}

	if err := logNormalizeProbabilities(stakeDataMinsMap); err != nil {
		log.Errorf("Failed to log normalize probabilities: %v", err)
		return false
	}

	foundSomething := false
	for addr, stakeData := range stakeDataMinsMap {
		if stakeData.Prob != nil {
			foundSomething = true
			log.Debugf("Address: %s, Log-normalized Probability: %f\n", addr.Hex(), stakeData.Prob)
		}
	}
	return foundSomething
}

func printAllStakes(stakeDataMinsMap map[common.Address]*UserStakeData) {
	for addr, minStake := range stakeDataMinsMap {
		log.Debugf("Address: %s, Min Stake: %s\n", addr.Hex(), minStake.StakeAmount.String())
	}
}

// IsTimeToVote determines if the current block has reached or exceeded the epoch end block.
// Returns true if voting time has arrived, false otherwise.
// BlockNumberer defines the interface for objects that provide a block number
type BlockNumberer interface {
	Number() *big.Int
}

func IsTimeToVote(endBlock *big.Int, blockHeader *types.Header) bool {
	log.Debugf("Checking voting time")

	// Validate inputs
	if endBlock == nil {
		log.Errorf("End block is nil")
		return false
	}
	if blockHeader == nil {
		log.Errorf("Current block is nil")
		return false
	}

	currentNum := blockHeader.Number
	if currentNum == nil {
		log.Errorf("Current block number is nil")
		return false
	}

	// Compare blocks and log result
	if endBlock.Cmp(currentNum) > 0 {
		log.Debugf("Not yet time to vote - Current block: %d, End block: %d", currentNum.Uint64(), endBlock.Uint64())
		return false
	}

	log.Infof("\nTime to vote - Current block: %d, End block: %d", currentNum.Uint64(), endBlock.Uint64())
	return true
}

func rewardWinningWallet(cProps *ConnectionProps, winner common.Address, totalMin *big.Int) error {
	log.Printf("Rewarding winning wallet: %s", winner.Hex())

	auth, err := NewTransactor(cProps)
	if err != nil {
		return fmt.Errorf("failed to create transactor: %v", err)
	}

	// Get the winner's balance before the reward
	balanceBefore, err := cProps.Client.BalanceAt(context.Background(), winner, nil)
	if err != nil {
		return fmt.Errorf("failed to get winner's balance before reward: %v", err)
	}

	// Convert balance before from wei to ETH
	weiToEthBefore := new(big.Float).SetInt(balanceBefore)
	balanceBeforeEth := new(big.Float).Quo(weiToEthBefore, big.NewFloat(1e18))

	// Get the contract balance (this will be the reward amount)
	rewardAmount, err := cProps.Client.BalanceAt(context.Background(), cProps.KtAddr, nil)
	if err != nil {
		return fmt.Errorf("failed to get contract balance: %v", err)
	}

	// Get total OC fees owed to subtract from reward amount (TTL-cached).
	tlOcFees, err := cachedTlOcFees(cProps)
	if err != nil {
		return fmt.Errorf("failed to get total OC fees: %v", err)
	}

	// Calculate reward amount as balance minus OC fees, set to 0 if negative
	rewardAmount = new(big.Int).Sub(rewardAmount, tlOcFees)
	if rewardAmount.Sign() < 0 {
		rewardAmount = big.NewInt(0)
	}

	if totalMin.Cmp(big.NewInt(0)) == 0 {
		rewardAmount = big.NewInt(0)
		log.Warn("No stakes - rewarding zero amount to prevent unintended transfer.")
	}

	// Convert reward amount from wei to ETH
	weiToEthReward := new(big.Float).SetInt(rewardAmount)
	rewardEth := new(big.Float).Quo(weiToEthReward, big.NewFloat(1e18))
	log.Printf("Contract balance (reward amount): %.6f ETH", rewardEth)

	// Call the rwd function to send the reward
	tx, err := cProps.Kt.Rwd(auth, winner, rewardAmount)
	if err != nil {
		if strings.Contains(err.Error(), "Epoch incomplete") {
			log.Warnf("Epoch incomplete - Not rewarding. Most likely another node got to it first. %v", err)
			return nil
		} else {
			return fmt.Errorf("failed to call rwd function: %v", err)
		}
	}

	log.Printf("Reward transaction sent: %s", tx.Hash().Hex())

	// Wait for the transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), cProps.Client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for reward transaction to be mined: %v", err)
	}

	log.Debugf("Reward transaction mined in block: %d", receipt.BlockNumber.Uint64())

	// Wait for additional blocks to pass
	err = WaitForBlocks(cProps)
	if err != nil {
		return fmt.Errorf("failed to wait for additional blocks: %v", err)
	}

	log.Debugf("Reward completed. %d blocks have passed.", cProps.BlocksToWait)

	// Verify the winner's balance after the reward
	balanceAfter, err := cProps.Client.BalanceAt(context.Background(), winner, nil)
	if err != nil {
		return fmt.Errorf("failed to get winner's balance after reward: %v", err)
	}

	// Convert balance after from wei to ETH
	weiToEthAfter := new(big.Float).SetInt(balanceAfter)
	balanceAfterEth := new(big.Float).Quo(weiToEthAfter, big.NewFloat(1e18))

	// Log all three: balance before, reward amount, and balance after
	log.Printf("Winner balance before: %.6f ETH | Awarded: %.6f ETH | New balance: %.6f ETH",
		balanceBeforeEth, rewardEth, balanceAfterEth)

	return nil
}

func WaitForBlocks(cProps *ConnectionProps) error {
	currentBlock, err := cProps.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get current block number: %v", err)
	}

	targetBlock := currentBlock + cProps.BlocksToWait

	log.Printf("Waiting for %d blocks to pass...", cProps.BlocksToWait)
	for currentBlock < targetBlock {
		time.Sleep(TimeToWaitForBlocks)
		currentBlock, err = cProps.Client.BlockNumber(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get current block number: %v", err)
		}

		status := fmt.Sprintf("Current block: %d, Target block: %d", currentBlock, targetBlock)
		fmt.Fprintf(os.Stdout, "\r\033[1;36m%s\033[0m", status) // Cyan text, reset color
		os.Stdout.Sync()                                        // Ensure it flushes immediately
	}
	log.Printf("")
	return nil
}

func NewTransactor(cProps *ConnectionProps) (*bind.TransactOpts, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(cProps.MyPrivateKey, cProps.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create trasnactor: %v", err)
	}

	if cProps.GasLimit > DefaultGasLimit {
		auth.GasLimit = cProps.GasLimit
	}

	return auth, nil
}

// getVoteCountAndRequired retrieves the current vote count and required votes for a winner.
// Returns the vote count, required votes, and an error if the operation fails.
func getVoteCountAndRequired(cProps *ConnectionProps, epochStartBlock *big.Int, winner common.Address) (voteCount uint16, voteRequired uint16, err error) {
	log.Debugf("Retrieving vote count and required votes")

	// Validate inputs
	if cProps == nil || cProps.Kt == nil {
		log.Errorf("Invalid ConnectionProps - KT instance: %v", cProps.Kt)
		return 0, 0, fmt.Errorf("invalid ConnectionProps: KT instance is nil")
	}
	if epochStartBlock == nil {
		log.Errorf("Epoch start block is nil")
		return 0, 0, fmt.Errorf("epoch start block is nil")
	}

	// Prepare call options
	callOpts := &bind.CallOpts{
		Context: context.Background(),
		Pending: false,
		From:    cProps.MyPubKey,
	}

	// Get vote count
	voteCount, err = cProps.Kt.BlockRwd(callOpts, epochStartBlock, winner)
	if err != nil {
		log.Errorf("Failed to get vote count for %s at block %d: %v", winner.Hex(), epochStartBlock.Uint64(), err)
		return 0, 0, fmt.Errorf("failed to get vote count: %w", err)
	}

	// Get required votes (TTL-cached: ConsensusReq is contract config that
	// changes via on-chain admin action and is otherwise stable).
	voteRequired, err = cachedConsensusReq(cProps)
	if err != nil {
		log.Errorf("Failed to get required votes: %v", err)
		return 0, 0, fmt.Errorf("failed to get required votes: %w", err)
	}

	// Log results
	log.Infof("Vote info - Block: %d, Count: %d, Required: %d", epochStartBlock.Uint64(), voteCount, voteRequired)
	return voteCount, voteRequired, nil
}

func vote(cProps *ConnectionProps, recipient common.Address, data string) error {
	auth, err := NewTransactor(cProps)
	if err != nil {
		return fmt.Errorf("failed to create function: %v", err)
	}

	// Call the vote function
	tx, err := cProps.Kt.Vote(auth, recipient, data)
	if err != nil {
		return fmt.Errorf("failed to vote: %v", err)
	}

	log.Debugf("Vote transaction sent: %s, %s", tx.Hash().Hex(), data)

	// Wait for the transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), cProps.Client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for vote transaction to be mined: %v", err)
	}

	log.Debugf("Vote transaction mined in block: %d", receipt.BlockNumber.Uint64())

	// Wait for additional blocks to pass
	err = WaitForBlocks(cProps)
	if err != nil {
		return fmt.Errorf("failed to wait for additional blocks: %v", err)
	}

	log.Printf("Vote function completed. %d blocks have passed.", cProps.BlocksToWait)
	return nil
}

func debugRawLogs(cProps *ConnectionProps, start, end uint64) {
	filter := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(start)),
		ToBlock:   big.NewInt(int64(end)),
		Addresses: []common.Address{cProps.KtAddr},
	}
	logs, err := cProps.Client.FilterLogs(context.Background(), filter)
	if err != nil {
		log.Errorf("Failed to fetch raw logs: %v", err)
		return
	}
	log.Infof("Found %d raw logs for contract %s from block %d to %d", len(logs), cProps.KtAddr.Hex(), start, end)
	for _, l := range logs {
		log.Infof("Raw log - Block: %d, Topics: %v, Data: %x", l.BlockNumber, l.Topics, l.Data)
	}
}

// GatherStakesAndWithdraws collects stake and withdrawal events for a KT contract from block startBlock to endBlock.
// Returns a map of address to block-specific stake data or an error if filtering fails.
func realGatherStakesAndWithdraws(cProps *ConnectionProps, kt Ktv2Interface, startBlock *big.Int, endBlock *big.Int) (map[common.Address]map[uint64]*UserStakeData, error) {
	log.Debugf("Gathering stakes and withdrawals")
	// Validate inputs
	if kt == nil {
		log.Errorf("KT contract instance is nil")
		return nil, fmt.Errorf("KT contract instance is nil")
	}
	if startBlock == nil || endBlock == nil {
		log.Errorf("Invalid block range - Start: %v, End: %v", startBlock, endBlock)
		return nil, fmt.Errorf("invalid block range: start and end blocks must be non-nil")
	}
	if startBlock.Cmp(endBlock) > 0 {
		log.Errorf("Start block %d exceeds end block %d", startBlock.Uint64(), endBlock.Uint64())
		return nil, fmt.Errorf("start block %d exceeds end block %d", startBlock.Uint64(), endBlock.Uint64())
	}
	// Ensure cache directory exists
	cacheDir := "cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Errorf("Failed to create cache directory: %v", err)
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Construct database file name using first 7 characters of contract address
	dbName := fmt.Sprintf("%s/%s.db", cacheDir, cProps.KtAddr.Hex()[:7])

	// Open DB
	db, err := bbolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Errorf("Failed to open database %s: %v", dbName, err)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	// Buckets:
	//   chunks — key = 8-byte BE chunkStart, value = gob ChunkEvents
	//   meta   — keys:
	//     "tip"            = 8-byte BE highest contiguously-processed block
	//     "schema_version" = 4-byte BE uint32; if absent or older than
	//                        cacheSchemaVersion, both buckets are wiped on
	//                        open so the node self-heals across upgrades.
	// Chunk endings are NOT in the key. A chunk for chunkStart contains all
	// events from [chunkStart, min(chunkStart+chunkSize-1, tip)]. When a new
	// call extends past the prior tip into the chunk, we fetch only the
	// uncovered tail and merge it into the existing chunk.
	if err := migrateOrInitCacheSchema(db); err != nil {
		log.Errorf("Failed to init/migrate cache schema: %v", err)
		return nil, fmt.Errorf("failed to init/migrate cache schema: %w", err)
	}
	chunkSize := uint64(cProps.ChunkSize)
	if chunkSize == 0 {
		chunkSize = uint64(DefaultChunkSize)
	}

	startU := startBlock.Uint64()
	endU := endBlock.Uint64()

	// Read tip.
	var tip uint64
	_ = db.View(func(tx *bbolt.Tx) error {
		if b := tx.Bucket([]byte("meta")); b != nil {
			if v := b.Get([]byte("tip")); len(v) == 8 {
				tip = binary.BigEndian.Uint64(v)
			}
		}
		return nil
	})
	log.Debugf("Cache tip on entry: %d", tip)

	chunkKey := func(chunkStart uint64) []byte {
		k := make([]byte, 8)
		binary.BigEndian.PutUint64(k, chunkStart)
		return k
	}
	loadChunk := func(chunkStart uint64) (ChunkEvents, bool, error) {
		var chunk ChunkEvents
		var found bool
		err := db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("chunks"))
			v := b.Get(chunkKey(chunkStart))
			if v == nil {
				return nil
			}
			if decErr := gob.NewDecoder(bytes.NewReader(v)).Decode(&chunk); decErr != nil {
				return decErr
			}
			found = true
			return nil
		})
		return chunk, found, err
	}
	storeChunk := func(chunkStart uint64, chunk ChunkEvents) error {
		return db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("chunks"))
			var buf bytes.Buffer
			if err := gob.NewEncoder(&buf).Encode(chunk); err != nil {
				return err
			}
			return b.Put(chunkKey(chunkStart), buf.Bytes())
		})
	}
	advanceTip := func(newTip uint64) error {
		if newTip <= tip {
			return nil
		}
		err := db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("meta"))
			buf := make([]byte, 8)
			binary.BigEndian.PutUint64(buf, newTip)
			return b.Put([]byte("tip"), buf)
		})
		if err == nil {
			tip = newTip
		}
		return err
	}

	fetchAndStore := func(chunkStart, fetchStart, chunkEnd uint64, existing ChunkEvents) error {
		if cProps.QueryDelay > 0 {
			time.Sleep(cProps.QueryDelay)
		}
		var s []StakeEvent
		var w []WithdrawEvent
		if err := queryChunkWithRetry(cProps, kt, fetchStart, chunkEnd, &s, &w); err != nil {
			return err
		}
		log.Infof("Fetched %d-%d: %d stakes, %d withdraws", fetchStart, chunkEnd, len(s), len(w))
		merged := ChunkEvents{
			StakeEvents:    append(existing.StakeEvents, s...),
			WithdrawEvents: append(existing.WithdrawEvents, w...),
		}
		if err := storeChunk(chunkStart, merged); err != nil {
			return fmt.Errorf("failed to store chunk %d: %w", chunkStart, err)
		}
		return advanceTip(chunkEnd)
	}

	var stakeEvents []StakeEvent
	var withdrawEvents []WithdrawEvent
	for chunkStart := startU; chunkStart <= endU; chunkStart += chunkSize {
		chunkEnd := chunkStart + chunkSize - 1
		if chunkEnd > endU {
			chunkEnd = endU
		}

		switch {
		case chunkEnd <= tip:
			// Fully cached.
			chunk, found, err := loadChunk(chunkStart)
			if err != nil || !found {
				if err != nil {
					log.Warnf("Failed to load cached chunk %d (expected hit, tip=%d): %v - re-querying", chunkStart, tip, err)
				} else {
					log.Warnf("Missing cached chunk %d (expected hit, tip=%d) - re-querying", chunkStart, tip)
				}
				if err := fetchAndStore(chunkStart, chunkStart, chunkEnd, ChunkEvents{}); err != nil {
					return nil, err
				}
				chunk, _, _ = loadChunk(chunkStart)
			} else {
				log.Infof("Cache HIT chunk %d (covers up to %d)", chunkStart, chunkEnd)
			}
			stakeEvents = append(stakeEvents, chunk.StakeEvents...)
			withdrawEvents = append(withdrawEvents, chunk.WithdrawEvents...)

		case chunkStart > tip:
			// New chunk: fetch the whole requested range from scratch.
			log.Infof("Fetching new chunk %d-%d (tip=%d)", chunkStart, chunkEnd, tip)
			if err := fetchAndStore(chunkStart, chunkStart, chunkEnd, ChunkEvents{}); err != nil {
				return nil, err
			}
			chunk, _, _ := loadChunk(chunkStart)
			stakeEvents = append(stakeEvents, chunk.StakeEvents...)
			withdrawEvents = append(withdrawEvents, chunk.WithdrawEvents...)

		default:
			// Partially cached: chunkStart ≤ tip < chunkEnd. Extend the chunk
			// by fetching only the uncovered tail (tip+1, chunkEnd).
			existing, found, err := loadChunk(chunkStart)
			if err != nil || !found {
				log.Warnf("Expected partial chunk %d in cache (tip=%d) but not found - re-fetching full chunk", chunkStart, tip)
				existing = ChunkEvents{}
				if err := fetchAndStore(chunkStart, chunkStart, chunkEnd, existing); err != nil {
					return nil, err
				}
			} else {
				log.Infof("Extending chunk %d: cached up to %d, fetching %d-%d", chunkStart, tip, tip+1, chunkEnd)
				if err := fetchAndStore(chunkStart, tip+1, chunkEnd, existing); err != nil {
					return nil, err
				}
			}
			chunk, _, _ := loadChunk(chunkStart)
			stakeEvents = append(stakeEvents, chunk.StakeEvents...)
			withdrawEvents = append(withdrawEvents, chunk.WithdrawEvents...)
		}
	}
	stakeDataMap := buildStakeDataMap(stakeEvents, withdrawEvents)
	log.Infof("Total addresses with events: %d", len(stakeDataMap))
	if len(stakeDataMap) == 0 {
		log.Warn("No events found in range - debugging with raw logs")
		debugRawLogs(cProps, startU, endU)
	}
	return stakeDataMap, nil
}

// cacheSchemaVersion identifies the current on-disk layout of the per-contract
// bbolt cache (cache/<addr>.db). Bump this whenever the storage layout
// changes in a way that requires existing entries to be invalidated.
//
// Versions:
//   - 1 (implicit; pre-Phase-3): chunks bucket only, keys = 16-byte BE
//     (chunkStart, chunkEnd). No meta bucket. Trailing-chunk keys shifted
//     every epoch — orphaned by the Phase 3 cache rewrite.
//   - 2 (current): chunks bucket with 8-byte BE chunkStart keys; meta
//     bucket with "tip" pointer and this "schema_version" marker.
const cacheSchemaVersion uint32 = 2

// migrateOrInitCacheSchema reads the schema_version marker from the meta
// bucket. If it's missing or older than cacheSchemaVersion, both buckets
// are dropped and recreated and the current version is written. The check
// runs once per process; on subsequent calls the marker matches and the
// function is a fast no-op that just ensures the buckets exist.
//
// This is how the node self-heals across upgrades — operators never have
// to delete cache files manually.
func migrateOrInitCacheSchema(db *bbolt.DB) error {
	return db.Update(func(tx *bbolt.Tx) error {
		// Read whatever schema_version may exist on the current meta bucket.
		var stored uint32
		var hasStored bool
		if meta := tx.Bucket([]byte("meta")); meta != nil {
			if v := meta.Get([]byte("schema_version")); len(v) == 4 {
				stored = binary.BigEndian.Uint32(v)
				hasStored = true
			}
		}
		if hasStored && stored == cacheSchemaVersion {
			// Already on the current schema — ensure both buckets exist
			// (defensive — they should already) and return.
			if _, err := tx.CreateBucketIfNotExists([]byte("chunks")); err != nil {
				return err
			}
			if _, err := tx.CreateBucketIfNotExists([]byte("meta")); err != nil {
				return err
			}
			return nil
		}

		// Missing or older — wipe and reinitialize. Deletes silently no-op
		// on absent buckets, which is what we want for a fresh DB.
		if err := tx.DeleteBucket([]byte("chunks")); err != nil && err != bbolt.ErrBucketNotFound {
			return err
		}
		if err := tx.DeleteBucket([]byte("meta")); err != nil && err != bbolt.ErrBucketNotFound {
			return err
		}
		if _, err := tx.CreateBucket([]byte("chunks")); err != nil {
			return err
		}
		meta, err := tx.CreateBucket([]byte("meta"))
		if err != nil {
			return err
		}
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, cacheSchemaVersion)
		if err := meta.Put([]byte("schema_version"), buf); err != nil {
			return err
		}
		if hasStored {
			log.Infof("Cache schema migrated: was v%d, now v%d (cache wiped and rebuilt)", stored, cacheSchemaVersion)
		} else {
			log.Infof("Cache schema initialized at v%d", cacheSchemaVersion)
		}
		return nil
	})
}

// buildStakeDataMap folds raw Staked / Withdrew events into per-address,
// per-block signed deltas. Per-block deltas may go negative (a withdraw
// at a block where the same wallet didn't also stake in that block);
// callers that need a non-negative running stake (findMinOverBlockRange)
// clamp the cumulative value, which is the correct place to enforce the
// on-chain invariant.
func buildStakeDataMap(stakeEvents []StakeEvent, withdrawEvents []WithdrawEvent) map[common.Address]map[uint64]*UserStakeData {
	stakeDataMap := make(map[common.Address]map[uint64]*UserStakeData)
	for _, e := range stakeEvents {
		addr := e.Addr
		if _, exists := stakeDataMap[addr]; !exists {
			stakeDataMap[addr] = make(map[uint64]*UserStakeData)
		}
		if stakeDataMap[addr][e.Block] == nil {
			stakeDataMap[addr][e.Block] = &UserStakeData{StakeAmount: big.NewInt(0)}
		}
		stakeDataMap[addr][e.Block].StakeAmount.Add(stakeDataMap[addr][e.Block].StakeAmount, e.Amount)
	}
	for _, e := range withdrawEvents {
		addr := e.Addr
		if _, exists := stakeDataMap[addr]; !exists {
			stakeDataMap[addr] = make(map[uint64]*UserStakeData)
		}
		if stakeDataMap[addr][e.Block] == nil {
			stakeDataMap[addr][e.Block] = &UserStakeData{StakeAmount: big.NewInt(0)}
		}
		stakeDataMap[addr][e.Block].StakeAmount.Sub(stakeDataMap[addr][e.Block].StakeAmount, e.Amount)
		// Do NOT clamp the per-block delta at zero — it must stay signed so
		// withdraws at blocks without a same-block stake aren't silently
		// erased. findMinOverBlockRange clamps the cumulative stake at zero,
		// which is the correct place to enforce the on-chain non-negative
		// invariant.
	}
	return stakeDataMap
}

// isQueryTooLargeError detects the family of RPC errors that recommend
// reducing the queried block range. Brittle: providers phrase it
// differently, so we substring-match on a few common shapes. Anything that
// doesn't match is treated as non-retriable and surfaces immediately.
func isQueryTooLargeError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "query") || strings.Contains(msg, "limit") || strings.Contains(msg, "large")
}

func queryChunkWithRetry(cProps *ConnectionProps, kt Ktv2Interface, start, end uint64, stakeOut *[]StakeEvent, withdrawOut *[]WithdrawEvent) error {
	// maxSplitLevels caps how many times we recursively halve a range when
	// the RPC returns "query too large". Three levels means a chunk can be
	// quartered (or eighthed under chunks). Effective floor = chunkSize / 8.
	const maxSplitLevels = 3
	var singleQuery func(s, e uint64) error

	singleQuery = func(s, e uint64) error {
		if cProps.QueryDelay > 0 {
			time.Sleep(cProps.QueryDelay)
		}
		opts := &bind.FilterOpts{
			Start:   s,
			End:     &e,
			Context: context.Background(),
		}
		log.Debugf("Querying stake events for block range %d-%d", s, e)
		stakeIter, err := kt.FilterStaked(opts)
		if err != nil {
			return fmt.Errorf("failed to filter stake events for %d-%d: %w", s, e, err)
		}
		for stakeIter.Next() {
			e := stakeIter.Event()
			if e == nil {
				continue
			}
			*stakeOut = append(*stakeOut, StakeEvent{
				Addr:   e.Arg0,
				Amount: e.Arg1,
				Block:  e.Raw.BlockNumber,
			})
		}
		if err := stakeIter.Error(); err != nil {
			return fmt.Errorf("stake iterator error for %d-%d: %w", s, e, err)
		}
		stakeIter.Close()
		if cProps.QueryDelay > 0 {
			time.Sleep(cProps.QueryDelay)
		}
		log.Debugf("Querying withdraw events for block range %d-%d", s, e)
		withdrawIter, err := kt.FilterWithdrew(opts)
		if err != nil {
			return fmt.Errorf("failed to filter withdrawal events for %d-%d: %w", s, e, err)
		}
		for withdrawIter.Next() {
			e := withdrawIter.Event()
			if e == nil {
				continue
			}
			*withdrawOut = append(*withdrawOut, WithdrawEvent{
				Addr:   e.Arg0,
				Amount: e.Arg1,
				Block:  e.Raw.BlockNumber,
			})
		}
		if err := withdrawIter.Error(); err != nil {
			return fmt.Errorf("withdraw iterator error for %d-%d: %w", s, e, err)
		}
		withdrawIter.Close()
		return nil
	}

	// queryWithSplit tries a single RPC call for [s, e]. On a retriable
	// "query too large" error it halves the range and recurses on each
	// half — up to maxSplitLevels deep. Non-retriable errors and depth
	// exhaustion propagate.
	var queryWithSplit func(s, e uint64, depth int) error
	queryWithSplit = func(s, e uint64, depth int) error {
		err := singleQuery(s, e)
		if err == nil {
			return nil
		}
		if !isQueryTooLargeError(err) {
			return err
		}
		if depth <= 0 {
			return fmt.Errorf("query split exceeded max depth at range %d-%d: %w", s, e, err)
		}
		if s >= e {
			// Already a single block; cannot split further.
			return err
		}
		mid := (s + e) / 2
		log.Warnf("Splitting chunk %d-%d into %d-%d and %d-%d due to error (depth left: %d): %v",
			s, e, s, mid, mid+1, e, depth-1, err)
		if err := queryWithSplit(s, mid, depth-1); err != nil {
			return err
		}
		return queryWithSplit(mid+1, e, depth-1)
	}

	return queryWithSplit(start, end, maxSplitLevels)
}

func GetContractCreationBlock(cProps *ConnectionProps) (uint64, error) {
	// The creation block never changes for a given contract. If we already
	// resolved it (or the user provided it via KT_START_BLOCK), reuse it
	// instead of re-running the binary search every epoch.
	if cProps.KtBlock != nil && cProps.KtBlock.Sign() > 0 {
		return cProps.KtBlock.Uint64(), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	latestBlock, err := cProps.Client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block number: %v", err)
	}

	low, high := uint64(0), latestBlock
	for low < high {
		mid := (low + high) / 2
		code, err := cProps.Client.CodeAt(ctx, cProps.KtAddr, big.NewInt(int64(mid)))
		if err != nil {
			return 0, fmt.Errorf("failed to get code at block %d: %v", mid, err)
		}

		if len(code) > 0 {
			high = mid // Potential creation block found, search lower
		} else {
			low = mid + 1 // No contract at this block, search higher
		}
	}

	// Check if contract exists at the found block
	code, err := cProps.Client.CodeAt(ctx, cProps.KtAddr, big.NewInt(int64(low)))
	if err != nil {
		return 0, fmt.Errorf("failed to get code at block %d: %v", low, err)
	}
	if len(code) == 0 {
		return 0, fmt.Errorf("contract creation block not found")
	}

	log.Infof("Contract creation block found at block %d", low)
	cProps.KtBlock = new(big.Int).SetUint64(low)
	return low, nil
}

// findMinOverBlockRange calculates the total and per-address minimum stake amounts over a block range.
// Returns the total minimum stake, a map of address to minimum stake data, and an error if any occurs.
func findMinOverBlockRange(epochStartBlock, endBlock uint64, stakeDataMap map[common.Address]map[uint64]*UserStakeData) (*big.Int, map[common.Address]*UserStakeData, error) {
	log.Debugf("Finding minimum stakes over block range")

	if epochStartBlock > endBlock {
		log.Errorf("Invalid block range - Start: %d, End: %d", epochStartBlock, endBlock)
		return nil, nil, fmt.Errorf("start block %d exceeds end block %d", epochStartBlock, endBlock)
	}
	if stakeDataMap == nil {
		log.Errorf("Stake data map is nil")
		return nil, nil, fmt.Errorf("stake data map is nil")
	}

	// Initialize results
	totalMin := big.NewInt(0)
	addressMins := make(map[common.Address]*UserStakeData)
	log.Infof("Processing block range - Start: %d, End: %d", epochStartBlock, endBlock)

	// If no addresses have events, return empty results
	if len(stakeDataMap) == 0 {
		log.Infof("No addresses with stake events found up to block %d", endBlock)
		return totalMin, addressMins, nil
	}

	// Process each address
	for addr, blockData := range stakeDataMap {
		// Sort blocks in ascending order
		var blocks []uint64
		for blk := range blockData {
			blocks = append(blocks, blk)
		}
		sort.Slice(blocks, func(i, j int) bool { return blocks[i] < blocks[j] })

		// Initialize
		currentStake := big.NewInt(0)
		var minStake *big.Int

		// Process all events in order
		var inRangeProcessed bool
		for _, block := range blocks {
			if blockData[block] == nil || blockData[block].StakeAmount == nil {
				log.Warnf("Nil stake data for %s at block %d", addr.Hex(), block)
				continue
			}
			// On the first in-range block, capture currentStake BEFORE applying
			// this block's delta — that value is the wallet's stake at the
			// instant the epoch window opens (i.e. their carried-in floor).
			if block >= epochStartBlock && !inRangeProcessed {
				if minStake == nil || currentStake.Cmp(minStake) < 0 {
					minStake = new(big.Int).Set(currentStake)
				}
				inRangeProcessed = true
			}
			// Apply this block's delta.
			currentStake.Add(currentStake, blockData[block].StakeAmount)
			if currentStake.Sign() < 0 {
				log.Warnf("Negative stake computed for %s at block %d: %s - Setting to zero", addr.Hex(), block, currentStake.String())
				currentStake.SetInt64(0)
			}
			if block < epochStartBlock {
				continue
			}
			// Post-event candidate (a withdraw can drop the running min further).
			if minStake == nil || currentStake.Cmp(minStake) < 0 {
				minStake = new(big.Int).Set(currentStake)
			}
			log.Debugf("Stake update for %s at %d - Current: %s, Min: %s", addr.Hex(), block, currentStake.String(), minStake.String())
		}

		// After all events: If no in-range events but prior stake >0, set min to that initial
		if minStake == nil && currentStake.Cmp(big.NewInt(0)) > 0 {
			minStake = new(big.Int).Set(currentStake)
			log.Debugf("No events in range for %s, using stake at end: %s", addr.Hex(), minStake.String())
		}

		// Include if min >0
		if minStake != nil && minStake.Cmp(big.NewInt(0)) > 0 {
			addressMins[addr] = &UserStakeData{StakeAmount: minStake}
			totalMin.Add(totalMin, minStake)
			log.Debugf("Minimum stake for %s: %s", addr.Hex(), minStake.String())
		} else if minStake != nil && minStake.Cmp(big.NewInt(0)) <= 0 {
			log.Debugf("Minimum stake for %s is %s - Excluding as invalid", addr.Hex(), minStake.String())
		}
	}

	log.Infof("Total minimum staked: %s", totalMin.String())
	return totalMin, addressMins, nil
}

// getNextNonce retrieves the next transaction nonce for the given address.
// Returns the nonce or 0 if retrieval fails.
func getNextNonce(client EthClient, addr common.Address) (uint64, error) {
	log.Debugf("Fetching next nonce")

	// Validate inputs
	if client == nil {
		log.Errorf("Ethereum client is nil")
		return 0, fmt.Errorf("ethereum client is nil")
	}
	if addr == (common.Address{}) {
		log.Errorf("Address is zero")
		return 0, fmt.Errorf("address is zero")
	}

	// Get nonce
	nonce, err := client.PendingNonceAt(context.Background(), addr)
	if err != nil {
		log.Errorf("Failed to get nonce for %s: %v", addr.Hex(), err)
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}

	log.Debugf("Next nonce for %s: %d", addr.Hex(), nonce)
	return nonce, nil
}
