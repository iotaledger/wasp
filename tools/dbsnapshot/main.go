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

func copyState(source indexedstore.IndexedStore, destination indexedstore.IndexedStore, trieRoot trie.Hash) error {
	//var buf bytes.Buffer

	fmt.Println("	Taking Snapshot")

	/*
		if err := source.TakeSnapshot(trieRoot, &buf); err != nil {
			return err
		}

		fmt.Println("	Copying Snapshot")

		if err := destination.RestoreSnapshot(trieRoot, &buf); err != nil {
			return err
		}*/

	sourceBlock, _ := source.BlockByTrieRoot(trieRoot)
	sourceState, _ := source.StateByTrieRoot(trieRoot)
	destDraft, _ := destination.NewStateDraft(sourceState.Timestamp(), sourceBlock.PreviousL1Commitment())

	sourceState.IterateSorted("", func(key kv.Key, value []byte) bool {
		destDraft.Set(key, value)
		return true
	})

	for k, _ := range sourceBlock.Mutations().Dels {
		destDraft.Del(k)
	}

	destination.SetLatest(trieRoot)

	return nil
}

func copyBlockWithState(source indexedstore.IndexedStore, destination indexedstore.IndexedStore, blockIndex uint32, start time.Time) {
	fmt.Printf("  Extracting Block %d\n", blockIndex)

	checkValidity := false

	sourceState, err := source.StateByIndex(blockIndex)
	if err != nil {
		panic(fmt.Errorf("failed to get state: %w", err))
	}

	srcMap := map[kv.Key][]byte{}
	dstMap := map[kv.Key][]byte{}

	if checkValidity {
		sourceState.IterateSorted("", func(key kv.Key, value []byte) bool {
			srcMap[key] = value
			return true
		})
	}

	if err := copyState(source, destination, sourceState.TrieRoot()); err != nil {
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
		}
	}

	fmt.Printf("  Block %d extracted, runtime: %v\n\n", blockIndex, time.Since(start))
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
	copyBlockWithState(sourceDB, destinationDB, 0, start)

	for _, block := range blocks.Blocks {
		copyBlockWithState(sourceDB, destinationDB, uint32(block), start)
	}

	fmt.Printf("Block extraction completed in %v\n", time.Since(start))
}
