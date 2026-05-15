package ktfunc

// Phase 1 (TDD) — reproduce the "any in-epoch event gives full weight" bug in
// findMinOverBlockRange. These tests are expected to fail on master and should
// pass once findMinOverBlockRange captures the pre-event currentStake on the
// first in-range block (Phase 2 fix). See plan
// /Users/joe/.claude/plans/please-do-an-audit-typed-oasis.md.
//
// Inputs to findMinOverBlockRange:
//   stakeDataMap[addr][block].StakeAmount is the per-block delta (positive for
//   stake events, negative for withdraw events). The function accumulates
//   currentStake across sorted blocks and tracks the minimum once a block
//   falls within [epochStart, epochEnd].
//
// Expected semantics: minStake is the wallet's smallest stake at any moment
// during the epoch, including the moment *before* an in-range event lands.
// A wallet that joins at epochStart had stake 0 immediately before the join
// event, so its true min for that epoch is 0.

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// makeDelta builds a per-block-delta entry for stakeDataMap.
func makeDelta(amount int64) *UserStakeData {
	return &UserStakeData{StakeAmount: big.NewInt(amount)}
}

// minStakeFor returns the calculated min stake for one address, or nil if the
// address was excluded.
func minStakeFor(
	t *testing.T,
	epochStart, epochEnd uint64,
	stakeDataMap map[common.Address]map[uint64]*UserStakeData,
	addr common.Address,
) *big.Int {
	t.Helper()
	_, addressMins, err := findMinOverBlockRange(epochStart, epochEnd, stakeDataMap)
	if err != nil {
		t.Fatalf("findMinOverBlockRange returned error: %v", err)
	}
	if v, ok := addressMins[addr]; ok && v != nil {
		return v.StakeAmount
	}
	return nil
}

// TestFindMinOverBlockRange_StakerJoinsAtEpochStart documents the bug where a
// staker who joins exactly at epochStart is credited with their full deposit
// as the epoch's minimum. The true minimum is 0 (the staker had no stake one
// block before, which is still inside the epoch window's first instant).
//
// Expected after Phase 2: addr is EXCLUDED (min=0 → excluded by line 1118).
// Current behavior: addr.StakeAmount == 1000.
func TestFindMinOverBlockRange_StakerJoinsAtEpochStart(t *testing.T) {
	addr := common.HexToAddress("0x000000000000000000000000000000000000000A")
	epochStart, epochEnd := uint64(100), uint64(200)

	stakeDataMap := map[common.Address]map[uint64]*UserStakeData{
		addr: {
			epochStart: makeDelta(1000), // joins exactly at start
		},
	}

	min := minStakeFor(t, epochStart, epochEnd, stakeDataMap, addr)
	if min != nil && min.Sign() != 0 {
		t.Errorf("FAIL (reproduces bug): expected staker to be excluded or have min=0 "+
			"because they had stake 0 immediately before their first in-range event, "+
			"got min=%s", min.String())
	}
}

// TestFindMinOverBlockRange_StakerToppedUpMidEpoch documents the case where a
// staker carries 100 from before the epoch and tops up by 900 mid-epoch. The
// true minimum during the epoch is 100 (the carried stake); the 900 top-up
// arrives later.
//
// Expected after Phase 2: addr min == 100.
// Current behavior: addr min == 1000 (post-event capture).
func TestFindMinOverBlockRange_StakerToppedUpMidEpoch(t *testing.T) {
	addr := common.HexToAddress("0x000000000000000000000000000000000000000B")
	epochStart, epochEnd := uint64(100), uint64(200)

	stakeDataMap := map[common.Address]map[uint64]*UserStakeData{
		addr: {
			50:  makeDelta(100), // pre-epoch baseline
			150: makeDelta(900), // mid-epoch top-up
		},
	}

	min := minStakeFor(t, epochStart, epochEnd, stakeDataMap, addr)
	if min == nil {
		t.Fatalf("expected addr to be included, but it was excluded")
	}
	want := big.NewInt(100)
	if min.Cmp(want) != 0 {
		t.Errorf("FAIL (reproduces bug): expected min=%s (the pre-event carried stake), got min=%s",
			want.String(), min.String())
	}
}

// TestFindMinOverBlockRange_StakerJoinsMidEpochOnly documents that a wallet
// that stakes mid-epoch with no prior stake should be excluded — their stake
// was 0 from epochStart up to the mid-epoch deposit.
//
// Expected after Phase 2: addr is EXCLUDED.
// Current behavior: addr min == 500.
func TestFindMinOverBlockRange_StakerJoinsMidEpochOnly(t *testing.T) {
	addr := common.HexToAddress("0x000000000000000000000000000000000000000C")
	epochStart, epochEnd := uint64(100), uint64(200)

	stakeDataMap := map[common.Address]map[uint64]*UserStakeData{
		addr: {
			150: makeDelta(500), // joins mid-epoch
		},
	}

	min := minStakeFor(t, epochStart, epochEnd, stakeDataMap, addr)
	if min != nil && min.Sign() != 0 {
		t.Errorf("FAIL (reproduces bug): expected staker to be excluded (min=0) "+
			"because they had no stake from epochStart through block 149, "+
			"got min=%s", min.String())
	}
}

// TestFindMinOverBlockRange_WithdrawMidEpoch pins correct behavior: a wallet
// that entered with 1000 and withdraws 500 mid-epoch had min=500. This passes
// today and should keep passing after Phase 2.
func TestFindMinOverBlockRange_WithdrawMidEpoch(t *testing.T) {
	addr := common.HexToAddress("0x000000000000000000000000000000000000000D")
	epochStart, epochEnd := uint64(100), uint64(200)

	stakeDataMap := map[common.Address]map[uint64]*UserStakeData{
		addr: {
			50:  makeDelta(1000), // entry pre-epoch
			150: makeDelta(-500), // withdraw mid-epoch
		},
	}

	min := minStakeFor(t, epochStart, epochEnd, stakeDataMap, addr)
	if min == nil {
		t.Fatalf("expected addr to be included, but it was excluded")
	}
	want := big.NewInt(500)
	if min.Cmp(want) != 0 {
		t.Errorf("expected min=%s after mid-epoch withdraw, got min=%s",
			want.String(), min.String())
	}
}

// TestFindMinOverBlockRange_NoInRangeEventsCarriedStake pins the line 1108
// fallback: a staker with only pre-epoch events keeps their carried stake as
// the min. Passes today; should keep passing after Phase 2.
func TestFindMinOverBlockRange_NoInRangeEventsCarriedStake(t *testing.T) {
	addr := common.HexToAddress("0x000000000000000000000000000000000000000E")
	epochStart, epochEnd := uint64(100), uint64(200)

	stakeDataMap := map[common.Address]map[uint64]*UserStakeData{
		addr: {
			50: makeDelta(100), // pre-epoch only
		},
	}

	min := minStakeFor(t, epochStart, epochEnd, stakeDataMap, addr)
	if min == nil {
		t.Fatalf("expected addr to be included (carried stake), but it was excluded")
	}
	want := big.NewInt(100)
	if min.Cmp(want) != 0 {
		t.Errorf("expected min=%s (carried stake), got min=%s",
			want.String(), min.String())
	}
}

// TestFindMinOverBlockRange_FreshMidEpochDepositorAgainstBaseline shows the
// fairness consequence of the bug: under linear normalization, a mid-epoch
// fresh depositor (min should be 0) ends up with the entire probability mass,
// shutting out baseline stakers entirely.
//
// Expected after Phase 2: mid-epoch fresh depositor (addrLate) is EXCLUDED;
// each baseline staker gets ~1/3 of the probability.
// Current behavior: addrLate dominates totalMin and absorbs ~all probability.
func TestFindMinOverBlockRange_FreshMidEpochDepositorAgainstBaseline(t *testing.T) {
	addr1 := common.HexToAddress("0x0000000000000000000000000000000000000001")
	addr2 := common.HexToAddress("0x0000000000000000000000000000000000000002")
	addr3 := common.HexToAddress("0x0000000000000000000000000000000000000003")
	addrLate := common.HexToAddress("0x00000000000000000000000000000000000000FF")
	epochStart, epochEnd := uint64(100), uint64(200)

	stakeDataMap := map[common.Address]map[uint64]*UserStakeData{
		addr1:    {50: makeDelta(1000)},
		addr2:    {50: makeDelta(1000)},
		addr3:    {50: makeDelta(1000)},
		addrLate: {150: makeDelta(1_000_000)}, // shows up mid-epoch with huge deposit
	}

	totalMin, addressMins, err := findMinOverBlockRange(epochStart, epochEnd, stakeDataMap)
	if err != nil {
		t.Fatalf("findMinOverBlockRange error: %v", err)
	}

	if _, included := addressMins[addrLate]; included {
		v := addressMins[addrLate].StakeAmount
		t.Errorf("FAIL (reproduces bug): mid-epoch fresh depositor should be excluded, "+
			"but is included with min=%s out of total=%s — they dominate the lottery",
			v.String(), totalMin.String())
	}
}
