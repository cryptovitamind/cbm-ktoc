package tests

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

// KeyPair holds a private key and its corresponding address
type KeyPair struct {
	PrivateKey *ecdsa.PrivateKey
	Address    *common.Address
}

// DeterministicPrivateKeys generates a slice of deterministic key pairs
func DeterministicPrivateKeys(amount int) []KeyPair {
	keyPairs := make([]KeyPair, amount)
	seed := "your_seed_here"

	for i := 0; i < amount; i++ {
		seedBytes := []byte(seed + string(rune(i)))
		privateKey, err := deterministicPrivateKey(seedBytes)
		if err != nil {
			log.Fatal(err)
		}

		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		keyPairs[i] = KeyPair{
			PrivateKey: privateKey,
			Address:    &address,
		}
	}

	return keyPairs
}

// deterministicPrivateKey generates a private key from a seed
func deterministicPrivateKey(seed []byte) (*ecdsa.PrivateKey, error) {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(seed)
	seedHash := hash.Sum(nil)
	return crypto.ToECDSA(seedHash)
}
