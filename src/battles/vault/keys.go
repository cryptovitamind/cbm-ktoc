package vault

// Vault keys. A key is 32 random bytes; the contract stores only
// keccak256(key) and pays out to whoever presents the preimage.

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// KeyLen is the byte length of a vault key.
const KeyLen = 32

// Key is the secret that unlocks one vault.
type Key [KeyLen]byte

// GenerateKey draws a fresh key from the OS random source.
func GenerateKey() (Key, error) {
	var k Key
	if _, err := rand.Read(k[:]); err != nil {
		return Key{}, fmt.Errorf("failed to draw a random key: %w", err)
	}
	return k, nil
}

// ParseKey accepts 64 hex characters, with or without a 0x prefix.
func ParseKey(s string) (Key, error) {
	b, err := hexBytes(s, KeyLen, "key")
	if err != nil {
		return Key{}, err
	}
	var k Key
	copy(k[:], b)
	return k, nil
}

// Hash is the value the vault is funded under: keccak256(abi.encodePacked(key)),
// which for a bytes32 is keccak256 of the raw 32 bytes.
func (k Key) Hash() common.Hash {
	return crypto.Keccak256Hash(k[:])
}

// Hex renders the key as 0x-prefixed lowercase hex.
func (k Key) Hex() string {
	return "0x" + hex.EncodeToString(k[:])
}

// ParseKeyHash accepts a 0x-prefixed 32-byte hash.
func ParseKeyHash(s string) (common.Hash, error) {
	if !strings.HasPrefix(s, "0x") {
		return common.Hash{}, fmt.Errorf("key hash must start with 0x")
	}
	b, err := hexBytes(s, common.HashLength, "key hash")
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(b), nil
}

func hexBytes(s string, n int, what string) ([]byte, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "0x")
	if len(s) != 2*n {
		return nil, fmt.Errorf("%s must be %d hex characters, got %d", what, 2*n, len(s))
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("%s is not valid hex: %w", what, err)
	}
	return b, nil
}
