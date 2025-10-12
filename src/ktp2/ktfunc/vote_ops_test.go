package ktfunc

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"ktp2/src/abis/ktv2"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKtv2 mocks the Ktv2 contract
type MockKtv2 struct {
	mock.Mock
}

// VoteToRemove mock
func (m *MockKtv2) VoteToRemove(opts *bind.TransactOpts, existingOC common.Address, data string) (*types.Transaction, error) {
	args := m.Called(opts, existingOC, data)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockKtv2) TotalOC(opts *bind.CallOpts) (uint16, error) {
	args := m.Called(opts)
	return args.Get(0).(uint16), args.Error(1)
}

// VoteToAdd mock
func (m *MockKtv2) VoteToAdd(opts *bind.TransactOpts, newOC common.Address, data string) (*types.Transaction, error) {
	args := m.Called(opts, newOC, data)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

// StartBlock mock
func (m *MockKtv2) StartBlock(opts *bind.CallOpts) (*big.Int, error) {
	args := m.Called(opts)
	return args.Get(0).(*big.Int), args.Error(1)
}

// EpochInterval mock
func (m *MockKtv2) EpochInterval(opts *bind.CallOpts) (uint16, error) {
	args := m.Called(opts)
	return uint16(args.Int(0)), args.Error(1)
}

// UserStks mock
func (m *MockKtv2) UserStks(opts *bind.CallOpts, address common.Address) (*big.Int, error) {
	args := m.Called(opts, address)
	return args.Get(0).(*big.Int), args.Error(1)
}

// Vote mock
func (m *MockKtv2) Vote(opts *bind.TransactOpts, recipient common.Address, data string) (*types.Transaction, error) {
	args := m.Called(opts, recipient, data)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

// Rwd mock
func (m *MockKtv2) Rwd(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	args := m.Called(opts, recipient, amount)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

// BlockRwd mock
func (m *MockKtv2) BlockRwd(opts *bind.CallOpts, blockNumber *big.Int, recipient common.Address) (uint16, error) {
	args := m.Called(opts, blockNumber, recipient)
	return uint16(args.Int(0)), args.Error(1)
}

// ConsensusReq mock
func (m *MockKtv2) ConsensusReq(opts *bind.CallOpts) (uint16, error) {
	args := m.Called(opts)
	return uint16(args.Int(0)), args.Error(1)
}

// FilterStaked mock
func (m *MockKtv2) FilterStaked(opts *bind.FilterOpts) (*ktv2.Ktv2StakedIterator, error) {
	args := m.Called(opts)
	return args.Get(0).(*ktv2.Ktv2StakedIterator), args.Error(1)
}

// FilterWithdrew mock
func (m *MockKtv2) FilterWithdrew(opts *bind.FilterOpts) (*ktv2.Ktv2WithdrewIterator, error) {
	args := m.Called(opts)
	return args.Get(0).(*ktv2.Ktv2WithdrewIterator), args.Error(1)
}

// Give mock
func (m *MockKtv2) Give(opts *bind.TransactOpts) (*types.Transaction, error) {
	args := m.Called(opts)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

// WithdrawOCFee mock
func (m *MockKtv2) WithdrawOCFee(opts *bind.TransactOpts, blocks []uint32) (*types.Transaction, error) {
	args := m.Called(opts, blocks)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

// ResetVoteToAdd mock
func (m *MockKtv2) ResetVoteToAdd(opts *bind.TransactOpts, newOC common.Address) (*types.Transaction, error) {
	args := m.Called(opts, newOC)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

// ResetVoteToRemove mock
func (m *MockKtv2) ResetVoteToRemove(opts *bind.TransactOpts, existingOC common.Address) (*types.Transaction, error) {
	args := m.Called(opts, existingOC)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

// SetEpochInterval mock
func (m *MockKtv2) SetEpochInterval(opts *bind.TransactOpts, newInterval uint16) (*types.Transaction, error) {
	args := m.Called(opts, newInterval)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

// TotalStk mock
func (m *MockKtv2) TotalStk(opts *bind.CallOpts) (*big.Int, error) {
	args := m.Called(opts)
	return args.Get(0).(*big.Int), args.Error(1)
}

// TotalGvn mock
func (m *MockKtv2) TotalGvn(opts *bind.CallOpts) (*big.Int, error) {
	args := m.Called(opts)
	return args.Get(0).(*big.Int), args.Error(1)
}

// TotalBurned mock
func (m *MockKtv2) TotalBurned(opts *bind.CallOpts) (*big.Int, error) {
	args := m.Called(opts)
	return args.Get(0).(*big.Int), args.Error(1)
}

// MaxBrnPrc mock
func (m *MockKtv2) MaxBrnPrc(opts *bind.CallOpts) (uint16, error) {
	args := m.Called(opts)
	return uint16(args.Int(0)), args.Error(1)
}

// DonationPrc mock
func (m *MockKtv2) DonationPrc(opts *bind.CallOpts) (uint16, error) {
	args := m.Called(opts)
	return uint16(args.Int(0)), args.Error(1)
}

// BurnFactor mock
func (m *MockKtv2) BurnFactor(opts *bind.CallOpts) (uint16, error) {
	args := m.Called(opts)
	return uint16(args.Int(0)), args.Error(1)
}

// V2 mock
func (m *MockKtv2) V2(opts *bind.CallOpts) (bool, error) {
	args := m.Called(opts)
	return args.Bool(0), args.Error(1)
}

// OcFee mock
func (m *MockKtv2) OcFee(opts *bind.CallOpts) (uint16, error) {
	args := m.Called(opts)
	return uint16(args.Int(0)), args.Error(1)
}

// OcFees mock
func (m *MockKtv2) OcFees(opts *bind.CallOpts, oc common.Address, blockNumber *big.Int) (*big.Int, error) {
	args := m.Called(opts, oc, blockNumber)
	return args.Get(0).(*big.Int), args.Error(1)
}

// TlOcFees mock
func (m *MockKtv2) TlOcFees(opts *bind.CallOpts) (*big.Int, error) {
	args := m.Called(opts)
	return args.Get(0).(*big.Int), args.Error(1)
}

// Stake mock
func (m *MockKtv2) Stake(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	args := m.Called(opts, amount)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

// Withdraw mock
func (m *MockKtv2) Withdraw(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	args := m.Called(opts, amount)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

// FilterRwd mock
func (m *MockKtv2) FilterRwd(opts *bind.FilterOpts) (*ktv2.Ktv2RwdIterator, error) {
	args := m.Called(opts)
	return args.Get(0).(*ktv2.Ktv2RwdIterator), args.Error(1)
}

// FilterVoted mock
func (m *MockKtv2) FilterVoted(opts *bind.FilterOpts) (*ktv2.Ktv2VotedIterator, error) {
	args := m.Called(opts)
	return args.Get(0).(*ktv2.Ktv2VotedIterator), args.Error(1)
}

// OcRwdrs mock
func (m *MockKtv2) OcRwdrs(opts *bind.CallOpts, oc common.Address) (bool, error) {
	args := m.Called(opts, oc)
	return args.Bool(0), args.Error(1)
}

// HasVoted mock
func (m *MockKtv2) HasVoted(opts *bind.CallOpts, voter common.Address, oc common.Address) (bool, error) {
	args := m.Called(opts, voter, oc)
	return args.Bool(0), args.Error(1)
}

// MockEthClient for ethclient
type MockEthClient struct {
	mock.Mock
}

// TransactionReceipt mock (used internally by bind.WaitMined)
func (m *MockEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	args := m.Called(ctx, txHash)
	return args.Get(0).(*types.Receipt), args.Error(1)
}

// CodeAt mock
func (m *MockEthClient) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, account, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}

// HeaderByNumber mock
func (m *MockEthClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	args := m.Called(ctx, number)
	return args.Get(0).(*types.Header), args.Error(1)
}

// BalanceAt mock
func (m *MockEthClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	args := m.Called(ctx, account, blockNumber)
	return args.Get(0).(*big.Int), args.Error(1)
}

// PendingNonceAt mock
func (m *MockEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	args := m.Called(ctx, account)
	return uint64(args.Int(0)), args.Error(1)
}

// BlockNumber mock
func (m *MockEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return uint64(args.Int(0)), args.Error(1)
}

// SuggestGasPrice mock
func (m *MockEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

// SendTransaction mock
func (m *MockEthClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

// TransactionByHash mock
func (m *MockEthClient) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(*types.Transaction), args.Bool(1), args.Error(2)
}

// FilterLogs mock
func (m *MockEthClient) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	return []types.Log{}, nil
}

// SubscribeFilterLogs mock
func (m *MockEthClient) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	return nil, nil
}

// TestValidateAddress tests the address validation helper.
func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name    string
		addrStr string
		want    common.Address
		err     error
	}{
		{
			name:    "valid address",
			addrStr: "0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d",
			want:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d"),
			err:     nil,
		},
		{
			name:    "invalid address",
			addrStr: "invalid",
			want:    common.Address{},
			err:     errors.New("Invalid Ethereum address length: invalid (must be 42 characters). Example: 0x22...822"),
		},
		{
			name:    "zero address",
			addrStr: "0x0000000000000000000000000000000000000000",
			want:    common.Address{},
			err:     errors.New("Invalid Ethereum address: 0x0000000000000000000000000000000000000000"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAddress(tt.addrStr)
			if tt.err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.err.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestVoteToAdd_Success tests successful VoteToAdd execution.
func TestVoteToAdd_Success(t *testing.T) {
	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{
		Kt:           mockKt,
		Client:       mockClient,
		BlocksToWait: 5,
	}
	myPubKey := common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d")
	cProps.MyPubKey = myPubKey
	targetAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	data := "test data"
	mockTx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), []byte{})
	mockReceipt := &types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(100)}

	// Save originals and defer restore
	originalNewTransactor := newTransactor
	originalWaitMined := waitMined
	originalWaitForBlocks := waitForBlocks
	defer func() {
		newTransactor = originalNewTransactor
		waitMined = originalWaitMined
		waitForBlocks = originalWaitForBlocks
	}()

	// Mock dependencies
	newTransactor = func(_ *ConnectionProps) (*bind.TransactOpts, error) {
		return &bind.TransactOpts{}, nil
	}
	waitMined = func(_ context.Context, _ bind.DeployBackend, tx *types.Transaction) (*types.Receipt, error) {
		return mockReceipt, nil
	}
	waitForBlocks = func(_ *ConnectionProps) error {
		return nil
	}

	mockClient.On("BalanceAt", mock.Anything, myPubKey, (*big.Int)(nil)).Return(big.NewInt(1000000000000000000), nil)
	mockKt.On("OcRwdrs", mock.Anything, targetAddr).Return(false, nil)
	mockKt.On("HasVoted", mock.Anything, myPubKey, targetAddr).Return(false, nil)
	mockClient.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20000000000), nil)

	mockKt.On("VoteToAdd", mock.Anything, targetAddr, data).Return(mockTx, nil)

	err := VoteToAdd(cProps, targetAddr, data)
	assert.NoError(t, err)
	mockKt.AssertExpectations(t)
}

// TestVoteToAdd_Revert tests VoteToAdd when transaction reverts.
func TestVoteToAdd_Revert(t *testing.T) {
	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{
		Kt:     mockKt,
		Client: mockClient,
	}
	myPubKey := common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d")
	cProps.MyPubKey = myPubKey
	targetAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	data := "test data"
	mockTx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), []byte{})
	mockReceipt := &types.Receipt{Status: types.ReceiptStatusFailed}

	// Save originals and defer restore
	originalNewTransactor := newTransactor
	originalWaitMined := waitMined
	defer func() {
		newTransactor = originalNewTransactor
		waitMined = originalWaitMined
	}()

	// Mock dependencies
	newTransactor = func(_ *ConnectionProps) (*bind.TransactOpts, error) {
		return &bind.TransactOpts{}, nil
	}
	waitMined = func(_ context.Context, _ bind.DeployBackend, tx *types.Transaction) (*types.Receipt, error) {
		return mockReceipt, nil
	}

	mockClient.On("BalanceAt", mock.Anything, myPubKey, (*big.Int)(nil)).Return(big.NewInt(1000000000000000000), nil)
	mockKt.On("OcRwdrs", mock.Anything, targetAddr).Return(false, nil)
	mockKt.On("HasVoted", mock.Anything, myPubKey, targetAddr).Return(false, nil)
	mockClient.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20000000000), nil)

	mockKt.On("VoteToAdd", mock.Anything, targetAddr, data).Return(mockTx, nil)

	err := VoteToAdd(cProps, targetAddr, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reverted with status 0")
	mockKt.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// TestVoteToAdd_ErrorInTransactor tests VoteToAdd when NewTransactor fails.
func TestVoteToAdd_ErrorInTransactor(t *testing.T) {
	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{
		Kt:     mockKt,
		Client: mockClient,
	}
	myPubKey := common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d")
	cProps.MyPubKey = myPubKey
	targetAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	data := "test data"

	// Save original and defer restore
	originalNewTransactor := newTransactor
	defer func() { newTransactor = originalNewTransactor }()

	// Mock pre-checks
	mockClient.On("BalanceAt", mock.Anything, myPubKey, (*big.Int)(nil)).Return(big.NewInt(1000000000000000000), nil)
	mockKt.On("OcRwdrs", mock.Anything, targetAddr).Return(false, nil)
	mockKt.On("HasVoted", mock.Anything, myPubKey, targetAddr).Return(false, nil)
	mockClient.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20000000000), nil)

	// Mock NewTransactor to fail
	newTransactor = func(_ *ConnectionProps) (*bind.TransactOpts, error) {
		return nil, errors.New("transactor error")
	}

	err := VoteToAdd(cProps, targetAddr, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create transactor")
	mockKt.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// Add similar tests for VoteToRemove...

// TestVoteToRemove_Success tests successful VoteToRemove execution.
func TestVoteToRemove_Success(t *testing.T) {
	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{
		Kt:           mockKt,
		Client:       mockClient,
		BlocksToWait: 5,
	}
	targetAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	data := "test data"
	mockTx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), []byte{})
	mockReceipt := &types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(100)}

	// Save originals and defer restore
	originalNewTransactor := newTransactor
	originalWaitMined := waitMined
	originalWaitForBlocks := waitForBlocks
	defer func() {
		newTransactor = originalNewTransactor
		waitMined = originalWaitMined
		waitForBlocks = originalWaitForBlocks
	}()

	// Mock dependencies
	newTransactor = func(_ *ConnectionProps) (*bind.TransactOpts, error) {
		return &bind.TransactOpts{}, nil
	}
	waitMined = func(_ context.Context, _ bind.DeployBackend, tx *types.Transaction) (*types.Receipt, error) {
		return mockReceipt, nil
	}
	waitForBlocks = func(_ *ConnectionProps) error {
		return nil
	}

	mockKt.On("VoteToRemove", mock.Anything, targetAddr, data).Return(mockTx, nil)

	err := VoteToRemove(cProps, targetAddr, data)
	assert.NoError(t, err)
	mockKt.AssertExpectations(t)
}

// TestVoteToRemove_Revert tests VoteToRemove when transaction reverts.
func TestVoteToRemove_Revert(t *testing.T) {
	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{
		Kt:     mockKt,
		Client: mockClient,
	}
	targetAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	data := "test data"
	mockTx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), []byte{})
	mockReceipt := &types.Receipt{Status: types.ReceiptStatusFailed}

	// Save originals and defer restore
	originalNewTransactor := newTransactor
	originalWaitMined := waitMined
	defer func() {
		newTransactor = originalNewTransactor
		waitMined = originalWaitMined
	}()

	// Mock dependencies
	newTransactor = func(_ *ConnectionProps) (*bind.TransactOpts, error) {
		return &bind.TransactOpts{}, nil
	}
	waitMined = func(_ context.Context, _ bind.DeployBackend, tx *types.Transaction) (*types.Receipt, error) {
		return mockReceipt, nil
	}

	mockKt.On("VoteToRemove", mock.Anything, targetAddr, data).Return(mockTx, nil)

	err := VoteToRemove(cProps, targetAddr, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reverted with status 0")
	mockKt.AssertExpectations(t)
}

// TestVoteToRemove_ErrorInTransactor tests VoteToRemove when NewTransactor fails.
func TestVoteToRemove_ErrorInTransactor(t *testing.T) {
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Kt: mockKt,
	}
	targetAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	data := "test data"

	// Save original and defer restore
	originalNewTransactor := newTransactor
	defer func() { newTransactor = originalNewTransactor }()

	// Mock NewTransactor to fail
	newTransactor = func(_ *ConnectionProps) (*bind.TransactOpts, error) {
		return nil, errors.New("transactor error")
	}

	err := VoteToRemove(cProps, targetAddr, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create transactor")
}

// TestVoteToAdd_ErrorInVoteCall tests VoteToAdd when contract call fails.
func TestVoteToAdd_ErrorInVoteCall(t *testing.T) {
	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{
		Kt:     mockKt,
		Client: mockClient,
	}
	myPubKey := common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d")
	cProps.MyPubKey = myPubKey
	targetAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	data := "test data"

	// Save original and defer restore
	originalNewTransactor := newTransactor
	defer func() { newTransactor = originalNewTransactor }()

	// Mock pre-checks
	mockClient.On("BalanceAt", mock.Anything, myPubKey, (*big.Int)(nil)).Return(big.NewInt(1000000000000000000), nil)
	mockKt.On("OcRwdrs", mock.Anything, targetAddr).Return(false, nil)
	mockKt.On("HasVoted", mock.Anything, myPubKey, targetAddr).Return(false, nil)
	mockClient.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20000000000), nil)

	newTransactor = func(_ *ConnectionProps) (*bind.TransactOpts, error) {
		return &bind.TransactOpts{}, nil
	}

	mockKt.On("VoteToAdd", mock.Anything, targetAddr, data).Return((*types.Transaction)(nil), errors.New("contract error"))

	err := VoteToAdd(cProps, targetAddr, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send vote to add transaction")
	mockKt.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// TestVoteToRemove_ErrorInVoteCall tests VoteToRemove when contract call fails.
func TestVoteToRemove_ErrorInVoteCall(t *testing.T) {
	mockKt := &MockKtv2{}
	cProps := &ConnectionProps{
		Kt: mockKt,
	}
	targetAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	data := "test data"

	// Save original and defer restore
	originalNewTransactor := newTransactor
	defer func() { newTransactor = originalNewTransactor }()

	newTransactor = func(_ *ConnectionProps) (*bind.TransactOpts, error) {
		return &bind.TransactOpts{}, nil
	}

	mockKt.On("VoteToRemove", mock.Anything, targetAddr, data).Return((*types.Transaction)(nil), errors.New("contract error"))

	err := VoteToRemove(cProps, targetAddr, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send vote to remove transaction")
	mockKt.AssertExpectations(t)
}

// TestVoteToAdd_ErrorInWaitMined tests VoteToAdd when waitMined fails.
func TestVoteToAdd_ErrorInWaitMined(t *testing.T) {
	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{
		Kt:     mockKt,
		Client: mockClient,
	}
	myPubKey := common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d")
	cProps.MyPubKey = myPubKey
	targetAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	data := "test data"
	mockTx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), []byte{})

	// Save originals and defer restore
	originalNewTransactor := newTransactor
	originalWaitMined := waitMined
	defer func() {
		newTransactor = originalNewTransactor
		waitMined = originalWaitMined
	}()

	// Mock pre-checks
	mockClient.On("BalanceAt", mock.Anything, myPubKey, (*big.Int)(nil)).Return(big.NewInt(1000000000000000000), nil)
	mockKt.On("OcRwdrs", mock.Anything, targetAddr).Return(false, nil)
	mockKt.On("HasVoted", mock.Anything, myPubKey, targetAddr).Return(false, nil)
	mockClient.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20000000000), nil)

	newTransactor = func(_ *ConnectionProps) (*bind.TransactOpts, error) {
		return &bind.TransactOpts{}, nil
	}
	waitMined = func(_ context.Context, _ bind.DeployBackend, tx *types.Transaction) (*types.Receipt, error) {
		return nil, errors.New("mining error")
	}

	mockKt.On("VoteToAdd", mock.Anything, targetAddr, data).Return(mockTx, nil)

	err := VoteToAdd(cProps, targetAddr, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to wait for vote to add transaction to be mined")
	mockKt.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// TestVoteToRemove_ErrorInWaitMined tests VoteToRemove when waitMined fails.
func TestVoteToRemove_ErrorInWaitMined(t *testing.T) {
	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{
		Kt:     mockKt,
		Client: mockClient,
	}
	targetAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	data := "test data"
	mockTx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), []byte{})

	// Save originals and defer restore
	originalNewTransactor := newTransactor
	originalWaitMined := waitMined
	defer func() {
		newTransactor = originalNewTransactor
		waitMined = originalWaitMined
	}()

	newTransactor = func(_ *ConnectionProps) (*bind.TransactOpts, error) {
		return &bind.TransactOpts{}, nil
	}
	waitMined = func(_ context.Context, _ bind.DeployBackend, tx *types.Transaction) (*types.Receipt, error) {
		return nil, errors.New("mining error")
	}

	mockKt.On("VoteToRemove", mock.Anything, targetAddr, data).Return(mockTx, nil)

	err := VoteToRemove(cProps, targetAddr, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to wait for vote to remove transaction to be mined")
	mockKt.AssertExpectations(t)
}

// TestVoteToAdd_ErrorInWaitForBlocks tests VoteToAdd when WaitForBlocks fails.
func TestVoteToAdd_ErrorInWaitForBlocks(t *testing.T) {
	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{
		Kt:     mockKt,
		Client: mockClient,
	}
	myPubKey := common.HexToAddress("0x742d35Cc6634C0532925a3b8D3fE0e9C6e776d3d")
	cProps.MyPubKey = myPubKey
	targetAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	data := "test data"
	mockTx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), []byte{})
	mockReceipt := &types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(100)}

	// Save originals and defer restore
	originalNewTransactor := newTransactor
	originalWaitMined := waitMined
	originalWaitForBlocks := waitForBlocks
	defer func() {
		newTransactor = originalNewTransactor
		waitMined = originalWaitMined
		waitForBlocks = originalWaitForBlocks
	}()

	// Mock pre-checks
	mockClient.On("BalanceAt", mock.Anything, myPubKey, (*big.Int)(nil)).Return(big.NewInt(1000000000000000000), nil)
	mockKt.On("OcRwdrs", mock.Anything, targetAddr).Return(false, nil)
	mockKt.On("HasVoted", mock.Anything, myPubKey, targetAddr).Return(false, nil)
	mockClient.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20000000000), nil)

	newTransactor = func(_ *ConnectionProps) (*bind.TransactOpts, error) {
		return &bind.TransactOpts{}, nil
	}
	waitMined = func(_ context.Context, _ bind.DeployBackend, tx *types.Transaction) (*types.Receipt, error) {
		return mockReceipt, nil
	}
	waitForBlocks = func(_ *ConnectionProps) error {
		return errors.New("wait blocks error")
	}

	mockKt.On("VoteToAdd", mock.Anything, targetAddr, data).Return(mockTx, nil)

	err := VoteToAdd(cProps, targetAddr, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to wait for additional blocks")
	mockKt.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// TestVoteToRemove_ErrorInWaitForBlocks tests VoteToRemove when WaitForBlocks fails.
func TestVoteToRemove_ErrorInWaitForBlocks(t *testing.T) {
	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	cProps := &ConnectionProps{
		Kt:     mockKt,
		Client: mockClient,
	}
	targetAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	data := "test data"
	mockTx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), []byte{})
	mockReceipt := &types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(100)}

	// Save originals and defer restore
	originalNewTransactor := newTransactor
	originalWaitMined := waitMined
	originalWaitForBlocks := waitForBlocks
	defer func() {
		newTransactor = originalNewTransactor
		waitMined = originalWaitMined
		waitForBlocks = originalWaitForBlocks
	}()

	newTransactor = func(_ *ConnectionProps) (*bind.TransactOpts, error) {
		return &bind.TransactOpts{}, nil
	}
	waitMined = func(_ context.Context, _ bind.DeployBackend, tx *types.Transaction) (*types.Receipt, error) {
		return mockReceipt, nil
	}
	waitForBlocks = func(_ *ConnectionProps) error {
		return errors.New("wait blocks error")
	}

	mockKt.On("VoteToRemove", mock.Anything, targetAddr, data).Return(mockTx, nil)

	err := VoteToRemove(cProps, targetAddr, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to wait for additional blocks")
	mockKt.AssertExpectations(t)
}
