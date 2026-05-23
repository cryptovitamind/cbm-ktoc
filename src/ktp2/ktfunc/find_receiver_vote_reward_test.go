package ktfunc

import (
	"errors"
	"math/big"
	"testing"

	"ktp2/src/abis/ktv2"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"os"
)

// MockStakedIterator mocks the Ktv2StakedIterator
type MockStakedIterator struct {
	events []*ktv2.Ktv2Staked
	index  int
}

func (m *MockStakedIterator) Next() bool {
	if m.index < len(m.events) {
		m.index++
		return true
	}
	return false
}

func (m *MockStakedIterator) Event() *ktv2.Ktv2Staked {
	if m.index > 0 && m.index <= len(m.events) {
		return m.events[m.index-1]
	}
	return nil
}

func (m *MockStakedIterator) Error() error {
	return nil
}

func (m *MockStakedIterator) Close() error {
	return nil
}

// MockWithdrewIterator mocks the Ktv2WithdrewIterator
type MockWithdrewIterator struct {
	events []*ktv2.Ktv2Withdrew
	index  int
}

func (m *MockWithdrewIterator) Next() bool {
	if m.index < len(m.events) {
		m.index++
		return true
	}
	return false
}

func (m *MockWithdrewIterator) Event() *ktv2.Ktv2Withdrew {
	if m.index > 0 && m.index <= len(m.events) {
		return m.events[m.index-1]
	}
	return nil
}

func (m *MockWithdrewIterator) Error() error {
	return nil
}

func (m *MockWithdrewIterator) Close() error {
	return nil
}

// TestVoteAndReward_NotTimeToVote tests when it's not time to vote yet.
func TestVoteAndReward_NotTimeToVote(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel) // Suppress logs

	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Client:       mockClient,
		Kt:           mockKt,
		KtAddr:       common.HexToAddress("0x1234567890123456789012345678901234567890"),
		MyPubKey:     common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		ChainID:      big.NewInt(1),
		BlocksToWait: 0, // Avoid waiting
	}

	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = privateKey

	// Mock current block < end block
	currentHeader := &types.Header{Number: big.NewInt(100)}
	mockClient.On("HeaderByNumber", mock.Anything, (*big.Int)(nil)).Return(currentHeader, nil)

	// Mock contract details
	startBlock := big.NewInt(50)
	interval := uint16(60) // end = 110
	mockKt.On("StartBlock", mock.Anything).Return(startBlock, nil)
	mockKt.On("EpochInterval", mock.Anything).Return(interval, nil)

	err := VoteAndReward(cProps)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockKt.AssertExpectations(t)
}

// TestVoteAndReward_NoStakes tests when there are no stakes, votes for dead address, no reward.
func TestVoteAndReward_NoStakes(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Client:       mockClient,
		Kt:           mockKt,
		KtAddr:            common.HexToAddress("0x1234567890123456789012345678901234567890"),
		MyPubKey:          common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		ChainID:           big.NewInt(1),
		BlocksToWait:      0,
		ChunkSize:         500,
		ConfirmationDepth: 1, // keep test's endBlock+1 expectation
	}

	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = privateKey

	// Mock current block > end block
	currentHeader := &types.Header{Number: big.NewInt(120)}
	mockClient.On("HeaderByNumber", mock.Anything, (*big.Int)(nil)).Return(currentHeader, nil)

	// Mock contract details
	startBlock := big.NewInt(50)
	endBlock := big.NewInt(110)
	interval := uint16(60)
	mockKt.On("StartBlock", mock.Anything).Return(startBlock, nil).Maybe()
	mockKt.On("EpochInterval", mock.Anything).Return(interval, nil).Maybe()

	// Mock end epoch balance
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, endBlock).Return(big.NewInt(1000000000000000000), nil)

	// Mock contract creation block (simplified to 0)
	mockClient.On("CodeAt", mock.Anything, cProps.KtAddr, mock.Anything).Return([]byte{0x60, 0x80}, nil)
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(120), nil)

	// Mock next block for voting
	nextHeader := &types.Header{Number: big.NewInt(111)}
	mockClient.On("HeaderByNumber", mock.Anything, big.NewInt(111)).Return(nextHeader, nil)

	// Mock empty stake and withdraw iterators
	emptyStakedIter := &MockStakedIterator{events: []*ktv2.Ktv2Staked{}}
	emptyWithdrewIter := &MockWithdrewIterator{events: []*ktv2.Ktv2Withdrew{}}

	mockKt.On("FilterStaked", mock.Anything).Return(emptyStakedIter, nil).Maybe()
	mockKt.On("FilterWithdrew", mock.Anything).Return(emptyWithdrewIter, nil).Maybe()

	// Since empty, totalMin=0, winner=zero
	zeroAddr := common.Address{}
	voteData := nextHeader.Hash().Hex()
	mockTx := types.NewTransaction(0, zeroAddr, big.NewInt(0), 0, big.NewInt(0), []byte{})
	mockKt.On("Vote", mock.Anything, zeroAddr, voteData).Return(mockTx, nil).Maybe()

	// Mock vote count < required
	mockKt.On("BlockRwd", mock.Anything, startBlock, zeroAddr).Return(uint16(0), nil).Maybe()
	mockKt.On("ConsensusReq", mock.Anything).Return(uint16(1), nil).Maybe()

	// Mock TransactionReceipt for WaitMined
	successReceipt := &types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(112)}
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(successReceipt, nil)

	// Clear entire cache directory to ensure fresh query and mocks are used
	_ = os.RemoveAll("cache")

	err := VoteAndReward(cProps)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockKt.AssertExpectations(t)
}

// TestVoteAndReward_WithStakesAndReward tests successful voting and rewarding with stakes.
func TestVoteAndReward_WithStakesAndReward(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Client:       mockClient,
		Kt:           mockKt,
		KtAddr:            common.HexToAddress("0x1234567890123456789012345678901234567890"),
		MyPubKey:          common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		ChainID:           big.NewInt(1),
		BlocksToWait:      0,
		ChunkSize:         500,
		ConfirmationDepth: 1, // keep test's endBlock+1 expectation
	}

	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = privateKey

	// Mock current block > end block
	currentHeader := &types.Header{Number: big.NewInt(120)}
	mockClient.On("HeaderByNumber", mock.Anything, (*big.Int)(nil)).Return(currentHeader, nil)

	// Mock contract details
	startBlock := big.NewInt(50)
	endBlock := big.NewInt(110)
	interval := uint16(60)
	mockKt.On("StartBlock", mock.Anything).Return(startBlock, nil)
	mockKt.On("EpochInterval", mock.Anything).Return(interval, nil)

	// Mock end epoch balance
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, endBlock).Return(big.NewInt(1000000000000000000), nil)

	// Mock contract creation block (0)
	mockClient.On("CodeAt", mock.Anything, cProps.KtAddr, mock.Anything).Return([]byte{0x60, 0x80}, nil)
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(120), nil)

	// Mock next block
	nextHeader := &types.Header{Number: big.NewInt(111)}
	mockClient.On("HeaderByNumber", mock.Anything, big.NewInt(111)).Return(nextHeader, nil)

	// Mock stake event: one staker at block 40 (pre-epoch) with 1000 wei.
	// The stake must land before epochStart (50) so the staker carries a
	// positive floor INTO the epoch — under the correct min-stake semantics,
	// a wallet that only stakes mid-epoch has min=0 and is excluded.
	stakerAddr := common.HexToAddress("0xabc123456789012345678901234567890123456")
	stakeEvent := &ktv2.Ktv2Staked{
		Arg0: stakerAddr,
		Arg1: big.NewInt(1000),
		Raw:  types.Log{BlockNumber: 40},
	}
	stakedIter := &MockStakedIterator{events: []*ktv2.Ktv2Staked{stakeEvent}}

	emptyWithdrewIter := &MockWithdrewIterator{events: []*ktv2.Ktv2Withdrew{}}

	mockKt.On("FilterStaked", mock.Anything).Return(stakedIter, nil)
	mockKt.On("FilterWithdrew", mock.Anything).Return(emptyWithdrewIter, nil)
	mockKt.On("Declines", mock.Anything, stakerAddr).Return(false, nil)

	// Winner will be stakerAddr, vote for it
	voteData := nextHeader.Hash().Hex()
	voteTx := types.NewTransaction(0, stakerAddr, big.NewInt(0), 0, big.NewInt(0), []byte{})
	mockKt.On("Vote", mock.Anything, stakerAddr, voteData).Return(voteTx, nil)

	// Mock votes enough for reward
	mockKt.On("BlockRwd", mock.Anything, startBlock, stakerAddr).Return(uint16(5), nil)
	mockKt.On("ConsensusReq", mock.Anything).Return(uint16(3), nil)

	// Mock tlOcFees for reward calculation (assuming no OC fees in this test)
	mockKt.On("TlOcFees", mock.Anything).Return(big.NewInt(0), nil)

	// Mock reward
	rewardAmount := big.NewInt(1000000000000000000) // 1 ETH
	rewardTx := types.NewTransaction(0, stakerAddr, rewardAmount, 0, big.NewInt(0), []byte{})
	mockKt.On("Rwd", mock.Anything, stakerAddr, rewardAmount).Return(rewardTx, nil)

	// Mock balances for reward
	beforeBalance := big.NewInt(0)
	mockClient.On("BalanceAt", mock.Anything, stakerAddr, (*big.Int)(nil)).Return(beforeBalance, nil) // First call before reward
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, (*big.Int)(nil)).Return(rewardAmount, nil)
	afterBalance := new(big.Int).Add(beforeBalance, rewardAmount)
	mockClient.On("BalanceAt", mock.Anything, stakerAddr, (*big.Int)(nil)).Return(afterBalance, nil) // Second call after

	// Mock TransactionReceipt for both vote and reward
	successReceipt := &types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(112)}
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(successReceipt, nil)

	// Clear entire cache directory to ensure fresh query and mocks are used
	_ = os.RemoveAll("cache")

	err := VoteAndReward(cProps)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockKt.AssertExpectations(t)
}

// TestVoteAndReward_GetCurrentBlockError tests error when getting current block fails.
func TestVoteAndReward_GetCurrentBlockError(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Client:   mockClient,
		Kt:       mockKt,
		KtAddr:   common.HexToAddress("0x123"),
		MyPubKey: common.HexToAddress("0x456"),
	}

	mockClient.On("HeaderByNumber", mock.Anything, (*big.Int)(nil)).Return((*types.Header)(nil), errors.New("header error"))

	err := VoteAndReward(cProps)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get current block")

	mockClient.AssertExpectations(t)
}

// TestVoteAndReward_StartBlockError tests error when getting start block fails.
func TestVoteAndReward_StartBlockError(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Client:       mockClient,
		Kt:           mockKt,
		KtAddr:       common.HexToAddress("0x123"),
		MyPubKey:     common.HexToAddress("0x456"),
		ChainID:      big.NewInt(1),
		BlocksToWait: 0,
	}

	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = privateKey

	currentHeader := &types.Header{Number: big.NewInt(100)}
	mockClient.On("HeaderByNumber", mock.Anything, (*big.Int)(nil)).Return(currentHeader, nil)

	mockKt.On("StartBlock", mock.Anything).Return((*big.Int)(nil), errors.New("start block error"))

	err := VoteAndReward(cProps)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get start block")

	mockClient.AssertExpectations(t)
	mockKt.AssertExpectations(t)
}

// TestRewardWinningWallet_OCFeesSubtraction tests that reward amount is correctly calculated as balance - tlOcFees.
func TestRewardWinningWallet_OCFeesSubtraction(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Client:       mockClient,
		Kt:           mockKt,
		KtAddr:       common.HexToAddress("0x1234567890123456789012345678901234567890"),
		MyPubKey:     common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		ChainID:      big.NewInt(1),
		BlocksToWait: 0,
	}

	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = privateKey

	winner := common.HexToAddress("0xabc123456789012345678901234567890123456")
	totalMin := big.NewInt(1000) // Non-zero to avoid zero reward

	// Mock contract balance: 2 ETH
	contractBalance := big.NewInt(2000000000000000000) // 2 ETH in wei
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, (*big.Int)(nil)).Return(contractBalance, nil)

	// Mock tlOcFees: 0.5 ETH
	tlOcFees := big.NewInt(500000000000000000) // 0.5 ETH in wei
	mockKt.On("TlOcFees", mock.Anything).Return(tlOcFees, nil)

	// Expected reward amount: 2 ETH - 0.5 ETH = 1.5 ETH
	expectedRewardAmount := big.NewInt(1500000000000000000) // 1.5 ETH in wei
	rewardTx := types.NewTransaction(0, winner, expectedRewardAmount, 0, big.NewInt(0), []byte{})
	mockKt.On("Rwd", mock.Anything, winner, expectedRewardAmount).Return(rewardTx, nil)

	// Mock winner's balance before and after
	beforeBalance := big.NewInt(1000000000000000000)                                                     // 1 ETH
	mockClient.On("BalanceAt", mock.Anything, winner, (*big.Int)(nil)).Return(beforeBalance, nil).Once() // Before reward
	afterBalance := new(big.Int).Add(beforeBalance, expectedRewardAmount)
	mockClient.On("BalanceAt", mock.Anything, winner, (*big.Int)(nil)).Return(afterBalance, nil).Once() // After reward

	// Mock BlockNumber for WaitForBlocks
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(110), nil).Maybe()

	// Mock transaction receipt
	successReceipt := &types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(112)}
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(successReceipt, nil)

	err := rewardWinningWallet(cProps, winner, totalMin)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockKt.AssertExpectations(t)
}

// TestRewardWinningWallet_OCFeesExceedBalance tests that reward amount is set to 0 when tlOcFees >= balance.
func TestRewardWinningWallet_OCFeesExceedBalance(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Client:       mockClient,
		Kt:           mockKt,
		KtAddr:       common.HexToAddress("0x1234567890123456789012345678901234567890"),
		MyPubKey:     common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		ChainID:      big.NewInt(1),
		BlocksToWait: 0,
	}

	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = privateKey

	winner := common.HexToAddress("0xabc123456789012345678901234567890123456")
	totalMin := big.NewInt(1000) // Non-zero to avoid zero reward

	// Mock contract balance: 0.5 ETH
	contractBalance := big.NewInt(500000000000000000) // 0.5 ETH in wei
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, (*big.Int)(nil)).Return(contractBalance, nil)

	// Mock tlOcFees: 1 ETH (exceeds balance)
	tlOcFees := big.NewInt(1000000000000000000) // 1 ETH in wei
	mockKt.On("TlOcFees", mock.Anything).Return(tlOcFees, nil)

	// Expected reward amount: 0 (since balance < tlOcFees)
	expectedRewardAmount := big.NewInt(0)
	rewardTx := types.NewTransaction(0, winner, expectedRewardAmount, 0, big.NewInt(0), []byte{})
	mockKt.On("Rwd", mock.Anything, winner, expectedRewardAmount).Return(rewardTx, nil)

	// Mock winner's balance before and after (no change since reward is 0)
	beforeBalance := big.NewInt(1000000000000000000)                                                     // 1 ETH
	mockClient.On("BalanceAt", mock.Anything, winner, (*big.Int)(nil)).Return(beforeBalance, nil).Once() // Before reward
	afterBalance := new(big.Int).Add(beforeBalance, expectedRewardAmount)                                // Still 1 ETH
	mockClient.On("BalanceAt", mock.Anything, winner, (*big.Int)(nil)).Return(afterBalance, nil).Once()  // After reward

	// Mock BlockNumber for WaitForBlocks (not called since BlocksToWait=0, but added for consistency)
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(110), nil).Maybe()

	// Mock transaction receipt
	successReceipt := &types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(112)}
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(successReceipt, nil)

	err := rewardWinningWallet(cProps, winner, totalMin)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockKt.AssertExpectations(t)
}

// ============================================================================
// Phase 5c — rewardWinningWallet error-variant tests.
//
// Existing tests above cover the OC-fee subtraction math. These cover the
// error paths: Rwd failures (with the special "Epoch incomplete" swallow),
// BalanceAt failures, TlOcFees failures.

// rewardSetup builds the minimum ConnectionProps + mocks needed to exercise
// rewardWinningWallet up to the Rwd call.
func rewardSetup(t *testing.T) (
	cProps *ConnectionProps,
	mockClient *MockEthClient,
	mockKt *MockKtv2,
	winner common.Address,
	totalMin *big.Int,
) {
	t.Helper()
	logrus.SetLevel(logrus.FatalLevel)

	mockClient = &MockEthClient{}
	mockKt = &MockKtv2{}
	cProps = &ConnectionProps{
		Client:       mockClient,
		Kt:           mockKt,
		KtAddr:       common.HexToAddress("0x1234567890123456789012345678901234567890"),
		MyPubKey:     common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		ChainID:      big.NewInt(1),
		BlocksToWait: 0,
	}
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = priv
	winner = common.HexToAddress("0xabc123456789012345678901234567890123456")
	totalMin = big.NewInt(1000)
	return
}

// TestRewardWinningWallet_RwdReturnsEpochIncompleteIsSilent — when Rwd
// returns an "Epoch incomplete" error (another node beat us to the reward),
// the function logs a warning and returns nil.
func TestRewardWinningWallet_RwdReturnsEpochIncompleteIsSilent(t *testing.T) {
	cProps, mockClient, mockKt, winner, totalMin := rewardSetup(t)

	mockClient.On("BalanceAt", mock.Anything, winner, (*big.Int)(nil)).Return(big.NewInt(0), nil).Once()
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, (*big.Int)(nil)).Return(big.NewInt(int64(2e18)), nil)
	mockKt.On("TlOcFees", mock.Anything).Return(big.NewInt(0), nil)
	mockKt.On("Rwd", mock.Anything, winner, mock.AnythingOfType("*big.Int")).
		Return((*types.Transaction)(nil), errors.New("execution reverted: Epoch incomplete"))

	err := rewardWinningWallet(cProps, winner, totalMin)
	assert.NoError(t, err, "Epoch incomplete should be swallowed, not propagated")
}

// TestRewardWinningWallet_RwdReturnsOtherErrorPropagates — any other Rwd
// error is wrapped and returned.
func TestRewardWinningWallet_RwdReturnsOtherErrorPropagates(t *testing.T) {
	cProps, mockClient, mockKt, winner, totalMin := rewardSetup(t)

	mockClient.On("BalanceAt", mock.Anything, winner, (*big.Int)(nil)).Return(big.NewInt(0), nil).Once()
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, (*big.Int)(nil)).Return(big.NewInt(int64(2e18)), nil)
	mockKt.On("TlOcFees", mock.Anything).Return(big.NewInt(0), nil)
	mockKt.On("Rwd", mock.Anything, winner, mock.AnythingOfType("*big.Int")).
		Return((*types.Transaction)(nil), errors.New("revert: only OC can call rwd"))

	err := rewardWinningWallet(cProps, winner, totalMin)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to call rwd function")
}

// TestRewardWinningWallet_BalanceAtFailureBeforeReward — looking up the
// winner's pre-reward balance fails. No tx should be sent.
func TestRewardWinningWallet_BalanceAtFailureBeforeReward(t *testing.T) {
	cProps, mockClient, _, winner, totalMin := rewardSetup(t)

	mockClient.On("BalanceAt", mock.Anything, winner, (*big.Int)(nil)).
		Return((*big.Int)(nil), errors.New("rpc: timeout")).Once()

	err := rewardWinningWallet(cProps, winner, totalMin)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "winner's balance before reward")
}

// TestRewardWinningWallet_TlOcFeesFailurePropagates — TlOcFees on the
// contract fails. Function must error out before sending Rwd.
func TestRewardWinningWallet_TlOcFeesFailurePropagates(t *testing.T) {
	cProps, mockClient, mockKt, winner, totalMin := rewardSetup(t)

	mockClient.On("BalanceAt", mock.Anything, winner, (*big.Int)(nil)).Return(big.NewInt(0), nil).Once()
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, (*big.Int)(nil)).Return(big.NewInt(int64(2e18)), nil)
	mockKt.On("TlOcFees", mock.Anything).Return((*big.Int)(nil), errors.New("rpc: connection refused"))

	err := rewardWinningWallet(cProps, winner, totalMin)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "total OC fees")
}

// ============================================================================
// Phase 6b — confirmation-depth tests.
//
// The lottery seed block is now `endBlock + cProps.ConfirmationDepth` (was
// always +1 pre-Phase-6b). These tests pin both the configurable behavior
// and the default-of-5 fallback so a future refactor can't quietly flip
// either back.

// TestCalculateVoteAndReward_UsesConfirmationDepthForSeedBlock — set
// ConfirmationDepth = 7; assert HeaderByNumber is queried for endBlock+7
// (not +1), and the hash from THAT block is what calcWinningWallet sees.
func TestCalculateVoteAndReward_UsesConfirmationDepthForSeedBlock(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Client:            mockClient,
		Kt:                mockKt,
		KtAddr:            common.HexToAddress("0x1234567890123456789012345678901234567890"),
		MyPubKey:          common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		ChainID:           big.NewInt(1),
		BlocksToWait:      0,
		ChunkSize:         500,
		ConfirmationDepth: 7,
	}
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = priv

	stakerAddr := common.HexToAddress("0xabc123456789012345678901234567890123456")
	startBlock := big.NewInt(50)
	endBlock := big.NewInt(110)
	seedBlockNum := new(big.Int).Add(endBlock, big.NewInt(7))

	stakeDataMins := map[common.Address]*UserStakeData{
		stakerAddr: {StakeAmount: big.NewInt(1000), Prob: new(big.Float).SetFloat64(1.0)},
	}

	// Capture the block hash that calcWinningWallet is called with.
	expectedSeedHash := common.HexToHash("0xdeadbeef00000000000000000000000000000000000000000000000000000000")
	var capturedHash common.Hash
	origCalc := calcWinningWallet
	SetCalculateWinningWallet(func(_ map[common.Address]*UserStakeData, h common.Hash) (common.Address, error) {
		capturedHash = h
		return stakerAddr, nil
	})
	defer SetCalculateWinningWallet(origCalc)

	// HeaderByNumber must be queried at endBlock+7, NOT endBlock+1.
	seedHeader := &types.Header{Number: seedBlockNum, ParentHash: expectedSeedHash}
	// The block hash exposed by Hash() depends on header contents; rather than
	// reconstruct it, assert the function asked for the right block number.
	mockClient.On("HeaderByNumber", mock.Anything, seedBlockNum).Return(seedHeader, nil)

	mockKt.On("Vote", mock.Anything, stakerAddr, mock.AnythingOfType("string")).Return(
		types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), []byte{}), nil)
	mockKt.On("BlockRwd", mock.Anything, startBlock, stakerAddr).Return(uint16(1), nil)
	mockKt.On("ConsensusReq", mock.Anything).Return(uint16(5), nil) // vote insufficient → no reward
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(
		&types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(112)}, nil)

	// WaitForBlocks polls BlockNumber once even when BlocksToWait=0.
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(200), nil).Maybe()

	_, err := calculateVoteAndReward(stakeDataMins, startBlock, endBlock, cProps, big.NewInt(1000))
	assert.NoError(t, err)
	mockClient.AssertCalled(t, "HeaderByNumber", mock.Anything, seedBlockNum)
	if capturedHash != seedHeader.Hash() {
		t.Errorf("calcWinningWallet got block hash %s, want %s (seed block %d's hash)",
			capturedHash.Hex(), seedHeader.Hash().Hex(), seedBlockNum.Uint64())
	}
}

// TestCalculateVoteAndReward_ZeroConfirmationDepthFallsBackToDefault —
// when cProps.ConfirmationDepth is 0, the code uses DefaultConfirmationDepth.
// Pin DefaultConfirmationDepth = 5.
func TestCalculateVoteAndReward_ZeroConfirmationDepthFallsBackToDefault(t *testing.T) {
	if DefaultConfirmationDepth != 5 {
		t.Errorf("DefaultConfirmationDepth should be 5 (the operator-facing convention); got %d",
			DefaultConfirmationDepth)
	}

	logrus.SetLevel(logrus.FatalLevel)
	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Client:       mockClient,
		Kt:           mockKt,
		KtAddr:       common.HexToAddress("0x1234567890123456789012345678901234567890"),
		MyPubKey:     common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		ChainID:      big.NewInt(1),
		BlocksToWait: 0,
		ChunkSize:    500,
		// ConfirmationDepth intentionally 0 → should use default.
	}
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = priv

	stakerAddr := common.HexToAddress("0xabc123456789012345678901234567890123456")
	endBlock := big.NewInt(110)
	expectedSeedNum := new(big.Int).Add(endBlock, big.NewInt(int64(DefaultConfirmationDepth)))

	stakeDataMins := map[common.Address]*UserStakeData{
		stakerAddr: {StakeAmount: big.NewInt(1000), Prob: new(big.Float).SetFloat64(1.0)},
	}

	origCalc := calcWinningWallet
	SetCalculateWinningWallet(func(_ map[common.Address]*UserStakeData, _ common.Hash) (common.Address, error) {
		return stakerAddr, nil
	})
	defer SetCalculateWinningWallet(origCalc)

	mockClient.On("HeaderByNumber", mock.Anything, expectedSeedNum).Return(
		&types.Header{Number: expectedSeedNum}, nil)
	mockKt.On("Vote", mock.Anything, stakerAddr, mock.AnythingOfType("string")).Return(
		types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), []byte{}), nil)
	mockKt.On("BlockRwd", mock.Anything, big.NewInt(50), stakerAddr).Return(uint16(1), nil)
	mockKt.On("ConsensusReq", mock.Anything).Return(uint16(5), nil)
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(
		&types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(115)}, nil)

	// WaitForBlocks polls BlockNumber once even when BlocksToWait=0.
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(200), nil).Maybe()

	_, err := calculateVoteAndReward(stakeDataMins, big.NewInt(50), endBlock, cProps, big.NewInt(1000))
	assert.NoError(t, err)
	mockClient.AssertCalled(t, "HeaderByNumber", mock.Anything, expectedSeedNum)
}
