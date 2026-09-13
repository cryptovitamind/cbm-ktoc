package ktfunc

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// callOptsAtBlock matches a CallOpts whose BlockNumber is exactly n, so a
// test can pin that a read was made against historic state, not latest.
func callOptsAtBlock(n uint64) interface{} {
	return mock.MatchedBy(func(o *bind.CallOpts) bool {
		return o != nil && o.BlockNumber != nil && o.BlockNumber.Uint64() == n
	})
}

// callOptsLatest matches a CallOpts with no BlockNumber (latest state).
func callOptsLatest() interface{} {
	return mock.MatchedBy(func(o *bind.CallOpts) bool { return o != nil && o.BlockNumber == nil })
}

func feeReserveSetup() (*ConnectionProps, *MockEthClient, *MockKtv2) {
	mockClient := &MockEthClient{}
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Client:   mockClient,
		Kt:       mockKt,
		KtAddr:   common.HexToAddress("0x1234567890123456789012345678901234567890"),
		MyPubKey: common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
		ChainID:  big.NewInt(1),
	}
	return cProps, mockClient, mockKt
}

func TestReadFeeReserve_WholeWhenBalanceCoversFees(t *testing.T) {
	cProps, mockClient, mockKt := feeReserveSetup()
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, (*big.Int)(nil)).Return(big.NewInt(1e18), nil)
	mockKt.On("TlOcFees", callOptsLatest()).Return(big.NewInt(4e17), nil)

	r, err := ReadFeeReserve(cProps, nil)
	if err != nil {
		t.Fatalf("ReadFeeReserve: %v", err)
	}
	assert.Equal(t, 0, r.Balance.Cmp(big.NewInt(1e18)), "balance")
	assert.Equal(t, 0, r.TlOcFees.Cmp(big.NewInt(4e17)), "tlOcFees")
	assert.Equal(t, 0, r.Reserve.Cmp(big.NewInt(6e17)), "reserve = balance - tlOcFees")
	assert.False(t, r.Underfunded())
	assert.Equal(t, 0, r.Deficit().Sign(), "no deficit when whole")
	assert.Equal(t, 0, r.Pot().Cmp(big.NewInt(6e17)), "pot is the positive reserve")
	assert.Contains(t, r.String(), FeeReserveOKLabel)
	assert.NotContains(t, r.String(), FeeReserveUnderfundedLabel)
}

// The SHI numbers from 2026-09-10: 0.001621 ETH balance against 0.004851 ETH
// of fee claims. The pot must clamp to zero and the deficit be exact.
func TestReadFeeReserve_UnderfundedWhenFeesExceedBalance(t *testing.T) {
	cProps, mockClient, mockKt := feeReserveSetup()
	balance, _ := new(big.Int).SetString("1621000000000000", 10)
	owed, _ := new(big.Int).SetString("4851000000000000", 10)
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, (*big.Int)(nil)).Return(balance, nil)
	mockKt.On("TlOcFees", callOptsLatest()).Return(owed, nil)

	r, err := ReadFeeReserve(cProps, nil)
	if err != nil {
		t.Fatalf("ReadFeeReserve: %v", err)
	}
	wantDeficit, _ := new(big.Int).SetString("3230000000000000", 10)
	assert.True(t, r.Underfunded())
	assert.Equal(t, 0, r.Deficit().Cmp(wantDeficit), "deficit = tlOcFees - balance")
	assert.Equal(t, 0, r.Pot().Sign(), "pot clamps to zero")
	assert.Contains(t, r.String(), FeeReserveUnderfundedLabel)
	assert.Contains(t, r.String(), "0.003230", "deficit is printed in ETH")
}

func TestReadFeeReserve_ReadsStateAtRequestedBlock(t *testing.T) {
	cProps, mockClient, mockKt := feeReserveSetup()
	at := big.NewInt(114)
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, at).Return(big.NewInt(5e18), nil)
	mockKt.On("TlOcFees", callOptsAtBlock(114)).Return(big.NewInt(1e18), nil)

	r, err := ReadFeeReserve(cProps, at)
	if err != nil {
		t.Fatalf("ReadFeeReserve: %v", err)
	}
	assert.Equal(t, 0, r.Reserve.Cmp(big.NewInt(4e18)))
	mockClient.AssertExpectations(t)
	mockKt.AssertExpectations(t)
}

func TestReadFeeReserve_PropagatesErrors(t *testing.T) {
	cProps, mockClient, mockKt := feeReserveSetup()
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, (*big.Int)(nil)).
		Return((*big.Int)(nil), errors.New("missing trie node")).Once()
	if _, err := ReadFeeReserve(cProps, nil); err == nil {
		t.Error("expected balance error to propagate")
	}

	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, (*big.Int)(nil)).Return(big.NewInt(1), nil)
	mockKt.On("TlOcFees", callOptsLatest()).Return((*big.Int)(nil), errors.New("rpc: timeout"))
	if _, err := ReadFeeReserve(cProps, nil); err == nil {
		t.Error("expected tlOcFees error to propagate")
	}
}

// EstimateFeeClaim must reproduce withdrawOCFee's arithmetic: migrateFees
// folds ocFees[oc][lastStartBlock] into pastOcFees, then the current epoch's
// fee is added once the epoch is complete, and the whole sum is paid
// all-or-nothing.

func claimSetup(t *testing.T, head uint64, lastStart int64) (*ConnectionProps, *MockKtv2, common.Address) {
	t.Helper()
	cProps, mockClient, mockKt := feeReserveSetup()
	oc := common.HexToAddress("0xf25d4199ed3ca881bfaf3a0801c4028cfc47359e")
	mockKt.On("StartBlock", mock.Anything).Return(big.NewInt(100), nil)
	mockKt.On("EpochInterval", mock.Anything).Return(uint16(10), nil)
	mockClient.On("BlockNumber", mock.Anything).Return(head, nil)
	mockKt.On("LastStartBlock", mock.Anything, oc).Return(big.NewInt(lastStart), nil)
	mockKt.On("PastOcFees", mock.Anything, oc).Return(big.NewInt(5), nil)
	mockKt.On("OcFees", mock.Anything, oc, big.NewInt(90)).Return(big.NewInt(4), nil).Maybe()
	mockKt.On("OcFees", mock.Anything, oc, big.NewInt(100)).Return(big.NewInt(3), nil).Maybe()
	return cProps, mockKt, oc
}

func TestEstimateFeeClaim_SumsPastUnmigratedAndCurrentEpoch(t *testing.T) {
	cProps, _, oc := claimSetup(t, 200, 90) // epoch 100..110 complete; last tx under epoch 90

	c, err := EstimateFeeClaim(cProps, oc)
	if err != nil {
		t.Fatalf("EstimateFeeClaim: %v", err)
	}
	assert.Equal(t, int64(5), c.Past.Int64())
	assert.Equal(t, int64(4), c.Unmigrated.Int64(), "ocFees under lastStartBlock migrates on the next call")
	assert.Equal(t, int64(3), c.CurrentEpoch.Int64(), "complete epoch's fee is paid too")
	assert.Equal(t, int64(12), c.Total.Int64())
}

func TestEstimateFeeClaim_CurrentEpochNotPaidUntilComplete(t *testing.T) {
	cProps, _, oc := claimSetup(t, 105, 90) // head inside epoch 100..110

	c, err := EstimateFeeClaim(cProps, oc)
	if err != nil {
		t.Fatalf("EstimateFeeClaim: %v", err)
	}
	assert.Equal(t, int64(0), c.CurrentEpoch.Int64())
	assert.Equal(t, int64(9), c.Total.Int64())
}

func TestEstimateFeeClaim_NothingToMigrateWhenLastStartIsCurrent(t *testing.T) {
	cProps, _, oc := claimSetup(t, 200, 100) // lastStartBlock == startBlock

	c, err := EstimateFeeClaim(cProps, oc)
	if err != nil {
		t.Fatalf("EstimateFeeClaim: %v", err)
	}
	assert.Equal(t, int64(0), c.Unmigrated.Int64(), "no migration when lastStartBlock is the current epoch")
	assert.Equal(t, int64(3), c.CurrentEpoch.Int64(), "current epoch fee counted exactly once")
	assert.Equal(t, int64(8), c.Total.Int64())
}

func TestEstimateFeeClaim_ZeroLastStartBlockMeansNeverTransacted(t *testing.T) {
	cProps, mockKt, oc := claimSetup(t, 200, 0)

	c, err := EstimateFeeClaim(cProps, oc)
	if err != nil {
		t.Fatalf("EstimateFeeClaim: %v", err)
	}
	assert.Equal(t, int64(0), c.Unmigrated.Int64())
	mockKt.AssertNotCalled(t, "OcFees", mock.Anything, oc, big.NewInt(0))
}
