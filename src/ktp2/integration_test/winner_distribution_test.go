package integration_test

// Phase 1 (TDD) — reproduces the fairness consequence of the
// findMinOverBlockRange bug at the public API surface.
//
// We drive the full winner-selection pipeline through
// ktfunc.VerifyWinnerCalculation, which internally runs
// findMinOverBlockRange + calculateProbsForEachWallet +
// defaultCalculateWinningWallet — i.e. the same code path
// VoteAndReward uses to pick a winner.
//
// Today, a wallet that joins exactly at epochStart (or tops up mid-epoch)
// is credited with its post-deposit stake as the epoch minimum. Under
// linear probability normalization that hands the wallet ~all of the
// probability mass.
//
// These tests will pass after the Phase 2 fix to findMinOverBlockRange.

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

func TestVoteAndReward_FreshMidEpochDepositorDoesNotAutoWin_Linear(t *testing.T) {
	// 9 baseline stakers (all entered pre-epoch with equal stake).
	// 1 newcomer who joins mid-epoch with a huge deposit. Under correct
	// semantics their min stake is 0 — they should NEVER win.

	stakeDataMap := make(map[common.Address]map[uint64]*ktfunc.UserStakeData)

	for i := 1; i <= 9; i++ {
		addr := common.HexToAddress(fmt.Sprintf("0x%040x", i))
		stakeDataMap[addr] = map[uint64]*ktfunc.UserStakeData{
			50: {StakeAmount: big.NewInt(1000)},
		}
	}
	newcomer := common.HexToAddress("0x00000000000000000000000000000000000000FF")
	stakeDataMap[newcomer] = map[uint64]*ktfunc.UserStakeData{
		150: {StakeAmount: big.NewInt(1_000_000)},
	}

	const samples = 200
	const epochStart, epochEnd = uint64(100), uint64(200)

	newcomerWins := 0
	for i := 0; i < samples; i++ {
		// Need a fresh map per call: VerifyWinnerCalculation mutates
		// the inner *UserStakeData (probabilities), and findMinOverBlockRange
		// reads/writes through it.
		fresh := cloneStakeMap(stakeDataMap)
		result, err := ktfunc.VerifyWinnerCalculation(fresh, epochStart, epochEnd, hashFromIndex(i), true /*useLinear*/)
		if err != nil {
			t.Fatalf("VerifyWinnerCalculation: %v", err)
		}
		if result.CalculatedWinner == newcomer {
			newcomerWins++
		}
	}

	// True semantics: newcomer's min is 0 → excluded → 0 wins.
	// Bug today: newcomer dominates probability mass under linear → ~all wins.
	maxAllowedWinPct := 25.0
	winPct := 100.0 * float64(newcomerWins) / float64(samples)
	if winPct > maxAllowedWinPct {
		t.Errorf("FAIL (reproduces bug): newcomer who only staked mid-epoch won %d/%d (%.1f%%) of epochs; "+
			"expected <%.0f%% because their true min stake during the epoch is 0",
			newcomerWins, samples, winPct, maxAllowedWinPct)
	}
}

func TestVoteAndReward_FreshMidEpochDepositorDoesNotAutoWin_Log(t *testing.T) {
	// Same scenario as above but with log-normalized probabilities. The
	// log dampens the imbalance, so the newcomer doesn't get *all* the
	// probability — but they still pick up a share they shouldn't have
	// (their true min is 0; they should be excluded).
	//
	// Production default is log normalization (linearProbs flag is off),
	// so this version corresponds to the live-node configuration.

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
		result, err := ktfunc.VerifyWinnerCalculation(fresh, epochStart, epochEnd, hashFromIndex(i), false /*useLinear=false → log*/)
		if err != nil {
			t.Fatalf("VerifyWinnerCalculation: %v", err)
		}
		if result.CalculatedWinner == newcomer {
			newcomerWins++
		}
	}

	// True semantics: newcomer is excluded → 0 wins.
	// Bug under log: newcomer still gets a measurable share of the lottery.
	// 10% gives plenty of headroom for noise; the newcomer should be at 0%.
	maxAllowedWinPct := 10.0
	winPct := 100.0 * float64(newcomerWins) / float64(samples)
	if winPct > maxAllowedWinPct {
		t.Errorf("FAIL (reproduces bug, log mode): newcomer who only staked mid-epoch won %d/%d (%.1f%%) of epochs; "+
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
		result, err := ktfunc.VerifyWinnerCalculation(fresh, epochStart, epochEnd, hashFromIndex(i), true /*linear*/)
		if err != nil {
			t.Fatalf("VerifyWinnerCalculation: %v", err)
		}
		if result.CalculatedWinner == addrC {
			cWins++
		}
	}

	// True semantics: A/B/C each have min=1000, so C should win ~33%.
	// Bug today: C's min becomes 1,001,000 → ~99% under linear.
	maxAllowedWinPct := 60.0
	winPct := 100.0 * float64(cWins) / float64(samples)
	if winPct > maxAllowedWinPct {
		t.Errorf("FAIL (reproduces bug): mid-epoch top-up gave C %d/%d (%.1f%%) wins; "+
			"expected ~33%% (well under %.0f%%) because C's true min is 1000, same as A and B",
			cWins, samples, winPct, maxAllowedWinPct)
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
