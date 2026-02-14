package ktfunc

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type GatherFunc func(*ConnectionProps, Ktv2Interface, *big.Int, *big.Int) (map[common.Address]map[uint64]*UserStakeData, error)
type CalcFunc func(map[common.Address]*UserStakeData, common.Hash) (common.Address, error)

// Rest of helpers unchanged...
func createMockStakeDataMins() map[common.Address]*UserStakeData {
	stakes := map[common.Address]*UserStakeData{}
	addresses := []string{"0x1111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222",
		"0x3333333333333333333333333333333333333333", "0x4444444444444444444444444444444444444444",
		"0x5555555555555555555555555555555555555555"}
	totalMin := big.NewInt(500) // 100 each for 5
	for _, addrStr := range addresses {
		addr := common.HexToAddress(addrStr)
		stakes[addr] = &UserStakeData{
			StakeAmount: big.NewInt(100),
		}
	}
	// Pre-calculate probs
	calculateProbsForEachWallet(stakes, totalMin, true)
	return stakes
}

func mockGatherStakesAndWithdraws(_ *ConnectionProps, _ Ktv2Interface, _ *big.Int, _ *big.Int) (map[common.Address]map[uint64]*UserStakeData, error) {
	// Return pre-built stake data
	stakeDataMap := make(map[common.Address]map[uint64]*UserStakeData)
	for addr, data := range createMockStakeDataMins() {
		stakeDataMap[addr] = map[uint64]*UserStakeData{
			100: {StakeAmount: data.StakeAmount}, // Dummy block
		}
	}
	return stakeDataMap, nil
}

func TestMultiVoterConsensus(t *testing.T) {
	epochStart := big.NewInt(100)
	epochInterval := uint16(10)
	consensusReq := uint16(3)
	winner := common.HexToAddress("0x1111111111111111111111111111111111111111")
	header111 := &types.Header{Number: big.NewInt(111)}

	dummyPriv, _ := crypto.HexToECDSA("fad9c8855b7dfdac54d5d47b0b1ef9e1a7a3c0b5a5d4e5f6a7b8c9d0e1f2a3b4")

	voterPubKeys := []common.Address{
		common.HexToAddress("0xa1"), common.HexToAddress("0xa2"), common.HexToAddress("0xa3"),
		common.HexToAddress("0xa4"), common.HexToAddress("0xa5"),
	}
	tests := []struct {
		name            string
		votersToSucceed int
		expectConsensus bool
		expectReward    bool
	}{
		{"3 voters succeed - consensus", 3, true, true},
		{"2 voters succeed - no consensus", 2, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockEthClient{}
			mockKt := &MockKtv2{}
			// Common setups
			mockClient.On("HeaderByNumber", mock.Anything, mock.Anything).Return(header111, nil).Maybe()
			mockClient.On("BlockNumber", mock.Anything).Return(uint64(111), nil).Maybe()
			mockClient.On("BalanceAt", mock.Anything, mock.AnythingOfType("common.Address"), mock.Anything).Return(big.NewInt(1e18), nil).Maybe()
			mockClient.On("CodeAt", mock.Anything, mock.AnythingOfType("common.Address"), mock.Anything).Return([]byte{0x01}, nil).Maybe()
			mockClient.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20e9), nil).Maybe()
			mockClient.On("PendingNonceAt", mock.Anything, mock.AnythingOfType("common.Address")).Return(uint64(0), nil).Maybe()
			mockClient.On("TransactionReceipt", mock.Anything, mock.AnythingOfType("common.Hash")).Return(&types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(111)}, nil).Maybe()
			mockKt.On("StartBlock", mock.Anything).Return(epochStart, nil)
			mockKt.On("EpochInterval", mock.Anything).Return(epochInterval, nil)
			mockKt.On("ConsensusReq", mock.Anything).Return(consensusReq, nil)
			mockKt.On("Declines", mock.Anything, mock.Anything).Return(false, nil)

			goodVotes := 0
			dummyTx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 21000, big.NewInt(20000000000), nil)
			originalGather := GatherStakesAndWithdraws
			GatherStakesAndWithdraws = mockGatherStakesAndWithdraws
			defer func() { GatherStakesAndWithdraws = originalGather }()
			originalCalc := calcWinningWallet
			defer func() { calcWinningWallet = originalCalc }()
			// Mock BlockRwd for winner (stateful)
			mockKt.On("BlockRwd", mock.Anything, epochStart, winner).Return(
				func(args mock.Arguments) uint16 {
					return uint16(goodVotes)
				},
				nil,
			).Maybe()
			// Mock BlockRwd for bad winners (always 0)
			mockKt.On("BlockRwd", mock.Anything, epochStart, mock.MatchedBy(func(addr common.Address) bool { return addr != winner })).Return(uint16(0), nil).Maybe()
			mockKt.On("BlockRwd", mock.Anything, mock.Anything, mock.AnythingOfType("common.Address")).Return(uint16(0), nil).Maybe()
			// Mock Vote
			mockKt.On("Vote", mock.Anything, mock.AnythingOfType("common.Address"), mock.AnythingOfType("string")).Run(func(args mock.Arguments) {
				rec := args.Get(1).(common.Address)
				if rec == winner {
					goodVotes++
				}
			}).Return(dummyTx, nil).Times(len(voterPubKeys))
			mockKt.On("Vote", mock.Anything, mock.AnythingOfType("common.Address"), mock.AnythingOfType("string")).Return(dummyTx, nil).Maybe()
			// Mock Rwd only for winner if expected
			if tt.expectReward {
				mockKt.On("Rwd", mock.Anything, winner, mock.AnythingOfType("*big.Int")).Return(dummyTx, nil).Once()
			}
			// Common props
			props := &ConnectionProps{
				Client:       mockClient,
				Kt:           mockKt,
				KtAddr:       common.HexToAddress("0xktaddr"),
				ChainID:      big.NewInt(1),
				BlocksToWait: 0,
				ChunkSize:    10,
				MyPrivateKey: dummyPriv,
			}
			for i := 0; i < len(voterPubKeys); i++ {
				if i < tt.votersToSucceed {
					calcWinningWallet = func(_ map[common.Address]*UserStakeData, h common.Hash) (common.Address, error) {
						return winner, nil
					}
				} else {
					badWinner := common.HexToAddress(fmt.Sprintf("0xB%x", i+10))
					calcWinningWallet = func(_ map[common.Address]*UserStakeData, h common.Hash) (common.Address, error) {
						return badWinner, nil
					}
				}
				props.MyPubKey = voterPubKeys[i]
				err := VoteAndReward(props)
				assert.NoError(t, err)
			}
			// Verify consensus
			count, req, err := getVoteCountAndRequired(props, epochStart, winner)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectConsensus, count >= req)
			mockKt.AssertExpectations(t)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestImproperVoterOverruled(t *testing.T) {
	dummyPriv, _ := crypto.HexToECDSA("fad9c8855b7dfdac54d5d47b0b1ef9e1a7a3c0b5a5d4e5f6a7b8c9d0e1f2a3b4")
	dummyTx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 21000, big.NewInt(20000000000), nil)

	epochStart := big.NewInt(100)
	epochInterval := uint16(10)
	consensusReq := uint16(3)
	winner := common.HexToAddress("0x1111111111111111111111111111111111111111")
	badWinner := common.HexToAddress("0x9999999999999999999999999999999999999999")
	goodVotes := 2 // Start with 2 votes for winner
	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	header111 := &types.Header{Number: big.NewInt(111)}
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(111), nil)
	mockClient.On("HeaderByNumber", mock.Anything, mock.Anything).Return(header111, nil)
	mockClient.On("BalanceAt", mock.Anything, mock.Anything, mock.Anything).Return(big.NewInt(1e18), nil)
	mockClient.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20e9), nil).Maybe()
	mockClient.On("PendingNonceAt", mock.Anything, mock.AnythingOfType("common.Address")).Return(uint64(0), nil).Maybe()
	mockClient.On("CodeAt", mock.Anything, mock.AnythingOfType("common.Address"), mock.Anything).Return([]byte{0x01}, nil)
	mockClient.On("TransactionReceipt", mock.Anything, mock.AnythingOfType("common.Hash")).Return(&types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(111)}, nil)
	mockKt.On("StartBlock", mock.Anything).Return(epochStart, nil)
	mockKt.On("EpochInterval", mock.Anything).Return(epochInterval, nil)
	mockKt.On("ConsensusReq", mock.Anything).Return(consensusReq, nil)
	mockKt.On("Declines", mock.Anything, mock.Anything).Return(false, nil)
	// Dynamic for winner
	mockKt.On("BlockRwd", mock.Anything, epochStart, winner).Return(
		func(args mock.Arguments) uint16 {
			return uint16(goodVotes)
		},
		nil,
	)
	mockKt.On("BlockRwd", mock.Anything, epochStart, badWinner).Return(uint16(1), nil)
	mockKt.On("Vote", mock.Anything, mock.AnythingOfType("common.Address"), mock.AnythingOfType("string")).Run(func(args mock.Arguments) {
		rec := args.Get(1).(common.Address)
		if rec == winner {
			goodVotes++
		}
	}).Return(dummyTx, nil)
	mockKt.On("Rwd", mock.Anything, winner, mock.AnythingOfType("*big.Int")).Return(dummyTx, nil)
	originalGather := GatherStakesAndWithdraws
	GatherStakesAndWithdraws = mockGatherStakesAndWithdraws
	defer func() { GatherStakesAndWithdraws = originalGather }()
	propsGood := &ConnectionProps{
		Client:       mockClient,
		Kt:           mockKt,
		KtAddr:       common.HexToAddress("0xktaddr"),
		MyPubKey:     common.HexToAddress("0xgood"),
		ChainID:      big.NewInt(1),
		BlocksToWait: 0,
		MyPrivateKey: dummyPriv,
	}
	calcWinningWallet = func(_ map[common.Address]*UserStakeData, _ common.Hash) (common.Address, error) {
		return winner, nil
	}
	defer func() { calcWinningWallet = defaultCalculateWinningWallet }()
	err := VoteAndReward(propsGood)
	assert.NoError(t, err)
	// Bad voter
	calcWinningWallet = func(_ map[common.Address]*UserStakeData, _ common.Hash) (common.Address, error) {
		return badWinner, nil
	}
	defer func() { calcWinningWallet = defaultCalculateWinningWallet }()
	propsBad := &ConnectionProps{
		Client:       mockClient,
		Kt:           mockKt,
		KtAddr:       common.HexToAddress("0xktaddr"),
		MyPubKey:     common.HexToAddress("0xbad"),
		ChainID:      big.NewInt(1),
		BlocksToWait: 0,
		MyPrivateKey: dummyPriv,
	}
	err = VoteAndReward(propsBad)
	assert.NoError(t, err)
	mockKt.AssertNumberOfCalls(t, "Vote", 2) // Both voted
	mockKt.AssertNumberOfCalls(t, "Rwd", 1)  // Only good rewarded
	mockKt.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestDeclinedStakerNotSelected(t *testing.T) {
	// This test should fail initially because declined stakers are included in probability calculation
	// After implementing filtering, it should pass

	// Two addresses, equal stake
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111") // will be sorted first
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222") // declined, will be sorted second

	stakeDataMins := map[common.Address]*UserStakeData{
		addr1: {StakeAmount: big.NewInt(100)},
		addr2: {StakeAmount: big.NewInt(100)}, // declined
	}
	totalMin := big.NewInt(200)

	// Mock cProps for filtering
	mockKt := &MockKtv2{}
	mockKt.On("Declines", mock.Anything, addr1).Return(false, nil) // addr1 not declined
	mockKt.On("Declines", mock.Anything, addr2).Return(true, nil)  // addr2 declined

	cProps := &ConnectionProps{
		Kt: mockKt,
	}

	// Filter declined stakers
	err := filterDeclinedStakers(stakeDataMins, cProps)
	assert.NoError(t, err)

	// Recalculate totalMin after filtering
	totalMin = big.NewInt(0)
	for _, data := range stakeDataMins {
		totalMin.Add(totalMin, data.StakeAmount)
	}

	// Calculate probs
	calculateProbsForEachWallet(stakeDataMins, totalMin, true)

	// Create a block hash that would select the second address if it existed
	// But since it's filtered out, it should select addr1
	hashBytes := make([]byte, 32)
	hashBytes[0] = 0x80 // set high bit
	blockHash := common.BytesToHash(hashBytes)

	winner, err := defaultCalculateWinningWallet(stakeDataMins, blockHash)
	assert.NoError(t, err)

	// After filtering, addr2 is removed, so only addr1 remains, winner should be addr1
	assert.Equal(t, addr1, winner, "Winner should be the non-declined staker")
	mockKt.AssertExpectations(t)
}
