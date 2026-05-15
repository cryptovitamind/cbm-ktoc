package integration_test

// Phase 1 (TDD) — reproduces the wasteful re-execution of
// GetContractCreationBlock's binary search on every VoteAndReward call.
// A contract's creation block does not change, so the second call
// should make zero CodeAt RPC calls.
//
// On master, the second call runs the same ~log₂(latestBlock) CodeAt
// queries as the first call. After Phase 3 caches the result on
// ConnectionProps, this test will pass.

import (
	"context"
	"math/big"
	"sync"
	"testing"

	"ktp2/src/ktp2/ktfunc"

	"github.com/ethereum/go-ethereum/common"
)

func TestGetContractCreationBlock_CachedAcrossCalls(t *testing.T) {
	const (
		latestBlock  = uint64(1_000_000)
		creationAt   = uint64(500_000)
		contractAddr = "0x000000000000000000000000000000000000BEEF"
	)

	var mu sync.Mutex
	codeAtCalls := 0

	fakeClient := &FakeEthClient{}
	fakeClient.BlockNumberFn = func(ctx context.Context) (uint64, error) {
		return latestBlock, nil
	}
	fakeClient.CodeAtFn = func(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
		mu.Lock()
		codeAtCalls++
		mu.Unlock()
		if blockNumber.Uint64() >= creationAt {
			return []byte{0x60, 0x80, 0x60, 0x40}, nil // non-empty bytecode
		}
		return []byte{}, nil
	}

	cProps := &ktfunc.ConnectionProps{
		Client: fakeClient,
		KtAddr: common.HexToAddress(contractAddr),
	}

	// First call: binary search expected.
	block1, err := ktfunc.GetContractCreationBlock(cProps)
	if err != nil {
		t.Fatalf("first GetContractCreationBlock: %v", err)
	}
	firstCount := codeAtCalls
	t.Logf("first call returned block=%d after %d CodeAt RPC calls", block1, firstCount)
	if firstCount == 0 {
		t.Fatalf("test setup wrong: expected first call to make at least one CodeAt call, got 0")
	}

	// Reset counter and call again.
	mu.Lock()
	codeAtCalls = 0
	mu.Unlock()

	block2, err := ktfunc.GetContractCreationBlock(cProps)
	if err != nil {
		t.Fatalf("second GetContractCreationBlock: %v", err)
	}
	if block1 != block2 {
		t.Errorf("first and second calls returned different blocks: %d vs %d", block1, block2)
	}

	if codeAtCalls > 0 {
		t.Errorf("FAIL (reproduces bug): second call to GetContractCreationBlock made %d CodeAt RPC calls; "+
			"expected 0 because the contract creation block is immutable and should be cached after the first lookup",
			codeAtCalls)
	}
}
