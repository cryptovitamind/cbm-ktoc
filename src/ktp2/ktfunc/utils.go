package ktfunc

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

// GetPublicAddress derives the Ethereum address from a private key
func GetPublicAddress(privateKey *ecdsa.PrivateKey) common.Address {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	return crypto.PubkeyToAddress(*publicKeyECDSA)
}

// printBalanceOfAddr logs the ETH balance of the specified address.
// Returns an error if the balance cannot be retrieved.
func PrintBalanceOfAddr(cProps *ConnectionProps) error {
	// Validate inputs
	if cProps.Client == nil {
		log.Errorf("Ethereum client is nil")
		return fmt.Errorf("ethereum client is nil")
	}
	if cProps == nil || cProps.MyPubKey == (common.Address{}) {
		log.Errorf("Invalid ConnectionProps - Public key: %v", cProps.MyPubKey)
		return fmt.Errorf("invalid ConnectionProps: public key is invalid")
	}

	// Get balance
	balance, err := cProps.Client.BalanceAt(context.Background(), cProps.MyPubKey, nil)
	if err != nil {
		log.Errorf("Failed to get balance for %s: %v", cProps.MyPubKey.Hex(), err)
		return fmt.Errorf("failed to get balance: %w", err)
	}

	// Log balance in ETH
	ethBalance := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
	log.Debugf("Balance of %s: %s ETH", cProps.MyPubKey.Hex(), ethBalance.String())
	return nil
}

// getChainId retrieves the blockchain's chain ID.
// Returns the chain ID or nil if retrieval fails.
func GetChainId(client *ethclient.Client) (*big.Int, error) {
	// Validate input
	if client == nil {
		log.Errorf("Ethereum client is nil")
		return nil, fmt.Errorf("ethereum client is nil")
	}

	// Get chain ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Errorf("Failed to get chain ID: %v", err)
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	log.Debugf("Chain ID: %s", chainID.String())
	return chainID, nil
}

// toAddr converts a hex string to a common.Address.
// Returns the address; no error return as it’s a simple conversion.
func ToAddr(addr string) common.Address {
	return common.HexToAddress(addr)
}
