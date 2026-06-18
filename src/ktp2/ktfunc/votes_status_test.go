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

// TestGatherEpochVotes_ReconstructsPerOCVotesWithResetNetting — two OCs vote,
// one resets and re-votes a different candidate, and an event from a different
// epoch is present. The reconstruction must reflect each OC's final active
// vote (no duplicates, prior epoch ignored).
func TestGatherEpochVotes_ReconstructsPerOCVotesWithResetNetting(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	chainID := big.NewInt(1337)
	xTx, addrX := signedTxFromKey(t, testKeyX, chainID)
	yTx, addrY := signedTxFromKey(t, testKeyY, chainID)

	startBlock := big.NewInt(1000)
	candA := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	candB := common.HexToAddress("0x00000000000000000000000000000000000000bb")
	candC := common.HexToAddress("0x00000000000000000000000000000000000000cc")

	h1 := common.HexToHash("0x01")
	h2 := common.HexToHash("0x02")
	h3 := common.HexToHash("0x03")
	h4 := common.HexToHash("0x04")
	h5 := common.HexToHash("0x05")

	events := []*ktv2.Ktv2Voted{
		{Arg0: big.NewInt(1000), Arg1: candA, Arg2: "data", Raw: types.Log{TxHash: h1, BlockNumber: 1100}}, // X -> A
		{Arg0: big.NewInt(1000), Arg1: candB, Arg2: "data", Raw: types.Log{TxHash: h2, BlockNumber: 1101}}, // Y -> B
		{Arg0: big.NewInt(1000), Arg1: candA, Arg2: "rst", Raw: types.Log{TxHash: h3, BlockNumber: 1102}},  // X resets A
		{Arg0: big.NewInt(1000), Arg1: candC, Arg2: "data", Raw: types.Log{TxHash: h4, BlockNumber: 1103}}, // X -> C
		{Arg0: big.NewInt(500), Arg1: candA, Arg2: "data", Raw: types.Log{TxHash: h5, BlockNumber: 1104}},  // different epoch
	}

	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	mockKt.On("FilterVoted", mock.Anything).Return(&mockVotedIter{events: events}, nil)
	mockClient.On("TransactionByHash", mock.Anything, h1).Return(xTx, false, nil)
	mockClient.On("TransactionByHash", mock.Anything, h2).Return(yTx, false, nil)
	mockClient.On("TransactionByHash", mock.Anything, h3).Return(xTx, false, nil)
	mockClient.On("TransactionByHash", mock.Anything, h4).Return(xTx, false, nil)

	cProps := &ConnectionProps{Kt: mockKt, Client: mockClient, ChunkSize: 10_000_000}

	votes, err := gatherEpochVotes(cProps, startBlock, 2000)
	assert.NoError(t, err)

	// X's final active vote is candC (A was reset); Y's is candB. X listed first.
	assert.Equal(t, []epochVote{
		{Voter: addrX, Candidate: candC},
		{Voter: addrY, Candidate: candB},
	}, votes)
}

// TestGatherEpochVotes_ResetOnlyVoterExcluded — an OC that votes then resets,
// with no re-vote, has no active vote and must not appear.
func TestGatherEpochVotes_ResetOnlyVoterExcluded(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	chainID := big.NewInt(1337)
	xTx, _ := signedTxFromKey(t, testKeyX, chainID)

	candA := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	h1 := common.HexToHash("0x11")
	h2 := common.HexToHash("0x12")

	events := []*ktv2.Ktv2Voted{
		{Arg0: big.NewInt(1000), Arg1: candA, Arg2: "data", Raw: types.Log{TxHash: h1, BlockNumber: 1100}},
		{Arg0: big.NewInt(1000), Arg1: candA, Arg2: "rst", Raw: types.Log{TxHash: h2, BlockNumber: 1101}},
	}

	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	mockKt.On("FilterVoted", mock.Anything).Return(&mockVotedIter{events: events}, nil)
	mockClient.On("TransactionByHash", mock.Anything, mock.Anything).Return(xTx, false, nil)

	cProps := &ConnectionProps{Kt: mockKt, Client: mockClient, ChunkSize: 10_000_000}

	votes, err := gatherEpochVotes(cProps, big.NewInt(1000), 2000)
	assert.NoError(t, err)
	assert.Empty(t, votes)
}

// TestPrintEpochVoteStatus_HappyPath drives the full status path end to end and
// asserts it completes without error.
func TestPrintEpochVoteStatus_HappyPath(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	chainID := big.NewInt(1337)
	xTx, _ := signedTxFromKey(t, testKeyX, chainID)

	candA := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	h1 := common.HexToHash("0x21")

	events := []*ktv2.Ktv2Voted{
		{Arg0: big.NewInt(1000), Arg1: candA, Arg2: "data", Raw: types.Log{TxHash: h1, BlockNumber: 1100}},
	}

	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	mockKt.On("StartBlock", mock.Anything).Return(big.NewInt(1000), nil)
	mockKt.On("EpochInterval", mock.Anything).Return(uint16(60), nil)
	mockKt.On("ConsensusReq", mock.Anything).Return(uint16(2), nil)
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(2000), nil)
	mockKt.On("FilterVoted", mock.Anything).Return(&mockVotedIter{events: events}, nil)
	mockClient.On("TransactionByHash", mock.Anything, h1).Return(xTx, false, nil)
	mockKt.On("BlockRwd", mock.Anything, big.NewInt(1000), candA).Return(uint16(1), nil)

	cProps := &ConnectionProps{Kt: mockKt, Client: mockClient, ChunkSize: 10_000_000}
	assert.NoError(t, PrintEpochVoteStatus(cProps))
}
