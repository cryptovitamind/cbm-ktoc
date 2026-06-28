package ktfunc

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// neverMines fakes a submitted transaction that never produces a receipt: it
// blocks until its context is cancelled, then reports the context error. This
// is exactly how bind.WaitMined behaves against a tx that was dropped from the
// mempool, underpriced, or stuck behind a nonce gap once a deadline fires. With
// context.Background() (no deadline) the real call would block forever.
func neverMines(ctx context.Context, _ bind.DeployBackend, _ *types.Transaction) (*types.Receipt, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

// runWithWatchdog runs fn in a goroutine and fails the test if it does not
// return within limit. A hang here means the node would have parked an operator
// inside a transaction wait until they restarted it.
func runWithWatchdog(t *testing.T, limit time.Duration, fn func() error) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- fn() }()
	select {
	case err := <-done:
		return err
	case <-time.After(limit):
		t.Fatal("call hung waiting for a transaction that never mines; it must time out and return")
		return nil
	}
}

// stubTxSeams points the newTransactor and waitMined package seams at test
// doubles: a no-op transactor and a tx that never mines. Cleanup restores them.
// Used by the manual-op watchdog tests so they exercise the bounded wait without
// real signing or a real RPC backend.
func stubTxSeams(t *testing.T) {
	t.Helper()
	origNewTransactor := newTransactor
	origWaitMined := waitMined
	t.Cleanup(func() {
		newTransactor = origNewTransactor
		waitMined = origWaitMined
	})
	newTransactor = func(_ *ConnectionProps) (*bind.TransactOpts, error) {
		return &bind.TransactOpts{}, nil
	}
	waitMined = neverMines
}

func txTimeoutProps(t *testing.T) (*ConnectionProps, *MockEthClient, *MockKtv2) {
	t.Helper()
	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Client:        mockClient,
		Kt:            mockKt,
		KtAddr:        common.HexToAddress("0x1234567890123456789012345678901234567890"),
		MyPubKey:      common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		ChainID:       big.NewInt(1),
		BlocksToWait:  0,
		TxMineTimeout: 200 * time.Millisecond, // bound the wait so the test is fast
	}
	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = privateKey
	return cProps, mockClient, mockKt
}

func dummyTx() *types.Transaction {
	return types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), []byte{})
}

// --- The shared chokepoint: every transaction wait now funnels through
// waitForTxMined, so these tests are the core guarantee that no call site can
// hang. ---

// TestWaitForTxMined_ReturnsErrorOnTimeout proves the bounded wait gives up and
// returns an error instead of polling forever when a tx never mines.
func TestWaitForTxMined_ReturnsErrorOnTimeout(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	originalWaitMined := waitMined
	defer func() { waitMined = originalWaitMined }()
	waitMined = neverMines

	cProps := &ConnectionProps{Client: &MockEthClient{}, TxMineTimeout: 150 * time.Millisecond}

	type result struct {
		receipt *types.Receipt
		err     error
	}
	done := make(chan result, 1)
	go func() {
		r, err := waitForTxMined(cProps, dummyTx())
		done <- result{r, err}
	}()
	select {
	case got := <-done:
		assert.Error(t, got.err, "must return an error when the tx never mines")
		assert.Nil(t, got.receipt)
	case <-time.After(5 * time.Second):
		t.Fatal("waitForTxMined hung; it must honor TxMineTimeout")
	}
}

// TestWaitForTxMined_ReturnsReceiptWhenMined confirms the normal path is
// unchanged: a mined tx yields its receipt with no error.
func TestWaitForTxMined_ReturnsReceiptWhenMined(t *testing.T) {
	originalWaitMined := waitMined
	defer func() { waitMined = originalWaitMined }()
	receipt := &types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(7)}
	waitMined = func(_ context.Context, _ bind.DeployBackend, _ *types.Transaction) (*types.Receipt, error) {
		return receipt, nil
	}

	cProps := &ConnectionProps{Client: &MockEthClient{}, TxMineTimeout: time.Minute}
	got, err := waitForTxMined(cProps, dummyTx())
	assert.NoError(t, err)
	assert.Equal(t, receipt, got)
}

// TestWaitForTxMined_ZeroTimeoutUsesDefault proves a node that never sets
// TxMineTimeout still gets a bounded wait via DefaultTxMineTimeout, not an
// infinite one.
func TestWaitForTxMined_ZeroTimeoutUsesDefault(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	originalWaitMined := waitMined
	originalDefault := DefaultTxMineTimeout
	defer func() {
		waitMined = originalWaitMined
		DefaultTxMineTimeout = originalDefault
	}()
	DefaultTxMineTimeout = 150 * time.Millisecond // shrink the fallback so the test is fast
	waitMined = neverMines

	cProps := &ConnectionProps{Client: &MockEthClient{}} // TxMineTimeout left at zero

	done := make(chan error, 1)
	go func() {
		_, err := waitForTxMined(cProps, dummyTx())
		done <- err
	}()
	select {
	case err := <-done:
		assert.Error(t, err, "zero TxMineTimeout must fall back to DefaultTxMineTimeout, not wait forever")
	case <-time.After(5 * time.Second):
		t.Fatal("waitForTxMined hung with zero TxMineTimeout; the default bound did not apply")
	}
}

// TestWaitForTxMined_PropagatesNonTimeoutError confirms a genuine RPC error is
// surfaced unchanged rather than masked as a timeout.
func TestWaitForTxMined_PropagatesNonTimeoutError(t *testing.T) {
	originalWaitMined := waitMined
	defer func() { waitMined = originalWaitMined }()
	sentinel := errors.New("rpc backend exploded")
	waitMined = func(_ context.Context, _ bind.DeployBackend, _ *types.Transaction) (*types.Receipt, error) {
		return nil, sentinel
	}

	cProps := &ConnectionProps{Client: &MockEthClient{}, TxMineTimeout: time.Minute}
	_, err := waitForTxMined(cProps, dummyTx())
	assert.ErrorIs(t, err, sentinel)
}

// --- Loop-path watchdogs: the two calls that hung in production. ---

// TestVote_ReturnsErrorWhenTxNeverMines proves the consensus loop does not hang
// after "Winner selected" when the vote transaction never mines. This is the
// production bug: operators saw the node stop at "Winner selected" and had to
// restart it. vote() must give up after TxMineTimeout and return an error so
// KeepRunning re-enters and re-submits.
func TestVote_ReturnsErrorWhenTxNeverMines(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	originalWaitMined := waitMined
	defer func() { waitMined = originalWaitMined }()
	waitMined = neverMines

	cProps, _, mockKt := txTimeoutProps(t)

	winner := common.HexToAddress("0xabc1230000000000000000000000000000000000")
	voteTx := types.NewTransaction(0, winner, big.NewInt(0), 0, big.NewInt(0), []byte{})
	mockKt.On("Vote", mock.Anything, winner, "seed").Return(voteTx, nil)

	err := runWithWatchdog(t, 5*time.Second, func() error {
		return vote(cProps, winner, "seed")
	})
	assert.Error(t, err, "vote must return an error when the tx never mines, not hang")
}

// TestRewardWinningWallet_ReturnsErrorWhenTxNeverMines proves the same for the
// reward leg of the loop, which carries the identical unbounded wait.
func TestRewardWinningWallet_ReturnsErrorWhenTxNeverMines(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)

	originalWaitMined := waitMined
	defer func() { waitMined = originalWaitMined }()
	waitMined = neverMines

	cProps, mockClient, mockKt := txTimeoutProps(t)

	winner := common.HexToAddress("0xabc1230000000000000000000000000000000000")
	totalMin := big.NewInt(1000)

	// Balances and fees needed before the reward tx is sent.
	mockClient.On("BalanceAt", mock.Anything, winner, (*big.Int)(nil)).Return(big.NewInt(0), nil)
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, (*big.Int)(nil)).Return(big.NewInt(1000000000000000000), nil)
	mockKt.On("TlOcFees", mock.Anything).Return(big.NewInt(0), nil)

	rewardTx := types.NewTransaction(0, winner, big.NewInt(0), 0, big.NewInt(0), []byte{})
	mockKt.On("Rwd", mock.Anything, winner, mock.Anything).Return(rewardTx, nil)

	err := runWithWatchdog(t, 5*time.Second, func() error {
		return rewardWinningWallet(cProps, winner, totalMin)
	})
	assert.Error(t, err, "rewardWinningWallet must return an error when the tx never mines, not hang")
}

// --- Manual-op watchdogs: the operator CLI paths that shared the same
// unbounded wait. They funnel through waitForTxMined too, so none can hang. ---

func TestVoteToRemove_ReturnsErrorWhenTxNeverMines(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	stubTxSeams(t)

	cProps, _, mockKt := txTimeoutProps(t)
	target := common.HexToAddress("0xbbb1230000000000000000000000000000000000")
	mockKt.On("VoteToRemove", mock.Anything, target, "data").Return(dummyTx(), nil)

	err := runWithWatchdog(t, 5*time.Second, func() error {
		return VoteToRemove(cProps, target, "data")
	})
	assert.Error(t, err)
}

func TestVoteToAdd_ReturnsErrorWhenTxNeverMines(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	stubTxSeams(t)

	cProps, mockClient, mockKt := txTimeoutProps(t)
	target := common.HexToAddress("0xbbb1230000000000000000000000000000000000")

	// Pre-checks must pass so execution reaches the bounded wait.
	mockClient.On("BalanceAt", mock.Anything, cProps.MyPubKey, (*big.Int)(nil)).Return(big.NewInt(1000000000000000000), nil)
	mockKt.On("OcRwdrs", mock.Anything, target).Return(false, nil)
	mockKt.On("HasVotedAdd", mock.Anything, cProps.MyPubKey, target).Return(false, nil)
	mockClient.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20000000000), nil)
	mockKt.On("VoteToAdd", mock.Anything, target, "data").Return(dummyTx(), nil)

	err := runWithWatchdog(t, 5*time.Second, func() error {
		return VoteToAdd(cProps, target, "data")
	})
	assert.Error(t, err)
}

func TestResetVoteToAdd_ReturnsErrorWhenTxNeverMines(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	stubTxSeams(t)

	cProps, _, mockKt := txTimeoutProps(t)
	target := common.HexToAddress("0xbbb1230000000000000000000000000000000000")
	mockKt.On("HasVotedAdd", mock.Anything, cProps.MyPubKey, target).Return(true, nil)
	mockKt.On("ResetVoteToAdd", mock.Anything, target).Return(dummyTx(), nil)

	err := runWithWatchdog(t, 5*time.Second, func() error {
		return ResetVoteToAdd(cProps, target)
	})
	assert.Error(t, err)
}

func TestResetVoteToRemove_ReturnsErrorWhenTxNeverMines(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	stubTxSeams(t)

	cProps, _, mockKt := txTimeoutProps(t)
	target := common.HexToAddress("0xbbb1230000000000000000000000000000000000")
	mockKt.On("HasVotedRemove", mock.Anything, cProps.MyPubKey, target).Return(true, nil)
	mockKt.On("ResetVoteToRemove", mock.Anything, target).Return(dummyTx(), nil)

	err := runWithWatchdog(t, 5*time.Second, func() error {
		return ResetVoteToRemove(cProps, target)
	})
	assert.Error(t, err)
}
