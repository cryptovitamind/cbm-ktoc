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

// ============================================================================
// Phase 4 (TDD) — reproduce the withdraw-erasure bug in buildStakeDataMap.
// Per-block delta clamp at find_receiver.go:998-1000 erases withdraw events
// at blocks where the same wallet didn't also stake in that block. Result:
// the node looks up cumulative stake without ever subtracting the withdraw,
// votes for wallets that have already unstaked, and gets stuck without
// consensus. See plan: /Users/joe/.claude/plans/please-do-an-audit-typed-oasis.md
//
// Realistic fixtures: wei-scale amounts (whole-ETH units) and block numbers
// in the live-chain range so the tests don't look synthetic.

// weiPerEth is 10^18 — the wei-to-ETH multiplier.
var weiPerEth = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

// eth returns n ETH expressed in wei.
func eth(n int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(n), weiPerEth)
}

// tenthEth returns n/10 ETH expressed in wei (e.g., tenthEth(4) = 0.4 ETH).
func tenthEth(n int64) *big.Int {
	e17 := new(big.Int).Exp(big.NewInt(10), big.NewInt(17), nil)
	return new(big.Int).Mul(big.NewInt(n), e17)
}

// TestBuildStakeDataMap_WithdrawAtDifferentBlockPreservesSignedDelta is the
// minimal helper-level reproduction. A wallet stakes 5 ETH at one block and
// withdraws all 5 ETH at a later block (no other event at the withdraw
// block). The per-block delta at the withdraw block should be -5 ETH; today
// it's clamped to 0.
func TestBuildStakeDataMap_WithdrawAtDifferentBlockPreservesSignedDelta(t *testing.T) {
	w := common.HexToAddress("0x000000000000000000000000000000000000ABcD")
	stakeEvents := []StakeEvent{
		{Addr: w, Amount: eth(5), Block: 18_000_000},
	}
	withdrawEvents := []WithdrawEvent{
		{Addr: w, Amount: eth(5), Block: 18_000_300},
	}

	got := buildStakeDataMap(stakeEvents, withdrawEvents)

	if got[w][18_000_000].StakeAmount.Cmp(eth(5)) != 0 {
		t.Errorf("stake-block delta wrong: got %s, want %s",
			got[w][18_000_000].StakeAmount.String(), eth(5).String())
	}
	want := new(big.Int).Neg(eth(5))
	if got[w][18_000_300].StakeAmount.Cmp(want) != 0 {
		t.Errorf("FAIL (reproduces bug): withdraw-block delta should be %s, got %s "+
			"(per-block clamp erases the withdraw)",
			want.String(), got[w][18_000_300].StakeAmount.String())
	}
}

// TestBuildStakeDataMap_PartialWithdrawAtDifferentBlock — wallet keeps part of
// their stake; the partial withdraw should still register as a negative delta.
func TestBuildStakeDataMap_PartialWithdrawAtDifferentBlock(t *testing.T) {
	w := common.HexToAddress("0x000000000000000000000000000000000000BeeF")
	stakeEvents := []StakeEvent{
		{Addr: w, Amount: eth(5), Block: 18_000_000},
	}
	withdrawEvents := []WithdrawEvent{
		{Addr: w, Amount: eth(2), Block: 18_000_300},
	}

	got := buildStakeDataMap(stakeEvents, withdrawEvents)

	want := new(big.Int).Neg(eth(2))
	if got[w][18_000_300].StakeAmount.Cmp(want) != 0 {
		t.Errorf("FAIL (reproduces bug): partial-withdraw delta should be %s, got %s",
			want.String(), got[w][18_000_300].StakeAmount.String())
	}
}

// TestBuildStakeDataMap_WithdrawAtSameBlockAsLargerStake pins existing correct
// behavior: stake + withdraw in the same block (two txs, one block) should
// net to the positive difference. Passes today; must still pass after the fix.
func TestBuildStakeDataMap_WithdrawAtSameBlockAsLargerStake(t *testing.T) {
	w := common.HexToAddress("0x000000000000000000000000000000000000C0DE")
	const blk = uint64(18_000_000)
	stakeEvents := []StakeEvent{
		{Addr: w, Amount: eth(5), Block: blk},
	}
	withdrawEvents := []WithdrawEvent{
		{Addr: w, Amount: eth(2), Block: blk},
	}

	got := buildStakeDataMap(stakeEvents, withdrawEvents)

	if got[w][blk].StakeAmount.Cmp(eth(3)) != 0 {
		t.Errorf("same-block stake+withdraw delta wrong: got %s, want %s",
			got[w][blk].StakeAmount.String(), eth(3).String())
	}
}

// TestBuildStakeDataMap_MultipleWithdrawsAcrossBlocks — staking once and
// then unwinding over several blocks (a realistic exit pattern). Each
// withdraw is at its own block, so each is independently clamped to zero
// under the bug.
func TestBuildStakeDataMap_MultipleWithdrawsAcrossBlocks(t *testing.T) {
	w := common.HexToAddress("0x000000000000000000000000000000000000DeFa")
	stakeEvents := []StakeEvent{
		{Addr: w, Amount: eth(10), Block: 18_000_000},
	}
	withdrawEvents := []WithdrawEvent{
		{Addr: w, Amount: eth(3), Block: 18_000_100},
		{Addr: w, Amount: eth(4), Block: 18_000_200},
		{Addr: w, Amount: eth(3), Block: 18_000_300},
	}

	got := buildStakeDataMap(stakeEvents, withdrawEvents)

	cases := []struct {
		block uint64
		want  *big.Int
	}{
		{18_000_100, new(big.Int).Neg(eth(3))},
		{18_000_200, new(big.Int).Neg(eth(4))},
		{18_000_300, new(big.Int).Neg(eth(3))},
	}
	for _, c := range cases {
		if got[w][c.block].StakeAmount.Cmp(c.want) != 0 {
			t.Errorf("FAIL (reproduces bug): block %d delta should be %s, got %s",
				c.block, c.want.String(), got[w][c.block].StakeAmount.String())
		}
	}
}

// TestFindMinOverBlockRange_UnstakerPreEpochIsExcluded is the marquee test.
// It directly mirrors the field report: a wallet that staked and then
// unstaked entirely before the epoch began should NOT appear in addressMins.
// Under the bug, the unstake is erased by the per-block clamp and the wallet
// looks like it still has its full stake going into the epoch — which is
// exactly why one node in the field voted for a wallet that "had unstaked".
func TestFindMinOverBlockRange_UnstakerPreEpochIsExcluded(t *testing.T) {
	stayer1 := common.HexToAddress("0x0000000000000000000000000000000000000001")
	stayer2 := common.HexToAddress("0x0000000000000000000000000000000000000002")
	unstaker := common.HexToAddress("0x000000000000000000000000000000000000Bad1")
	const epochStart, epochEnd = uint64(18_000_500), uint64(18_001_100)

	// stakeDataMap is what realGatherStakesAndWithdraws would produce for
	// this set of events — we build it via the same buildStakeDataMap path
	// the production code uses, so the test exercises the actual bug.
	stakeEvents := []StakeEvent{
		{Addr: stayer1, Amount: eth(1), Block: 18_000_000},
		{Addr: stayer2, Amount: eth(1), Block: 18_000_000},
		{Addr: unstaker, Amount: eth(1), Block: 18_000_000},
	}
	withdrawEvents := []WithdrawEvent{
		{Addr: unstaker, Amount: eth(1), Block: 18_000_200}, // unstakes pre-epoch
	}
	stakeDataMap := buildStakeDataMap(stakeEvents, withdrawEvents)

	_, addressMins, err := findMinOverBlockRange(epochStart, epochEnd, stakeDataMap)
	if err != nil {
		t.Fatalf("findMinOverBlockRange error: %v", err)
	}

	if _, included := addressMins[unstaker]; included {
		v := addressMins[unstaker].StakeAmount
		t.Errorf("FAIL (reproduces field report): a wallet that fully unstaked pre-epoch "+
			"should be excluded from the lottery, but appears in addressMins with min=%s. "+
			"This is the bug that caused a node to vote for a wallet that had unstaked.",
			v.String())
	}
	for _, addr := range []common.Address{stayer1, stayer2} {
		if v, ok := addressMins[addr]; !ok || v.StakeAmount.Cmp(eth(1)) != 0 {
			got := "missing"
			if ok {
				got = v.StakeAmount.String()
			}
			t.Errorf("stayer %s should have min=%s, got %s", addr.Hex(), eth(1).String(), got)
		}
	}
}

// TestFindMinOverBlockRange_FullUnstakeMidEpochEliminatesEligibility — same
// scenario but the unstake happens mid-epoch. Wallet has full stake going
// into the epoch, drops to zero mid-way; correct minStake is 0 (excluded).
func TestFindMinOverBlockRange_FullUnstakeMidEpochEliminatesEligibility(t *testing.T) {
	w := common.HexToAddress("0x000000000000000000000000000000000000Bad2")
	const epochStart, epochEnd = uint64(18_000_500), uint64(18_001_100)

	stakeEvents := []StakeEvent{
		{Addr: w, Amount: eth(1), Block: 18_000_000}, // pre-epoch
	}
	withdrawEvents := []WithdrawEvent{
		{Addr: w, Amount: eth(1), Block: 18_000_700}, // mid-epoch
	}
	stakeDataMap := buildStakeDataMap(stakeEvents, withdrawEvents)

	_, addressMins, err := findMinOverBlockRange(epochStart, epochEnd, stakeDataMap)
	if err != nil {
		t.Fatalf("findMinOverBlockRange error: %v", err)
	}

	if v, included := addressMins[w]; included {
		t.Errorf("FAIL (reproduces bug): mid-epoch full unstake should exclude the wallet, "+
			"got min=%s", v.StakeAmount.String())
	}
}

// TestFindMinOverBlockRange_PartialWithdrawMidEpochReducesMin — wallet stakes
// 1 ETH, withdraws 0.4 ETH mid-epoch. Correct minStake is 0.6 ETH (the
// post-withdraw stake).
func TestFindMinOverBlockRange_PartialWithdrawMidEpochReducesMin(t *testing.T) {
	w := common.HexToAddress("0x000000000000000000000000000000000000Bad3")
	const epochStart, epochEnd = uint64(18_000_500), uint64(18_001_100)

	stakeEvents := []StakeEvent{
		{Addr: w, Amount: eth(1), Block: 18_000_000},
	}
	withdrawEvents := []WithdrawEvent{
		{Addr: w, Amount: tenthEth(4), Block: 18_000_700}, // 0.4 ETH out
	}
	stakeDataMap := buildStakeDataMap(stakeEvents, withdrawEvents)

	_, addressMins, err := findMinOverBlockRange(epochStart, epochEnd, stakeDataMap)
	if err != nil {
		t.Fatalf("findMinOverBlockRange error: %v", err)
	}

	want := tenthEth(6) // 0.6 ETH
	v, included := addressMins[w]
	if !included {
		t.Fatalf("FAIL: wallet should still be eligible after partial withdraw")
	}
	if v.StakeAmount.Cmp(want) != 0 {
		t.Errorf("FAIL (reproduces bug): after partial mid-epoch withdraw, min should be %s, got %s",
			want.String(), v.StakeAmount.String())
	}
}
