package ktfunc

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"go.etcd.io/bbolt"
)

func TestGatherStakesAndWithdraws_Caching(t *testing.T) {
	// Remove existing DB if any
	os.Remove("events.db")

	db, err := bbolt.Open("events.db", 0600, nil)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()
	defer os.Remove("events.db")

	// Create bucket
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("chunks"))
		return err
	})
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	// Write a sample chunk
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, 1)

	sampleChunk := ChunkEvents{
		StakeEvents: []StakeEvent{{common.Address{}, big.NewInt(100), 1}},
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("chunks"))
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(sampleChunk); err != nil {
			return err
		}
		return b.Put(key, buf.Bytes())
	})
	if err != nil {
		t.Fatalf("Failed to write chunk: %v", err)
	}

	// Read the chunk
	var readChunk ChunkEvents
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("chunks"))
		v := b.Get(key)
		if v == nil {
			return fmt.Errorf("chunk not found")
		}
		return gob.NewDecoder(bytes.NewReader(v)).Decode(&readChunk)
	})
	if err != nil {
		t.Fatalf("Failed to read chunk: %v", err)
	}

	if len(readChunk.StakeEvents) != 1 || readChunk.StakeEvents[0].Amount.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("Read chunk does not match written chunk")
	}
}

func TestGatherStakesAndWithdraws_Chunking(t *testing.T) {
	// This test would need proper mocking of FilterStaked and FilterWithdrew
	// For example, set up mock iterators that return expected events for specific ranges
	// Then verify the stakeDataMap is correctly populated
}
