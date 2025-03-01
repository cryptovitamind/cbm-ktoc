package tests

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"ktp2/src/abis"
	"ktp2/src/ktp2/ktfunc"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
)

func StakeTokensToKt(cProps *ktfunc.ConnectionProps, privateKey *ecdsa.PrivateKey, amount *big.Int) error {
	testToken, err := GetTestToken(cProps)
	if err != nil {
		return fmt.Errorf("failed to initialize token contract: %w", err)
	}

	// Prepare transaction options
	auth, err := ktfunc.NewTransactor(cProps)
	if err != nil {
		return err
	}

	if err := executeIncreaseAllowance(cProps, testToken, auth, amount); err != nil {
		return err
	}

	if err := executeStake(cProps, auth, amount); err != nil {
		return err
	}

	return logTotalStaked(cProps, privateKey)
}

func executeIncreaseAllowance(cProps *ktfunc.ConnectionProps, token *abis.Shib, auth *bind.TransactOpts, amount *big.Int) error {
	tx, err := token.IncreaseAllowance(auth, cProps.KtAddr, amount)
	if err != nil {
		return fmt.Errorf("failed to increase allowance: %w", err)
	}

	log.Infof("IncreaseAllowance sent: %s", tx.Hash().String())
	return waitAndCheckTx(cProps, tx)
}

func executeStake(cProps *ktfunc.ConnectionProps, auth *bind.TransactOpts, amount *big.Int) error {
	tx, err := cProps.Kt.Stake(auth, amount)
	if err != nil {
		return fmt.Errorf("failed to stake tokens: %w", err)
	}

	log.Infof("Stake sent: %s", tx.Hash().String())
	return waitAndCheckTx(cProps, tx)
}

func waitAndCheckTx(cProps *ktfunc.ConnectionProps, tx *types.Transaction) error {
	receipt, err := bind.WaitMined(context.Background(), cProps.Client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for reward transaction to be mined: %v", err)
	}

	if receipt.Status == 0 {
		log.Error("Transaction failed")
		return fmt.Errorf("transaction reverted")
	}
	log.Info("Transaction succeeded")
	return nil
}

func logTotalStaked(props *ktfunc.ConnectionProps, privateKey *ecdsa.PrivateKey) error {
	callOpts := &bind.CallOpts{
		Context: context.Background(),
		From:    ktfunc.GetPublicAddress(privateKey),
	}

	stakeAmt, err := props.Kt.TotalStk(callOpts)
	if err != nil {
		return fmt.Errorf("failed to get total staked amount: %w", err)
	}

	log.Infof("Total staked amount: %s", stakeAmt.String())
	return nil
}

func GetTestToken(cProps *ktfunc.ConnectionProps) (*abis.Shib, error) {
	token, err := abis.NewShib(ktfunc.ToAddr(cProps.Addresses.TknAddr), cProps.Backend)
	if err != nil {
		return nil, err
	}
	return token, nil
}
