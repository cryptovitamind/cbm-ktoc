package ktfunc

// Phase 6f — tests for the refactored WaitForBlocks (subscribe + polling
// fallback + deadline guard).

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// fakeSubscription is a minimal ethereum.Subscription that lets the test
// signal an error or just sit silent until Unsubscribe is called.
type fakeSubscription struct {
	errCh chan error
}

func newFakeSubscription() *fakeSubscription {
	return &fakeSubscription{errCh: make(chan error, 1)}
}

func (s *fakeSubscription) Unsubscribe()        { close(s.errCh) }
func (s *fakeSubscription) Err() <-chan error   { return s.errCh }
func (s *fakeSubscription) raise(err error)     { s.errCh <- err }

// TestWaitForBlocks_SubscriptionPathReturnsWhenTargetReached pins the
// happy path: SubscribeNewHead returns a working subscription, headers
// stream in, and once we see one at or past the target the function
// returns nil with zero BlockNumber polls.
func TestWaitForBlocks_SubscriptionPathReturnsWhenTargetReached(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{Client: mockClient, BlocksToWait: 3}

	// startBlock = 100, target = 103
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Once()

	sub := newFakeSubscription()
	var capturedCh chan<- *types.Header
	mockClient.On("SubscribeNewHead", mock.Anything, mock.AnythingOfType("chan<- *types.Header")).Return(
		(ethereum.Subscription)(sub), nil).Run(func(args mock.Arguments) {
		capturedCh = args.Get(1).(chan<- *types.Header)
	})

	// Stream headers in a goroutine.
	go func() {
		// Wait briefly for WaitForBlocks to subscribe.
		for capturedCh == nil {
			time.Sleep(time.Millisecond)
		}
		capturedCh <- &types.Header{Number: big.NewInt(101)}
		capturedCh <- &types.Header{Number: big.NewInt(102)}
		capturedCh <- &types.Header{Number: big.NewInt(103)} // target
	}()

	err := WaitForBlocks(cProps)
	assert.NoError(t, err)
	// Subscription path SHOULD avoid all the polling-side BlockNumber calls.
	// (One BlockNumber at start to fix the target is mocked .Once() above.)
	mockClient.AssertNumberOfCalls(t, "BlockNumber", 1)
}

// TestWaitForBlocks_FallsBackToPollingWhenSubscribeFails pins the
// HTTP-only-RPC fallback: SubscribeNewHead returns an error, the
// function falls through to the polling path. We mock BlockNumber to
// jump from start to past target on the second call.
func TestWaitForBlocks_FallsBackToPollingWhenSubscribeFails(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	// Shorten the poll interval so the test runs in milliseconds.
	origInterval := TimeToWaitForBlocks
	TimeToWaitForBlocks = 5 * time.Millisecond
	defer func() { TimeToWaitForBlocks = origInterval }()

	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{Client: mockClient, BlocksToWait: 1}

	// First BlockNumber: start = 50. Second (after poll sleep): 51 → done.
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(50), nil).Once()
	mockClient.On("SubscribeNewHead", mock.Anything, mock.Anything).Return(
		(ethereum.Subscription)(nil), assert.AnError) // unsupported
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(51), nil)

	err := WaitForBlocks(cProps)
	assert.NoError(t, err)
}

// TestWaitForBlocks_ZeroBlocksToWaitIsNoop — pin a small but important
// case: BlocksToWait=0 means "don't wait, return immediately."
func TestWaitForBlocks_ZeroBlocksToWaitIsNoop(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{Client: mockClient, BlocksToWait: 0}
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Once()

	err := WaitForBlocks(cProps)
	assert.NoError(t, err)
	mockClient.AssertNumberOfCalls(t, "BlockNumber", 1)
	// SubscribeNewHead must NOT be called when there's nothing to wait for.
	mockClient.AssertNotCalled(t, "SubscribeNewHead", mock.Anything, mock.Anything)
}

// TestWaitForBlocks_SubscriptionMidStreamErrorPropagates — subscription
// is established, then errors out before reaching the target. Function
// falls back to polling for the remainder. Pin that the path completes
// (rather than hanging or returning the sub-error directly).
func TestWaitForBlocks_SubscriptionMidStreamErrorPropagatesToFallback(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	origInterval := TimeToWaitForBlocks
	TimeToWaitForBlocks = 5 * time.Millisecond
	defer func() { TimeToWaitForBlocks = origInterval }()

	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{Client: mockClient, BlocksToWait: 1}
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Once()

	sub := newFakeSubscription()
	mockClient.On("SubscribeNewHead", mock.Anything, mock.Anything).Return(
		(ethereum.Subscription)(sub), nil)

	// Raise an error on the sub immediately — function should fall back to polling.
	go func() {
		time.Sleep(20 * time.Millisecond)
		sub.raise(assert.AnError)
	}()

	// Polling fallback: BlockNumber returns target.
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(101), nil)

	err := WaitForBlocks(cProps)
	assert.NoError(t, err)
}
