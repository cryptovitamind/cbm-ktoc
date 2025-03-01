package ktfunc

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func createUserStakeData2(stake string, prob float64) *UserStakeData {

	stakeInt, _ := new(big.Int).SetString(stake, 10)
	probFloat := new(big.Float).SetFloat64(prob)
	return &UserStakeData{
		StakeAmount: stakeInt,
		Prob:        probFloat,
	}
}

func TestCalculateWinningWallet_NilMap(t *testing.T) {
	winner, err := calcWinningWallet(nil, common.Hash{})
	if err == nil || err.Error() != "stake data map is nil" {
		t.Errorf("Expected error 'stake data map is nil', got %v", err)
	}
	if winner != (common.Address{}) {
		t.Errorf("Expected zero address, got %s", winner.Hex())
	}
}

func TestCalculateWinningWallet_EmptyMap(t *testing.T) {
	stakeDataMinsMap := make(map[common.Address]*UserStakeData)
	winner, err := calcWinningWallet(stakeDataMinsMap, common.Hash{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if winner != (common.Address{}) {
		t.Errorf("Expected zero address, got %s", winner.Hex())
	}
}

func TestCalculateWinningWallet_SingleAddress(t *testing.T) {
	addr := common.HexToAddress("0x1")
	stakeDataMinsMap := map[common.Address]*UserStakeData{
		addr: createUserStakeData2("100", 1.0),
	}
	// Random number doesn't affect result since prob=1 covers the entire range [0,1)
	winner, err := calcWinningWallet(stakeDataMinsMap, common.Hash{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if winner != addr {
		t.Errorf("Expected winner %s, got %s", addr.Hex(), winner.Hex())
	}
}

func TestCalculateWinningWallet_EqualProbabilities(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	stakeDataMinsMap := map[common.Address]*UserStakeData{
		addr1: createUserStakeData2("100", 0.5),
		addr2: createUserStakeData2("100", 0.5),
	}

	// Sorted order: "0x1" < "0x2"
	// Cumulative probs: addr1=0.5, addr2=1.0

	// Test with randomFloat = 0.0 (all zeros hash)
	// randFloat = 0 / 2^256 = 0.0 < 0.5, should select addr1
	winnerLow, err := calcWinningWallet(stakeDataMinsMap, common.Hash{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if winnerLow != addr1 {
		t.Errorf("Expected winner %s, got %s", addr1.Hex(), winnerLow.Hex())
	}

	// Test with randomFloat = 0.5
	// randInt = 2^255, so randFloat = 2^255 / 2^256 = 0.5
	// 0.5 >= 0.5 (false), 0.5 < 1.0 (true), should select addr2
	randomNumberHigh := common.Hash{0x80} // 0x80 followed by 31 zero bytes
	winnerHigh, err := calcWinningWallet(stakeDataMinsMap, randomNumberHigh)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if winnerHigh != addr2 {
		t.Errorf("Expected winner %s, got %s", addr2.Hex(), winnerHigh.Hex())
	}
}

func TestCalculateWinningWallet_UnequalProbabilities(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	addr3 := common.HexToAddress("0x3")
	stakeDataMinsMap := map[common.Address]*UserStakeData{
		addr1: createUserStakeData2("100", 0.2),
		addr2: createUserStakeData2("150", 0.3),
		addr3: createUserStakeData2("250", 0.5),
	}

	// Sorted order: "0x1" < "0x2" < "0x3"
	// Cumulative probs: addr1=0.2, addr2=0.5, addr3=1.0

	// randomFloat = 0.0 < 0.2, select addr1
	winner1, err := calcWinningWallet(stakeDataMinsMap, common.Hash{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if winner1 != addr1 {
		t.Errorf("Expected winner %s, got %s", addr1.Hex(), winner1.Hex())
	}

	// randomFloat ≈ 0.25
	// randInt = 0.25 * 2^256, approximate with 2^254
	// randomFloat = 2^254 / 2^256 = 0.25
	// 0.25 > 0.2, 0.25 < 0.5, select addr2
	var bytes [32]byte
	bytes[0] = 0x40 // 2^254, since 0x40 * 256^31 = 2^254
	randomNumberMid := common.Hash(bytes)
	winner2, err := calcWinningWallet(stakeDataMinsMap, randomNumberMid)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if winner2 != addr2 {
		t.Errorf("Expected winner %s, got %s", addr2.Hex(), winner2.Hex())
	}

	// randomFloat = 0.5
	// 0.5 > 0.2, 0.5 >= 0.5 (false), 0.5 < 1.0, select addr3
	winner3, err := calcWinningWallet(stakeDataMinsMap, common.Hash{0x80})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if winner3 != addr3 {
		t.Errorf("Expected winner %s, got %s", addr3.Hex(), winner3.Hex())
	}
}

func TestCalculateWinningWallet_ProbsLessThanOne(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	stakeDataMinsMap := map[common.Address]*UserStakeData{
		addr1: createUserStakeData2("100", 0.4),
		addr2: createUserStakeData2("100", 0.4),
	}

	// Sorted order: "0x1" < "0x2"
	// Cumulative probs: addr1=0.4, addr2=0.8
	// randomFloat = 0.9 > 0.8, should fallback to addr2

	// Set randomFloat ≈ 0.9
	// randInt ≈ 0.9 * 2^256
	// Approximate with 0xE6... (since 0xE6/0xFF ≈ 0.9)
	var bytes [32]byte
	bytes[0] = 0xE6
	randomNumber := common.Hash(bytes)
	winner, err := calcWinningWallet(stakeDataMinsMap, randomNumber)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if winner != addr2 {
		t.Errorf("Expected winner %s (last address), got %s", addr2.Hex(), winner.Hex())
	}
}

func TestCalculateWinningWallet_InvalidStakeData(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	stakeDataMinsMap := map[common.Address]*UserStakeData{
		addr1: nil,
		addr2: createUserStakeData2("100", 1.0),
	}

	// Should skip addr1, select addr2
	winner, err := calcWinningWallet(stakeDataMinsMap, common.Hash{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if winner != addr2 {
		t.Errorf("Expected winner %s, got %s", addr2.Hex(), winner.Hex())
	}

	// Test with nil Prob
	stakeDataMinsMap[addr1] = &UserStakeData{StakeAmount: big.NewInt(100), Prob: nil}
	winner, err = calcWinningWallet(stakeDataMinsMap, common.Hash{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if winner != addr2 {
		t.Errorf("Expected winner %s, got %s", addr2.Hex(), winner.Hex())
	}
}
