package ktfunc

import (
	"context"
	"fmt"
	"ktp2/src/abis"
	"math/big"
	"os"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
)

// Define as a variable holding a function
var calcWinningWallet = defaultCalculateWinningWallet

// Exported for testing purposes if needed
func SetCalculateWinningWallet(f func(map[common.Address]*UserStakeData, common.Hash) (common.Address, error)) {
	calcWinningWallet = f
}

// VoteAndReward determines the winning wallet, votes for it, and rewards it if conditions are met.
// Returns the winning wallet address or a zero address if no winner is determined.
func VoteAndReward(cProps *ConnectionProps) error {
	log.Debugf("Voting and rewarding")

	// Get current block
	currentBlock, err := getCurrentBlock(cProps)
	if err != nil {
		log.Errorf("Failed to get current block: %v", err)
		return fmt.Errorf("failed to get current block: %w", err)
	}

	// Get KT target block numbers for the epoch
	startBlock, endBlock, err := getStartAndEndEpochBlocks(cProps)
	if err != nil {
		log.Errorf("Failed to get epoch blocks: %v", err)
		return fmt.Errorf("failed to get epoch blocks: %w", err)
	}

	// Check if it's time to vote
	if !IsTimeToVote(endBlock, currentBlock) {
		status := fmt.Sprintf("Not time to vote yet - Current block: %d, End block: %d", currentBlock.NumberU64(), endBlock)
		fmt.Fprintf(os.Stdout, "\r\033[1;36m%s\033[0m", status) // Cyan text, reset color
		os.Stdout.Sync()                                        // Ensure it flushes immediately
		return nil                                              // Not an error, just not time yet
	}

	// Log the end-of-epoch ETH balance
	if err := printEndEpochKtEthBalance(cProps, endBlock); err != nil {
		log.Warnf("Failed to print end epoch balance: %v", err)
	}

	// Gather stake and withdrawal events
	stakeDataMap, err := gatherStakesAndWithdraws(cProps.Kt, cProps.KtBlock, endBlock)
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
		log.Warn("No stakes found.")
		return nil
	}

	// Print minimum stakes
	printAllStakes(stakeDataMinsMap)

	// Calculate probabilities for each wallet
	if !calculateProbsForEachWallet(stakeDataMinsMap, totalMin) {
		log.Warn("No valid probabilities calculated")
		return nil
	}

	// Vote and potentially reward the winner
	winner, err := calculateVoteAndReward(stakeDataMinsMap, startBlock, endBlock, cProps)
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

func calculateVoteAndReward(stakeDataMinsMap map[common.Address]*UserStakeData, epochStartBlock, endEpochBlockNumber *big.Int, cProps *ConnectionProps) (common.Address, error) {
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

	const oneExtraBlock = 1
	nextBlockNumber := new(big.Int).Add(endEpochBlockNumber, big.NewInt(oneExtraBlock))
	log.Printf("Epoch start block: %d, Next block: %d", epochStartBlock.Uint64(), nextBlockNumber.Uint64())

	// Fetch next block
	nextBlock, err := cProps.Client.BlockByNumber(context.Background(), nextBlockNumber)
	if err != nil {
		log.Errorf("Failed to get next block %d: %v", nextBlockNumber.Uint64(), err)
		return common.Address{}, fmt.Errorf("failed to get next block: %w", err)
	}
	log.Infof("Next block: %d", nextBlock.NumberU64())

	// Calculate winning wallet
	winner, err := calcWinningWallet(stakeDataMinsMap, nextBlock.Hash())
	if err != nil {
		log.Errorf("Failed to calculate winning wallet: %v", err)
		return common.Address{}, fmt.Errorf("failed to calculate winning wallet: %w", err)
	}
	if winner == (common.Address{}) {
		log.Warn("No winner determined")
		return winner, nil
	}
	log.Infof("Winner selected: %s", winner.Hex())

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
		if err := rewardWinningWallet(cProps, winner); err != nil {
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
func getCurrentBlock(cProps *ConnectionProps) (*types.Block, error) {
	log.Debug("Fetching current block")

	// Fetch the latest block (nil block number means latest)
	block, err := cProps.Client.BlockByNumber(context.Background(), nil)
	if err != nil {
		log.Errorf("Failed to retrieve current block: %v", err)
		return nil, fmt.Errorf("failed to get current block: %w", err)
	}

	// Log block details with pretty formatting
	log.Debugf("Block number: %d", block.NumberU64())
	log.Debugf("Block hash: %s", block.Hash().Hex())

	return block, nil
}

// defaultCalculateWinningWallet selects the winning wallet based on stake probabilities and a random seed.
// Returns the winning address or a zero address if no winner can be determined.
func defaultCalculateWinningWallet(stakeDataMinsMap map[common.Address]*UserStakeData, randomNumber common.Hash) (common.Address, error) {
	LogOperationStart("Calculating winning wallet")

	// Validate inputs
	if stakeDataMinsMap == nil {
		log.Warn("Stake data map is nil - No stakes to process")
		return common.Address{}, fmt.Errorf("stake data map is nil")
	}
	if len(stakeDataMinsMap) == 0 {
		log.Info("No stakes found in the map")
		return common.Address{}, nil
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

	log.Info("No valid addresses available - Returning zero address")
	return common.Address{}, nil
}

func calculateProbsForEachWallet(stakeDataMinsMap map[common.Address]*UserStakeData, totalMin *big.Int) bool {
	foundSomething := false

	for addr, stakeData := range stakeDataMinsMap {
		foundSomething = true

		stakeFloat := new(big.Float).SetInt(stakeData.StakeAmount)
		totalFloat := new(big.Float).SetInt(totalMin)

		stakeData.Prob = new(big.Float).Quo(stakeFloat, totalFloat)
		log.Debugf("Address: %s, Probability: %f\n", addr.Hex(), stakeData.Prob)
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

func IsTimeToVote(endBlock *big.Int, block BlockNumberer) bool {
	log.Debugf("Checking voting time")

	// Validate inputs
	if endBlock == nil {
		log.Errorf("End block is nil")
		return false
	}
	if block == nil {
		log.Errorf("Current block is nil")
		return false
	}

	currentNum := block.Number()
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

func rewardWinningWallet(cProps *ConnectionProps, winner common.Address) error {
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

	// Convert reward amount from wei to ETH
	weiToEthReward := new(big.Float).SetInt(rewardAmount)
	rewardEth := new(big.Float).Quo(weiToEthReward, big.NewFloat(1e18))
	log.Printf("Contract balance (reward amount): %.6f ETH", rewardEth)

	// Call the rwd function to send the reward
	tx, err := cProps.Kt.Rwd(auth, winner, rewardAmount)
	if err != nil {
		return fmt.Errorf("failed to call rwd function: %v", err)
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

	// Get required votes
	voteRequired, err = cProps.Kt.ConsensusReq(callOpts)
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

// gatherStakesAndWithdraws collects stake and withdrawal events for a KT contract from block startBlock to endBlock.
// Returns a map of address to block-specific stake data or an error if filtering fails.
func gatherStakesAndWithdraws(kt *abis.Ktv2, startBlock *big.Int, endBlock *big.Int) (map[common.Address]map[uint64]*UserStakeData, error) {
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

	// Initialize stake data map
	stakeDataMap := make(map[common.Address]map[uint64]*UserStakeData)
	log.Infof("Collecting events from block %d to %d", startBlock.Uint64(), endBlock.Uint64())

	// Set up filter options from block 0 to endBlock
	startBlockUint64 := startBlock.Uint64()
	endBlockUint64 := endBlock.Uint64()
	opts := &bind.FilterOpts{
		Start:   startBlockUint64,
		End:     &endBlockUint64,
		Context: context.Background(),
	}

	// Filter stake events
	stakeIter, err := kt.FilterStaked(opts)
	if err != nil {
		log.Errorf("Failed to filter stake events: %v", err)
		return nil, fmt.Errorf("failed to filter stake events: %w", err)
	}
	defer stakeIter.Close()

	for stakeIter.Next() {
		event := stakeIter.Event
		if event == nil {
			log.Warn("Encountered nil stake event")
			continue
		}

		addr := event.Arg0
		amount := event.Arg1
		blockNum := event.Raw.BlockNumber

		if _, exists := stakeDataMap[addr]; !exists {
			stakeDataMap[addr] = make(map[uint64]*UserStakeData)
		}
		if stakeDataMap[addr][blockNum] == nil {
			stakeDataMap[addr][blockNum] = &UserStakeData{StakeAmount: big.NewInt(0)}
		}
		stakeDataMap[addr][blockNum].StakeAmount.Add(stakeDataMap[addr][blockNum].StakeAmount, amount)
		log.Debugf("Stake event - Address: %s, Amount: %s, Block: %d", addr.Hex(), amount.String(), blockNum)
	}

	if err := stakeIter.Error(); err != nil {
		log.Errorf("Stake iterator error: %v", err)
		return nil, fmt.Errorf("stake iterator encountered an error: %w", err)
	}

	// Filter withdrawal events
	withdrawIter, err := kt.FilterWithdrew(opts)
	if err != nil {
		log.Errorf("Failed to filter withdrawal events: %v", err)
		return nil, fmt.Errorf("failed to filter withdrawal events: %w", err)
	}
	defer withdrawIter.Close()

	for withdrawIter.Next() {
		event := withdrawIter.Event
		if event == nil {
			log.Warn("Encountered nil withdrawal event")
			continue
		}

		addr := event.Arg0
		amount := event.Arg1
		blockNum := event.Raw.BlockNumber

		if _, exists := stakeDataMap[addr]; !exists {
			stakeDataMap[addr] = make(map[uint64]*UserStakeData)
		}
		if stakeDataMap[addr][blockNum] == nil {
			stakeDataMap[addr][blockNum] = &UserStakeData{StakeAmount: big.NewInt(0)}
		}
		stakeDataMap[addr][blockNum].StakeAmount.Sub(stakeDataMap[addr][blockNum].StakeAmount, amount)
		log.Infof("Withdrawal event - Address: %s, Amount: %s, Block: %d", addr.Hex(), amount.String(), blockNum)
	}

	if err := withdrawIter.Error(); err != nil {
		log.Errorf("Withdrawal iterator error: %v", err)
		return nil, fmt.Errorf("withdrawal iterator encountered an error: %w", err)
	}

	log.Infof("Total addresses with events: %d", len(stakeDataMap))
	return stakeDataMap, nil
}

func GetContractCreationBlock(cProps *ConnectionProps) (uint64, error) {
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
		// THis isn't working but it's close: return big.Int(0), fmt.Errorf("failed to get code at block %d: %v", low, err)
		return 0, fmt.Errorf("contract creation block not found")
	}
	if len(code) == 0 {
		return 0, fmt.Errorf("contract creation block not found")
	}

	log.Infof("Contract creation block found at block %d", low)
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

		currentStake := big.NewInt(0)
		var minStake *big.Int // nil until we hit the range

		// Process events in order
		for _, block := range blocks {
			if blockData[block] == nil || blockData[block].StakeAmount == nil {
				log.Warnf("Nil stake data for %s at block %d", addr.Hex(), block)
				continue
			}

			// Update current stake with the net change
			currentStake.Add(currentStake, blockData[block].StakeAmount)
			// Ensure stake doesn’t go negative (per contract logic)
			if currentStake.Sign() < 0 {
				log.Warnf("Negative stake computed for %s at block %d: %s - Setting to zero", addr.Hex(), block, currentStake.String())
				currentStake.SetInt64(0)
			}

			// If within the range, update minStake
			if block >= epochStartBlock && block <= endBlock {
				if minStake == nil || currentStake.Cmp(minStake) < 0 {
					minStake = new(big.Int).Set(currentStake)
				}
				log.Debugf("Stake update for %s at %d - Current: %s, Min: %s", addr.Hex(), block, currentStake.String(), minStake.String())
			}
		}

		// Handle users with no events in range but prior stake
		if minStake == nil && currentStake.Cmp(big.NewInt(0)) > 0 {
			minStake = new(big.Int).Set(currentStake)
			log.Debugf("No events in range for %s, using stake at end: %s", addr.Hex(), minStake.String())
		}

		// Include in results if minimum stake is positive
		if minStake != nil && minStake.Cmp(big.NewInt(0)) > 0 {
			addressMins[addr] = &UserStakeData{StakeAmount: minStake}
			totalMin.Add(totalMin, minStake)
			log.Debugf("Minimum stake for %s: %s", addr.Hex(), minStake.String())
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
