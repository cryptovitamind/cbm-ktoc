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
	found := calculateProbsForEachWallet(nil, big.NewInt(0), true)
	if found {
		t.Errorf("Expected found=false for nil map, got true")
	}
}

// Test with an empty map
func TestCalculateProbsForEachWallet_EmptyMap(t *testing.T) {
	stakeDataMinsMap := make(map[common.Address]*UserStakeData)
	found := calculateProbsForEachWallet(stakeDataMinsMap, big.NewInt(0), true)
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

	found := calculateProbsForEachWallet(stakeDataMinsMap, totalMin, true)
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

	found := calculateProbsForEachWallet(stakeDataMinsMap, totalMin, true)
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

	found := calculateProbsForEachWallet(stakeDataMinsMap, totalMin, true)
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

// Test with totalMin = 0 (log normalization still works)
func TestCalculateProbsForEachWallet_TotalMinZero(t *testing.T) {
	addr := common.HexToAddress("0x1")
	stakeDataMinsMap := map[common.Address]*UserStakeData{
		addr: createUserStakeData1("100"),
	}
	totalMin := big.NewInt(0)

	found := calculateProbsForEachWallet(stakeDataMinsMap, totalMin, false)
	if !found {
		t.Errorf("Expected found=true, got false")
	}

	prob, _ := stakeDataMinsMap[addr].Prob.Float64()
	expectedProb := 1.0 // log(100) / log(100) = 1.0
	if math.Abs(prob-expectedProb) > 1e-6 {
		t.Errorf("Expected probability %f, got %f", expectedProb, prob)
	}
}

// Test log-normalized probabilities with a single address
func TestCalculateProbsForEachWallet_LogNormalized_SingleAddress(t *testing.T) {
	addr := common.HexToAddress("0x1")
	stakeDataMinsMap := map[common.Address]*UserStakeData{
		addr: createUserStakeData1("100"),
	}
	totalMin := big.NewInt(100)

	found := calculateProbsForEachWallet(stakeDataMinsMap, totalMin, false)
	if !found {
		t.Errorf("Expected found=true, got false")
	}

	prob, _ := stakeDataMinsMap[addr].Prob.Float64()
	expectedProb := 1.0 // log(100) / log(100) = 1.0
	if math.Abs(prob-expectedProb) > 1e-6 {
		t.Errorf("Expected probability %f, got %f", expectedProb, prob)
	}
}

// Test log-normalized probabilities with multiple addresses
func TestCalculateProbsForEachWallet_LogNormalized_MultipleAddresses(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	addr3 := common.HexToAddress("0x3")
	stakeDataMinsMap := map[common.Address]*UserStakeData{
		addr1: createUserStakeData1("100"),
		addr2: createUserStakeData1("200"),
		addr3: createUserStakeData1("300"),
	}
	totalMin := big.NewInt(600)

	found := calculateProbsForEachWallet(stakeDataMinsMap, totalMin, false)
	if !found {
		t.Errorf("Expected found=true, got false")
	}

	// Calculate expected probabilities
	log100 := math.Log(100)
	log200 := math.Log(200)
	log300 := math.Log(300)
	sumLog := log100 + log200 + log300
	expectedProb1 := log100 / sumLog
	expectedProb2 := log200 / sumLog
	expectedProb3 := log300 / sumLog

	// Check probabilities with tolerance
	prob1, _ := stakeDataMinsMap[addr1].Prob.Float64()
	if math.Abs(prob1-expectedProb1) > 1e-6 {
		t.Errorf("Expected probability %f for addr1, got %f", expectedProb1, prob1)
	}
	prob2, _ := stakeDataMinsMap[addr2].Prob.Float64()
	if math.Abs(prob2-expectedProb2) > 1e-6 {
		t.Errorf("Expected probability %f for addr2, got %f", expectedProb2, prob2)
	}
	prob3, _ := stakeDataMinsMap[addr3].Prob.Float64()
	if math.Abs(prob3-expectedProb3) > 1e-6 {
		t.Errorf("Expected probability %f for addr3, got %f", expectedProb3, prob3)
	}
}

// Test log-normalized probabilities with a zero stake amount
func TestCalculateProbsForEachWallet_LogNormalized_ZeroStake(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	stakeDataMinsMap := map[common.Address]*UserStakeData{
		addr1: createUserStakeData1("0"),
		addr2: createUserStakeData1("100"),
	}
	totalMin := big.NewInt(100)

	found := calculateProbsForEachWallet(stakeDataMinsMap, totalMin, false)
	if !found {
		t.Errorf("Expected found=true, got false")
	}

	prob1, _ := stakeDataMinsMap[addr1].Prob.Float64()
	if prob1 != 0.0 {
		t.Errorf("Expected probability 0.0 for addr1, got %f", prob1)
	}
	prob2, _ := stakeDataMinsMap[addr2].Prob.Float64()
	expectedProb2 := 1.0 // log(100) / log(100) = 1.0
	if math.Abs(prob2-expectedProb2) > 1e-6 {
		t.Errorf("Expected probability %f for addr2, got %f", expectedProb2, prob2)
	}
}
