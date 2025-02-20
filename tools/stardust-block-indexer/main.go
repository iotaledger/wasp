// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <chain-db-dir> <dest-index-file> [lastBlockIdx]", os.Args[0])
	}

	targetChainDBDir := os.Args[1]
	destIndexFile := os.Args[2]

	var lastBlockIdx uint32
	if len(os.Args) > 3 {
		lo.Must(fmt.Sscanf(os.Args[3], "%d", &lastBlockIdx))
	}

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

	var lastState old_state.State
	if lastBlockIdx == 0 {
		lastState = lo.Must(targetStore.LatestState())
	} else {
		lastState = lo.Must(targetStore.StateByIndex(lastBlockIdx))
	}

	totalBlocksCount := lastState.BlockIndex() + 1

	fmt.Printf("Last block index: %v\n", lastState.BlockIndex())

	startTime := time.Now()
	printStats := newStatsPrinter(totalBlocksCount)

	log.Printf("Creating empty index file at %v...", destIndexFile)
	f := lo.Must(os.Create(destIndexFile))
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()

	log.Printf("Allocating memory for %v entries of index...", totalBlocksCount)
	indexEntries := make([]old_trie.Hash, 0, totalBlocksCount)

	log.Printf("Building index for blocks in range [0; %v]...", lastState.BlockIndex())
	reverseIterateBlocks(targetStore, lastBlockIdx, func(trieRoot old_trie.Hash, block old_state.Block) bool {
		printStats(block.StateIndex)
		indexEntries = append(indexEntries, trieRoot)

		return ctx.Err() == nil
	})

	fmt.Println()
	fmt.Printf("Elapsed time: %v\n", time.Since(startTime))

	log.Printf("Saving index to %v...", destIndexFile)

	for i := len(indexEntries) - 1; i >= 0; i-- {
		if _, err := w.Write(indexEntries[i].Bytes()); err != nil {
			panic(err)
		}
	}

	log.Printf("Index saved to %v", destIndexFile)
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
				panic(fmt.Errorf("state block index %d has no previous L1 commitment", state.BlockIndex()))
			}

			// done
			break
		}

		trieRoot = prevL1Commitment.TrieRoot()
		state = lo.Must(s.StateByTrieRoot(trieRoot))
	}
}

func reverseIterateBlocks(s old_indexedstore.IndexedStore, fromIdx uint32, f func(trieRoot old_trie.Hash, block old_state.Block) bool) {
	var block old_state.Block
	var trieRoot old_trie.Hash

	if fromIdx == 0 {
		block = lo.Must(s.LatestBlock())
		trieRoot = block.TrieRoot()
	} else {
		block = lo.Must(s.BlockByIndex(fromIdx))
		trieRoot = block.TrieRoot()
	}

	for {
		if !f(trieRoot, block) {
			return
		}

		prevL1Commitment := block.PreviousL1Commitment()
		if prevL1Commitment == nil {
			if block.StateIndex() != 0 {
				// Just double-checking
				panic(fmt.Errorf("block index %d has no previous L1 commitment", block.StateIndex()))
			}

			// done
			break
		}

		trieRoot = prevL1Commitment.TrieRoot()
		block = lo.Must(s.BlockByTrieRoot(trieRoot))
	}
}

func periodicAction(period time.Duration, lastActionTime *time.Time, action func()) {
	if time.Since(*lastActionTime) >= period {
		action()
		*lastActionTime = time.Now()
	}
}

func newStatsPrinter(totalBlocksCount uint32) func(getBlockIndex func() uint32) {
	blocksLeft := totalBlocksCount

	var estimateRunTime time.Duration
	var avgSpeed int
	var currentSpeed int
	prevBlocksLeft := blocksLeft
	startTime := time.Now()
	lastEstimateUpdateTime := time.Now()

	return func(getBlockIndex func() uint32) {
		blocksLeft--

		const period = time.Second
		periodicAction(period, &lastEstimateUpdateTime, func() {
			if getBlockIndex() != blocksLeft {
				// Just double-checking
				panic(fmt.Errorf("state block index %d does not match expected block index %d", getBlockIndex(), blocksLeft))
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
