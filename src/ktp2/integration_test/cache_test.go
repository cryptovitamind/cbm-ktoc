package integration_test

// Cache-correctness tests. These reproduce the original RPC-overuse report
// (the node re-scanned all blockchain history every vote cycle and rebuilt the
// entire stake/withdraw set even though old data never changes) and pin the
// fixes:
//
//   - The tip-pointer cache must not re-query blocks already cached by an
//     earlier call when the requested end-block shifts.
//   - A reorg at the cached tip must wipe and rebuild; a near-head tip must not
//     be hash-checked (so transient RPC inconsistencies don't trigger a storm).
//
// They drive ktfunc.GatherStakesAndWithdraws and observe the block ranges
// queried from a fake Ktv2.

import (
	"context"
	"math/big"
	"os"
	"sync"
	"testing"

	"ktp2/src/abis/ktv2"
	"ktp2/src/ktp2/ktfunc"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestGatherStakesAndWithdraws_DoesNotRefetchPreviouslyCachedBlocks(t *testing.T) {
	// Isolate cache in a temp directory. ktfunc writes to ./cache/<addr>.db
	// relative to CWD, so chdir into a temp dir for the duration of the test.
	tmp := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	var mu sync.Mutex
	var stakedRanges, withdrewRanges []FilterRange

	fake := &FakeKtv2{}
	fake.FilterStakedFn = func(opts *bind.FilterOpts) (ktfunc.StakedIterator, error) {
		mu.Lock()
		stakedRanges = append(stakedRanges, FilterRange{Start: opts.Start, End: *opts.End})
		mu.Unlock()
		return &stakedIter{events: []ktv2.Ktv2Staked{}}, nil
	}
	fake.FilterWithdrewFn = func(opts *bind.FilterOpts) (ktfunc.WithdrewIterator, error) {
		mu.Lock()
		withdrewRanges = append(withdrewRanges, FilterRange{Start: opts.Start, End: *opts.End})
		mu.Unlock()
		return &withdrewIter{events: []ktv2.Ktv2Withdrew{}}, nil
	}

	cProps := &ktfunc.ConnectionProps{
		KtAddr:    common.HexToAddress("0x000000000000000000000000000000000000ABCD"),
		ChunkSize: 500,
		Client:    &FakeEthClient{}, // only used by debug paths we won't hit
	}

	// First scan: blocks [100, 950]. End is intentionally a non-clean
	// chunk boundary so the trailing chunk's key is (600, 950).
	if _, err := ktfunc.GatherStakesAndWithdraws(cProps, fake, big.NewInt(100), big.NewInt(950)); err != nil {
		t.Fatalf("first GatherStakesAndWithdraws: %v", err)
	}
	firstCallStakedCount := len(stakedRanges)
	firstCallWithdrewCount := len(withdrewRanges)
	t.Logf("first call: %d FilterStaked, %d FilterWithdrew queries", firstCallStakedCount, firstCallWithdrewCount)

	// Reset counters for the second call.
	stakedRanges = nil
	withdrewRanges = nil

	// Second scan: blocks [100, 1500]. Most of this range overlaps with
	// the first call and should be served from cache. Today, the chunk
	// (600, 1099) is keyed differently from the cached (600, 950), so
	// the node re-queries the entire (600, 1099) range from the node.
	if _, err := ktfunc.GatherStakesAndWithdraws(cProps, fake, big.NewInt(100), big.NewInt(1500)); err != nil {
		t.Fatalf("second GatherStakesAndWithdraws: %v", err)
	}
	t.Logf("second call: %d FilterStaked, %d FilterWithdrew queries", len(stakedRanges), len(withdrewRanges))

	// Assertion: blocks already fully cached during the first call should
	// not be re-queried in the second call. Block 800 lies firmly inside
	// the first call's cached range (the (600, 950) chunk).
	const definitelyCachedBlock = uint64(800)
	rangeContains := func(r FilterRange, b uint64) bool { return r.Start <= b && r.End >= b }

	for _, r := range stakedRanges {
		if rangeContains(r, definitelyCachedBlock) {
			t.Errorf("FAIL (reproduces bug): second-call FilterStaked re-queried range %d-%d which contains "+
				"block %d already cached by the first call",
				r.Start, r.End, definitelyCachedBlock)
		}
	}
	for _, r := range withdrewRanges {
		if rangeContains(r, definitelyCachedBlock) {
			t.Errorf("FAIL (reproduces bug): second-call FilterWithdrew re-queried range %d-%d which contains "+
				"block %d already cached by the first call",
				r.Start, r.End, definitelyCachedBlock)
		}
	}
}

// TestGatherStakesAndWithdraws_DetectsReorgAndRebuildsCache — the cache stores
// tipHash alongside the tip pointer. If the chain's hash at the (buried) tip
// changes between calls (i.e. a reorg dropped or rewrote events past the cached
// point), the second call must wipe the chunks bucket and re-fetch.
func TestGatherStakesAndWithdraws_DetectsReorgAndRebuildsCache(t *testing.T) {
	tmp := t.TempDir()
	oldwd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	contractAddr := common.HexToAddress("0x000000000000000000000000000000000000FACE")
	cProps := &ktfunc.ConnectionProps{
		KtAddr:    contractAddr,
		ChunkSize: 500,
	}

	hashBefore := common.HexToHash("0xaaaa000000000000000000000000000000000000000000000000000000000000")
	fakeClient := &FakeEthClient{
		HeaderByNumberFn: func(_ context.Context, n *big.Int) (*types.Header, error) {
			return &types.Header{Number: new(big.Int).Set(n), ParentHash: hashBefore}, nil
		},
	}
	cProps.Client = fakeClient

	var stakedQueries int
	fakeKt := &FakeKtv2{}
	fakeKt.FilterStakedFn = func(opts *bind.FilterOpts) (ktfunc.StakedIterator, error) {
		stakedQueries++
		return &stakedIter{}, nil
	}
	fakeKt.FilterWithdrewFn = func(opts *bind.FilterOpts) (ktfunc.WithdrewIterator, error) {
		return &withdrewIter{}, nil
	}

	if _, err := ktfunc.GatherStakesAndWithdraws(cProps, fakeKt, big.NewInt(100), big.NewInt(599)); err != nil {
		t.Fatalf("first gather: %v", err)
	}
	if stakedQueries == 0 {
		t.Fatalf("test setup wrong: expected at least one FilterStaked call on first scan")
	}

	// Simulate reorg: header now reports a different ParentHash for the
	// same block number, flipping its Hash().
	hashAfter := common.HexToHash("0xbbbb000000000000000000000000000000000000000000000000000000000000")
	fakeClient.HeaderByNumberFn = func(_ context.Context, n *big.Int) (*types.Header, error) {
		return &types.Header{Number: new(big.Int).Set(n), ParentHash: hashAfter}, nil
	}
	stakedQueries = 0

	if _, err := ktfunc.GatherStakesAndWithdraws(cProps, fakeKt, big.NewInt(100), big.NewInt(599)); err != nil {
		t.Fatalf("second gather: %v", err)
	}
	if stakedQueries == 0 {
		t.Errorf("expected reorg detection to wipe cache and re-fetch; got %d new queries", stakedQueries)
	}
}

// TestGatherStakesAndWithdraws_NoReorgKeepsCache — same shape but the
// header hash is stable; the second call should hit the cache (zero
// new FilterStaked queries).
func TestGatherStakesAndWithdraws_NoReorgKeepsCache(t *testing.T) {
	tmp := t.TempDir()
	oldwd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	contractAddr := common.HexToAddress("0x000000000000000000000000000000000000C0DE")
	cProps := &ktfunc.ConnectionProps{
		KtAddr:    contractAddr,
		ChunkSize: 500,
	}

	stableParent := common.HexToHash("0xcccc000000000000000000000000000000000000000000000000000000000000")
	fakeClient := &FakeEthClient{
		HeaderByNumberFn: func(_ context.Context, n *big.Int) (*types.Header, error) {
			return &types.Header{Number: new(big.Int).Set(n), ParentHash: stableParent}, nil
		},
	}
	cProps.Client = fakeClient

	var stakedQueries int
	fakeKt := &FakeKtv2{}
	fakeKt.FilterStakedFn = func(opts *bind.FilterOpts) (ktfunc.StakedIterator, error) {
		stakedQueries++
		return &stakedIter{}, nil
	}
	fakeKt.FilterWithdrewFn = func(opts *bind.FilterOpts) (ktfunc.WithdrewIterator, error) {
		return &withdrewIter{}, nil
	}

	if _, err := ktfunc.GatherStakesAndWithdraws(cProps, fakeKt, big.NewInt(100), big.NewInt(599)); err != nil {
		t.Fatalf("first gather: %v", err)
	}
	stakedQueries = 0

	if _, err := ktfunc.GatherStakesAndWithdraws(cProps, fakeKt, big.NewInt(100), big.NewInt(599)); err != nil {
		t.Fatalf("second gather: %v", err)
	}
	if stakedQueries != 0 {
		t.Errorf("expected cache to serve the repeat call; got %d new FilterStaked queries", stakedQueries)
	}
}

// TestGatherStakesAndWithdraws_NearHeadTipNotHashChecked — when the cached tip
// is within reorgSafetyDepth of the chain head, its block hash must NOT be
// recorded. So even if that block's hash later changes (the kind of momentary
// inconsistency a load-balanced RPC returns near head), the next call must NOT
// false-trigger a reorg wipe — it serves from cache. This is what keeps a node
// from re-fetching every chunk (the "too many requests" storm) at vote time,
// when the tip sits only a few blocks behind head.
func TestGatherStakesAndWithdraws_NearHeadTipNotHashChecked(t *testing.T) {
	tmp := t.TempDir()
	oldwd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	contractAddr := common.HexToAddress("0x000000000000000000000000000000000000BEEF")
	cProps := &ktfunc.ConnectionProps{
		KtAddr:    contractAddr,
		ChunkSize: 500,
	}

	// Head is only 6 blocks past the tip (599) — well inside reorgSafetyDepth
	// (32), so the tip counts as near-head and its hash is never recorded.
	hashBefore := common.HexToHash("0xaaaa000000000000000000000000000000000000000000000000000000000000")
	fakeClient := &FakeEthClient{
		BlockNumberFn: func(_ context.Context) (uint64, error) { return 605, nil },
		HeaderByNumberFn: func(_ context.Context, n *big.Int) (*types.Header, error) {
			return &types.Header{Number: new(big.Int).Set(n), ParentHash: hashBefore}, nil
		},
	}
	cProps.Client = fakeClient

	var stakedQueries int
	fakeKt := &FakeKtv2{}
	fakeKt.FilterStakedFn = func(opts *bind.FilterOpts) (ktfunc.StakedIterator, error) {
		stakedQueries++
		return &stakedIter{}, nil
	}
	fakeKt.FilterWithdrewFn = func(opts *bind.FilterOpts) (ktfunc.WithdrewIterator, error) {
		return &withdrewIter{}, nil
	}

	if _, err := ktfunc.GatherStakesAndWithdraws(cProps, fakeKt, big.NewInt(100), big.NewInt(599)); err != nil {
		t.Fatalf("first gather: %v", err)
	}

	// Flip the near-head block's hash, exactly the transient a load-balanced
	// RPC might return. Because no hash was recorded, this must not wipe.
	hashAfter := common.HexToHash("0xbbbb000000000000000000000000000000000000000000000000000000000000")
	fakeClient.HeaderByNumberFn = func(_ context.Context, n *big.Int) (*types.Header, error) {
		return &types.Header{Number: new(big.Int).Set(n), ParentHash: hashAfter}, nil
	}
	stakedQueries = 0

	if _, err := ktfunc.GatherStakesAndWithdraws(cProps, fakeKt, big.NewInt(100), big.NewInt(599)); err != nil {
		t.Fatalf("second gather: %v", err)
	}
	if stakedQueries != 0 {
		t.Errorf("near-head tip must not be hash-checked; expected cache hit, got %d new FilterStaked queries", stakedQueries)
	}
}
