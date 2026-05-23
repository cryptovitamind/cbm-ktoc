package ktfunc

// Phase 6c — TTL cache for "current" contract state reads.
//
// StartBlock / EpochInterval / ConsensusReq / TlOcFees are queried every
// VoteAndReward invocation but change very rarely on-chain. With the
// node running in a loop (default WaitDuration = 1 min between epochs)
// this was 4-5 round trips per loop iteration that didn't need to
// happen. A 30-second TTL closes that gap with no consensus risk —
// even if state changes mid-window, the staleness is bounded.
//
// Important: this cache ONLY applies to the "current state" reads (i.e.
// nil BlockNumber in the call opts). State-at-block queries (e.g.,
// `Kt.StartBlock(BlockNumber: prevBlock)` in rwd.go) bypass the cache.

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

const (
	contractStateCacheTTL = 30 * time.Second
	gasPriceCacheTTL      = 60 * time.Second
)

// cachedValue is a single-entry TTL cache safe for concurrent access. Zero
// value is a fresh cache with no entry. Generic so each call site keeps
// its own type-safe value.
type cachedValue[T any] struct {
	mu        sync.Mutex
	value     T
	expiresAt time.Time
	hasValue  bool
}

// Get returns the cached value and true if a non-expired entry exists.
func (c *cachedValue[T]) Get() (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.hasValue && time.Now().Before(c.expiresAt) {
		return c.value, true
	}
	var zero T
	return zero, false
}

// Set stores a value with the given TTL.
func (c *cachedValue[T]) Set(v T, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = v
	c.expiresAt = time.Now().Add(ttl)
	c.hasValue = true
}

// cachedStartBlock returns the contract's current StartBlock, hitting the
// chain at most once per contractStateCacheTTL. Use this for "what's the
// current epoch start?" — NOT for `StartBlock(BlockNumber: X)` queries.
func cachedStartBlock(cProps *ConnectionProps) (*big.Int, error) {
	if v, ok := cProps.cachedStartBlock.Get(); ok {
		return new(big.Int).Set(v), nil
	}
	v, err := cProps.Kt.StartBlock(&bind.CallOpts{Context: context.Background(), From: cProps.MyPubKey})
	if err != nil {
		return nil, err
	}
	cProps.cachedStartBlock.Set(new(big.Int).Set(v), contractStateCacheTTL)
	return v, nil
}

func cachedEpochInterval(cProps *ConnectionProps) (uint16, error) {
	if v, ok := cProps.cachedEpochInterval.Get(); ok {
		return v, nil
	}
	v, err := cProps.Kt.EpochInterval(&bind.CallOpts{Context: context.Background(), From: cProps.MyPubKey})
	if err != nil {
		return 0, err
	}
	cProps.cachedEpochInterval.Set(v, contractStateCacheTTL)
	return v, nil
}

func cachedConsensusReq(cProps *ConnectionProps) (uint16, error) {
	if v, ok := cProps.cachedConsensusReq.Get(); ok {
		return v, nil
	}
	v, err := cProps.Kt.ConsensusReq(&bind.CallOpts{Context: context.Background(), From: cProps.MyPubKey})
	if err != nil {
		return 0, err
	}
	cProps.cachedConsensusReq.Set(v, contractStateCacheTTL)
	return v, nil
}

func cachedTlOcFees(cProps *ConnectionProps) (*big.Int, error) {
	if v, ok := cProps.cachedTlOcFees.Get(); ok {
		return new(big.Int).Set(v), nil
	}
	v, err := cProps.Kt.TlOcFees(&bind.CallOpts{Context: context.Background(), From: cProps.MyPubKey})
	if err != nil {
		return nil, err
	}
	cProps.cachedTlOcFees.Set(new(big.Int).Set(v), contractStateCacheTTL)
	return v, nil
}

// cachedSuggestGasPrice returns the eth client's gas-price suggestion,
// re-querying at most once per gasPriceCacheTTL. Gas prices fluctuate
// continuously but the consumers here use the result only for INFO logging
// (kt_props.go) or tx-construction defaults (vote_ops.go), so a 60-second
// staleness is fine and saves several RPC calls per loop iteration.
func cachedSuggestGasPrice(cProps *ConnectionProps) (*big.Int, error) {
	if v, ok := cProps.cachedGasPrice.Get(); ok {
		return new(big.Int).Set(v), nil
	}
	v, err := cProps.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	cProps.cachedGasPrice.Set(new(big.Int).Set(v), gasPriceCacheTTL)
	return v, nil
}
