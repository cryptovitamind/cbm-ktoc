package ktfunc

import (
	"context"
	"fmt"
	"ktp2/src/abis/ktv2"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
)

var newTransactor = NewTransactor
var waitForBlocks = WaitForBlocks
var waitMined = bind.WaitMined

// ValidateAddress parses and validates an Ethereum address string.
// Returns the address and nil error if valid and non-zero; otherwise, an error.
func ValidateAddress(addrStr string) (common.Address, error) {
	if len(addrStr) != 42 {
		return common.Address{}, fmt.Errorf("Invalid Ethereum address length: %s (must be 42 characters). Example: 0x22...822", addrStr)
	}
	addr := common.HexToAddress(addrStr)
	if addr == (common.Address{}) {
		return addr, fmt.Errorf("Invalid Ethereum address: %s", addrStr)
	}
	return addr, nil
}

// VoteToRemove calls the contract's VoteToRemove function for the given target address and data.
// Follows existing transaction patterns: auth, transact, wait for mining, wait for blocks.
func VoteToRemove(cProps *ConnectionProps, targetAddr common.Address, data string) error {
	log.Infof("Initiating vote to remove OC: %s with data: %q", targetAddr.Hex(), data)

	auth, err := newTransactor(cProps)
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}

	tx, err := cProps.Kt.VoteToRemove(auth, targetAddr, data)
	if err != nil {
		return fmt.Errorf("failed to send vote to remove transaction: %w", err)
	}

	log.Printf("Vote to remove transaction sent: %s", tx.Hash().Hex())

	receipt, err := waitMined(context.Background(), cProps.Client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for vote to remove transaction to be mined: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Errorf("Vote to remove transaction reverted: %s", tx.Hash().Hex())
		return fmt.Errorf("vote to remove transaction reverted with status %d", receipt.Status)
	}

	log.Debugf("Vote to remove transaction mined in block: %d", receipt.BlockNumber.Uint64())

	err = waitForBlocks(cProps)
	if err != nil {
		return fmt.Errorf("failed to wait for additional blocks: %w", err)
	}

	log.Printf("Vote to remove %s completed successfully after %d blocks.", targetAddr.Hex(), cProps.BlocksToWait)
	return nil
}

// VoteToAdd calls the contract's VoteToAdd function for the given target address and data.
// Follows existing transaction patterns: auth, transact, wait for mining, wait for blocks.
func VoteToAdd(cProps *ConnectionProps, targetAddr common.Address, data string) error {
	log.Infof("Initiating vote to add OC: %s with data: %q", targetAddr.Hex(), data)

	// Pre-checks for debugging
	callOpts := &bind.CallOpts{Context: context.Background(), From: cProps.MyPubKey}

	// Check wallet balance
	balance, err := cProps.Client.BalanceAt(context.Background(), cProps.MyPubKey, nil)
	if err != nil {
		log.Warnf("Failed to get wallet balance: %v", err)
	} else {
		ethBalance := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
		log.Infof("Wallet balance: %.6f ETH", ethBalance)
		if balance.Cmp(big.NewInt(0)) == 0 {
			return fmt.Errorf("wallet has zero balance, cannot pay for gas")
		}
	}

	// Check if target is already an OC
	isOC, err := cProps.Kt.OcRwdrs(callOpts, targetAddr)
	if err != nil {
		log.Warnf("Failed to check if target is OC: %v", err)
	} else if isOC {
		log.Warnf("Target address %s is already an OC, cannot add duplicate", targetAddr.Hex())
		return fmt.Errorf("target is already an OC")
	} else {
		log.Infof("Target %s is not an OC", targetAddr.Hex())
	}

	// Check if already voted for this add
	hasVotedForAdd, err := cProps.Kt.HasVotedAdd(callOpts, cProps.MyPubKey, targetAddr)
	if err != nil {
		log.Warnf("Failed to check if already voted to add: %v", err)
	} else if hasVotedForAdd {
		log.Warnf("Already voted to add %s", targetAddr.Hex())
		return fmt.Errorf("already voted to add this OC")
	} else {
		log.Infof("Not yet voted to add %s", targetAddr.Hex())
	}

	// Check data
	if data == "" {
		log.Warnf("Message to send with vote is empty, contract does not require a message.")
	} else {
		log.Infof("Data provided: %q", data)
	}

	// Estimate gas
	gasPrice, err := cProps.Client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Warnf("Failed to get gas price: %v", err)
	} else {
		log.Infof("Suggested gas price: %s wei", gasPrice.String())
	}

	auth, err := newTransactor(cProps)
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}

	tx, err := cProps.Kt.VoteToAdd(auth, targetAddr, data)
	if err != nil {
		return fmt.Errorf("failed to send vote to add transaction: %w", err)
	}

	log.Printf("Vote to add transaction sent: %s", tx.Hash().Hex())

	receipt, err := waitMined(context.Background(), cProps.Client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for vote to add transaction to be mined: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Errorf("Vote to add transaction reverted: %s", tx.Hash().Hex())
		return fmt.Errorf("vote to add transaction reverted with status %d", receipt.Status)
	}

	log.Debugf("Vote to add transaction mined in block: %d", receipt.BlockNumber.Uint64())

	err = waitForBlocks(cProps)
	if err != nil {
		return fmt.Errorf("failed to wait for additional blocks: %w", err)
	}

	log.Printf("Vote to add %s completed successfully after %d blocks.", targetAddr.Hex(), cProps.BlocksToWait)
	return nil
}

// ResetVoteToAdd calls the contract's ResetVoteToAdd function for the given target address.
// Follows existing transaction patterns: auth, transact, wait for mining, wait for blocks.
func ResetVoteToAdd(cProps *ConnectionProps, targetAddr common.Address) error {
	log.Infof("Initiating reset vote to add OC: %s", targetAddr.Hex())

	// Pre-check if already voted for this add
	callOpts := &bind.CallOpts{Context: context.Background(), From: cProps.MyPubKey}

	hasVotedForAdd, err := cProps.Kt.HasVotedAdd(callOpts, cProps.MyPubKey, targetAddr)
	if err != nil {
		log.Warnf("Failed to check if voted to add: %v", err)
	} else if !hasVotedForAdd {
		log.Warnf("Not voted to add %s", targetAddr.Hex())
		return fmt.Errorf("not voted to add this OC")
	} else {
		log.Infof("Voted to add %s, proceeding to reset", targetAddr.Hex())
	}

	auth, err := newTransactor(cProps)
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}

	tx, err := cProps.Kt.ResetVoteToAdd(auth, targetAddr)
	if err != nil {
		return fmt.Errorf("failed to send reset vote to add transaction: %w", err)
	}

	log.Printf("Reset vote to add transaction sent: %s", tx.Hash().Hex())

	receipt, err := waitMined(context.Background(), cProps.Client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for reset vote to add transaction to be mined: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Errorf("Reset vote to add transaction reverted: %s", tx.Hash().Hex())
		return fmt.Errorf("reset vote to add transaction reverted with status %d", receipt.Status)
	}

	log.Debugf("Reset vote to add transaction mined in block: %d", receipt.BlockNumber.Uint64())

	err = waitForBlocks(cProps)
	if err != nil {
		return fmt.Errorf("failed to wait for additional blocks: %w", err)
	}

	log.Printf("Reset vote to add %s completed successfully after %d blocks.", targetAddr.Hex(), cProps.BlocksToWait)
	return nil
}

// ResetVoteToRemove calls the contract's ResetVoteToRemove function for the given target address.
// Follows existing transaction patterns: auth, transact, wait for mining, wait for blocks.
func ResetVoteToRemove(cProps *ConnectionProps, targetAddr common.Address) error {
	log.Infof("Initiating reset vote to remove OC: %s", targetAddr.Hex())

	// Pre-check if already voted for this remove
	callOpts := &bind.CallOpts{Context: context.Background(), From: cProps.MyPubKey}

	hasVotedForRemove, err := cProps.Kt.HasVotedRemove(callOpts, cProps.MyPubKey, targetAddr)
	if err != nil {
		log.Warnf("Failed to check if voted to remove: %v", err)
	} else if !hasVotedForRemove {
		log.Warnf("Not voted to remove %s", targetAddr.Hex())
		return fmt.Errorf("not voted to remove this OC")
	} else {
		log.Infof("Voted to remove %s, proceeding to reset", targetAddr.Hex())
	}

	auth, err := newTransactor(cProps)
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}

	tx, err := cProps.Kt.ResetVoteToRemove(auth, targetAddr)
	if err != nil {
		return fmt.Errorf("failed to send reset vote to remove transaction: %w", err)
	}

	log.Printf("Reset vote to remove transaction sent: %s", tx.Hash().Hex())

	receipt, err := waitMined(context.Background(), cProps.Client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for reset vote to remove transaction to be mined: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Errorf("Reset vote to remove transaction reverted: %s", tx.Hash().Hex())
		return fmt.Errorf("reset vote to remove transaction reverted with status %d", receipt.Status)
	}

	log.Debugf("Reset vote to remove transaction mined in block: %d", receipt.BlockNumber.Uint64())

	err = waitForBlocks(cProps)
	if err != nil {
		return fmt.Errorf("failed to wait for additional blocks: %w", err)
	}

	log.Printf("Reset vote to remove %s completed successfully after %d blocks.", targetAddr.Hex(), cProps.BlocksToWait)
	return nil
}

// PrintVoteEvents retrieves and prints all VotedToAdd and VotedToRemove events between the specified block range.
func PrintVoteEvents(cProps *ConnectionProps, fromBlock *big.Int, toBlock *big.Int) error {
	log.Infof("Retrieving vote events from block %s to %s", fromBlock.String(), toBlock.String())

	from := fromBlock.Uint64()
	to := toBlock.Uint64()
	opts := &bind.FilterOpts{
		Start:   from,
		End:     &to,
		Context: context.Background(),
	}

	filterer, err := ktv2.NewKtv2Filterer(cProps.KtAddr, cProps.Client)
	if err != nil {
		return fmt.Errorf("failed to create filterer: %w", err)
	}

	// FilterVotedToAdd looks like this:
	// func (_Ktv2 *Ktv2Filterer) FilterVotedToAdd(opts *bind.FilterOpts, voter []common.Address, newOC []common.Address) (*Ktv2VotedToAddIterator, error) {
	addIterator, err := filterer.FilterVotedToAdd(opts, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to filter VotedToAdd events: %w", err)
	}
	defer addIterator.Close()

	log.Info("VotedToAdd events:")
	for addIterator.Next() {
		if addIterator.Error() != nil {
			log.Errorf("error iterating VotedToAdd: %v", addIterator.Error())
			break
		}
		event := addIterator.Event
		log.Infof("Block: %d, Tx: %s, Voter: %s, NewOC: %s, Data: %q",
			event.Raw.BlockNumber, event.Raw.TxHash.Hex(), event.Voter.Hex(), event.NewOC.Hex(), event.Data)
	}

	// Filter VotedToRemove events
	removeIterator, err := filterer.FilterVotedToRemove(opts, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to filter VotedToRemove events: %w", err)
	}
	defer removeIterator.Close()

	log.Info("VotedToRemove events:")
	for removeIterator.Next() {
		if removeIterator.Error() != nil {
			log.Errorf("error iterating VotedToRemove: %v", removeIterator.Error())
			break
		}
		event := removeIterator.Event
		log.Infof("Block: %d, Tx: %s, Voter: %s, ExistingOC: %s, Data: %q",
			event.Raw.BlockNumber, event.Raw.TxHash.Hex(), event.Voter.Hex(), event.ExistingOC.Hex(), event.Data)
	}

	log.Info("Vote events retrieval completed.")
	return nil
}
