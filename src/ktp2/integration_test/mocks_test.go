package integration_test

// Test-only fakes for the integration_test package. We use the "embedded
// nil interface" trick so the structs satisfy ktfunc.Ktv2Interface and
// ktfunc.EthClient without writing 50+ stub methods: any method not
// overridden will panic at call time, which is exactly what we want — it
// flags an un-mocked dependency loudly during a test rather than silently
// returning a zero value.

import (
	"context"
	"math/big"

	"ktp2/src/abis/ktv2"
	"ktp2/src/ktp2/ktfunc"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// FilterRange captures a (start, end) range a test observed being queried.
type FilterRange struct {
	Start uint64
	End   uint64
}

// FakeKtv2 implements ktfunc.Ktv2Interface. Override the hook fields you
// care about; calling any non-overridden method will panic (because the
// embedded interface is nil).
type FakeKtv2 struct {
	ktfunc.Ktv2Interface // nil; any non-overridden call will panic

	FilterStakedFn   func(opts *bind.FilterOpts) (ktfunc.StakedIterator, error)
	FilterWithdrewFn func(opts *bind.FilterOpts) (ktfunc.WithdrewIterator, error)
}

func (f *FakeKtv2) FilterStaked(opts *bind.FilterOpts) (ktfunc.StakedIterator, error) {
	return f.FilterStakedFn(opts)
}

func (f *FakeKtv2) FilterWithdrew(opts *bind.FilterOpts) (ktfunc.WithdrewIterator, error) {
	return f.FilterWithdrewFn(opts)
}

// FakeEthClient implements ktfunc.EthClient with overridable hooks.
type FakeEthClient struct {
	ktfunc.EthClient // nil; any non-overridden call will panic

	CodeAtFn         func(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
	BlockNumberFn    func(ctx context.Context) (uint64, error)
	FilterLogsFn     func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
	HeaderByNumberFn func(ctx context.Context, number *big.Int) (*types.Header, error)
}

func (f *FakeEthClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if f.HeaderByNumberFn != nil {
		return f.HeaderByNumberFn(ctx, number)
	}
	// Default: pretend the chain advanced past whatever was asked. Returning
	// a header with the requested number gives any tipHash-capture path
	// something deterministic to work with; tests that care about the
	// returned hash should set HeaderByNumberFn explicitly.
	if number == nil {
		return &types.Header{Number: big.NewInt(0)}, nil
	}
	return &types.Header{Number: new(big.Int).Set(number)}, nil
}

func (f *FakeEthClient) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	return f.CodeAtFn(ctx, account, blockNumber)
}

func (f *FakeEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	if f.BlockNumberFn != nil {
		return f.BlockNumberFn(ctx)
	}
	// Default: pretend the chain head is far ahead of any block a test gathers,
	// so cached tips count as "buried" and the reorg-detection hash gets
	// recorded. Tests that need the head near a tip (to exercise the near-head
	// skip) set BlockNumberFn explicitly.
	return 10_000_000, nil
}

// FilterLogs is wired so `realGatherStakesAndWithdraws`'s debug-path
// (`debugRawLogs`) doesn't panic if the test happens to produce an empty
// event set. Defaults to returning an empty slice.
func (f *FakeEthClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	if f.FilterLogsFn != nil {
		return f.FilterLogsFn(ctx, q)
	}
	return []types.Log{}, nil
}

// stakedIter / withdrewIter are minimal in-memory iterator implementations.
type stakedIter struct {
	events []ktv2.Ktv2Staked
	i      int
}

func (it *stakedIter) Next() bool {
	if it.i >= len(it.events) {
		return false
	}
	it.i++
	return true
}
func (it *stakedIter) Event() *ktv2.Ktv2Staked { return &it.events[it.i-1] }
func (it *stakedIter) Error() error            { return nil }
func (it *stakedIter) Close() error            { return nil }

type withdrewIter struct {
	events []ktv2.Ktv2Withdrew
	i      int
}

func (it *withdrewIter) Next() bool {
	if it.i >= len(it.events) {
		return false
	}
	it.i++
	return true
}
func (it *withdrewIter) Event() *ktv2.Ktv2Withdrew { return &it.events[it.i-1] }
func (it *withdrewIter) Error() error              { return nil }
func (it *withdrewIter) Close() error              { return nil }
