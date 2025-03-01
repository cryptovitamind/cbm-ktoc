package ktfunc

import (
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
	expectedProb := 1.0 // 100 / 100
	if prob != expectedProb {
		t.Errorf("Expected probability %f, got %f", expectedProb, prob)
	}
}

// Test with multiple addresses
func TestCalculateProbsForEachWallet_MultipleAddresses(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	addr3 := common.HexToAddress("0x3")
	stakeDataMinsMap := map[common.Address]*UserStakeData{
		addr1: createUserStakeData1("100"),
		addr2: createUserStakeData1("200"),
		addr3: createUserStakeData1("300"),
	}
	totalMin := big.NewInt(600) // 100 + 200 + 300

	found := calculateProbsForEachWallet(stakeDataMinsMap, totalMin)
	if !found {
		t.Errorf("Expected found=true, got false")
	}

	// Check probabilities
	prob1, _ := stakeDataMinsMap[addr1].Prob.Float64()
	if prob1 != 100.0/600.0 {
		t.Errorf("Expected probability %f for addr1, got %f", 100.0/600.0, prob1)
	}
	prob2, _ := stakeDataMinsMap[addr2].Prob.Float64()
	if prob2 != 200.0/600.0 {
		t.Errorf("Expected probability %f for addr2, got %f", 200.0/600.0, prob2)
	}
	prob3, _ := stakeDataMinsMap[addr3].Prob.Float64()
	if prob3 != 300.0/600.0 {
		t.Errorf("Expected probability %f for addr3, got %f", 300.0/600.0, prob3)
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

// Test with totalMin = 0 (division by zero case)
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

	// Division by zero in big.Float.Quo results in +Inf
	prob := stakeDataMinsMap[addr].Prob
	if !prob.IsInf() { // Expect positive infinity
		t.Errorf("Expected probability +Inf for totalMin=0, got %v", prob)
	}
}
