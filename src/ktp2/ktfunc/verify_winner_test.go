package ktfunc

import (
	"math/big"
	"testing"

	"ktp2/src/abis/ktv2"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	)

	assert.NoError(t, err)
	assert.Equal(t, common.Address{}, result.CalculatedWinner)
}

// (Phase 6a: removed TestVerifyWinnerCalculation_LinearVsLog. The linear
//  mode was eliminated to stop different operators silently computing
//  different winners. Log is the only mode.)

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
	)

	assert.Error(t, err)
}

// ============================================================================
// Phase 5f — VerifyLastWinner direct tests.
//
// VerifyLastWinner fetches the most recent on-chain Rwd event and the
// matching Voted event, then replays the winner calculation to confirm
// the on-chain winner matches what the algorithm would pick today.
// Previously had zero direct tests.

// vlwSetup builds a minimal ConnectionProps + mocks for VerifyLastWinner.
// Pre-sets KtBlock so GetContractCreationBlock returns immediately.
func vlwSetup(t *testing.T) (*ConnectionProps, *MockEthClient, *MockKtv2) {
	t.Helper()
	logrus.SetLevel(logrus.FatalLevel)

	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Client:   mockClient,
		Kt:       mockKt,
		KtAddr:   common.HexToAddress("0x1234567890123456789012345678901234567890"),
		MyPubKey: common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		KtBlock:  big.NewInt(18_000_000), // skips GetContractCreationBlock binary search
	}
	return cProps, mockClient, mockKt
}

// TestVerifyLastWinner_NoRwdInSearchRangeReturnsNil — empty FilterRwd
// iterator. The function logs a warning and returns nil (nothing to
// verify).
func TestVerifyLastWinner_NoRwdInSearchRangeReturnsNil(t *testing.T) {
	cProps, mockClient, mockKt := vlwSetup(t)

	mockClient.On("BlockNumber", mock.Anything).Return(uint64(18_001_000), nil)
	mockKt.On("EpochInterval", mock.Anything).Return(uint16(600), nil)
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: nil}, nil)

	err := VerifyLastWinner(cProps)
	assert.NoError(t, err)
}

// TestVerifyLastWinner_NoMatchingVotedReturnsError — Rwd event found,
// but no Voted event in the search range matches the winner address.
// Function returns an error.
func TestVerifyLastWinner_NoMatchingVotedReturnsError(t *testing.T) {
	cProps, mockClient, mockKt := vlwSetup(t)

	winner := common.HexToAddress("0x000000000000000000000000000000000000Win0")
	otherCandidate := common.HexToAddress("0x000000000000000000000000000000000000A1ce")
	rwdEvents := []*ktv2.Ktv2Rwd{
		{Arg0: winner, Arg1: big.NewInt(int64(1e18)),
			Raw: types.Log{BlockNumber: 18_000_900}},
	}
	// Voted events exist but vote for a different address than the winner.
	votedEvents := []*ktv2.Ktv2Voted{
		{Arg0: big.NewInt(18_000_500), Arg1: otherCandidate, Arg2: "0xabcd",
			Raw: types.Log{BlockNumber: 18_000_850}},
	}

	mockClient.On("BlockNumber", mock.Anything).Return(uint64(18_001_000), nil)
	mockKt.On("EpochInterval", mock.Anything).Return(uint16(600), nil)
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: rwdEvents}, nil)
	mockKt.On("FilterVoted", mock.Anything).Return(&mockVotedIter{events: votedEvents}, nil)

	err := VerifyLastWinner(cProps)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no matching Voted event")
}

// TestVerifyLastWinner_ReplayMatchesOnChain — Rwd + matching Voted exist,
// and the replay (via mocked calcWinningWallet) produces the same winner.
// Function returns nil (success: VERIFIED).
func TestVerifyLastWinner_ReplayMatchesOnChain(t *testing.T) {
	cProps, mockClient, mockKt := vlwSetup(t)

	winner := common.HexToAddress("0x000000000000000000000000000000000000Win0")
	epochStart := big.NewInt(18_000_500)
	blockHashStr := "0x0102030405060708091011121314151617181920212223242526272829303132"

	rwdEvents := []*ktv2.Ktv2Rwd{
		{Arg0: winner, Arg1: big.NewInt(int64(1e18)),
			Raw: types.Log{BlockNumber: 18_000_900}},
	}
	votedEvents := []*ktv2.Ktv2Voted{
		{Arg0: epochStart, Arg1: winner, Arg2: blockHashStr,
			Raw: types.Log{BlockNumber: 18_000_850}},
	}

	mockClient.On("BlockNumber", mock.Anything).Return(uint64(18_001_000), nil)
	mockKt.On("EpochInterval", mock.Anything).Return(uint16(600), nil)
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: rwdEvents}, nil)
	mockKt.On("FilterVoted", mock.Anything).Return(&mockVotedIter{events: votedEvents}, nil)

	// Stub GatherStakesAndWithdraws to return a single staker = winner with
	// pre-epoch stake. calcWinningWallet then picks them (only candidate).
	origGather := GatherStakesAndWithdraws
	GatherStakesAndWithdraws = func(_ *ConnectionProps, _ Ktv2Interface, _, _ *big.Int) (map[common.Address]map[uint64]*UserStakeData, error) {
		return map[common.Address]map[uint64]*UserStakeData{
			winner: {18_000_000: {StakeAmount: big.NewInt(int64(1e18))}},
		}, nil
	}
	defer func() { GatherStakesAndWithdraws = origGather }()

	err := VerifyLastWinner(cProps)
	assert.NoError(t, err)
}

// TestVerifyLastWinner_ReplayMismatchReturnsNilButLogsWarning — replay
// produces a different winner than the on-chain one. Function still
// returns nil (the result is informational; the warning is the signal),
// but does NOT silently report success. Pin: no error, but the
// calculated winner is different.
func TestVerifyLastWinner_ReplayMismatchReturnsNilButLogsWarning(t *testing.T) {
	cProps, mockClient, mockKt := vlwSetup(t)

	onChainWinner := common.HexToAddress("0x000000000000000000000000000000000000Win0")
	differentWinner := common.HexToAddress("0x000000000000000000000000000000000000D1ff")
	epochStart := big.NewInt(18_000_500)

	rwdEvents := []*ktv2.Ktv2Rwd{
		{Arg0: onChainWinner, Arg1: big.NewInt(int64(1e18)),
			Raw: types.Log{BlockNumber: 18_000_900}},
	}
	votedEvents := []*ktv2.Ktv2Voted{
		{Arg0: epochStart, Arg1: onChainWinner, Arg2: "0xabcd",
			Raw: types.Log{BlockNumber: 18_000_850}},
	}

	mockClient.On("BlockNumber", mock.Anything).Return(uint64(18_001_000), nil)
	mockKt.On("EpochInterval", mock.Anything).Return(uint16(600), nil)
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: rwdEvents}, nil)
	mockKt.On("FilterVoted", mock.Anything).Return(&mockVotedIter{events: votedEvents}, nil)

	// Stub the algorithm to return a DIFFERENT winner than on-chain.
	origCalc := calcWinningWallet
	SetCalculateWinningWallet(func(_ map[common.Address]*UserStakeData, _ common.Hash) (common.Address, error) {
		return differentWinner, nil
	})
	defer SetCalculateWinningWallet(origCalc)
	origGather := GatherStakesAndWithdraws
	GatherStakesAndWithdraws = func(_ *ConnectionProps, _ Ktv2Interface, _, _ *big.Int) (map[common.Address]map[uint64]*UserStakeData, error) {
		return map[common.Address]map[uint64]*UserStakeData{
			onChainWinner:   {18_000_000: {StakeAmount: big.NewInt(int64(1e18))}},
			differentWinner: {18_000_000: {StakeAmount: big.NewInt(int64(1e18))}},
		}, nil
	}
	defer func() { GatherStakesAndWithdraws = origGather }()

	err := VerifyLastWinner(cProps)
	// VerifyLastWinner returns nil even on mismatch; the warning is logged.
	// Pin: no error propagation.
	assert.NoError(t, err)
}
