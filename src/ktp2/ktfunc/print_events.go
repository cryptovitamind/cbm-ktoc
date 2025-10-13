package ktfunc

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"go.etcd.io/bbolt"
)

// PrintEvents reads and prints the contents of the events database for debugging purposes.
// It uses the provided contract address to determine the database file name.
func PrintEvents(contractAddr common.Address) error {
	// Ensure cache directory exists
	cacheDir := "cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Construct database file name using first 7 characters of contract address
	dbName := fmt.Sprintf("%s/%s.db", cacheDir, contractAddr.Hex()[:7])

	// Open the database file
	db, err := bbolt.Open(dbName, 0600, nil)
	if err != nil {
		return fmt.Errorf("failed to open database %s: %v", dbName, err)
	}
	defer db.Close()

	// Print header
	fmt.Println("Database Contents:")
	fmt.Println("------------------")

	// Iterate over all buckets and keys
	err = db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("chunks"))
		if bucket == nil {
			fmt.Println("No 'chunks' bucket found in database.")
			return nil
		}

		fmt.Println("Bucket: chunks")
		return bucket.ForEach(func(k, v []byte) error {
			var chunk ChunkEvents
			err := gob.NewDecoder(bytes.NewReader(v)).Decode(&chunk)
			if err != nil {
				fmt.Printf("  Key: %x, Error decoding data: %v\n", k, err)
				return nil
			}

			fmt.Printf("  Chunk starting at block: %d\n", binary.BigEndian.Uint64(k))
			if len(chunk.StakeEvents) > 0 {
				fmt.Println("    Stake Events:")
				fmt.Println("      Address                                      | Amount (Wei)          | Block")
				fmt.Println("      ---------------------------------------------|-----------------------|-------")
				for _, event := range chunk.StakeEvents {
					fmt.Printf("      %s | %23s | %d\n", event.Addr.Hex(), event.Amount.String(), event.Block)
				}
			} else {
				fmt.Println("    No Stake Events in this chunk.")
			}

			if len(chunk.WithdrawEvents) > 0 {
				fmt.Println("    Withdraw Events:")
				fmt.Println("      Address                                      | Amount (Wei)          | Block")
				fmt.Println("      ---------------------------------------------|-----------------------|-------")
				for _, event := range chunk.WithdrawEvents {
					fmt.Printf("      %s | %23s | %d\n", event.Addr.Hex(), event.Amount.String(), event.Block)
				}
			} else {
				fmt.Println("    No Withdraw Events in this chunk.")
			}
			return nil
		})
	})
	if err != nil {
		return fmt.Errorf("failed to iterate over database contents: %v", err)
	}

	fmt.Println("------------------")
	return nil
}
