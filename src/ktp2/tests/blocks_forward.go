package tests

import (
	"context"
	"fmt"
	"ktp2/src/ktp2/ktfunc"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
)

func KeepMovingBlocks(cProps *ktfunc.ConnectionProps, gasLimit uint64) {
	numBlocks := int64(1)
	for {
		jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
		time.Sleep(3*time.Second + jitter)

		err := MoveBlocksForward(cProps, &numBlocks, gasLimit)
		if err != nil {
			log.Printf("Error in MoveBlocksForward: %v", err)
		}
	}
}

func MoveBlocksForward(cProps *ktfunc.ConnectionProps, numBlocks *int64, gasLimit uint64) error {
	ktfunc.LogOperationStart("Moving blocks forward")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	startBlock, err := cProps.Client.BlockNumber(ctx)
	if err != nil {
		log.Errorf("Failed to get start block: %v", err)
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	targetBlock := startBlock + uint64(*numBlocks)
	log.Infof("Starting block: %d", startBlock)
	log.Infof("Target block: %d", targetBlock)

	gasPrice, err := cProps.Client.SuggestGasPrice(ctx)
	if err != nil {
		log.Errorf("Gas price fetch failed: %v", err)
		return fmt.Errorf("failed to get gas price: %w", err)
	}
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(12))
	gasPrice = gasPrice.Div(gasPrice, big.NewInt(10)) // 1.2x suggested price
	log.Infof("Adjusted gas price: %s wei", gasPrice.String())

	if gasLimit == 0 {
		gasLimit = ktfunc.DefaultGasLimit
		log.Warn(fmt.Sprintf("Gas limit unspecified: Using default %d", ktfunc.DefaultGasLimit))
	} else {
		log.Infof("Using gas limit: %d", gasLimit)
	}

	currentBlock := startBlock
	maxRetries := 5
	retries := 0

	for currentBlock < targetBlock {
		if retries >= maxRetries {
			return fmt.Errorf("failed to advance blocks after %d retries", maxRetries)
		}

		nonce, err := cProps.Client.PendingNonceAt(ctx, cProps.MyPubKey)
		if err != nil {
			log.Errorf("Failed to get nonce: %v", err)
			return fmt.Errorf("failed to get nonce: %w", err)
		}

		tx := types.NewTransaction(
			nonce,
			cProps.MyPubKey,
			big.NewInt(0),
			gasLimit,
			gasPrice,
			nil,
		)

		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(cProps.ChainID), cProps.MyPrivateKey)
		if err != nil {
			log.Errorf("Transaction signing failed for nonce %d: %v", nonce, err)
			return fmt.Errorf("failed to sign transaction (nonce %d): %w", nonce, err)
		}

		err = cProps.Client.SendTransaction(ctx, signedTx)
		if err != nil {
			if strings.Contains(err.Error(), "already known") {
				log.Warnf("Transaction send failed for nonce %d: %v - Retrying with next nonce", nonce, err)
				time.Sleep(1 * time.Second)
				retries++
				continue
			}
			log.Errorf("Transaction send failed for nonce %d: %v", nonce, err)
			return fmt.Errorf("failed to send transaction (nonce %d): %w", nonce, err)
		}
		log.Infof("Transaction sent: %s", signedTx.Hash().Hex())

		receipt, err := waitForTx(ctx, cProps.Client, signedTx.Hash())
		if err != nil {
			log.Errorf("Transaction confirmation failed for nonce %d: %v", nonce, err)
			retries++
			time.Sleep(1 * time.Second)
			continue
		}
		log.Infof("Block advanced to: %d (nonce %d)", receipt.BlockNumber.Uint64(), nonce)

		currentBlock, err = cProps.Client.BlockNumber(ctx)
		if err != nil {
			log.Errorf("Failed to update current block: %v", err)
			return fmt.Errorf("failed to get updated block number: %w", err)
		}
		log.Infof("Current block: %d", currentBlock)
		retries = 0
	}

	log.Info("Block advancement complete")
	if err := printCurrentBlockNumber(cProps); err != nil {
		log.Warnf("Failed to print final block number: %v", err)
	}

	return nil
}

// waitForTx waits for a transaction to be mined and returns its receipt.
func waitForTx(ctx context.Context, client ktfunc.EthClient, txHash common.Hash) (*types.Receipt, error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Warnf("Transaction wait canceled: %s", txHash.Hex())
			return nil, fmt.Errorf("context canceled while waiting for transaction %s", txHash.Hex())
		case <-ticker.C:
			receipt, err := client.TransactionReceipt(ctx, txHash)
			if err == nil {
				if receipt.Status == 0 {
					log.Warnf("Transaction failed: %s (status: 0)", txHash.Hex())
					return nil, fmt.Errorf("transaction %s failed (status: 0)", txHash.Hex())
				}
				log.Infof("Transaction confirmed: %s", txHash.Hex())
				return receipt, nil
			}
			if err != ethereum.NotFound {
				log.Errorf("Receipt fetch failed: %v", err)
				return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
			}
			// Transaction not yet mined; continue polling
		}
	}
}

// printCurrentBlockNumber retrieves and logs the current block number.
func printCurrentBlockNumber(cProps *ktfunc.ConnectionProps) error {
	ctx := context.Background()
	currentBlock, err := cProps.Client.BlockNumber(ctx)
	if err != nil {
		log.Errorf("Failed to get current block number: %v", err)
		return fmt.Errorf("failed to get current block number: %w", err)
	}
	log.Infof("Final block number: %d", currentBlock)
	return nil
}
