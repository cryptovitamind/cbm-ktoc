package vault

// Vault keys. A key is 32 random bytes; the contract stores only
// keccak256(key) and pays out to whoever presents the preimage.

import (
	"github.com/ethereum/go-ethereum/common"
)

// KeyLen is the byte length of a vault key.
const KeyLen = 32

// Key is the secret that unlocks one vault.
type Key [KeyLen]byte

// GenerateKey draws a fresh key from the OS random source.
func GenerateKey() (Key, error) {
	return Key{}, errNotImplemented
}

// ParseKey accepts 64 hex characters, with or without a 0x prefix.
func ParseKey(s string) (Key, error) {
	return Key{}, errNotImplemented
}

// Hash is the value the vault is funded under: keccak256(abi.encodePacked(key)),
// which for a bytes32 is keccak256 of the raw 32 bytes.
func (k Key) Hash() common.Hash {
	return common.Hash{}
}

// Hex renders the key as 0x-prefixed lowercase hex.
func (k Key) Hex() string {
	return ""
}

// ParseKeyHash accepts a 0x-prefixed 32-byte hash.
func ParseKeyHash(s string) (common.Hash, error) {
	return common.Hash{}, errNotImplemented
}
