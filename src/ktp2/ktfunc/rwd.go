package ktfunc

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"

	"encoding/binary"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

type FeeInfo struct {
	Block uint64
	Fee   *big.Int
}

// GetOwedEpochBlocks retrieves the unique epoch start blocks where the given address accrued OC fees,
// by filtering Voted and Rwd events and checking if the transaction was from the address.
func GetOwedEpochBlocks(cProps *ConnectionProps, addr common.Address, startBlock, endBlock uint64) ([]uint64, error) {
	uniqueBlocks := make(map[uint64]struct{})

	chunkSize := uint64(50000) // Adjust based on node limits; safer than 500 for logs
	for currentStart := startBlock; currentStart <= endBlock; currentStart += chunkSize {
		currentEnd := currentStart + chunkSize - 1
		if currentEnd > endBlock {
			currentEnd = endBlock
		}
		endPtr := currentEnd
		opts := &bind.FilterOpts{
			Start:   currentStart,
			End:     &endPtr,
			Context: context.Background(),
		}

		// Filter Voted events
		log.Debugf("Querying Voted events for block range %d-%d", currentStart, currentEnd)
		votedIter, err := cProps.Kt.FilterVoted(opts)
		if err != nil {
			if strings.Contains(err.Error(), "query returned more than") || strings.Contains(err.Error(), "log limit") {
				// If log limit exceeded, reduce chunk size and retry
				chunkSize /= 2
				if chunkSize < 1000 {
					return nil, fmt.Errorf("log query failed even with small chunks: %v", err)
				}
				currentStart -= chunkSize // Retry this chunk with smaller size
				continue
			}
			return nil, fmt.Errorf("failed to filter Voted events: %v", err)
		}
		for votedIter.Next() {
			event := votedIter.Event
			// Get tx details to check sender
			tx, isPending, err := cProps.Client.TransactionByHash(context.Background(), event.Raw.TxHash)
			if err != nil {
				log.Warnf("Failed to get tx %s: %v", event.Raw.TxHash.Hex(), err)
				continue
			}
			if isPending || tx == nil {
				continue
			}
			signer := types.LatestSignerForChainID(tx.ChainId())
			sender, err := types.Sender(signer, tx)
			if err != nil {
				log.Warnf("Failed to get sender for tx %s: %v", event.Raw.TxHash.Hex(), err)
				continue
			}
			if sender == addr {
				uniqueBlocks[event.Raw.BlockNumber] = struct{}{}
			}
		}
		if err := votedIter.Error(); err != nil {
			return nil, fmt.Errorf("error iterating Voted events: %v", err)
		}
		votedIter.Close()

		// Filter Rwd events
		log.Debugf("Querying Rwd events for block range %d-%d", currentStart, currentEnd)
		rwdIter, err := cProps.Kt.FilterRwd(opts)
		if err != nil {
			if strings.Contains(err.Error(), "query returned more than") || strings.Contains(err.Error(), "log limit") {
				chunkSize /= 2
				if chunkSize < 1000 {
					return nil, fmt.Errorf("log query failed even with small chunks: %v", err)
				}
				currentStart -= chunkSize
				continue
			}
			return nil, fmt.Errorf("failed to filter Rwd events: %v", err)
		}
		for rwdIter.Next() {
			event := rwdIter.Event
			// Get tx details to check sender
			tx, isPending, err := cProps.Client.TransactionByHash(context.Background(), event.Raw.TxHash)
			if err != nil {
				log.Warnf("Failed to get tx %s: %v", event.Raw.TxHash.Hex(), err)
				continue
			}
			if isPending || tx == nil {
				continue
			}
			signer := types.LatestSignerForChainID(tx.ChainId())
			sender, err := types.Sender(signer, tx)
			if err != nil {
				log.Warnf("Failed to get sender for tx %s: %v", event.Raw.TxHash.Hex(), err)
				continue
			}
			if sender == addr {
				// For Rwd, query startBlock at the block before the event
				prevBlockNum := big.NewInt(0).Sub(new(big.Int).SetUint64(event.Raw.BlockNumber), big.NewInt(1))
				callOpts := &bind.CallOpts{
					Context:     context.Background(),
					BlockNumber: prevBlockNum,
				}
				start, err := cProps.Kt.StartBlock(callOpts)
				if err != nil {
					log.Warnf("Failed to query startBlock at block %d: %v", prevBlockNum.Uint64(), err)
					continue
				}
				uniqueBlocks[start.Uint64()] = struct{}{}
			}
		}
		if err := rwdIter.Error(); err != nil {
			return nil, fmt.Errorf("error iterating Rwd events: %v", err)
		}
		rwdIter.Close()
	}

	var blocks []uint64
	for b := range uniqueBlocks {
		blocks = append(blocks, b)
	}
	sort.Slice(blocks, func(i, j int) bool { return blocks[i] < blocks[j] })
	return blocks, nil
}

func GetOCFeesOwed(cProps *ConnectionProps, startEndBlocks string) (*big.Float, error) {
	log.Printf("Querying OC fees owed to wallet %s for blocks %s", cProps.MyPubKey, startEndBlocks)
	// Parse start and end blocks from string
	startBlock, endBlock, err := ParseStartEndBlocks(startEndBlocks)
	if err != nil {
		log.Fatalf("Invalid start:end blocks format: %s", startEndBlocks)
		return nil, err
	}
	addr := ToAddr(cProps.Addresses.MyPublicKey)
	// Open DB once
	db, err := openFeesDB(cProps)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// Get relevant epoch blocks
	epochBlocks, err := GetOwedEpochBlocks(cProps, addr, startBlock, endBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get owed epoch blocks: %v", err)
	}
	log.Debugf("Found %d potential epoch blocks with fees", len(epochBlocks))
	// Sum fees for those blocks
	totalFees := big.NewInt(0)
	var nonZeroFees []FeeInfo
	numBlocks := len(epochBlocks)
	processed := 0
	progressInterval := 100
	if numBlocks < progressInterval {
		progressInterval = 0 // No progress for small number
	}
	if progressInterval > 0 {
		fmt.Printf("Processing %d potential fee blocks...", numBlocks)
	}
	for _, block := range epochBlocks {
		fee, err := getOcFee(db, cProps, addr, block)
		if err != nil {
			if progressInterval > 0 {
				fmt.Println() // Ensure newline
			}
			return nil, fmt.Errorf("failed to fetch fee for block %d: %v", block, err)
		}
		if fee.Cmp(big.NewInt(0)) > 0 {
			nonZeroFees = append(nonZeroFees, FeeInfo{Block: block, Fee: fee})
		}
		totalFees.Add(totalFees, fee)
		processed++
		if progressInterval > 0 && processed%progressInterval == 0 {
			fmt.Printf("\rProcessing: %d / %d blocks", processed, numBlocks)
		}
	}
	if progressInterval > 0 {
		fmt.Printf("\rProcessing: %d / %d blocks\n", numBlocks, numBlocks)
	}
	weiToEth := new(big.Float).SetInt(totalFees)
	totalEth := new(big.Float).Quo(weiToEth, big.NewFloat(1e18))
	for _, f := range nonZeroFees {
		log.Printf("Fee for block %d: %.6f ETH", f.Block, new(big.Float).Quo(new(big.Float).SetInt(f.Fee), big.NewFloat(1e18)))
	}
	if totalFees.Cmp(big.NewInt(0)) > 0 {
		log.Printf("Total fees across blocks %d to %d: %s wei", startBlock, endBlock, totalFees.String())
	}
	PrintKtBalance(cProps)
	PrintBalanceOfAddr(cProps, cProps.MyPubKey)
	log.Printf("Total OC fees owed: %.6f ETH", totalEth)
	return totalEth, nil
}

// startBlock, endBlock, err := ParseStartEndBlocks(startEndBlocks)
func ParseStartEndBlocks(startEndBlocks string) (uint64, uint64, error) {
	parts := strings.Split(startEndBlocks, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid start:end blocks format, expected 'start:end'")
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
	PrintBalanceOfAddr(cProps, cProps.MyPubKey)
	var blocksUint32 []uint32
	var err error
	if blocks == "" {
		// Auto-discover unpaid blocks
		latest, err := cProps.Client.BlockNumber(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get latest block: %v", err)
		}
		epochBlocks, err := GetOwedEpochBlocks(cProps, caller, 0, latest)
		if err != nil {
			return fmt.Errorf("failed to get owed epoch blocks: %v", err)
		}
		log.Debugf("Found %d potential epoch blocks", len(epochBlocks))
		// Open DB for checking fees
		db, err := openFeesDB(cProps)
		if err != nil {
			return err
		}
		var unpaid []uint64
		for _, b := range epochBlocks {
			fee, err := getOcFee(db, cProps, caller, b)
			if err != nil {
				log.Warnf("Failed to check fee for block %d: %v", b, err)
				continue
			}
			if fee.Cmp(big.NewInt(0)) > 0 {
				unpaid = append(unpaid, b)
			}
		}
		sort.Slice(unpaid, func(i, j int) bool { return unpaid[i] < unpaid[j] })
		for _, b := range unpaid {
			blocksUint32 = append(blocksUint32, uint32(b))
		}
		log.Printf("Auto-discovered %d unpaid blocks: %v", len(blocksUint32), blocksUint32)
		db.Close()
	} else {
		// Parse provided blocks
		blocksUint32, err = parseWithdrawBlocks(blocks)
		if err != nil {
			log.Fatalf("invalid block format: %v", err)
			return fmt.Errorf("invalid block format: %v", err)
		}
		log.Printf("Parsed blocks: %v", blocksUint32)
	}

	log.Debugf("Get caller's balance before withdrawal")
	balanceBefore, err := cProps.Client.BalanceAt(context.Background(), caller, nil)
	if err != nil {
		return fmt.Errorf("failed to get caller's balance before withdrawal: %v", err)
	}
	// Open DB once (if not already for auto)
	log.Debugf("Preparing to withdraw fees for %d blocks", len(blocksUint32))
	db, err := openFeesDB(cProps)
	if err != nil {
		return err
	}
	log.Debug("Opened fees DB")
	defer db.Close()
	// Check fees owed for these blocks (optional, for debugging)
	log.Printf("Verifying fees owed for %d blocks before withdrawal...", len(blocksUint32))
	totalFeesOwed := big.NewInt(0)
	numBlocks := len(blocksUint32)
	processed := 0
	progressInterval := 100
	if numBlocks < progressInterval {
		progressInterval = 0 // No progress for small number of blocks
	}
	if progressInterval > 0 {
		fmt.Printf("Checking fees for %d blocks...", numBlocks)
	}
	for _, block := range blocksUint32 {
		fee, err := getOcFee(db, cProps, caller, uint64(block))
		if err != nil {
			log.Warnf("Failed to query ocFees for block %d: %v", block, err)
			continue
		}
		totalFeesOwed.Add(totalFeesOwed, fee)
		processed++
		if progressInterval > 0 && processed%progressInterval == 0 {
			fmt.Printf("\rChecking fees: %d / %d blocks", processed, numBlocks)
		}
	}
	if progressInterval > 0 {
		fmt.Printf("\rChecking fees: %d / %d blocks\n", numBlocks, numBlocks)
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
	// Update cache to set fees to 0 for withdrawn blocks
	err = updateOcFeesToZero(db, caller, blocksUint32)
	if err != nil {
		log.Warnf("Failed to update cache after withdrawal: %v", err)
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
	PrintBalanceOfAddr(cProps, cProps.MyPubKey)
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
		if val < 1 || val > 5000000 {
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

// openFeesDB opens the fees database and creates the bucket if needed.
func openFeesDB(cProps *ConnectionProps) (*bbolt.DB, error) {
	// Ensure cache directory exists
	cacheDir := "cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Errorf("Failed to create cache directory: %v", err)
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	log.Debugf("Cache directory ensured: %s", cacheDir)
	// Construct database file name using first 7 characters of contract address
	dbName := fmt.Sprintf("%s/fees_%s.db", cacheDir, cProps.KtAddr.Hex()[:7])
	log.Debugf("Database file name constructed: %s", dbName)
	// Open DB
	db, err := bbolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Errorf("Failed to open database %s: %v", dbName, err)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	log.Debugf("Database opened successfully: %s", dbName)
	// Create bucket if not exists
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("fees"))
		return err
	})
	if err != nil {
		db.Close()
		log.Errorf("Failed to create bucket: %v", err)
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}
	log.Debugf("Bucket 'fees' created or verified in database")
	return db, nil
}

// getOcFee retrieves the OC fee for a given address and block, using cache if available.
// Assumes db is already open.
func getOcFee(db *bbolt.DB, cProps *ConnectionProps, addr common.Address, block uint64) (*big.Int, error) {
	// Generate key: address_bytes + "_" + block (as 8-byte big-endian)
	key := make([]byte, common.AddressLength+1+8)
	copy(key, addr.Bytes())
	key[common.AddressLength] = '_'
	binary.BigEndian.PutUint64(key[common.AddressLength+1:], block)
	var fee *big.Int
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("fees"))
		v := b.Get(key)
		if v != nil {
			fee = new(big.Int).SetBytes(v)
			log.Debugf("Loaded cached fee for address %s block %d: %s", addr.Hex(), block, fee.String())
			return nil
		}
		return nil // Not found, proceed to query
	})
	if err != nil {
		return nil, err
	}
	if fee != nil {
		return fee, nil
	}
	// Not in cache, query the node
	log.Debugf("Cache miss for address %s block %d, querying node", addr.Hex(), block)
	fee, err = cProps.Kt.OcFees(&bind.CallOpts{Context: context.Background()}, addr, big.NewInt(int64(block)))
	if err != nil {
		return nil, err
	}
	// Store in cache
	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("fees"))
		return b.Put(key, fee.Bytes())
	})
	if err != nil {
		log.Warnf("Failed to cache fee for address %s block %d: %v", addr.Hex(), block, err)
		// Continue anyway, don't fail the query
	}
	return fee, nil
}

// updateOcFeesToZero sets the cached OC fees to zero for the given blocks after withdrawal.
// Assumes db is already open.
func updateOcFeesToZero(db *bbolt.DB, addr common.Address, blocks []uint32) error {
	zero := big.NewInt(0)
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("fees"))
		if b == nil {
			return nil // Bucket doesn't exist, nothing to update
		}
		for _, blk := range blocks {
			block := uint64(blk)
			key := make([]byte, common.AddressLength+1+8)
			copy(key, addr.Bytes())
			key[common.AddressLength] = '_'
			binary.BigEndian.PutUint64(key[common.AddressLength+1:], block)
			if err := b.Put(key, zero.Bytes()); err != nil {
				return err
			}
			log.Debugf("Updated cache to zero for address %s block %d", addr.Hex(), block)
		}
		return nil
	})
	if err != nil {
		log.Errorf("Failed to update cache to zero: %v", err)
		return err
	}
	return nil
}
