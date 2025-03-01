package tests

import (
	"context"
	"fmt"
	"ktp2/src/ktp2/ktfunc"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

// MockEthClient is a mock implementation of ethclient.Client for testing.
type MockEthClient struct {
	blockNumber   uint64
	nonce         uint64
	gasPrice      *big.Int
	receipts      map[common.Hash]*types.Receipt
	txErrors      map[common.Hash]error
	sendTxError   error
	blockNumError error
}

func (m *MockEthClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return big.NewInt(1000000000000000000), nil
}

func (m *MockEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return m.nonce, nil
}

func (m *MockEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	if m.gasPrice == nil {
		return nil, fmt.Errorf("gas price unavailable")
	}
	return m.gasPrice, nil
}

func (m *MockEthClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	if m.sendTxError != nil {
		return m.sendTxError
	}
	hash := tx.Hash()
	m.receipts[hash] = &types.Receipt{
		Status:      1, // Success
		BlockNumber: big.NewInt(int64(m.blockNumber)),
		TxHash:      hash,
		GasUsed:     21000,
	}
	m.blockNumber++
	return nil
}

func (m *MockEthClient) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	return []byte{1, 2, 3}, nil
}

func (m *MockEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	if err, exists := m.txErrors[txHash]; exists {
		return nil, err
	}
	if receipt, exists := m.receipts[txHash]; exists {
		return receipt, nil
	}
	return nil, ethereum.NotFound
}

func (m *MockEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	if m.blockNumError != nil {
		return nil, m.blockNumError
	}

	// Create a minimal Header
	header := &types.Header{
		Number:      number,                                                                                 // Set the block number
		Difficulty:  big.NewInt(0),                                                                          // Required field
		Time:        uint64(time.Now().Unix()),                                                              // Dummy timestamp
		ParentHash:  common.Hash{},                                                                          // Zero hash as placeholder
		UncleHash:   common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"), // Empty uncles hash
		Coinbase:    common.Address{},                                                                       // Zero address
		Root:        common.Hash{},                                                                          // Zero hash
		TxHash:      common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"), // Empty txs hash
		ReceiptHash: common.Hash{},                                                                          // Zero hash
	}

	// Use types.NewBlockWithHeader to create a Block
	block := types.NewBlockWithHeader(header)
	return block, nil
}

func (m *MockEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	if m.blockNumError != nil {
		return 0, m.blockNumError
	}
	return m.blockNumber, nil
}

// TestParseStartEndBlocks is included for completeness, assuming it’s in another file but not here.
func TestMoveBlocksForward(t *testing.T) {
	privKey, _ := crypto.GenerateKey()
	cProps := &ktfunc.ConnectionProps{
		Client: &MockEthClient{
			blockNumber: 1000,
			nonce:       0,
			gasPrice:    big.NewInt(1000000000), // 1 Gwei
			receipts:    make(map[common.Hash]*types.Receipt),
		},
		MyPubKey:     crypto.PubkeyToAddress(privKey.PublicKey),
		MyPrivateKey: privKey,
		ChainID:      big.NewInt(1337),
	}

	tests := []struct {
		name        string
		numBlocks   int64
		gasLimit    uint64
		clientSetup func(*MockEthClient)
		expectError bool
		errorMsg    string
	}{
		{
			name:      "move one block",
			numBlocks: 1,
			gasLimit:  24000,
			clientSetup: func(m *MockEthClient) {
				// Default setup
			},
			expectError: false,
		},
		{
			name:      "gas price error",
			numBlocks: 1,
			gasLimit:  24000,
			clientSetup: func(m *MockEthClient) {
				m.gasPrice = nil
			},
			expectError: true,
			errorMsg:    "failed to get gas price: <nil>",
		},
		{
			name:      "send transaction error",
			numBlocks: 1,
			gasLimit:  24000,
			clientSetup: func(m *MockEthClient) {
				m.sendTxError = fmt.Errorf("network error")
			},
			expectError: true,
			errorMsg:    "failed to advance blocks after 5 retries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := cProps.Client.(*MockEthClient)
			mockClient.blockNumber = 1000
			mockClient.receipts = make(map[common.Hash]*types.Receipt)
			mockClient.sendTxError = nil
			tt.clientSetup(mockClient)

			err := MoveBlocksForward(cProps, &tt.numBlocks, tt.gasLimit)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.ErrorContains(t, err, "failed to get gas price")
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, uint64(1001), mockClient.blockNumber, "block should advance by 1")
			}
		})
	}
}

// TestWaitForTx tests the transaction waiting logic.
func TestWaitForTx(t *testing.T) {
	client := &MockEthClient{
		receipts: make(map[common.Hash]*types.Receipt),
	}
	txHash := common.HexToHash("0x1234")

	tests := []struct {
		name        string
		setup       func(*MockEthClient)
		ctx         context.Context
		expectError bool
	}{
		{
			name: "successful confirmation",
			setup: func(m *MockEthClient) {
				m.receipts[txHash] = &types.Receipt{Status: 1, BlockNumber: big.NewInt(1000), TxHash: txHash}
			},
			ctx:         context.Background(),
			expectError: false,
		},
		{
			name: "transaction failed",
			setup: func(m *MockEthClient) {
				m.receipts[txHash] = &types.Receipt{Status: 0, BlockNumber: big.NewInt(1000), TxHash: txHash}
			},
			ctx:         context.Background(),
			expectError: true,
		},
		{
			name: "timeout",
			setup: func(m *MockEthClient) {
				// Leave receipts empty to simulate delay
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()
				return ctx
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(client)
			receipt, err := waitForTx(tt.ctx, client, txHash)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, receipt)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, receipt)
				assert.Equal(t, txHash, receipt.TxHash)
			}
		})
	}
}

// TestPrintCurrentBlockNumber tests the block number printing function.
func TestPrintCurrentBlockNumber(t *testing.T) {
	cProps := &ktfunc.ConnectionProps{
		Client: &MockEthClient{
			blockNumber: 1000,
		},
	}

	tests := []struct {
		name        string
		setup       func(*MockEthClient)
		expectError bool
	}{
		{
			name: "successful fetch",
			setup: func(m *MockEthClient) {
				m.blockNumError = nil
			},
			expectError: false,
		},
		{
			name: "fetch error",
			setup: func(m *MockEthClient) {
				m.blockNumError = fmt.Errorf("network error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(cProps.Client.(*MockEthClient))
			err := printCurrentBlockNumber(cProps)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
