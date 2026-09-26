package vault

import (
	"math/big"
	"testing"

	"ktp2/src/abis/artifacts"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	bank   = common.HexToAddress(bankA)
	sender = common.HexToAddress("0x9999999999999999999999999999999999999999")
)

// giveCalldata is the real give() selector from the Ktv2 ABI, so the
// verifier is pinned to the contract and not to a hand-typed constant.
func giveCalldata(t *testing.T) []byte {
	t.Helper()
	art, err := artifacts.Load(artifacts.Ktv2)
	if err != nil {
		t.Fatalf("load Ktv2 artifact: %v", err)
	}
	m, ok := art.ABI.Methods["give"]
	if !ok {
		t.Fatal("Ktv2 ABI has no give()")
	}
	return m.ID
}

func mkTx(to *common.Address, value *big.Int, data []byte) *types.Transaction {
	return types.NewTx(&types.LegacyTx{Nonce: 1, To: to, Value: value, Gas: 100000, GasPrice: big.NewInt(1), Data: data})
}

func minedAt(block uint64, status uint64) *types.Receipt {
	return &types.Receipt{Status: status, BlockNumber: new(big.Int).SetUint64(block)}
}

func TestGiveSelector_MatchesKtv2ABI(t *testing.T) {
	want := giveCalldata(t)
	if string(giveSelector) != string(want) {
		t.Fatalf("giveSelector = %x, Ktv2 give() selector = %x", giveSelector, want)
	}
}

func TestEvaluateMatch_AcceptsQualifyingGive(t *testing.T) {
	tx := mkTx(&bank, big.NewInt(2e16), giveCalldata(t))
	r := EvaluateMatch(tx, minedAt(200, types.ReceiptStatusSuccessful), sender, true, big.NewInt(1e16), 100)
	if !r.OK || r.Reason != MatchOK {
		t.Fatalf("expected OK, got %+v", r)
	}
	if r.Bank != bank || r.Sender != sender || r.Amount.Cmp(big.NewInt(2e16)) != 0 || r.Block != 200 {
		t.Fatalf("result fields wrong: %+v", r)
	}
}

func TestEvaluateMatch_ExactMinimumQualifies(t *testing.T) {
	tx := mkTx(&bank, big.NewInt(1e16), giveCalldata(t))
	r := EvaluateMatch(tx, minedAt(200, types.ReceiptStatusSuccessful), sender, true, big.NewInt(1e16), 100)
	if !r.OK {
		t.Fatalf("an exact match must qualify, got %+v", r)
	}
}

func TestEvaluateMatch_Rejections(t *testing.T) {
	give := giveCalldata(t)
	other := common.HexToAddress(bankB)
	cases := []struct {
		name    string
		tx      *types.Transaction
		receipt *types.Receipt
		isBank  bool
		want    string
	}{
		{"not mined", mkTx(&bank, big.NewInt(1e16), give), nil, true, MatchNotMined},
		{"reverted", mkTx(&bank, big.NewInt(1e16), give), minedAt(200, types.ReceiptStatusFailed), true, MatchReverted},
		{"not a bank", mkTx(&other, big.NewInt(1e16), give), minedAt(200, types.ReceiptStatusSuccessful), false, MatchNotABank},
		{"plain transfer", mkTx(&bank, big.NewInt(1e16), nil), minedAt(200, types.ReceiptStatusSuccessful), true, MatchNotGive},
		{"other method", mkTx(&bank, big.NewInt(1e16), []byte{1, 2, 3, 4}), minedAt(200, types.ReceiptStatusSuccessful), true, MatchNotGive},
		{"too small", mkTx(&bank, big.NewInt(1e16-1), give), minedAt(200, types.ReceiptStatusSuccessful), true, MatchTooSmall},
		{"same block as funding", mkTx(&bank, big.NewInt(1e16), give), minedAt(100, types.ReceiptStatusSuccessful), true, MatchTooEarly},
		{"before funding", mkTx(&bank, big.NewInt(1e16), give), minedAt(99, types.ReceiptStatusSuccessful), true, MatchTooEarly},
		{"contract creation", mkTx(nil, big.NewInt(1e16), give), minedAt(200, types.ReceiptStatusSuccessful), false, MatchMissingContract},
	}
	for _, tc := range cases {
		r := EvaluateMatch(tc.tx, tc.receipt, sender, tc.isBank, big.NewInt(1e16), 100)
		if r.OK || r.Reason != tc.want {
			t.Errorf("%s: got OK=%v reason=%q, want reason %q", tc.name, r.OK, r.Reason, tc.want)
		}
	}
}
