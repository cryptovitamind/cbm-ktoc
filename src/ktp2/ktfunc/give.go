package ktfunc

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	log "github.com/sirupsen/logrus"
)

// give sends a specified amount of ETH to the KT contract, signing the
// transaction with the supplied privateKey (NOT the operator's key on
// cProps.MyPrivateKey). main.go's giveETH() depends on this so each test
// wallet sends from its own balance.
func Give(cProps *ConnectionProps, privateKey *ecdsa.PrivateKey, amount *big.Int) error {
	log.Debugf("Giving ETH to contract")

	// Validate inputs. cProps is checked first and on its own, so the
	// follow-up Client/Kt accesses can't fault.
	if cProps == nil {
		log.Errorf("ConnectionProps is nil")
		return fmt.Errorf("invalid ConnectionProps: cProps is nil")
	}
	if cProps.Client == nil || cProps.Kt == nil {
		log.Errorf("Invalid ConnectionProps - Client: %v, KT: %v", cProps.Client, cProps.Kt)
		return fmt.Errorf("invalid ConnectionProps: client or KT instance is nil")
	}
	if privateKey == nil {
		log.Errorf("Private key is nil")
		return fmt.Errorf("private key is nil")
	}
	if amount == nil || amount.Sign() < 0 {
		log.Errorf("Invalid amount: %v", amount)
		return fmt.Errorf("amount must be non-nil and non-negative")
	}

	// Prepare transaction authorization signed with the SUPPLIED privateKey
	// (not cProps.MyPrivateKey). Callers pass per-wallet keys expecting
	// each tx to originate from the corresponding wallet.
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, cProps.ChainID)
	if err != nil {
		return fmt.Errorf("failed to create function: %v", err)
	}
	if cProps.GasLimit > DefaultGasLimit {
		auth.GasLimit = cProps.GasLimit
	}

	auth.Value = new(big.Int).Set(amount) // Ensure a copy to avoid modifying input
	log.Infof("Sending amount: %s ETH from %s", new(big.Float).Quo(new(big.Float).SetInt(amount), big.NewFloat(1e18)).String(), auth.From.Hex())

	// Execute the Give transaction
	tx, err := cProps.Kt.Give(auth)
	if err != nil {
		log.Errorf("Failed to send give transaction: %v", err)
		return fmt.Errorf("failed to send give transaction: %w", err)
	}
	log.Infof("Transaction sent: %s", tx.Hash().Hex())

	// Wait for the transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), cProps.Client, tx)
	if err != nil {
		log.Errorf("Failed to mine transaction: %v", err)
		return fmt.Errorf("failed to mine transaction: %w", err)
	}

	// Check transaction status
	if receipt.Status == 0 {
		log.Warnf("Transaction failed - Hash: %s", tx.Hash().Hex())
		return fmt.Errorf("transaction failed: %s", tx.Hash().Hex())
	}
	log.Infof("Transaction succeeded - Block: %d", receipt.BlockNumber.Uint64())
	return nil
}
