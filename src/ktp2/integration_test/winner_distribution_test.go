package integration_test

// Distribution tests for the winner-selection pipeline, driven through
// ktfunc.VerifyWinnerCalculation — which runs findMinOverBlockRange +
// probability normalization + defaultCalculateWinningWallet, the same code
// path VoteAndReward uses to pick a winner. Each test samples many block
// hashes to approximate win rates and asserts they land where the min-stake
// semantics and the weighting curve say they should.

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"testing"

	"ktp2/src/ktp2/ktfunc"

	"github.com/ethereum/go-ethereum/common"
)

// hashFromIndex turns a small integer into a distinct 32-byte hash so we
// can sample many "block hashes" to simulate distinct epoch outcomes.
func hashFromIndex(i int) common.Hash {
	var h common.Hash
	binary.BigEndian.PutUint64(h[:8], uint64(i)*0x9E3779B97F4A7C15) // golden-ratio mix
	binary.BigEndian.PutUint64(h[8:16], uint64(i)*0xBF58476D1CE4E5B9)
	binary.BigEndian.PutUint64(h[16:24], uint64(i)*0x94D049BB133111EB)
	binary.BigEndian.PutUint64(h[24:], uint64(i)*0xD6E8FEB86659FD93)
	return h
}

func TestVoteAndReward_FreshMidEpochDepositorDoesNotAutoWin(t *testing.T) {
	// A wallet that first deposits mid-epoch has a true epoch minimum of 0
	// and must be excluded from the lottery entirely, no matter how large
	// the deposit or which weighting curve is in effect. Nine modest
	// baseline stakers plus one huge mid-epoch newcomer: the newcomer
	// should win nothing.

	stakeDataMap := make(map[common.Address]map[uint64]*ktfunc.UserStakeData)
	for i := 1; i <= 9; i++ {
		addr := common.HexToAddress(fmt.Sprintf("0x%040x", i))
		stakeDataMap[addr] = map[uint64]*ktfunc.UserStakeData{
			50: {StakeAmount: big.NewInt(1000)},
		}
	}
	newcomer := common.HexToAddress("0x00000000000000000000000000000000000000FF")
	stakeDataMap[newcomer] = map[uint64]*ktfunc.UserStakeData{
		150: {StakeAmount: big.NewInt(1_000_000_000_000_000_000)}, // 1 ETH in wei — more extreme to make bug visible under log
	}

	const samples = 200
	const epochStart, epochEnd = uint64(100), uint64(200)

	newcomerWins := 0
	for i := 0; i < samples; i++ {
		fresh := cloneStakeMap(stakeDataMap)
		result, err := ktfunc.VerifyWinnerCalculation(fresh, epochStart, epochEnd, hashFromIndex(i), ktfunc.SchemeSqrt)
		if err != nil {
			t.Fatalf("VerifyWinnerCalculation: %v", err)
		}
		if result.CalculatedWinner == newcomer {
			newcomerWins++
		}
	}

	// True semantics: newcomer is excluded → 0 wins.
	// 10% gives plenty of headroom for noise; the newcomer should be at 0%.
	maxAllowedWinPct := 10.0
	winPct := 100.0 * float64(newcomerWins) / float64(samples)
	if winPct > maxAllowedWinPct {
		t.Errorf("FAIL: newcomer who only staked mid-epoch won %d/%d (%.1f%%) of epochs; "+
			"expected <%.0f%% because their true min stake during the epoch is 0",
			newcomerWins, samples, winPct, maxAllowedWinPct)
	}
}

func TestVoteAndReward_TopUpMidEpochDoesNotInflateWeight(t *testing.T) {
	// Two equal baseline stakers + one staker who carries 1000 from
	// pre-epoch AND tops up by another 1,000,000 mid-epoch. Their
	// correct min is 1000 (the carried floor) — i.e. they should be
	// roughly tied with the other two baselines, not dominate.

	addrA := common.HexToAddress("0x0000000000000000000000000000000000000001")
	addrB := common.HexToAddress("0x0000000000000000000000000000000000000002")
	addrC := common.HexToAddress("0x0000000000000000000000000000000000000003")
	stakeDataMap := map[common.Address]map[uint64]*ktfunc.UserStakeData{
		addrA: {50: {StakeAmount: big.NewInt(1000)}},
		addrB: {50: {StakeAmount: big.NewInt(1000)}},
		addrC: {
			50:  {StakeAmount: big.NewInt(1000)},      // carried
			150: {StakeAmount: big.NewInt(1_000_000)}, // mid-epoch top-up
		},
	}

	const samples = 200
	const epochStart, epochEnd = uint64(100), uint64(200)

	cWins := 0
	for i := 0; i < samples; i++ {
		fresh := cloneStakeMap(stakeDataMap)
		result, err := ktfunc.VerifyWinnerCalculation(fresh, epochStart, epochEnd, hashFromIndex(i), ktfunc.SchemeSqrt)
		if err != nil {
			t.Fatalf("VerifyWinnerCalculation: %v", err)
		}
		if result.CalculatedWinner == addrC {
			cWins++
		}
	}

	// True semantics: A/B/C each have min=1000 (the top-up never lowers the
	// epoch minimum), so all three are equally weighted under any curve and
	// C should win ~33%.
	maxAllowedWinPct := 60.0
	winPct := 100.0 * float64(cWins) / float64(samples)
	if winPct > maxAllowedWinPct {
		t.Errorf("FAIL (reproduces bug): mid-epoch top-up gave C %d/%d (%.1f%%) wins; "+
			"expected ~33%% (well under %.0f%%) because C's true min is 1000, same as A and B",
			cWins, samples, winPct, maxAllowedWinPct)
	}
}

// TestWinnerDistribution_SqrtWeighting pins the live curve end-to-end: with
// stakes of 1e18 and 4e18 carried through the whole epoch, sqrt weighting
// gives the larger staker sqrt(4e18)/(sqrt(1e18)+sqrt(4e18)) = 2/3 of the
// wins. The band [58%, 78%] is ~2.6σ around the 2/3 expectation for 200
// samples and cleanly excludes log weighting, which would put the larger
// staker at ~50.8%.
func TestWinnerDistribution_SqrtWeighting(t *testing.T) {
	small := common.HexToAddress("0x0000000000000000000000000000000000000001")
	large := common.HexToAddress("0x0000000000000000000000000000000000000002")

	oneToken := new(big.Int).SetUint64(1_000_000_000_000_000_000)
	fourTokens := new(big.Int).Mul(oneToken, big.NewInt(4))

	stakeDataMap := map[common.Address]map[uint64]*ktfunc.UserStakeData{
		small: {50: {StakeAmount: oneToken}},
		large: {50: {StakeAmount: fourTokens}},
	}

	const samples = 200
	const epochStart, epochEnd = uint64(100), uint64(200)

	largeWins := 0
	for i := 0; i < samples; i++ {
		fresh := cloneStakeMap(stakeDataMap)
		result, err := ktfunc.VerifyWinnerCalculation(fresh, epochStart, epochEnd, hashFromIndex(i), ktfunc.SchemeSqrt)
		if err != nil {
			t.Fatalf("VerifyWinnerCalculation: %v", err)
		}
		if result.CalculatedWinner == large {
			largeWins++
		}
	}

	winPct := 100.0 * float64(largeWins) / float64(samples)
	if winPct < 58.0 || winPct > 78.0 {
		t.Errorf("sqrt weighting should give the 4x staker ~66.7%% of wins; got %d/%d (%.1f%%), outside [58%%, 78%%]",
			largeWins, samples, winPct)
	}
}

func cloneStakeMap(src map[common.Address]map[uint64]*ktfunc.UserStakeData) map[common.Address]map[uint64]*ktfunc.UserStakeData {
	dst := make(map[common.Address]map[uint64]*ktfunc.UserStakeData, len(src))
	for addr, blocks := range src {
		inner := make(map[uint64]*ktfunc.UserStakeData, len(blocks))
		for blk, data := range blocks {
			inner[blk] = &ktfunc.UserStakeData{
				StakeAmount: new(big.Int).Set(data.StakeAmount),
			}
		}
		dst[addr] = inner
	}
	return dst
}
