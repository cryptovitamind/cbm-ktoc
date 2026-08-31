package ktfunc

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// sqrtStakeMap builds a stake map from decimal stake strings keyed by address.
func sqrtStakeMap(stakes map[common.Address]string) map[common.Address]*UserStakeData {
	m := make(map[common.Address]*UserStakeData, len(stakes))
	for addr, s := range stakes {
		m[addr] = createUserStakeData1(s)
	}
	return m
}

// TestSqrtWeighting_MultipleAddresses pins the sqrt formula:
// prob_i = sqrt(stake_i) / sum(sqrt(stake_j)).
func TestSqrtWeighting_MultipleAddresses(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	addr3 := common.HexToAddress("0x3")
	m := sqrtStakeMap(map[common.Address]string{
		addr1: "100",
		addr2: "200",
		addr3: "300",
	})

	if err := normalizeProbabilities(m, SchemeSqrt); err != nil {
		t.Fatalf("normalizeProbabilities: %v", err)
	}

	sqrt100 := math.Sqrt(100)
	sqrt200 := math.Sqrt(200)
	sqrt300 := math.Sqrt(300)
	sum := sqrt100 + sqrt200 + sqrt300

	for _, tc := range []struct {
		addr     common.Address
		expected float64
	}{
		{addr1, sqrt100 / sum},
		{addr2, sqrt200 / sum},
		{addr3, sqrt300 / sum},
	} {
		prob, _ := m[tc.addr].Prob.Float64()
		if math.Abs(prob-tc.expected) > 1e-6 {
			t.Errorf("addr %s: expected probability %f, got %f", tc.addr.Hex(), tc.expected, prob)
		}
	}
}

// TestSqrtWeighting_SingleAddress: a lone staker always has probability 1.
func TestSqrtWeighting_SingleAddress(t *testing.T) {
	addr := common.HexToAddress("0x1")
	m := sqrtStakeMap(map[common.Address]string{addr: "100"})

	if err := normalizeProbabilities(m, SchemeSqrt); err != nil {
		t.Fatalf("normalizeProbabilities: %v", err)
	}

	prob, _ := m[addr].Prob.Float64()
	if prob != 1.0 {
		t.Errorf("Expected probability 1.0, got %f", prob)
	}
}

// TestSqrtWeighting_ZeroStakeGetsZeroProb: zero stakes are excluded with
// probability exactly 0; the remaining staker takes the whole pot.
func TestSqrtWeighting_ZeroStakeGetsZeroProb(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	m := sqrtStakeMap(map[common.Address]string{
		addr1: "0",
		addr2: "100",
	})

	if err := normalizeProbabilities(m, SchemeSqrt); err != nil {
		t.Fatalf("normalizeProbabilities: %v", err)
	}

	prob1, _ := m[addr1].Prob.Float64()
	if prob1 != 0.0 {
		t.Errorf("Expected probability 0.0 for zero stake, got %f", prob1)
	}
	prob2, _ := m[addr2].Prob.Float64()
	if prob2 != 1.0 {
		t.Errorf("Expected probability 1.0 for sole staker, got %f", prob2)
	}
}

// TestSqrtWeighting_TinyStakePositive: a 1-wei stake against a 1-token whale
// keeps a strictly positive share of sqrt(1)/(sqrt(1)+sqrt(1e18)) ≈ 1e-9.
// Unlike log weighting, the tiny share is genuinely tiny.
func TestSqrtWeighting_TinyStakePositive(t *testing.T) {
	tiny := common.HexToAddress("0x1")
	whale := common.HexToAddress("0x2")
	m := sqrtStakeMap(map[common.Address]string{
		tiny:  "1",
		whale: "1000000000000000000",
	})

	if err := normalizeProbabilities(m, SchemeSqrt); err != nil {
		t.Fatalf("normalizeProbabilities: %v", err)
	}

	tinyProb, _ := m[tiny].Prob.Float64()
	whaleProb, _ := m[whale].Prob.Float64()

	if tinyProb <= 0 {
		t.Errorf("tiny stake excluded (prob=%g); expected a strictly positive share", tinyProb)
	}
	if tinyProb >= 1e-8 {
		t.Errorf("tiny stake prob %g too large; sqrt share of 1 wei vs 1e18 is ~1e-9", tinyProb)
	}
	if whaleProb <= tinyProb {
		t.Errorf("whale prob (%g) should exceed tiny prob (%g)", whaleProb, tinyProb)
	}
	if total := tinyProb + whaleProb; math.Abs(total-1.0) > 1e-9 {
		t.Errorf("probabilities should sum to 1, got %g", total)
	}
}

// TestSqrtWeighting_Deterministic: probabilities must be bit-identical across
// repeated computations of the same stake set. Winner selection walks a
// cumulative probability, so even 1-ULP drift between operators is a consensus
// hazard; the sqrt path sums weights in sorted-address order to prevent it.
func TestSqrtWeighting_Deterministic(t *testing.T) {
	build := func() map[common.Address]*UserStakeData {
		m := make(map[common.Address]*UserStakeData, 20)
		for i := 1; i <= 20; i++ {
			addr := common.HexToAddress(fmt.Sprintf("0x%040x", i*7919))
			// Spread stakes across many magnitudes, wei-scale included.
			stake := new(big.Int).Exp(big.NewInt(3), big.NewInt(int64(i)), nil)
			stake.Mul(stake, big.NewInt(1_000_003))
			m[addr] = &UserStakeData{StakeAmount: stake}
		}
		return m
	}

	reference := build()
	if err := normalizeProbabilities(reference, SchemeSqrt); err != nil {
		t.Fatalf("normalizeProbabilities: %v", err)
	}

	for run := 0; run < 10; run++ {
		fresh := build()
		if err := normalizeProbabilities(fresh, SchemeSqrt); err != nil {
			t.Fatalf("normalizeProbabilities run %d: %v", run, err)
		}
		for addr, want := range reference {
			got := fresh[addr]
			if got.Prob.Cmp(want.Prob) != 0 {
				t.Fatalf("run %d: addr %s prob drifted: %s vs %s",
					run, addr.Hex(), got.Prob.Text('p', -1), want.Prob.Text('p', -1))
			}
		}
	}
}

// TestSqrtWeighting_WeiScaleResolution: stakes that differ by less than a
// float64 ULP at wei scale must still get distinct, correctly ordered weights.
// 1e10 is far below the ~2e11 float64 ULP spacing at 1e27, so an
// implementation that squeezes stakes through float64 collapses the two.
func TestSqrtWeighting_WeiScaleResolution(t *testing.T) {
	smaller := common.HexToAddress("0x1")
	larger := common.HexToAddress("0x2")

	base, _ := new(big.Int).SetString("1000000000000000000000000000", 10) // 1e27
	bumped := new(big.Int).Add(base, big.NewInt(10_000_000_000))          // +1e10

	m := map[common.Address]*UserStakeData{
		smaller: {StakeAmount: base},
		larger:  {StakeAmount: bumped},
	}

	if err := normalizeProbabilities(m, SchemeSqrt); err != nil {
		t.Fatalf("normalizeProbabilities: %v", err)
	}

	if m[larger].Prob.Cmp(m[smaller].Prob) <= 0 {
		t.Errorf("larger stake must get strictly larger probability: larger=%s smaller=%s",
			m[larger].Prob.Text('p', -1), m[smaller].Prob.Text('p', -1))
	}
}
