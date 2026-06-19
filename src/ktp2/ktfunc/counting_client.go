package ktfunc

// RPC call counting.
//
// Node runners pay per JSON-RPC call to Alchemy/Infura, and the common
// complaint is "too many calls". CountingClient wraps the Ethereum client and
// tallies every call by its JSON-RPC method name, so the node can log exactly
// what it is sending and we can confirm the on-disk cache is keeping the
// expensive eth_getLogs / eth_call counts down.
//
// It wraps BOTH roles the node uses the client for: the direct EthClient calls
// AND the contract-binding backend (bind.ContractBackend). The binding backend
// is where eth_getLogs (FilterStaked/Voted/Rwd) and eth_call (contract reads)
// actually go, which is the bulk of the traffic.

import (
	"context"
	"math/big"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
)

// fullClient is the union of the two roles the node uses the RPC client for.
// Both *ethclient.Client and the simulated backend client satisfy it.
type fullClient interface {
	EthClient
	bind.ContractBackend
}

// CountingClient wraps a fullClient and counts calls by JSON-RPC method name.
// Safe for concurrent use.
type CountingClient struct {
	inner  fullClient
	mu     sync.Mutex
	counts map[string]int64
}

// NewCountingClient wraps an Ethereum client to count its RPC calls.
func NewCountingClient(inner fullClient) *CountingClient {
	return &CountingClient{inner: inner, counts: make(map[string]int64)}
}

func (c *CountingClient) inc(method string) {
	c.mu.Lock()
	c.counts[method]++
	c.mu.Unlock()
}

// Snapshot returns a copy of the current per-method counts and their total.
func (c *CountingClient) Snapshot() (map[string]int64, int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]int64, len(c.counts))
	var total int64
	for k, v := range c.counts {
		out[k] = v
		total += v
	}
	return out, total
}

// Reset zeroes the counters (e.g. at the start of each loop iteration).
func (c *CountingClient) Reset() {
	c.mu.Lock()
	c.counts = make(map[string]int64)
	c.mu.Unlock()
}

// LogSummary logs a one-line, method-by-method breakdown sorted by call count,
// most expensive first. prefix labels the window (e.g. "RPC this cycle").
func (c *CountingClient) LogSummary(prefix string) {
	counts, total := c.Snapshot()
	if total == 0 {
		log.Infof("%s: 0 RPC calls", prefix)
		return
	}
	type kv struct {
		method string
		n      int64
	}
	pairs := make([]kv, 0, len(counts))
	for m, n := range counts {
		pairs = append(pairs, kv{m, n})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].n != pairs[j].n {
			return pairs[i].n > pairs[j].n
		}
		return pairs[i].method < pairs[j].method
	})
	parts := ""
	for _, p := range pairs {
		parts += " " + p.method + "=" + itoa(p.n)
	}
	log.Infof("%s: %d RPC calls |%s", prefix, total, parts)
}

func itoa(n int64) string { return new(big.Int).SetInt64(n).String() }

// ---- EthClient + bind.ContractBackend, each counting then delegating ----

func (c *CountingClient) CodeAt(ctx context.Context, account common.Address, block *big.Int) ([]byte, error) {
	c.inc("eth_getCode")
	return c.inner.CodeAt(ctx, account, block)
}

func (c *CountingClient) CallContract(ctx context.Context, call ethereum.CallMsg, block *big.Int) ([]byte, error) {
	c.inc("eth_call")
	return c.inner.CallContract(ctx, call, block)
}

func (c *CountingClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	c.inc("eth_getBlockByNumber")
	return c.inner.HeaderByNumber(ctx, number)
}

func (c *CountingClient) BalanceAt(ctx context.Context, account common.Address, block *big.Int) (*big.Int, error) {
	c.inc("eth_getBalance")
	return c.inner.BalanceAt(ctx, account, block)
}

func (c *CountingClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	c.inc("eth_getTransactionCount")
	return c.inner.PendingNonceAt(ctx, account)
}

func (c *CountingClient) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	c.inc("eth_getCode")
	return c.inner.PendingCodeAt(ctx, account)
}

func (c *CountingClient) BlockNumber(ctx context.Context) (uint64, error) {
	c.inc("eth_blockNumber")
	return c.inner.BlockNumber(ctx)
}

func (c *CountingClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	c.inc("eth_gasPrice")
	return c.inner.SuggestGasPrice(ctx)
}

func (c *CountingClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	c.inc("eth_maxPriorityFeePerGas")
	return c.inner.SuggestGasTipCap(ctx)
}

func (c *CountingClient) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	c.inc("eth_estimateGas")
	return c.inner.EstimateGas(ctx, call)
}

func (c *CountingClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	c.inc("eth_sendRawTransaction")
	return c.inner.SendTransaction(ctx, tx)
}

func (c *CountingClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	c.inc("eth_getTransactionReceipt")
	return c.inner.TransactionReceipt(ctx, txHash)
}

func (c *CountingClient) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	c.inc("eth_getTransactionByHash")
	return c.inner.TransactionByHash(ctx, hash)
}

func (c *CountingClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	c.inc("eth_getLogs")
	return c.inner.FilterLogs(ctx, q)
}

func (c *CountingClient) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	c.inc("eth_subscribe(logs)")
	return c.inner.SubscribeFilterLogs(ctx, q, ch)
}

func (c *CountingClient) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	c.inc("eth_subscribe(newHeads)")
	return c.inner.SubscribeNewHead(ctx, ch)
}
