package vault

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestGenerateKey_IsRandomAndFullLength(t *testing.T) {
	a, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	b, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	if a == b {
		t.Fatal("two generated keys are identical")
	}
	if a == (Key{}) {
		t.Fatal("generated key is all zero")
	}
}

// The contract computes keccak256(abi.encodePacked(key)); for a bytes32 that
// is keccak256 of the raw bytes. The Go side must match it bit for bit or no
// funded vault could ever be unlocked.
func TestKeyHash_MatchesSolidityPackedKeccak(t *testing.T) {
	var k Key
	for i := range k {
		k[i] = byte(i)
	}
	want := crypto.Keccak256Hash(k[:])
	if got := k.Hash(); got != want {
		t.Fatalf("Hash() = %s, want keccak256(key bytes) = %s", got.Hex(), want.Hex())
	}
}

func TestKeyHex_RoundTripsThroughParseKey(t *testing.T) {
	k, _ := GenerateKey()
	hex := k.Hex()
	if !strings.HasPrefix(hex, "0x") || len(hex) != 2+2*KeyLen {
		t.Fatalf("Hex() = %q, want 0x + 64 hex chars", hex)
	}
	back, err := ParseKey(hex)
	if err != nil {
		t.Fatalf("ParseKey(%q): %v", hex, err)
	}
	if back != k {
		t.Fatal("ParseKey(Hex()) != original key")
	}
	bare, err := ParseKey(strings.TrimPrefix(hex, "0x"))
	if err != nil || bare != k {
		t.Fatalf("ParseKey without 0x prefix: %v", err)
	}
}

func TestParseKey_RejectsBadInput(t *testing.T) {
	for _, s := range []string{"", "0x", "0xabc", strings.Repeat("g", 64), "0x" + strings.Repeat("ab", 31), "0x" + strings.Repeat("ab", 33)} {
		if _, err := ParseKey(s); err == nil {
			t.Errorf("ParseKey(%q) accepted invalid input", s)
		}
	}
}

func TestParseKeyHash_AcceptsHashRejectsGarbage(t *testing.T) {
	h := common.HexToHash("0x1234")
	got, err := ParseKeyHash(h.Hex())
	if err != nil || got != h {
		t.Fatalf("ParseKeyHash(%s) = %s, %v", h.Hex(), got.Hex(), err)
	}
	for _, s := range []string{"", "1234", "0x12", "0x" + strings.Repeat("zz", 32)} {
		if _, err := ParseKeyHash(s); err == nil {
			t.Errorf("ParseKeyHash(%q) accepted invalid input", s)
		}
	}
}
