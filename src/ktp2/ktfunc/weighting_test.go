package ktfunc

import (
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// TestNormalizeProbabilities_UnknownSchemeErrors: an unrecognized scheme must
// error and leave the map untouched, never silently fall back to a curve.
func TestNormalizeProbabilities_UnknownSchemeErrors(t *testing.T) {
	addr := common.HexToAddress("0x1")
	m := sqrtStakeMap(map[common.Address]string{addr: "100"})

	err := normalizeProbabilities(m, WeightingScheme("banana"))
	if err == nil {
		t.Fatal("expected error for unknown weighting scheme, got nil")
	}
	if m[addr].Prob != nil {
		t.Errorf("map must be untouched on unknown scheme; got prob %v", m[addr].Prob)
	}
}

// TestCalculateProbsForEachWallet_UsesSqrtNotLog pins the curve behind the
// production wrapper: live selection must be sqrt-weighted. The outer stakes'
// log and sqrt shares differ by ~0.05, so the test discriminates the curves.
func TestCalculateProbsForEachWallet_UsesSqrtNotLog(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	addr3 := common.HexToAddress("0x3")
	m := sqrtStakeMap(map[common.Address]string{
		addr1: "100",
		addr2: "200",
		addr3: "300",
	})

	found := calculateProbsForEachWallet(m, big.NewInt(600))
	if !found {
		t.Fatal("Expected found=true, got false")
	}

	sqrtSum := math.Sqrt(100) + math.Sqrt(200) + math.Sqrt(300)
	logSum := math.Log1p(100) + math.Log1p(200) + math.Log1p(300)

	// The middle stake's sqrt and log shares happen to nearly coincide
	// (0.3411 vs 0.3394), so only the outer stakes discriminate the curves;
	// the exact-share assertion below still pins all three.
	for _, tc := range []struct {
		addr             common.Address
		stake            float64
		discriminatesLog bool
	}{
		{addr1, 100, true}, {addr2, 200, false}, {addr3, 300, true},
	} {
		prob, _ := m[tc.addr].Prob.Float64()
		wantSqrt := math.Sqrt(tc.stake) / sqrtSum
		wantLog := math.Log1p(tc.stake) / logSum
		if math.Abs(prob-wantSqrt) > 1e-6 {
			t.Errorf("addr %s: live path must be sqrt-weighted; expected %f, got %f",
				tc.addr.Hex(), wantSqrt, prob)
		}
		if tc.discriminatesLog && math.Abs(prob-wantLog) < 0.01 {
			t.Errorf("addr %s: prob %f matches the log curve; live path must not use log",
				tc.addr.Hex(), prob)
		}
	}
}

// ResolveVerifyWeighting precedence: explicit flag > env > default.

func TestResolveVerifyWeighting_Default(t *testing.T) {
	scheme, err := ResolveVerifyWeighting("", string(DefaultVerifyWeighting))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scheme != SchemeSqrt {
		t.Errorf("expected default %q, got %q", SchemeSqrt, scheme)
	}
}

func TestResolveVerifyWeighting_EnvWins(t *testing.T) {
	scheme, err := ResolveVerifyWeighting(string(SchemeLog), string(DefaultVerifyWeighting))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scheme != SchemeLog {
		t.Errorf("expected env value %q, got %q", SchemeLog, scheme)
	}
}

func TestResolveVerifyWeighting_FlagBeatsEnv(t *testing.T) {
	scheme, err := ResolveVerifyWeighting(string(SchemeSqrt), string(SchemeLog))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scheme != SchemeLog {
		t.Errorf("expected explicit flag %q to beat env, got %q", SchemeLog, scheme)
	}
}

// TestResolveVerifyWeighting_FlagAtDefaultLetsEnvWin pins the sentinel quirk
// shared with ResolveWaitDuration: a flag left at (or explicitly set to) the
// default is indistinguishable from an omitted flag, so the env value applies.
func TestResolveVerifyWeighting_FlagAtDefaultLetsEnvWin(t *testing.T) {
	scheme, err := ResolveVerifyWeighting(string(SchemeLog), string(SchemeSqrt))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scheme != SchemeLog {
		t.Errorf("flag at default must let env win; expected %q, got %q", SchemeLog, scheme)
	}
}

func TestResolveVerifyWeighting_InvalidErrors(t *testing.T) {
	if _, err := ResolveVerifyWeighting("banana", string(DefaultVerifyWeighting)); err == nil {
		t.Error("expected error for invalid env value, got nil")
	}
	if _, err := ResolveVerifyWeighting("", "banana"); err == nil {
		t.Error("expected error for invalid flag value, got nil")
	}
}
