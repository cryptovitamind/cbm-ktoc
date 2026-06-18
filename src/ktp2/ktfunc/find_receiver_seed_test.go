package ktfunc

// Tests pinning the consensus-critical property that the lottery seed block
// is a FIXED offset past the epoch end (endBlock + SeedOffset) and does NOT
// depend on the per-operator confirmationDepth. confirmationDepth only delays
// submission. If two operators on different depths seeded from different
// blocks they would compute different winners and split the vote.

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// seedTestHeader returns a header whose Hash() is distinct per block number,
// so a test can tell which block the seed was sampled from.
func seedTestHeader(n uint64) *types.Header {
	return &types.Header{Number: new(big.Int).SetUint64(n), Extra: []byte("seedtest")}
}

func TestCalculateVoteAndReward_SeedBlockIsConstantRegardlessOfDepth(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	startBlock := big.NewInt(50)
	endBlock := big.NewInt(110)

	// Run calculateVoteAndReward with a stub winner-calculator that records the
	// seed hash it was handed and then aborts (returns an error) so the heavy
	// vote/reward path never runs.
	runWithDepth := func(depth uint64) common.Hash {
		mockClient := &MockEthClient{}
		mockKt := &MockKtv2{}
		cProps := &ConnectionProps{
			Client:            mockClient,
			Kt:                mockKt,
			ConfirmationDepth: depth,
			MyPubKey:          common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		}

		// Provide headers for every block number either the current or a
		// regressed code path could request: the seed block (endBlock+SeedOffset),
		// the confirmation block for this depth (seed+depth), and the
		// would-be-buggy seed at endBlock+depth.
		seedBlock := endBlock.Uint64() + SeedOffset
		for _, n := range []uint64{seedBlock, seedBlock + depth, endBlock.Uint64() + depth} {
			mockClient.On("HeaderByNumber", mock.Anything, new(big.Int).SetUint64(n)).
				Return(seedTestHeader(n), nil).Maybe()
		}

		var captured common.Hash
		orig := calcWinningWallet
		SetCalculateWinningWallet(func(_ map[common.Address]*UserStakeData, h common.Hash) (common.Address, error) {
			captured = h
			return common.Address{}, errors.New("captured seed; stop before voting")
		})
		defer func() { calcWinningWallet = orig }()

		_, _ = calculateVoteAndReward(map[common.Address]*UserStakeData{}, startBlock, endBlock, cProps, big.NewInt(0))
		return captured
	}

	expected := seedTestHeader(endBlock.Uint64() + SeedOffset).Hash()
	h5 := runWithDepth(5)
	h50 := runWithDepth(50)

	assert.Equal(t, expected, h5, "depth=5 must seed from endBlock+SeedOffset")
	assert.Equal(t, h5, h50, "seed block must not change with confirmationDepth")
}

func TestCalculateVoteAndReward_WaitsForSeedPlusDepth(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	startBlock := big.NewInt(50)
	endBlock := big.NewInt(110)
	depth := uint64(7)
	seedBlock := endBlock.Uint64() + SeedOffset // 115
	requiredBlock := seedBlock + depth          // 122

	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Client:            mockClient,
		Kt:                mockKt,
		ConfirmationDepth: depth,
		MyPubKey:          common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
	}

	mockClient.On("HeaderByNumber", mock.Anything, new(big.Int).SetUint64(requiredBlock)).
		Return(seedTestHeader(requiredBlock), nil)
	mockClient.On("HeaderByNumber", mock.Anything, new(big.Int).SetUint64(seedBlock)).
		Return(seedTestHeader(seedBlock), nil)
	// Permissive fallback for any other block the code might probe.
	mockClient.On("HeaderByNumber", mock.Anything, mock.Anything).
		Return(seedTestHeader(0), nil).Maybe()

	orig := calcWinningWallet
	SetCalculateWinningWallet(func(_ map[common.Address]*UserStakeData, _ common.Hash) (common.Address, error) {
		return common.Address{}, errors.New("stop before voting")
	})
	defer func() { calcWinningWallet = orig }()

	_, _ = calculateVoteAndReward(map[common.Address]*UserStakeData{}, startBlock, endBlock, cProps, big.NewInt(0))

	// The node must wait for the seed block to be buried by confirmationDepth
	// blocks — i.e. it must observe block (seedBlock + depth) before voting.
	mockClient.AssertCalled(t, "HeaderByNumber", mock.Anything, new(big.Int).SetUint64(requiredBlock))
}
