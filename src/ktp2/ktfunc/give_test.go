package ktfunc

// Phase 5c — Give tests.
//
// Give sends ETH to the KT contract. Previously had no direct tests.
// One test below (TestGive_PrivateKeyArgIsActuallyUsed) is **expected
// to fail on master** — it documents a real latent bug found during the
// audit: the privateKey argument is checked for nil but otherwise
// unused. The function always signs with cProps.MyPrivateKey via
// NewTransactor(cProps). main.go's giveETH() loops through 10 test
// wallets passing each one's private key — but every call actually
// signs as the operator. This goes in as a separate TDD-style commit
// pair (this commit lands the failing test; a fix commit follows).

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func giveSetup(t *testing.T) (
	cProps *ConnectionProps,
	mockClient *MockEthClient,
	mockKt *MockKtv2,
) {
	t.Helper()
	logrus.SetLevel(logrus.FatalLevel)

	mockClient = &MockEthClient{}
	mockKt = &MockKtv2{}
	cProps = &ConnectionProps{
		Client:   mockClient,
		Kt:       mockKt,
		KtAddr:   common.HexToAddress("0x1234567890123456789012345678901234567890"),
		MyPubKey: common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		ChainID:  big.NewInt(1),
	}
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	cProps.MyPrivateKey = priv
	return
}

// TestGive_NilConnectionPropsErrors — passing nil cProps should return an
// error gracefully. **Fails on master** with a nil-pointer panic because
// the early-validation log line dereferences cProps.Client AFTER the
// short-circuited cProps==nil check (give.go:19-21). The error path is
// effectively a crash. Fix: guard the log against nil cProps.
func TestGive_NilConnectionPropsErrors(t *testing.T) {
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	assert.NotPanics(t, func() {
		err := Give(nil, priv, big.NewInt(1e18))
		assert.Error(t, err)
	})
}

func TestGive_NilPrivateKeyArgErrors(t *testing.T) {
	cProps, _, _ := giveSetup(t)
	err := Give(cProps, nil, big.NewInt(1e18))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "private key is nil")
}

func TestGive_NegativeAmountErrors(t *testing.T) {
	cProps, _, _ := giveSetup(t)
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	err := Give(cProps, priv, big.NewInt(-1))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-negative")
}

func TestGive_NilAmountErrors(t *testing.T) {
	cProps, _, _ := giveSetup(t)
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")
	err := Give(cProps, priv, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-nil")
}

func TestGive_NewTransactorFailsWhenChainIDIsNil(t *testing.T) {
	cProps, _, _ := giveSetup(t)
	cProps.ChainID = nil // forces bind.NewKeyedTransactorWithChainID to fail
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")

	err := Give(cProps, priv, big.NewInt(1e18))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create function")
}

func TestGive_KtGiveErrorPropagates(t *testing.T) {
	cProps, _, mockKt := giveSetup(t)
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")

	mockKt.On("Give", mock.Anything).Return((*types.Transaction)(nil), errors.New("revert: paused"))

	err := Give(cProps, priv, big.NewInt(1e18))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send give transaction")
}

func TestGive_TxReceiptFailedStatusReturnsError(t *testing.T) {
	cProps, mockClient, mockKt := giveSetup(t)
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")

	tx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 21000, big.NewInt(1e9), nil)
	mockKt.On("Give", mock.Anything).Return(tx, nil)
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(
		&types.Receipt{Status: types.ReceiptStatusFailed, BlockNumber: big.NewInt(100)}, nil)

	err := Give(cProps, priv, big.NewInt(1e18))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction failed")
}

func TestGive_Success(t *testing.T) {
	cProps, mockClient, mockKt := giveSetup(t)
	priv, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a")

	tx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 21000, big.NewInt(1e9), nil)
	mockKt.On("Give", mock.Anything).Return(tx, nil)
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(
		&types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(100)}, nil)

	err := Give(cProps, priv, big.NewInt(1e18))
	assert.NoError(t, err)
}

// TestGive_PrivateKeyArgIsActuallyUsed — Give takes a `privateKey` argument
// but never uses it beyond a nil check; NewTransactor(cProps) signs with
// cProps.MyPrivateKey. Callers in main.go's giveETH() loop pass different
// private keys per test wallet expecting Give to sign with that wallet's
// key, but the operator's key signs every tx instead.
//
// This test asserts the contract Give() is called with auth.From matching
// the address derived from the `privateKey` ARGUMENT, not from
// cProps.MyPrivateKey. **Fails on master** — surfacing this real bug.
func TestGive_PrivateKeyArgIsActuallyUsed(t *testing.T) {
	cProps, mockClient, mockKt := giveSetup(t)

	// Different private key than cProps.MyPrivateKey, with a known address.
	otherPriv, _ := crypto.HexToECDSA(testKeyX)
	otherAddr := crypto.PubkeyToAddress(otherPriv.PublicKey)
	operatorAddr := crypto.PubkeyToAddress(cProps.MyPrivateKey.PublicKey)

	// Sanity: the two addresses differ. (If they didn't, this test would
	// be vacuous.)
	assert.NotEqual(t, otherAddr, operatorAddr, "test keys must differ")

	tx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 21000, big.NewInt(1e9), nil)
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(
		&types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(100)}, nil)

	// Capture the auth.From that Kt.Give was actually called with.
	var capturedFrom common.Address
	mockKt.On("Give", mock.AnythingOfType("*bind.TransactOpts")).
		Run(func(args mock.Arguments) {
			opts := args.Get(0).(*bind.TransactOpts)
			capturedFrom = opts.From
		}).Return(tx, nil)

	err := Give(cProps, otherPriv, big.NewInt(1e18))
	assert.NoError(t, err)

	assert.Equal(t, otherAddr, capturedFrom,
		"Give should sign with the privateKey argument (%s), but signed with the operator key (%s) instead. "+
			"main.go's giveETH() loops over test wallets passing each one's privateKey expecting it to be honored.",
		otherAddr.Hex(), capturedFrom.Hex())
}
