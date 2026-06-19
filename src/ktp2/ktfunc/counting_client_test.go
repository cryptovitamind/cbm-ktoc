package ktfunc

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

// noopClient implements fullClient with zero-value returns, for counting tests.
type noopClient struct{}

func (noopClient) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) { return nil, nil }
func (noopClient) CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error) {
	return nil, nil
}
func (noopClient) HeaderByNumber(context.Context, *big.Int) (*types.Header, error)       { return nil, nil }
func (noopClient) BalanceAt(context.Context, common.Address, *big.Int) (*big.Int, error) { return nil, nil }
func (noopClient) PendingNonceAt(context.Context, common.Address) (uint64, error)        { return 0, nil }
func (noopClient) PendingCodeAt(context.Context, common.Address) ([]byte, error)         { return nil, nil }
func (noopClient) BlockNumber(context.Context) (uint64, error)                           { return 0, nil }
func (noopClient) SuggestGasPrice(context.Context) (*big.Int, error)                     { return nil, nil }
func (noopClient) SuggestGasTipCap(context.Context) (*big.Int, error)                    { return nil, nil }
func (noopClient) EstimateGas(context.Context, ethereum.CallMsg) (uint64, error)         { return 0, nil }
func (noopClient) SendTransaction(context.Context, *types.Transaction) error             { return nil }
func (noopClient) TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	return nil, nil
}
func (noopClient) TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error) {
	return nil, false, nil
}
func (noopClient) FilterLogs(context.Context, ethereum.FilterQuery) ([]types.Log, error) {
	return nil, nil
}
func (noopClient) SubscribeFilterLogs(context.Context, ethereum.FilterQuery, chan<- types.Log) (ethereum.Subscription, error) {
	return nil, nil
}
func (noopClient) SubscribeNewHead(context.Context, chan<- *types.Header) (ethereum.Subscription, error) {
	return nil, nil
}

func TestCountingClient_CountsByMethod(t *testing.T) {
	c := NewCountingClient(noopClient{})

	_, _ = c.BlockNumber(context.Background())
	_, _ = c.BlockNumber(context.Background())
	_, _ = c.FilterLogs(context.Background(), ethereum.FilterQuery{})
	_, _ = c.BalanceAt(context.Background(), common.Address{}, nil)
	_, _ = c.HeaderByNumber(context.Background(), big.NewInt(1))
	_, _ = c.CallContract(context.Background(), ethereum.CallMsg{}, nil)

	counts, total := c.Snapshot()
	assert.Equal(t, int64(6), total)
	assert.Equal(t, int64(2), counts["eth_blockNumber"])
	assert.Equal(t, int64(1), counts["eth_getLogs"])
	assert.Equal(t, int64(1), counts["eth_getBalance"])
	assert.Equal(t, int64(1), counts["eth_getBlockByNumber"])
	assert.Equal(t, int64(1), counts["eth_call"])

	c.Reset()
	_, afterReset := c.Snapshot()
	assert.Equal(t, int64(0), afterReset)
}
