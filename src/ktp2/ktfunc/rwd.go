package ktfunc

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	log "github.com/sirupsen/logrus"
)

func GetOCFeesOwed(cProps *ConnectionProps, startEndBlocks string) (*big.Float, error) {
	log.Printf("Querying OC fees owed to wallet %s for blocks %s", cProps.MyPubKey, startEndBlocks)

	// Parse start and end blocks from string
	startBlock, endBlock, err := ParseStartEndBlocks(startEndBlocks)
	if err != nil {
		log.Fatalf("Invalid start:end blocks format: %s", startEndBlocks)
		return nil, err
	}

	fetchFee := func(block uint64) (*big.Int, error) {
		return cProps.Kt.OcFees(&bind.CallOpts{Context: context.Background()}, ToAddr(cProps.Addresses.MyPublicKey), big.NewInt(int64(block)))
	}
	totalFees, err := sumFeesOverBlocks(fetchFee, startBlock, endBlock)
	if err != nil {
		return nil, err
	}

	weiToEth := new(big.Float).SetInt(totalFees)
	totalEth := new(big.Float).Quo(weiToEth, big.NewFloat(1e18))
	if totalFees.Cmp(big.NewInt(0)) > 0 {
		log.Printf("Total fees across blocks %d to %d: %s wei", startBlock, endBlock, totalFees.String())
	}

	PrintKtBalance(cProps)
	PrintBalanceOfAddr(cProps)
	log.Printf("Total OC fees owed: %.6f ETH", totalEth)
	return totalEth, nil
}

// startBlock, endBlock, err := ParseStartEndBlocks(startEndBlocks)
func ParseStartEndBlocks(startEndBlocks string) (uint64, uint64, error) {
	parts := strings.Split(startEndBlocks, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid start:end blocks format, expected 'start-end'")
	}

	start, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start block: %v", err)
	}

	end, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end block: %v", err)
	}

	return uint64(start), uint64(end), nil
}

func WithdrawOCFees(cProps *ConnectionProps, blocks string) error {
	log.Printf("Withdrawing OC fees for blocks: %s", blocks)

	// Print initial contract balance
	PrintKtBalance(cProps)

	// Get caller address
	caller := ToAddr(cProps.Addresses.MyPublicKey)
	PrintBalanceOfAddr(cProps)

	// Parse blocks into uint32 slice
	blocksUint32, err := parseWithdrawBlocks(blocks)
	if err != nil {
		log.Fatalf("invalid block format: %v", err)
		return fmt.Errorf("invalid block format: %v", err)
	}
	log.Printf("Parsed blocks: %v", blocksUint32)

	// Get caller's balance before withdrawal
	balanceBefore, err := cProps.Client.BalanceAt(context.Background(), caller, nil)
	if err != nil {
		return fmt.Errorf("failed to get caller's balance before withdrawal: %v", err)
	}

	// Check fees owed for these blocks (optional, for debugging)
	totalFeesOwed := big.NewInt(0)
	for _, block := range blocksUint32 {
		fee, err := cProps.Kt.OcFees(&bind.CallOpts{Context: context.Background()}, caller, big.NewInt(int64(block)))
		if err != nil {
			log.Warnf("Failed to query ocFees for block %d: %v", block, err)
			continue
		}
		totalFeesOwed.Add(totalFeesOwed, fee)
	}
	weiToEthOwed := new(big.Float).SetInt(totalFeesOwed)
	owedEth := new(big.Float).Quo(weiToEthOwed, big.NewFloat(1e18))
	log.Printf("Total fees owed for blocks: %.6f ETH", owedEth)
	if owedEth.Cmp(big.NewFloat(0)) <= 0 {
		log.Infof("No fees owed for blocks: %v", blocks)
		return nil
	}
	// Create an authenticated transactor
	auth, err := NewTransactor(cProps)
	if err != nil {
		return fmt.Errorf("failed to create transactor: %v", err)
	}

	// Call the withdrawOCFee function
	tx, err := cProps.Kt.WithdrawOCFee(auth, blocksUint32)
	if err != nil {
		return fmt.Errorf("failed to call withdrawOCFee: %v", err)
	}

	log.Printf("Withdraw transaction sent: %s", tx.Hash().Hex())

	// Wait for the transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), cProps.Client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for withdraw transaction to be mined: %v", err)
	}

	log.Debugf("Withdraw transaction mined in block: %d", receipt.BlockNumber.Uint64())

	// Wait for additional blocks
	err = WaitForBlocks(cProps)
	if err != nil {
		return fmt.Errorf("failed to wait for additional blocks: %v", err)
	}

	// Get caller's balance after withdrawal
	balanceAfter, err := cProps.Client.BalanceAt(context.Background(), caller, nil)
	if err != nil {
		return fmt.Errorf("failed to get caller's balance after withdrawal: %v", err)
	}

	// Calculate net balance change (including gas)
	amountWithdrawn := new(big.Int).Sub(balanceAfter, balanceBefore)
	weiToEthWithdrawn := new(big.Float).SetInt(amountWithdrawn)
	withdrawnEth := new(big.Float).Quo(weiToEthWithdrawn, big.NewFloat(1e18))

	// Convert balances to ETH
	weiToEthBefore := new(big.Float).SetInt(balanceBefore)
	balanceBeforeEth := new(big.Float).Quo(weiToEthBefore, big.NewFloat(1e18))
	weiToEthAfter := new(big.Float).SetInt(balanceAfter)
	balanceAfterEth := new(big.Float).Quo(weiToEthAfter, big.NewFloat(1e18))

	// Log gas cost (approximate)
	gasCost := new(big.Int).Mul(tx.GasPrice(), big.NewInt(int64(receipt.GasUsed)))
	weiToEthGas := new(big.Float).SetInt(gasCost)
	gasEth := new(big.Float).Quo(weiToEthGas, big.NewFloat(1e18))

	// Log the result
	log.Printf("OC fees withdrawn successfully for %d blocks | Amount received: %.6f ETH | Gas cost: %.6f ETH | Balance before: %.6f ETH | Balance after: %.6f ETH",
		len(blocksUint32), withdrawnEth, gasEth, balanceBeforeEth, balanceAfterEth)

	// Print final balances
	PrintKtBalance(cProps)
	PrintBalanceOfAddr(cProps)

	return nil
}

func parseWithdrawBlocks(blocks string) ([]uint32, error) {
	log.Printf("Parsing withdraw blocks: %q", blocks) // Use %q for quoted string to see exact input
	if blocks == "" {
		return nil, fmt.Errorf("block string cannot be empty")
	}

	parts := strings.Split(blocks, ",")

	result := make([]uint32, 0, len(parts))
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return nil, fmt.Errorf("block %d is empty", i+1)
		}

		val, err := strconv.ParseUint(trimmed, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid block number '%s' at position %d: %v", trimmed, i+1, err)
		}

		if val < 1 || val > 500000000 {
			return nil, fmt.Errorf("block number %d at position %d is out of range (1-5000000)", val, i+1)
		}

		log.Printf("Parsed block %d: %d", i+1, val)
		result = append(result, uint32(val))
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid blocks parsed")
	}

	return result, nil
}

type FeeFetcher func(block uint64) (*big.Int, error)

func sumFeesOverBlocks(fetchFee FeeFetcher, start, end uint64) (*big.Int, error) {
	total := big.NewInt(0)
	for block := start; block <= end; block++ {
		fee, err := fetchFee(block)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch fee for block %d: %v", block, err)
		}
		if fee.Cmp(big.NewInt(0)) > 0 {
			log.Printf("Fee for block %d: %.6f ETH", block, new(big.Float).Quo(new(big.Float).SetInt(fee), big.NewFloat(1e18)))
		}
		total.Add(total, fee)
	}
	return total, nil
}
