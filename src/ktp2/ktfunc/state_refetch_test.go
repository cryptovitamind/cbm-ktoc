package ktfunc

// Pins that contract state which gates a transaction or the lottery seed is
// read fresh on every use, NOT served from a stale cache. A stale read of any
// of these makes the node act on an already-rewarded epoch or miscompute the
// reward, and the on-chain tx then reverts.

import (
	"math/big"
	"os"
	"testing"

	"ktp2/src/abis/ktv2"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// voteAndRewardTestProps builds a ConnectionProps + mocks wired so that
// VoteAndReward runs end-to-end through a no-stakes epoch (winner = dead
// address, no reward). Returns the props and its two mocks so a test can
// assert on call counts. The seed block is endBlock+SeedOffset and the node
// waits ConfirmationDepth(1) more blocks.
func voteAndRewardTestProps(t *testing.T) (*ConnectionProps, *MockEthClient, *MockKtv2) {
	t.Helper()
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
		ConfirmationDepth: 1,
	}
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = priv

	currentHeader := &types.Header{Number: big.NewInt(120)}
	mockClient.On("HeaderByNumber", mock.Anything, (*big.Int)(nil)).Return(currentHeader, nil)
	mockClient.On("HeaderByNumber", mock.Anything, mock.Anything).Return(&types.Header{}, nil).Maybe()

	endBlock := big.NewInt(110) // startBlock(50) + interval(60)
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, endBlock).Return(big.NewInt(1e18), nil)
	mockClient.On("CodeAt", mock.Anything, cProps.KtAddr, mock.Anything).Return([]byte{0x60, 0x80}, nil)
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(120), nil)

	emptyStaked := &MockStakedIterator{events: []*ktv2.Ktv2Staked{}}
	emptyWithdrew := &MockWithdrewIterator{events: []*ktv2.Ktv2Withdrew{}}
	mockKt.On("FilterStaked", mock.Anything).Return(emptyStaked, nil).Maybe()
	mockKt.On("FilterWithdrew", mock.Anything).Return(emptyWithdrew, nil).Maybe()

	zeroAddr := common.Address{}
	mockKt.On("Vote", mock.Anything, zeroAddr, mock.AnythingOfType("string")).Return(
		types.NewTransaction(0, zeroAddr, big.NewInt(0), 0, big.NewInt(0), []byte{}), nil).Maybe()
	mockKt.On("BlockRwd", mock.Anything, mock.Anything, zeroAddr).Return(uint16(0), nil).Maybe()
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(
		&types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(112)}, nil).Maybe()

	return cProps, mockClient, mockKt
}

func TestVoteAndReward_RefetchesStartBlockAndIntervalEachCall(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	_ = os.RemoveAll("cache")
	t.Cleanup(func() { _ = os.RemoveAll("cache") })

	cProps, _, mockKt := voteAndRewardTestProps(t)
	mockKt.On("StartBlock", mock.Anything).Return(big.NewInt(50), nil)
	mockKt.On("EpochInterval", mock.Anything).Return(uint16(60), nil)
	mockKt.On("ConsensusReq", mock.Anything).Return(uint16(1), nil).Maybe()

	for i := 0; i < 2; i++ {
		if err := VoteAndReward(cProps); err != nil {
			t.Fatalf("VoteAndReward call %d: %v", i, err)
		}
	}

	// Both reads gate the epoch boundary / seed, so each must hit the chain
	// once per cycle — never served from a cross-cycle cache.
	mockKt.AssertNumberOfCalls(t, "StartBlock", 2)
	mockKt.AssertNumberOfCalls(t, "EpochInterval", 2)
}

func TestRewardWinningWallet_RefetchesTlOcFeesPerCall(t *testing.T) {
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
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = priv

	winner := common.HexToAddress("0xabc123456789012345678901234567890123456")
	mockClient.On("BalanceAt", mock.Anything, winner, (*big.Int)(nil)).Return(big.NewInt(0), nil)
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, (*big.Int)(nil)).Return(big.NewInt(1e18), nil)
	mockKt.On("TlOcFees", mock.Anything).Return(big.NewInt(0), nil)
	mockKt.On("Rwd", mock.Anything, winner, mock.Anything).Return(
		types.NewTransaction(0, winner, big.NewInt(0), 0, big.NewInt(0), []byte{}), nil)
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(
		&types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(112)}, nil)
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(120), nil).Maybe()

	for i := 0; i < 2; i++ {
		if err := rewardWinningWallet(cProps, winner, big.NewInt(1000)); err != nil {
			t.Fatalf("rewardWinningWallet call %d: %v", i, err)
		}
	}

	// tlOcFees feeds the reward amount; a stale value reverts the reward tx,
	// so it must be read fresh every reward.
	mockKt.AssertNumberOfCalls(t, "TlOcFees", 2)
}
