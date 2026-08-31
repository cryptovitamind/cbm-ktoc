package ktfunc

// Tests for the log weighting curve. Log is no longer used for live winner
// selection; it is retained so -verifyLastWinner can replay epochs that were
// rewarded by builds that selected winners with log weights. These tests pin
// that the retained implementation still computes exactly log(1+stake) shares.

import (
	"math"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// Log-weighted share of a lone staker is 1: log(101) / log(101).
func TestLogWeighting_SingleAddress(t *testing.T) {
	addr := common.HexToAddress("0x1")
	m := sqrtStakeMap(map[common.Address]string{addr: "100"})

	if err := normalizeProbabilities(m, SchemeLog); err != nil {
		t.Fatalf("normalizeProbabilities: %v", err)
	}

	prob, _ := m[addr].Prob.Float64()
	expectedProb := 1.0
	if math.Abs(prob-expectedProb) > 1e-6 {
		t.Errorf("Expected probability %f, got %f", expectedProb, prob)
	}
}

// Multiple addresses: prob_i = log(1+stake_i) / sum(log(1+stake_j)).
func TestLogWeighting_MultipleAddresses(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	addr3 := common.HexToAddress("0x3")
	m := sqrtStakeMap(map[common.Address]string{
		addr1: "100",
		addr2: "200",
		addr3: "300",
	})

	if err := normalizeProbabilities(m, SchemeLog); err != nil {
		t.Fatalf("normalizeProbabilities: %v", err)
	}

	// Expected probabilities use log1p to match logNormalizeProbabilities,
	// which uses log(1+stake) so a 1-wei stake doesn't collapse to 0.
	log100 := math.Log1p(100)
	log200 := math.Log1p(200)
	log300 := math.Log1p(300)
	sumLog := log100 + log200 + log300

	for _, tc := range []struct {
		addr     common.Address
		expected float64
	}{
		{addr1, log100 / sumLog},
		{addr2, log200 / sumLog},
		{addr3, log300 / sumLog},
	} {
		prob, _ := m[tc.addr].Prob.Float64()
		if math.Abs(prob-tc.expected) > 1e-6 {
			t.Errorf("addr %s: expected probability %f, got %f", tc.addr.Hex(), tc.expected, prob)
		}
	}
}

// Zero stakes get probability exactly 0 under log weighting.
func TestLogWeighting_ZeroStake(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	m := sqrtStakeMap(map[common.Address]string{
		addr1: "0",
		addr2: "100",
	})

	if err := normalizeProbabilities(m, SchemeLog); err != nil {
		t.Fatalf("normalizeProbabilities: %v", err)
	}

	prob1, _ := m[addr1].Prob.Float64()
	if prob1 != 0.0 {
		t.Errorf("Expected probability 0.0 for addr1, got %f", prob1)
	}
	prob2, _ := m[addr2].Prob.Float64()
	expectedProb2 := 1.0
	if math.Abs(prob2-expectedProb2) > 1e-6 {
		t.Errorf("Expected probability %f for addr2, got %f", expectedProb2, prob2)
	}
}

// TestLogWeighting_TinyStakeNotExcluded pins the log(1+x) offset: a wallet
// with a 1-wei stake must end up with a small but strictly positive
// probability, even against a whale (plain log(1) would be 0 and silently
// exclude the wallet).
func TestLogWeighting_TinyStakeNotExcluded(t *testing.T) {
	tiny := common.HexToAddress("0x1")
	whale := common.HexToAddress("0x2")
	m := sqrtStakeMap(map[common.Address]string{
		tiny:  "1",                   // 1 wei
		whale: "1000000000000000000", // 1 ETH
	})

	if err := normalizeProbabilities(m, SchemeLog); err != nil {
		t.Fatalf("normalizeProbabilities: %v", err)
	}

	tinyProb, _ := m[tiny].Prob.Float64()
	whaleProb, _ := m[whale].Prob.Float64()

	if tinyProb <= 0 {
		t.Errorf("tiny-stake wallet was excluded (prob=%g); expected a small but positive share", tinyProb)
	}
	if whaleProb <= tinyProb {
		t.Errorf("whale prob (%g) should still exceed tiny prob (%g) under log weighting", whaleProb, tinyProb)
	}
	// Combined probs should sum to ~1.
	if total := tinyProb + whaleProb; math.Abs(total-1.0) > 1e-9 {
		t.Errorf("probabilities should sum to 1, got %g", total)
	}
}
