package ktfunc

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func manualOpsProps(t *testing.T) (*ConnectionProps, *MockKtv2, *MockEthClient) {
	t.Helper()
	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{
		Kt:           mockKt,
		Client:       mockClient,
		KtAddr:       common.HexToAddress("0x1234567890123456789012345678901234567890"),
		MyPubKey:     common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		ChainID:      big.NewInt(1),
		BlocksToWait: 0,
	}
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = priv
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(
		&types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(10)}, nil)
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Maybe()
	return cProps, mockKt, mockClient
}

func TestVoteForAddress_VotesWithOverrideMarkerAndReportsStatus(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	cProps, mockKt, _ := manualOpsProps(t)
	target := common.HexToAddress("0x00000000000000000000000000000000000000aa")

	voteTx := types.NewTransaction(0, target, big.NewInt(0), 0, big.NewInt(0), nil)
	mockKt.On("Vote", mock.Anything, target, manualVoteData).Return(voteTx, nil)
	mockKt.On("StartBlock", mock.Anything).Return(big.NewInt(1000), nil)
	mockKt.On("BlockRwd", mock.Anything, big.NewInt(1000), target).Return(uint16(2), nil)
	mockKt.On("ConsensusReq", mock.Anything).Return(uint16(2), nil)

	assert.NoError(t, VoteForAddress(cProps, target))
	// Manual votes are tagged so they're distinguishable from algorithmic ones.
	mockKt.AssertCalled(t, "Vote", mock.Anything, target, manualVoteData)
}

func TestResetLotteryVote_CallsContractResetVote(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	cProps, mockKt, _ := manualOpsProps(t)
	target := common.HexToAddress("0x00000000000000000000000000000000000000bb")

	resetTx := types.NewTransaction(0, target, big.NewInt(0), 0, big.NewInt(0), nil)
	mockKt.On("ResetVote", mock.Anything, target).Return(resetTx, nil)

	assert.NoError(t, ResetLotteryVote(cProps, target))
	mockKt.AssertCalled(t, "ResetVote", mock.Anything, target)
}
