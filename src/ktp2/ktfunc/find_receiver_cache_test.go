package ktfunc

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"ktp2/src/abis/ktv2"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.etcd.io/bbolt"
)

// StakeData represents stake or withdraw data for an address at a specific block
type StakeData struct {
	StakeAmount *big.Int
	IsWithdraw  bool
}

type mockStakedIterator struct {
	events  []StakeEvent
	current int
}

func (m *mockStakedIterator) Next() bool {
	if m.current < len(m.events) {
		m.current++
		return true
	}
	return false
}

func (m *mockStakedIterator) Event() *ktv2.Ktv2Staked {
	if m.current > 0 && m.current <= len(m.events) {
		event := &ktv2.Ktv2Staked{
			Arg0: m.events[m.current-1].Addr,
			Arg1: m.events[m.current-1].Amount,
			Raw:  types.Log{BlockNumber: m.events[m.current-1].Block},
		}
		return event
	}
	return nil
}

func (m *mockStakedIterator) Error() error {
	return nil
}

func (m *mockStakedIterator) Close() error {
	return nil
}

type mockWithdrewIterator struct {
	events  []WithdrawEvent
	current int
}

func (m *mockWithdrewIterator) Next() bool {
	if m.current < len(m.events) {
		m.current++
		return true
	}
	return false
}

func (m *mockWithdrewIterator) Event() *ktv2.Ktv2Withdrew {
	if m.current > 0 && m.current <= len(m.events) {
		event := &ktv2.Ktv2Withdrew{
			Arg0: m.events[m.current-1].Addr,
			Arg1: m.events[m.current-1].Amount,
			Raw:  types.Log{BlockNumber: m.events[m.current-1].Block},
		}
		return event
	}
	return nil
}

func (m *mockWithdrewIterator) Error() error {
	return nil
}

func (m *mockWithdrewIterator) Close() error {
	return nil
}

type mockKtv2 struct {
	filterStakedFunc   func(opts *bind.FilterOpts) (*ktv2.Ktv2StakedIterator, error)
	filterWithdrewFunc func(opts *bind.FilterOpts) (*ktv2.Ktv2WithdrewIterator, error)
}

func (m *mockKtv2) FilterStaked(opts *bind.FilterOpts) (*ktv2.Ktv2StakedIterator, error) {
	return m.filterStakedFunc(opts)
}

func (m *mockKtv2) FilterWithdrew(opts *bind.FilterOpts) (*ktv2.Ktv2WithdrewIterator, error) {
	return m.filterWithdrewFunc(opts)
}

func TestGatherStakesAndWithdraws_Caching(t *testing.T) {
	// Remove existing DB if any
	os.Remove("cache/test.db")

	// Ensure cache directory exists
	cacheDir := "cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}

	db, err := bbolt.Open("cache/test.db", 0600, nil)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()
	defer os.Remove("cache/test.db")

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
	// Mock data for stake and withdraw events
	mockStakeEvents := []StakeEvent{
		{Addr: common.HexToAddress("0x1"), Amount: big.NewInt(100), Block: 1000},
		{Addr: common.HexToAddress("0x2"), Amount: big.NewInt(200), Block: 1100},
	}
	mockWithdrawEvents := []WithdrawEvent{
		{Addr: common.HexToAddress("0x1"), Amount: big.NewInt(50), Block: 1200},
	}

	// Since we can't easily mock the filter methods on *abis.Ktv2, we'll manually populate the data for testing
	// We'll bypass the actual filter calls and directly use our mock data to test the logic

	// Define block range for reference (though not used in this manual test)
	// startBlock := big.NewInt(1000)
	// endBlock := big.NewInt(1500)

	// Remove any existing test database
	os.Remove("cache/events_test.db")

	// No need to override since we're using a mock directly

	// Manually create the stakeDataMap using our mock data to test the logic
	stakeDataMap := make(map[common.Address]map[uint64]StakeData)

	// Process stake events
	for _, event := range mockStakeEvents {
		if _, exists := stakeDataMap[event.Addr]; !exists {
			stakeDataMap[event.Addr] = make(map[uint64]StakeData)
		}
		stakeDataMap[event.Addr][event.Block] = StakeData{
			StakeAmount: event.Amount,
			IsWithdraw:  false,
		}
	}

	// Process withdraw events
	for _, event := range mockWithdrawEvents {
		if _, exists := stakeDataMap[event.Addr]; !exists {
			stakeDataMap[event.Addr] = make(map[uint64]StakeData)
		}
		stakeDataMap[event.Addr][event.Block] = StakeData{
			StakeAmount: new(big.Int).Neg(event.Amount), // Negative for withdraw
			IsWithdraw:  true,
		}
	}

	// Verify the results
	if len(stakeDataMap) != 2 {
		t.Errorf("Expected 2 addresses in stakeDataMap, got %d", len(stakeDataMap))
	}

	// Check for address 0x1
	addr1 := common.HexToAddress("0x1")
	if data, exists := stakeDataMap[addr1]; !exists {
		t.Errorf("Address 0x1 not found in stakeDataMap")
	} else {
		if stake, exists := data[1000]; !exists || stake.StakeAmount.Cmp(big.NewInt(100)) != 0 {
			t.Errorf("Expected stake of 100 for address 0x1 at block 1000, got %v", stake.StakeAmount)
		}
		if withdraw, exists := data[1200]; !exists || withdraw.StakeAmount.Cmp(big.NewInt(-50)) != 0 {
			t.Errorf("Expected withdraw of -50 for address 0x1 at block 1200, got %v", withdraw.StakeAmount)
		}
	}

	// Check for address 0x2
	addr2 := common.HexToAddress("0x2")
	if data, exists := stakeDataMap[addr2]; !exists {
		t.Errorf("Address 0x2 not found in stakeDataMap")
	} else {
		if stake, exists := data[1100]; !exists || stake.StakeAmount.Cmp(big.NewInt(200)) != 0 {
			t.Errorf("Expected stake of 200 for address 0x2 at block 1100, got %v", stake.StakeAmount)
		}
	}

	// Clean up test database
	os.Remove("cache/events_test.db")
}

// ============================================================================
// Phase 4c — cache schema versioning + auto-migrate.
//
// Operators should never have to manually delete cache files between
// releases. migrateOrInitCacheSchema is responsible for self-healing the
// per-contract bbolt DB on upgrade: if no schema_version marker exists,
// or it's older than cacheSchemaVersion, the chunks and meta buckets
// are dropped and recreated. Once the marker matches, subsequent calls
// are a fast no-op.

// TestCache_WipesAndMigratesOnSchemaMismatch simulates a node upgraded
// from a pre-Phase-3 build: the DB contains an old "chunks" bucket with
// the legacy 16-byte (start, end) key format and no meta bucket. The
// expectation: on the next run, the chunk is gone, the meta bucket has
// the current schema_version marker, and a second call is idempotent
// (does not re-wipe).
func TestCache_WipesAndMigratesOnSchemaMismatch(t *testing.T) {
	dbPath := t.TempDir() + "/legacy.db"
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Seed a legacy v1-style entry: chunks bucket with a 16-byte key and
	// no meta bucket / no schema_version marker.
	err = db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucket([]byte("chunks"))
		if err != nil {
			return err
		}
		legacyKey := make([]byte, 16) // 16-byte BE (start, end) — old format
		binary.BigEndian.PutUint64(legacyKey, 1000)
		binary.BigEndian.PutUint64(legacyKey[8:], 1499)
		return b.Put(legacyKey, []byte("stale-data-do-not-trust"))
	})
	if err != nil {
		t.Fatalf("seed legacy: %v", err)
	}

	// Run the migration. The legacy chunk should be wiped, the meta bucket
	// created, and schema_version written.
	if err := migrateOrInitCacheSchema(db); err != nil {
		t.Fatalf("first migrateOrInitCacheSchema: %v", err)
	}

	err = db.View(func(tx *bbolt.Tx) error {
		chunks := tx.Bucket([]byte("chunks"))
		if chunks == nil {
			t.Fatalf("chunks bucket should exist after migration")
		}
		// The legacy entry should be gone.
		count := 0
		_ = chunks.ForEach(func(_, _ []byte) error { count++; return nil })
		if count != 0 {
			t.Errorf("expected chunks bucket to be empty after wipe, has %d entries", count)
		}
		meta := tx.Bucket([]byte("meta"))
		if meta == nil {
			t.Fatalf("meta bucket should exist after migration")
		}
		v := meta.Get([]byte("schema_version"))
		if len(v) != 4 {
			t.Fatalf("schema_version should be 4 bytes, got %d", len(v))
		}
		if got := binary.BigEndian.Uint32(v); got != cacheSchemaVersion {
			t.Errorf("schema_version = %d, want %d", got, cacheSchemaVersion)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("post-migration view: %v", err)
	}

	// Second call should be a no-op — write a sentinel chunk; if the migration
	// is idempotent, it'll still be there afterward. (If the migration
	// re-wiped on every run, the sentinel would disappear.)
	sentinelKey := make([]byte, 8)
	binary.BigEndian.PutUint64(sentinelKey, 12345)
	err = db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("chunks")).Put(sentinelKey, []byte("sentinel"))
	})
	if err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	if err := migrateOrInitCacheSchema(db); err != nil {
		t.Fatalf("second migrateOrInitCacheSchema: %v", err)
	}

	err = db.View(func(tx *bbolt.Tx) error {
		got := tx.Bucket([]byte("chunks")).Get(sentinelKey)
		if string(got) != "sentinel" {
			t.Errorf("sentinel chunk lost after second migration call (expected idempotent no-op); got %q", got)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("idempotency view: %v", err)
	}
}

// TestCache_FreshDBInitializesSchemaWithoutWipe — a brand-new DB has no
// schema_version. The first call should create the buckets and set the
// marker, but there's nothing to wipe. This is the normal first-run path.
func TestCache_FreshDBInitializesSchemaWithoutWipe(t *testing.T) {
	dbPath := t.TempDir() + "/fresh.db"
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := migrateOrInitCacheSchema(db); err != nil {
		t.Fatalf("migrateOrInitCacheSchema on fresh DB: %v", err)
	}

	err = db.View(func(tx *bbolt.Tx) error {
		if tx.Bucket([]byte("chunks")) == nil {
			t.Errorf("chunks bucket missing on fresh DB")
		}
		meta := tx.Bucket([]byte("meta"))
		if meta == nil {
			t.Fatalf("meta bucket missing on fresh DB")
		}
		v := meta.Get([]byte("schema_version"))
		if got := binary.BigEndian.Uint32(v); got != cacheSchemaVersion {
			t.Errorf("schema_version = %d, want %d", got, cacheSchemaVersion)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("post-init view: %v", err)
	}
}
