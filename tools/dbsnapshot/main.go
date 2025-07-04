package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/samber/lo"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/trie"
)

func openSourceDB(dbPath string) indexedstore.IndexedStore {
	dbConn := lo.Must(rocksdb.OpenDBReadOnly(dbPath))
	db := database.New(dbPath, rocksdb.New(dbConn), hivedb.EngineRocksDB, false, func() bool {
		return false
	})

	store := db.KVStore()
	rebasedDBStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(store))

	return rebasedDBStore
}

func createDestinationDB(dbPath string) indexedstore.IndexedStore {
	dbConn := lo.Must(database.NewRocksDB(dbPath, database.CacheSizeDefault))
	db := database.New(dbPath, rocksdb.New(dbConn), hivedb.EngineRocksDB, false, func() bool {
		return false
	})

	store := db.KVStore()
	rebasedDBStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(store))

	return rebasedDBStore
}

func copyAsSnapshot(source indexedstore.IndexedStore, destination indexedstore.IndexedStore, trieRoot trie.Hash) error {
	var buf bytes.Buffer

	fmt.Println("	Taking Snapshot")

	if err := source.TakeSnapshot(trieRoot, &buf); err != nil {
		return err
	}

	fmt.Println("	Copying Snapshot")

	if err := destination.RestoreSnapshot(trieRoot, &buf); err != nil {
		return err
	}

	return nil
}

func copyState(source indexedstore.IndexedStore, destination indexedstore.IndexedStore, trieRoot trie.Hash, previousL1Commitment *state.L1Commitment, validationMap map[kv.Key][]byte) (*state.L1Commitment, error) {
	sourceState, _ := source.StateByTrieRoot(trieRoot)

	stateDraft, err := destination.NewStateDraft(sourceState.Timestamp(), previousL1Commitment)
	if err != nil {
		return nil, err
	}

	sourceKeys := map[kv.Key]bool{}

	sourceState.IterateKeys("", func(key kv.Key) bool {
		sourceKeys[key] = true
		return true
	})
	
	sourceState.Iterate("", func(key kv.Key, value []byte) bool {
		stateDraft.Set(key, value)
		validationMap[key] = value
		return true
	})

	newBlock := destination.Commit(stateDraft)
	destination.SetLatest(trieRoot)

	return newBlock.L1Commitment(), nil
}

func copyBlockWithState(source indexedstore.IndexedStore, destination indexedstore.IndexedStore, blockIndex uint32, start time.Time, previousL1Commitment *state.L1Commitment) *state.L1Commitment {
	fmt.Printf("  Extracting Block %d\n", blockIndex)

	checkValidity := true

	sourceState, err := source.StateByIndex(blockIndex)
	if err != nil {
		panic(fmt.Errorf("failed to get state: %w", err))
	}

	srcMap := map[kv.Key][]byte{}
	dstMap := map[kv.Key][]byte{}

	var newBlock *state.L1Commitment

	if newBlock, err = copyState(source, destination, sourceState.TrieRoot(), previousL1Commitment, srcMap); err != nil {
		panic(fmt.Errorf("failed to copy snapshot: %w", err))
	}

	if checkValidity {
		destinationState, err := destination.StateByIndex(blockIndex)
		if err != nil {
			panic(fmt.Errorf("failed to get state dst: %w", err))
		}

		destinationState.IterateSorted("", func(key kv.Key, value []byte) bool {
			dstMap[key] = value
			return true
		})

		if !compareMaps(srcMap, dstMap) {
			panic(fmt.Errorf("snapshot failed for block: %d", blockIndex))
		} else {
			fmt.Printf("  Successfully copied block %d\n", blockIndex)
		}
	}

	fmt.Printf("  Block %d extracted, runtime: %v\n\n", blockIndex, time.Since(start))

	return newBlock
}

func compareMaps(map1, map2 map[kv.Key][]byte) bool {
	// Check if lengths are different
	if len(map1) != len(map2) {
		return false
	}

	// Compare each key-value pair
	for key, value1 := range map1 {
		value2, exists := map2[key]
		if !exists {
			return false
		}

		// Use bytes.Equal to compare []byte slices
		if !bytes.Equal(value1, value2) {
			return false
		}
	}

	return true
}

func main() {
	start := time.Now()
	sourceDBPath, destinationDBPath, blocks := readArgs()

	fmt.Printf("Block Extraction Configuration:\n")
	fmt.Printf("  Source DB: %s\n", sourceDBPath)
	fmt.Printf("  Target DB: %s\n", destinationDBPath)
	fmt.Printf("  Blocks to extract: %v\n", blocks.Blocks)
	fmt.Printf("  Total blocks: %d\n\n", len(blocks.Blocks))

	sourceDB := openSourceDB(sourceDBPath)
	destinationDB := createDestinationDB(destinationDBPath)

	b, _ := sourceDB.LatestState()
	fmt.Printf("%v", b)

	block0, err := sourceDB.BlockByIndex(0)
	if err != nil {
		panic(err)
	}

	err = copyAsSnapshot(sourceDB, destinationDB, block0.TrieRoot())
	if err != nil {
		panic(err)
	}

	l1Commitment := block0.L1Commitment()
	for _, block := range blocks.Blocks {
		l1Commitment = copyBlockWithState(sourceDB, destinationDB, uint32(block), start, l1Commitment)
	}

	fmt.Printf("Block extraction completed in %v\n", time.Since(start))
}
