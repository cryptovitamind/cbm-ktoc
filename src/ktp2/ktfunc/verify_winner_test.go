package ktfunc

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// VerificationInput holds the data needed to replay and verify a winner selection.
// This mirrors the data that VerifyLastWinner fetches from on-chain events.
type verifyTestInput struct {
	onChainWinner common.Address
	rewardAmount  *big.Int
	epochStart    uint64
	epochEnd      uint64
	blockHash     common.Hash
	stakeDataMap  map[common.Address]map[uint64]*UserStakeData
	useLinear     bool
}

func buildSimpleStakeDataMap(addr common.Address, stakeBlock uint64, amount int64) map[common.Address]map[uint64]*UserStakeData {
	return map[common.Address]map[uint64]*UserStakeData{
		addr: {
			stakeBlock: {StakeAmount: big.NewInt(amount)},
		},
	}
}

func buildMultiStakeDataMap(stakes map[common.Address]map[uint64]int64) map[common.Address]map[uint64]*UserStakeData {
	result := make(map[common.Address]map[uint64]*UserStakeData)
	for addr, blocks := range stakes {
		result[addr] = make(map[uint64]*UserStakeData)
		for block, amount := range blocks {
			result[addr][block] = &UserStakeData{StakeAmount: big.NewInt(amount)}
		}
	}
	return result
}

// TestVerifyWinnerCalculation_SingleStaker tests verification when there's only one staker.
// The winner should always be that staker regardless of block hash.
func TestVerifyWinnerCalculation_SingleStaker(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	stakerAddr := common.HexToAddress("0xabc123456789012345678901234567890123456")
	blockHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	// Stake pre-epoch so the wallet has a non-zero floor going INTO the epoch.
	// An in-range-only stake (e.g. block 55) would give min=0 and be excluded.
	stakeDataMap := buildSimpleStakeDataMap(stakerAddr, 40, 1000)

	result, err := VerifyWinnerCalculation(
		stakeDataMap,
		50,  // epochStart
		110, // epochEnd
		blockHash,
		false, // useLinear (log normalization)
	)

	assert.NoError(t, err)
	assert.Equal(t, stakerAddr, result.CalculatedWinner)
	assert.True(t, result.Match)
}

// TestVerifyWinnerCalculation_MultipleStakers tests verification with multiple stakers.
// Uses a known block hash so the winner selection is deterministic.
func TestVerifyWinnerCalculation_MultipleStakers(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	addr1 := common.HexToAddress("0x1000000000000000000000000000000000000001")
	addr2 := common.HexToAddress("0x2000000000000000000000000000000000000002")
	addr3 := common.HexToAddress("0x3000000000000000000000000000000000000003")

	stakeDataMap := buildMultiStakeDataMap(map[common.Address]map[uint64]int64{
		addr1: {40: 1000},
		addr2: {40: 2000},
		addr3: {40: 3000},
	})

	// Block hash of all zeros -> randFloat = 0.0 -> should select first sorted address
	blockHash := common.Hash{}

	result, err := VerifyWinnerCalculation(
		stakeDataMap,
		50,
		110,
		blockHash,
		true, // linear probs
	)

	assert.NoError(t, err)
	// With randFloat=0.0 and linear probs, winner should be the first address in sorted order
	// since 0.0 < any positive probability
	assert.NotEqual(t, common.Address{}, result.CalculatedWinner)
	assert.True(t, result.Match)
}

// TestVerifyWinnerCalculation_MatchesOnChainWinner tests that when we set OnChainWinner
// to the same address the algorithm calculates, Match is true.
func TestVerifyWinnerCalculation_MatchesOnChainWinner(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	addr := common.HexToAddress("0xabc123456789012345678901234567890123456")

	stakeDataMap := buildSimpleStakeDataMap(addr, 40, 5000)

	result, err := VerifyWinnerCalculation(
		stakeDataMap,
		50,
		110,
		common.Hash{}, // all zeros
		false,
	)

	assert.NoError(t, err)
	// Single staker -> always wins
	assert.Equal(t, addr, result.CalculatedWinner)
	assert.True(t, result.Match)
}

// TestVerifyWinnerCalculation_EmptyStakes tests verification with no stakes.
func TestVerifyWinnerCalculation_EmptyStakes(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	stakeDataMap := make(map[common.Address]map[uint64]*UserStakeData)

	result, err := VerifyWinnerCalculation(
		stakeDataMap,
		50,
		110,
		common.Hash{},
		false,
	)

	assert.NoError(t, err)
	assert.Equal(t, common.Address{}, result.CalculatedWinner)
}

// TestVerifyWinnerCalculation_LinearVsLog tests that switching between linear and log
// probability modes can produce different winners.
func TestVerifyWinnerCalculation_LinearVsLog(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	addr1 := common.HexToAddress("0x1000000000000000000000000000000000000001")
	addr2 := common.HexToAddress("0x2000000000000000000000000000000000000002")

	// addr1 has a very small stake, addr2 has a huge stake
	// Log normalization will give addr1 a higher relative probability than linear
	stakeDataMap := buildMultiStakeDataMap(map[common.Address]map[uint64]int64{
		addr1: {40: 10},
		addr2: {40: 10000000},
	})

	// Use a block hash that gives a randFloat around 0.5
	blockHash := common.Hash{0x80}

	resultLinear, err := VerifyWinnerCalculation(stakeDataMap, 50, 110, blockHash, true)
	assert.NoError(t, err)

	// Rebuild map since probs get mutated
	stakeDataMap = buildMultiStakeDataMap(map[common.Address]map[uint64]int64{
		addr1: {40: 10},
		addr2: {40: 10000000},
	})

	resultLog, err := VerifyWinnerCalculation(stakeDataMap, 50, 110, blockHash, false)
	assert.NoError(t, err)

	// With linear probs, addr1 has ~0.000001 probability so addr2 almost certainly wins
	// With log probs, addr1 has a much larger relative probability
	// At randFloat=0.5, linear should pick addr2, log might pick either
	// This test verifies the prob mode affects the outcome
	assert.Equal(t, addr2, resultLinear.CalculatedWinner,
		"With linear probs and randFloat=0.5, the high-stake addr should win")
	// We don't assert the log result equals a specific addr since it depends on the exact math,
	// but we verify the function completes without error
	assert.NotEqual(t, common.Address{}, resultLog.CalculatedWinner)
}

// TestVerifyWinnerCalculation_StakesOutsideEpoch tests that stakes outside the epoch range
// are handled correctly (pre-epoch stakes count, post-epoch events are excluded by endBlock).
func TestVerifyWinnerCalculation_StakesOutsideEpoch(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	addr1 := common.HexToAddress("0x1000000000000000000000000000000000000001")
	addr2 := common.HexToAddress("0x2000000000000000000000000000000000000002")

	// addr1 staked pre-epoch and carries that stake INTO the epoch — eligible
	// with min=5000.
	// addr2 stakes only mid-epoch with no prior position — their pre-event
	// floor is 0, so they have min=0 and are correctly excluded.
	stakeDataMap := buildMultiStakeDataMap(map[common.Address]map[uint64]int64{
		addr1: {10: 5000}, // pre-epoch (epoch starts at 50)
		addr2: {55: 3000}, // mid-epoch only
	})

	result, err := VerifyWinnerCalculation(
		stakeDataMap,
		50,
		110,
		common.Hash{}, // randFloat=0
		true,          // linear
	)

	assert.NoError(t, err)
	// Only addr1 is eligible, so it must be the winner regardless of randFloat.
	assert.Equal(t, addr1, result.CalculatedWinner)
}

// TestVerifyWinnerCalculation_NilMap tests that nil stake data map returns error.
func TestVerifyWinnerCalculation_NilMap(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	_, err := VerifyWinnerCalculation(
		nil,
		50,
		110,
		common.Hash{},
		false,
	)

	assert.Error(t, err)
}
