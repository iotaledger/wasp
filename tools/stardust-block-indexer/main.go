// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	hivedb "github.com/iotaledger/hive.go/kvstore/database"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
	"github.com/samber/lo"

	old_kvstore "github.com/iotaledger/hive.go/kvstore"
	old_database "github.com/nnikolash/wasp-types-exported/packages/database"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <chain-db-dir> <dest-index-file>", os.Args[0])
	}

	targetChainDBDir := os.Args[1]
	destIndexFile := os.Args[2]

	targetChainDBDir = lo.Must(filepath.Abs(targetChainDBDir))
	destIndexFile = lo.Must(filepath.Abs(destIndexFile))

	if strings.HasPrefix(destIndexFile, targetChainDBDir) {
		log.Fatalf("destination file cannot reside inside source database folder")
	}

	if _, err := os.Stat(destIndexFile); !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("destination file already exists: %v", destIndexFile)
	}

	targetKVS := ConnectToDB(targetChainDBDir)
	targetStore := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(targetKVS))

	var almostLastState old_state.State
	timeForGettingAlmostLastState := measureTime(func() {
		almostLastState = lo.Must(targetStore.StateByIndex(lo.Must(targetStore.LatestBlockIndex()) - uint32(1000)))
	})

	var firstState old_state.State
	timeForGettingFirstState := measureTime(func() {
		firstState = lo.Must(targetStore.StateByIndex(0))
	})

	log.Printf("Time for getting almost last state by index: %v\n", timeForGettingAlmostLastState)
	log.Printf("Time for getting first state by index: %v\n", timeForGettingFirstState)

	almostLastStateHash := almostLastState.TrieRoot()
	timeForGettingAlmostLastState = measureTime(func() {
		lo.Must(targetStore.StateByTrieRoot(almostLastStateHash))
	})

	firstStateHash := firstState.TrieRoot()
	timeForGettingFirstState = measureTime(func() {
		lo.Must(targetStore.StateByTrieRoot(firstStateHash))
	})

	log.Printf("Time for getting almost last state by hash: %v\n", timeForGettingAlmostLastState)
	log.Printf("Time for getting first state by hash: %v\n", timeForGettingFirstState)

	latestState := lo.Must(targetStore.LatestState())

	fmt.Printf("Last block index: %v\n", latestState.BlockIndex())

	startTime := time.Now()
	printStats := newStatsPrinter(latestState)

	reverseIterateStates(targetStore, func(trieRoot old_trie.Hash, state old_state.State) bool {
		printStats(state)

		return true
	})

	fmt.Println()
	fmt.Printf("Elapsed time: %v\n", time.Since(startTime))
}

func ConnectToDB(dbDir string) old_kvstore.KVStore {
	log.Printf("Connecting to DB in %v\n", dbDir)

	rocksDatabase := lo.Must(rocksdb.OpenDBReadOnly(dbDir,
		rocksdb.IncreaseParallelism(runtime.NumCPU()-1),
		rocksdb.Custom([]string{
			"periodic_compaction_seconds=43200",
			"level_compaction_dynamic_level_bytes=true",
			"keep_log_file_num=2",
			"max_log_file_size=50000000", // 50MB per log file
		}),
	))

	db := old_database.New(
		dbDir,
		rocksdb.New(rocksDatabase),
		hivedb.EngineRocksDB,
		true,
		func() bool { panic("should not be called") },
	)

	kvs := db.KVStore()

	return kvs
}

func reverseIterateStates(s old_indexedstore.IndexedStore, f func(trieRoot old_trie.Hash, state old_state.State) bool) {
	state := lo.Must(s.LatestState())
	trieRoot := state.TrieRoot()

	for {
		if !f(trieRoot, state) {
			return
		}

		prevL1Commitment := state.PreviousL1Commitment()
		if prevL1Commitment == nil {
			if state.BlockIndex() != 0 {
				// Just double-checking
				panic(fmt.Errorf("iterating the chain: state block index %d has no previous L1 commitment", state.BlockIndex()))
			}

			// done
			break
		}

		trieRoot = prevL1Commitment.TrieRoot()
		state = lo.Must(s.StateByTrieRoot(trieRoot))
	}
}

func periodicAction(period time.Duration, lastActionTime *time.Time, action func()) {
	if time.Since(*lastActionTime) >= period {
		action()
		*lastActionTime = time.Now()
	}
}

func newStatsPrinter(latestState old_state.State) func(state old_state.State) {
	totalBlocksCount := latestState.BlockIndex() + 1
	blocksLeft := totalBlocksCount

	var estimateRunTime time.Duration
	var avgSpeed int
	var currentSpeed int
	prevBlocksLeft := blocksLeft
	startTime := time.Now()
	lastEstimateUpdateTime := time.Now()

	return func(state old_state.State) {
		blocksLeft--

		const period = time.Second
		periodicAction(period, &lastEstimateUpdateTime, func() {
			if state.BlockIndex() != blocksLeft {
				// Just double-checking
				panic(fmt.Errorf("blocks left: state block index %d does not match expected block index %d", state.BlockIndex(), blocksLeft))
			}

			totalBlocksProcessed := totalBlocksCount - blocksLeft
			relProgress := float64(totalBlocksProcessed) / float64(totalBlocksCount)
			estimateRunTime = time.Duration(float64(time.Since(startTime)) / relProgress)
			avgSpeed = int(float64(totalBlocksProcessed) / time.Since(startTime).Seconds())

			recentBlocksProcessed := prevBlocksLeft - blocksLeft
			currentSpeed = int(float64(recentBlocksProcessed) / period.Seconds())
			prevBlocksLeft = blocksLeft
		})

		fmt.Printf("\rBlocks left: %v. Speed: %v blocks/sec. Avg speed: %v blocks/sec. Estimate time left: %v           ",
			blocksLeft, currentSpeed, avgSpeed, estimateRunTime)
	}
}

func measureTime(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Since(start)
}
