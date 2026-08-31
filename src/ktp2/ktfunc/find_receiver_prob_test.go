package ktfunc

import (
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// Helper function to create UserStakeData with a given stake amount
func createUserStakeData1(stake string) *UserStakeData {
	stakeInt, _ := new(big.Int).SetString(stake, 10)
	return &UserStakeData{
		StakeAmount: stakeInt,
		Prob:        nil, // Prob starts as nil; function will set it
	}
}

// Test with a nil map
func TestCalculateProbsForEachWallet_NilMap(t *testing.T) {
	found := calculateProbsForEachWallet(nil, big.NewInt(0))
	if found {
		t.Errorf("Expected found=false for nil map, got true")
	}
}

// Test with an empty map
func TestCalculateProbsForEachWallet_EmptyMap(t *testing.T) {
	stakeDataMinsMap := make(map[common.Address]*UserStakeData)
	found := calculateProbsForEachWallet(stakeDataMinsMap, big.NewInt(0))
	if found {
		t.Errorf("Expected found=false for empty map, got true")
	}
}

// Test with a single address
func TestCalculateProbsForEachWallet_SingleAddress(t *testing.T) {
	addr := common.HexToAddress("0x1")
	stakeDataMinsMap := map[common.Address]*UserStakeData{
		addr: createUserStakeData1("100"),
	}
	totalMin := big.NewInt(100)

	found := calculateProbsForEachWallet(stakeDataMinsMap, totalMin)
	if !found {
		t.Errorf("Expected found=true, got false")
	}

	prob, _ := stakeDataMinsMap[addr].Prob.Float64()
	expectedProb := 1.0 // sole staker always has the whole probability mass
	if prob != expectedProb {
		t.Errorf("Expected probability %f, got %f", expectedProb, prob)
	}
}

// Test with a zero stake amount
func TestCalculateProbsForEachWallet_ZeroStake(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	stakeDataMinsMap := map[common.Address]*UserStakeData{
		addr1: createUserStakeData1("0"),
		addr2: createUserStakeData1("100"),
	}
	totalMin := big.NewInt(100)

	found := calculateProbsForEachWallet(stakeDataMinsMap, totalMin)
	if !found {
		t.Errorf("Expected found=true, got false")
	}

	prob1, _ := stakeDataMinsMap[addr1].Prob.Float64()
	if prob1 != 0.0 {
		t.Errorf("Expected probability 0.0 for addr1, got %f", prob1)
	}
	prob2, _ := stakeDataMinsMap[addr2].Prob.Float64()
	if prob2 != 1.0 {
		t.Errorf("Expected probability 1.0 for addr2, got %f", prob2)
	}
}

// Test with totalMin = 0 (the weighting does not divide by totalMin, so a
// zero total must not break normalization)
func TestCalculateProbsForEachWallet_TotalMinZero(t *testing.T) {
	addr := common.HexToAddress("0x1")
	stakeDataMinsMap := map[common.Address]*UserStakeData{
		addr: createUserStakeData1("100"),
	}
	totalMin := big.NewInt(0)

	found := calculateProbsForEachWallet(stakeDataMinsMap, totalMin)
	if !found {
		t.Errorf("Expected found=true, got false")
	}

	prob, _ := stakeDataMinsMap[addr].Prob.Float64()
	expectedProb := 1.0 // sole staker always has the whole probability mass
	if math.Abs(prob-expectedProb) > 1e-6 {
		t.Errorf("Expected probability %f, got %f", expectedProb, prob)
	}
}
