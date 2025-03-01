package tests

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"ktp2/src/ktp2/ktfunc"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

func SendSomeEth(cProps *ktfunc.ConnectionProps, privateKey *ecdsa.PrivateKey, amount float64) {
	recipientAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	amountWei := big.NewFloat(amount)
	amountWei.Mul(amountWei, big.NewFloat(10e18))
	amountWeiInt, _ := amountWei.Int(nil)

	sendETH(cProps, recipientAddress, amountWeiInt)
}

func SendSomeTestTokens(cProps *ktfunc.ConnectionProps, privateKey *ecdsa.PrivateKey, amount float64) {
	recipientAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	fmt.Printf("Sending %f tokens to %s\n", amount, recipientAddress)

	amountTokens := big.NewFloat(amount)
	amountTokens.Mul(amountTokens, big.NewFloat(1e18)) // Assuming 18 decimals for the test token
	amountTokensInt, _ := amountTokens.Int(nil)

	sendTestTokens(cProps, recipientAddress, amountTokensInt)
}

func PrintBalances(cProps *ktfunc.ConnectionProps, address common.Address) {
	fmt.Printf("Address %s:\n", address.Hex())

	// Get ETH balance
	ethBalance, err := cProps.Client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Fatalf("Failed to get ETH balance: %v", err)
	}

	// Convert Wei to ETH
	ethBalanceFloat := new(big.Float).Quo(new(big.Float).SetInt(ethBalance), big.NewFloat(1e18))

	// Get test token balance
	testToken, err := GetTestToken(cProps)
	if err != nil {
		log.Fatalf("Failed to get test token: %v", err)
	}

	tokenBalance, err := testToken.BalanceOf(&bind.CallOpts{}, address)
	if err != nil {
		log.Fatalf("Failed to get test token balance: %v", err)
	}

	// Convert token amount to float (assuming 18 decimals)
	tokenBalanceFloat := new(big.Float).Quo(new(big.Float).SetInt(tokenBalance), big.NewFloat(1e18))

	fmt.Printf("   ETH: %f\n", ethBalanceFloat)
	fmt.Printf("   Test Tokens: %f\n", tokenBalanceFloat)
}

func sendETH(cProps *ktfunc.ConnectionProps, to common.Address, amount *big.Int) {
	fmt.Printf("Sending %d Wei to %s...\n", amount, to.Hex())

	nonce, err := cProps.Client.PendingNonceAt(context.Background(), cProps.MyPubKey)
	if err != nil {
		log.Fatal(err)
	}

	gasLimit := uint64(cProps.GasLimit)
	gasPrice, err := cProps.Client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(cProps.ChainID), cProps.MyPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = cProps.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ETH transfer sent: %s\n", signedTx.Hash().Hex())
}

func sendTestTokens(cProps *ktfunc.ConnectionProps, to common.Address, amount *big.Int) {
	testToken, err := GetTestToken(cProps)
	if err != nil {
		log.Fatal(err)
	}

	// Prepare transaction options
	auth, err := ktfunc.NewTransactor(cProps)
	if err != nil {
		log.Fatalf("Failed to create transactor: %v", err)
	}

	// Get the current nonce
	nonce, err := cProps.Client.PendingNonceAt(context.Background(), cProps.MyPubKey)
	if err != nil {
		log.Fatal(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))

	// Get the suggested gas price and increase it slightly
	gasPrice, err := cProps.Client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	auth.GasPrice = new(big.Int).Mul(gasPrice, big.NewInt(110))
	auth.GasPrice = auth.GasPrice.Div(auth.GasPrice, big.NewInt(100)) // 110% of suggested gas price

	tx, err := testToken.Transfer(auth, to, amount)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test token transfer sent: %s\n", tx.Hash().Hex())

	// Wait for the transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), cProps.Client, tx)
	if err != nil {
		log.Fatalf("Failed to wait for transaction to be mined: %v", err)
	}

	if receipt.Status == 0 {
		log.Fatalf("Transaction failed: %s", tx.Hash().Hex())
	}

	fmt.Printf("Test token transfer confirmed in block %d\n", receipt.BlockNumber)
}
