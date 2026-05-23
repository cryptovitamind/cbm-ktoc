package ktfunc

import (
	"math/big"
	"testing"

	"ktp2/src/abis/ktv2"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestParseStartEndBlocks(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedStart uint64
		expectedEnd   uint64
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "valid input",
			input:         "100:200",
			expectedStart: 100,
			expectedEnd:   200,
			expectError:   false,
		},
		{
			name:          "zero start",
			input:         "0:100",
			expectedStart: 0,
			expectedEnd:   100,
			expectError:   false,
		},
		{
			name:        "non-numeric values",
			input:       "abc:def",
			expectError: true,
			errorMsg:    "invalid start block: strconv.ParseUint: parsing \"abc\": invalid syntax",
		},
		{
			name:        "missing end",
			input:       "100:",
			expectError: true,
			errorMsg:    "invalid end block: strconv.ParseUint: parsing \"\": invalid syntax",
		},
		{
			name:        "missing start",
			input:       ":200",
			expectError: true,
			errorMsg:    "invalid start block: strconv.ParseUint: parsing \"\": invalid syntax",
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
			errorMsg:    "invalid start:end blocks format, expected 'start:end'",
		},
		{
			name:        "extra parts",
			input:       "100:200:300",
			expectError: true,
			errorMsg:    "invalid start:end blocks format, expected 'start:end'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ParseStartEndBlocks(tt.input)
			if tt.expectError {
				assert.Error(t, err, "expected an error for input: %s", tt.input)
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error(), "error message mismatch")
				}
			} else {
				assert.NoError(t, err, "unexpected error for input: %s", tt.input)
				assert.Equal(t, tt.expectedStart, start, "start block mismatch")
				assert.Equal(t, tt.expectedEnd, end, "end block mismatch")
			}
		})
	}
}

func TestParseWithdrawBlocks(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []uint32
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid input",
			input:       "100,200,300",
			expected:    []uint32{100, 200, 300},
			expectError: false,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
			errorMsg:    "block string cannot be empty",
		},
		{
			name:        "non-numeric value",
			input:       "100,abc,200",
			expectError: true,
			errorMsg:    "invalid block number 'abc' at position 2: strconv.ParseUint: parsing \"abc\": invalid syntax",
		},
		{
			name:        "below range",
			input:       "0",
			expectError: true,
			errorMsg:    "block number 0 at position 1 is out of range (1-5000000)",
		},
		{
			name:        "above range",
			input:       "5000001",
			expectError: true,
			errorMsg:    "block number 5000001 at position 1 is out of range (1-5000000)",
		},
		{
			name:        "empty part",
			input:       "100,,200",
			expectError: true,
			errorMsg:    "block 2 is empty",
		},
		{
			name:        "spaces in input",
			input:       " 100 , 200 ",
			expected:    []uint32{100, 200},
			expectError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseWithdrawBlocks(tt.input)
			if tt.expectError {
				assert.Error(t, err, "expected an error for input: %s", tt.input)
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error(), "error message mismatch")
				}
			} else {
				assert.NoError(t, err, "unexpected error for input: %s", tt.input)
				assert.Equal(t, tt.expected, result, "result mismatch")
			}
		})
	}
}

// ============================================================================
// Phase 5a — GetOwedEpochBlocks tests.
//
// This is the same shape as the Phase 4 withdraw-erasure bug: raw events
// (Voted, Rwd) get folded into a derived structure (unique epoch blocks).
// The function uses tx sender recovery to filter "did THIS address vote?".
// Highest-priority area for surfacing latent bugs.

// mockVotedIter is a minimal in-memory VotedIterator for tests.
type mockVotedIter struct {
	events []*ktv2.Ktv2Voted
	i      int
	err    error
}

func (m *mockVotedIter) Next() bool {
	if m.i >= len(m.events) {
		return false
	}
	m.i++
	return true
}
func (m *mockVotedIter) Event() *ktv2.Ktv2Voted { return m.events[m.i-1] }
func (m *mockVotedIter) Error() error           { return m.err }
func (m *mockVotedIter) Close() error           { return nil }

// mockRwdIter is a minimal in-memory RwdIterator for tests.
type mockRwdIter struct {
	events []*ktv2.Ktv2Rwd
	i      int
	err    error
}

func (m *mockRwdIter) Next() bool {
	if m.i >= len(m.events) {
		return false
	}
	m.i++
	return true
}
func (m *mockRwdIter) Event() *ktv2.Ktv2Rwd { return m.events[m.i-1] }
func (m *mockRwdIter) Error() error         { return m.err }
func (m *mockRwdIter) Close() error         { return nil }

// signedTxFromKey signs a no-op transaction with the given hex private key
// and returns both the tx (so its sender can be recovered) and the address
// of that signer. Uses LatestSignerForChainID to match what GetOwedEpochBlocks
// uses in production.
func signedTxFromKey(t *testing.T, hexKey string, chainID *big.Int) (*types.Transaction, common.Address) {
	t.Helper()
	priv, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		t.Fatalf("HexToECDSA: %v", err)
	}
	addr := crypto.PubkeyToAddress(priv.PublicKey)
	tx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 21000, big.NewInt(1e9), nil)
	signer := types.LatestSignerForChainID(chainID)
	signed, err := types.SignTx(tx, signer, priv)
	if err != nil {
		t.Fatalf("SignTx: %v", err)
	}
	return signed, addr
}

// Two deterministic test keys.
const (
	testKeyX = "0101010101010101010101010101010101010101010101010101010101010101"
	testKeyY = "0202020202020202020202020202020202020202020202020202020202020202"
)

// TestGetOwedEpochBlocks_ReturnsBlocksForMatchingSender — the basic flow:
// folder filters Voted events by tx sender. Two votes from X, one from Y;
// query for X. Expect X's two block numbers back, Y's filtered out.
func TestGetOwedEpochBlocks_ReturnsBlocksForMatchingSender(t *testing.T) {
	chainID := big.NewInt(1337)
	xTx, addrX := signedTxFromKey(t, testKeyX, chainID)
	yTx, _ := signedTxFromKey(t, testKeyY, chainID)

	txHashX1 := common.HexToHash("0x1111000000000000000000000000000000000000000000000000000000000000")
	txHashY := common.HexToHash("0x2222000000000000000000000000000000000000000000000000000000000000")
	txHashX2 := common.HexToHash("0x3333000000000000000000000000000000000000000000000000000000000000")

	votedEvents := []*ktv2.Ktv2Voted{
		{Raw: types.Log{TxHash: txHashX1, BlockNumber: 18_000_100}},
		{Raw: types.Log{TxHash: txHashY, BlockNumber: 18_000_400}},
		{Raw: types.Log{TxHash: txHashX2, BlockNumber: 18_000_700}},
	}

	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	mockKt.On("FilterVoted", mock.Anything).Return(&mockVotedIter{events: votedEvents}, nil)
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: nil}, nil)
	mockClient.On("TransactionByHash", mock.Anything, txHashX1).Return(xTx, false, nil)
	mockClient.On("TransactionByHash", mock.Anything, txHashY).Return(yTx, false, nil)
	mockClient.On("TransactionByHash", mock.Anything, txHashX2).Return(xTx, false, nil)

	cProps := &ConnectionProps{Kt: mockKt, Client: mockClient, ChunkSize: 10_000_000}

	blocks, err := GetOwedEpochBlocks(cProps, addrX, 18_000_000, 18_001_000)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []uint64{18_000_100, 18_000_700}, blocks)
}

// TestGetOwedEpochBlocks_DedupsRepeatedBlocksFromSameSender — sender voted
// twice in the same block (two txs same block, same epoch). The function
// stores into a map[uint64]struct{} so the block appears once.
func TestGetOwedEpochBlocks_DedupsRepeatedBlocksFromSameSender(t *testing.T) {
	chainID := big.NewInt(1337)
	xTx, addrX := signedTxFromKey(t, testKeyX, chainID)

	txHashA := common.HexToHash("0xaa00000000000000000000000000000000000000000000000000000000000000")
	txHashB := common.HexToHash("0xbb00000000000000000000000000000000000000000000000000000000000000")

	votedEvents := []*ktv2.Ktv2Voted{
		{Raw: types.Log{TxHash: txHashA, BlockNumber: 18_000_500}},
		{Raw: types.Log{TxHash: txHashB, BlockNumber: 18_000_500}}, // same block
	}

	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	mockKt.On("FilterVoted", mock.Anything).Return(&mockVotedIter{events: votedEvents}, nil)
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: nil}, nil)
	mockClient.On("TransactionByHash", mock.Anything, txHashA).Return(xTx, false, nil)
	mockClient.On("TransactionByHash", mock.Anything, txHashB).Return(xTx, false, nil)

	cProps := &ConnectionProps{Kt: mockKt, Client: mockClient, ChunkSize: 10_000_000}

	blocks, err := GetOwedEpochBlocks(cProps, addrX, 18_000_000, 18_001_000)
	assert.NoError(t, err)
	assert.Equal(t, []uint64{18_000_500}, blocks)
}

// TestGetOwedEpochBlocks_SkipsPendingTxs — when TransactionByHash returns
// isPending=true, the event is skipped (the vote isn't yet confirmed on
// chain so we don't credit the sender).
func TestGetOwedEpochBlocks_SkipsPendingTxs(t *testing.T) {
	chainID := big.NewInt(1337)
	xTx, addrX := signedTxFromKey(t, testKeyX, chainID)
	pendingTx, _ := signedTxFromKey(t, testKeyX, chainID) // X is also pending

	txHashConfirmed := common.HexToHash("0xcc00000000000000000000000000000000000000000000000000000000000000")
	txHashPending := common.HexToHash("0xdd00000000000000000000000000000000000000000000000000000000000000")

	votedEvents := []*ktv2.Ktv2Voted{
		{Raw: types.Log{TxHash: txHashConfirmed, BlockNumber: 18_000_100}},
		{Raw: types.Log{TxHash: txHashPending, BlockNumber: 18_000_200}},
	}

	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	mockKt.On("FilterVoted", mock.Anything).Return(&mockVotedIter{events: votedEvents}, nil)
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: nil}, nil)
	mockClient.On("TransactionByHash", mock.Anything, txHashConfirmed).Return(xTx, false, nil)
	mockClient.On("TransactionByHash", mock.Anything, txHashPending).Return(pendingTx, true, nil)

	cProps := &ConnectionProps{Kt: mockKt, Client: mockClient, ChunkSize: 10_000_000}

	blocks, err := GetOwedEpochBlocks(cProps, addrX, 18_000_000, 18_001_000)
	assert.NoError(t, err)
	assert.Equal(t, []uint64{18_000_100}, blocks)
}

// TestGetOwedEpochBlocks_SkipsTxWhenTransactionByHashErrors — RPC returns
// an error for one tx lookup. That event is skipped; the iterator
// continues; the other events are still processed.
func TestGetOwedEpochBlocks_SkipsTxWhenTransactionByHashErrors(t *testing.T) {
	chainID := big.NewInt(1337)
	xTx, addrX := signedTxFromKey(t, testKeyX, chainID)

	txHashOk := common.HexToHash("0xee00000000000000000000000000000000000000000000000000000000000000")
	txHashErr := common.HexToHash("0xff00000000000000000000000000000000000000000000000000000000000000")

	votedEvents := []*ktv2.Ktv2Voted{
		{Raw: types.Log{TxHash: txHashOk, BlockNumber: 18_000_100}},
		{Raw: types.Log{TxHash: txHashErr, BlockNumber: 18_000_200}},
	}

	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	mockKt.On("FilterVoted", mock.Anything).Return(&mockVotedIter{events: votedEvents}, nil)
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: nil}, nil)
	mockClient.On("TransactionByHash", mock.Anything, txHashOk).Return(xTx, false, nil)
	mockClient.On("TransactionByHash", mock.Anything, txHashErr).Return((*types.Transaction)(nil), false, assert.AnError)

	cProps := &ConnectionProps{Kt: mockKt, Client: mockClient, ChunkSize: 10_000_000}

	blocks, err := GetOwedEpochBlocks(cProps, addrX, 18_000_000, 18_001_000)
	assert.NoError(t, err)
	assert.Equal(t, []uint64{18_000_100}, blocks)
}

// TestGetOwedEpochBlocks_RwdEventBlockComesFromStartBlockLookup pins that
// for Rwd events the function does NOT use event.Raw.BlockNumber directly
// for the epoch block; it calls Kt.StartBlock(prevBlock) and uses that.
// A future refactor mistakenly using event.Raw.BlockNumber would silently
// pay fees for the wrong epoch.
func TestGetOwedEpochBlocks_RwdEventBlockComesFromStartBlockLookup(t *testing.T) {
	chainID := big.NewInt(1337)
	xTx, addrX := signedTxFromKey(t, testKeyX, chainID)

	rwdTxHash := common.HexToHash("0x9900000000000000000000000000000000000000000000000000000000000000")
	const rwdBlockNum = uint64(18_000_900)
	const epochStartBlock = uint64(18_000_400)

	rwdEvents := []*ktv2.Ktv2Rwd{
		{Raw: types.Log{TxHash: rwdTxHash, BlockNumber: rwdBlockNum}},
	}

	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	mockKt.On("FilterVoted", mock.Anything).Return(&mockVotedIter{events: nil}, nil)
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: rwdEvents}, nil)
	mockClient.On("TransactionByHash", mock.Anything, rwdTxHash).Return(xTx, false, nil)
	mockKt.On("StartBlock", mock.MatchedBy(func(opts *bind.CallOpts) bool {
		// Must be queried at (rwdBlockNum - 1).
		return opts != nil && opts.BlockNumber != nil && opts.BlockNumber.Uint64() == rwdBlockNum-1
	})).Return(big.NewInt(int64(epochStartBlock)), nil)

	cProps := &ConnectionProps{Kt: mockKt, Client: mockClient, ChunkSize: 10_000_000}

	blocks, err := GetOwedEpochBlocks(cProps, addrX, 18_000_000, 18_001_000)
	assert.NoError(t, err)
	assert.Equal(t, []uint64{epochStartBlock}, blocks,
		"Rwd event at block %d should resolve to epoch start %d via StartBlock(prevBlock); "+
			"if the test sees %d in the result it means a refactor regressed to using event.Raw.BlockNumber",
		rwdBlockNum, epochStartBlock, rwdBlockNum)
}

// TestGetOwedEpochBlocks_VotedAndRwdEventsAreBothCollected — both event
// types contribute to the same result set. This pins the union behavior
// so a future split into "voted-only" or "rwd-only" paths gets caught.
func TestGetOwedEpochBlocks_VotedAndRwdEventsAreBothCollected(t *testing.T) {
	chainID := big.NewInt(1337)
	xTx, addrX := signedTxFromKey(t, testKeyX, chainID)

	voteTxHash := common.HexToHash("0xaaa0000000000000000000000000000000000000000000000000000000000000")
	rwdTxHash := common.HexToHash("0xbbb0000000000000000000000000000000000000000000000000000000000000")

	votedEvents := []*ktv2.Ktv2Voted{
		{Raw: types.Log{TxHash: voteTxHash, BlockNumber: 18_000_300}},
	}
	rwdEvents := []*ktv2.Ktv2Rwd{
		{Raw: types.Log{TxHash: rwdTxHash, BlockNumber: 18_000_800}},
	}

	mockKt := &MockKtv2{}
	mockClient := &MockEthClient{}
	mockKt.On("FilterVoted", mock.Anything).Return(&mockVotedIter{events: votedEvents}, nil)
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: rwdEvents}, nil)
	mockClient.On("TransactionByHash", mock.Anything, voteTxHash).Return(xTx, false, nil)
	mockClient.On("TransactionByHash", mock.Anything, rwdTxHash).Return(xTx, false, nil)
	mockKt.On("StartBlock", mock.Anything).Return(big.NewInt(18_000_500), nil)

	cProps := &ConnectionProps{Kt: mockKt, Client: mockClient, ChunkSize: 10_000_000}

	blocks, err := GetOwedEpochBlocks(cProps, addrX, 18_000_000, 18_001_000)
	assert.NoError(t, err)
	// Voted contributes 18_000_300 (its raw block); Rwd contributes 18_000_500
	// (the StartBlock looked up at rwdBlock-1).
	assert.ElementsMatch(t, []uint64{18_000_300, 18_000_500}, blocks)
}

// TestGetOwedEpochBlocks_ZeroChunkSizeReturnsError pins the up-front guard.
func TestGetOwedEpochBlocks_ZeroChunkSizeReturnsError(t *testing.T) {
	cProps := &ConnectionProps{
		Kt:        &MockKtv2{},
		Client:    &MockEthClient{},
		ChunkSize: 0,
	}
	_, err := GetOwedEpochBlocks(cProps, common.Address{}, 0, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Chunk size cannot be zero")
}
