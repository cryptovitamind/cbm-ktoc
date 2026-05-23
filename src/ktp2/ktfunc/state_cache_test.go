package ktfunc

// Phase 6c — tests for the contract-state TTL cache.

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCachedStartBlock_HitsContractOnceWithinTTL(t *testing.T) {
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Kt:       mockKt,
		MyPubKey: common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
	}
	mockKt.On("StartBlock", mock.Anything).Return(big.NewInt(18_000_000), nil).Once()

	for i := 0; i < 5; i++ {
		v, err := cachedStartBlock(cProps)
		assert.NoError(t, err)
		assert.Equal(t, uint64(18_000_000), v.Uint64())
	}
	mockKt.AssertNumberOfCalls(t, "StartBlock", 1)
}

func TestCachedEpochInterval_HitsContractOnceWithinTTL(t *testing.T) {
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Kt:       mockKt,
		MyPubKey: common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
	}
	mockKt.On("EpochInterval", mock.Anything).Return(uint16(600), nil).Once()

	for i := 0; i < 3; i++ {
		v, err := cachedEpochInterval(cProps)
		assert.NoError(t, err)
		assert.Equal(t, uint16(600), v)
	}
	mockKt.AssertNumberOfCalls(t, "EpochInterval", 1)
}

func TestCachedConsensusReq_HitsContractOnceWithinTTL(t *testing.T) {
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Kt:       mockKt,
		MyPubKey: common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
	}
	mockKt.On("ConsensusReq", mock.Anything).Return(uint16(3), nil).Once()

	for i := 0; i < 3; i++ {
		v, err := cachedConsensusReq(cProps)
		assert.NoError(t, err)
		assert.Equal(t, uint16(3), v)
	}
	mockKt.AssertNumberOfCalls(t, "ConsensusReq", 1)
}

func TestCachedTlOcFees_HitsContractOnceWithinTTL(t *testing.T) {
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Kt:       mockKt,
		MyPubKey: common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
	}
	mockKt.On("TlOcFees", mock.Anything).Return(big.NewInt(1_000_000), nil).Once()

	for i := 0; i < 3; i++ {
		v, err := cachedTlOcFees(cProps)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1_000_000), v.Uint64())
	}
	mockKt.AssertNumberOfCalls(t, "TlOcFees", 1)
}

func TestCachedValue_ExpiresAfterTTL(t *testing.T) {
	var c cachedValue[int]
	c.Set(42, 10*time.Millisecond)
	v, ok := c.Get()
	assert.True(t, ok)
	assert.Equal(t, 42, v)

	time.Sleep(15 * time.Millisecond)
	_, ok = c.Get()
	assert.False(t, ok, "value should have expired")
}

func TestCachedStartBlock_PropagatesContractError(t *testing.T) {
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Kt:       mockKt,
		MyPubKey: common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
	}
	mockKt.On("StartBlock", mock.Anything).Return((*big.Int)(nil), assert.AnError)
	_, err := cachedStartBlock(cProps)
	assert.Error(t, err)
	// Failure should NOT populate the cache; second call hits the contract again.
	_, err = cachedStartBlock(cProps)
	assert.Error(t, err)
	mockKt.AssertNumberOfCalls(t, "StartBlock", 2)
}
