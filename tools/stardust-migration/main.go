// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	"github.com/samber/lo"

	old_iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"

	"github.com/iotaledger/wasp/tools/stardust-migration/blockindex"
	"github.com/iotaledger/wasp/tools/stardust-migration/cli"
	"github.com/iotaledger/wasp/tools/stardust-migration/db"
	"github.com/iotaledger/wasp/tools/stardust-migration/migrations"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
)

// NOTE: Every record type should be explicitly included in migration
// NOTE: All migration is node at once or just abandoned. There is no option to continue.
// TODO: Do we start from block 0 or N+1 where N last old block?
// TODO: Do we prune old block? Are we going to do migration from origin? If not, have we pruned blocks with old schemas?
// TODO: What to do with foundry prefixes?
// TODO: From where to get new chain ID?
// TODO: Need to migrate ALL trie roots to support tracing.
// TODO: New state draft might be huge, but it is stored in memory - might be an issue.

func main() {
	if len(os.Args) < 4 {
		log.Fatalf("usage: %s <src-chain-db-dir> <dest-chain-db-dir> <new-chain-id>", os.Args[0])
	}

	srcChainDBDir := os.Args[1]
	destChainDBDir := os.Args[2]
	newChainIDStr := os.Args[3]

	srcChainDBDir = lo.Must(filepath.Abs(srcChainDBDir))
	destChainDBDir = lo.Must(filepath.Abs(destChainDBDir))

	if strings.HasPrefix(destChainDBDir, srcChainDBDir) {
		log.Fatalf("destination database cannot reside inside source database folder")
	}

	srcKVS := db.Connect(srcChainDBDir)
	srcStore := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(srcKVS))
	srcState := lo.Must(srcStore.LatestState())

	if newChainIDStr == "dummy" {
		// just for easier testing
		newChainIDStr = "0x00000000000000000000000000000000000000000000000000000000000000ff"
	}
	oldChainID := old_isc.ChainID(GetAnchorOutput(srcState).AliasID)
	newChainID := lo.Must(isc.ChainIDFromString(newChainIDStr))

	lo.Must0(os.MkdirAll(destChainDBDir, 0o755))

	entries := lo.Must(os.ReadDir(destChainDBDir))
	if len(entries) > 0 {
		// TODO: Disabled this check now, so you can run the migrator multiple times for testing
		// log.Fatalf("destination directory is not empty: %v", destChainDBDir)
	}

	destKVS := db.Create(destChainDBDir)
	destStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVS))
	destStateDraft := destStore.NewOriginStateDraft()

	//migrateAllBlocks(srcStore, destStore)

	v := migrations.MigrateRootContract(srcState, destStateDraft)
	migrations.MigrateAccountsContract(v, srcState, destStateDraft, oldChainID, newChainID)
	migrations.MigrateBlocklogContract(srcState, destStateDraft)
	// migrations.MigrateGovernanceContract(srcState, destStateDraft)
	migrations.MigrateEVMContract(srcState, destStateDraft)

	newBlock := destStore.Commit(destStateDraft)
	destStore.SetLatest(newBlock.TrieRoot())
	destKVS.Flush()
}

// migrateAllBlocks calls migration functions for all mutations of each block.
func migrateAllBlocks(srcStore old_indexedstore.IndexedStore, destStore indexedstore.IndexedStore) {
	forEachBlock(srcStore, func(blockIndex uint32, blockHash old_trie.Hash, block *old_state.Block) {

	})
}

// forEachBlock iterates over all blocks.
// It uses index file index.bin if it is present, otherwise it uses indexing on-the-fly with blockindex.BlockIndexer.
// If index file does not have enough entries, it retrieves the rest of the blocks without indexing.
func forEachBlock(srcStore old_indexedstore.IndexedStore, f func(blockIndex uint32, blockHash old_trie.Hash, block *old_state.Block)) {
	totalBlocksCount := lo.Must(srcStore.LatestBlockIndex()) + 1
	printProgress := newProgressPrinter(totalBlocksCount)

	const indexFilePath = "index.bin"
	cli.Logf("Trying to read index from %v", indexFilePath)

	blockTrieRoots, indexFileFound := blockindex.ReadIndexFromFile(indexFilePath)
	if indexFileFound {
		if len(blockTrieRoots) > int(totalBlocksCount) {
			panic(fmt.Errorf("index file contains more entries than there are blocks: %v > %v", len(blockTrieRoots), totalBlocksCount))
		}
		if len(blockTrieRoots) < int(totalBlocksCount) {
			cli.Logf("Index file contains less entries than there are blocks - last %v blocks will be retrieves without indexing: %v < %v",
				len(blockTrieRoots), totalBlocksCount, totalBlocksCount-uint32(len(blockTrieRoots)))
		}

		for i, trieRoot := range blockTrieRoots {
			printProgress(func() uint32 { return uint32(i) })
			block := lo.Must(srcStore.BlockByTrieRoot(trieRoot))
			f(uint32(i), trieRoot, &block)
		}

		cli.Logf("Retrieving next blocks without indexing...")

		for i := uint32(len(blockTrieRoots)); i < totalBlocksCount; i++ {
			printProgress(func() uint32 { return uint32(i) })
			block := lo.Must(srcStore.BlockByIndex(i))
			f(uint32(i), block.TrieRoot(), &block)
		}

		return
	}

	cli.Logf("Index file NOT found at %v, using on-the-fly indexing", indexFilePath)

	// Index file is not available - using on-the-fly indexer
	indexer := blockindex.LoadOrCreate(srcStore)
	printIndexerStats(indexer, srcStore)

	for i := uint32(0); i < totalBlocksCount; i++ {
		printProgress(func() uint32 { return i })

		block, trieRoot := indexer.BlockByIndex(i)
		f(i, trieRoot, &block)
	}
}

func measureTime(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Since(start)
}

func measureTimeAndPrint(descr string, f func()) {
	d := measureTime(f)
	cli.Logf("%v: %v\n", descr, d)
}

func printIndexerStats(indexer *blockindex.BlockIndexer, s old_state.Store) {
	latestBlockIndex := lo.Must(s.LatestBlockIndex())
	measureTimeAndPrint("Time for retrieving block 0", func() { indexer.BlockByIndex(0) })
	measureTimeAndPrint("Time for retrieving block 100", func() { indexer.BlockByIndex(100) })
	measureTimeAndPrint("Time for retrieving block 10000", func() { indexer.BlockByIndex(10000) })
	measureTimeAndPrint("Time for retrieving block 1000000", func() { indexer.BlockByIndex(1000000) })
	measureTimeAndPrint(fmt.Sprintf("Time for retrieving block %v", latestBlockIndex-1000), func() { indexer.BlockByIndex(latestBlockIndex - 1000) })
	measureTimeAndPrint(fmt.Sprintf("Time for retrieving block %v", latestBlockIndex), func() { indexer.BlockByIndex(latestBlockIndex) })
}

func periodicAction(period time.Duration, lastActionTime *time.Time, action func()) {
	if time.Since(*lastActionTime) >= period {
		action()
		*lastActionTime = time.Now()
	}
}

func newProgressPrinter(totalBlocksCount uint32) (printProgress func(getBlockIndex func() uint32)) {
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
			totalBlocksProcessed := totalBlocksCount - blocksLeft
			relProgress := float64(totalBlocksProcessed) / float64(totalBlocksCount)
			estimateRunTime = time.Duration(float64(time.Since(startTime)) / relProgress)
			avgSpeed = int(float64(totalBlocksProcessed) / time.Since(startTime).Seconds())

			recentBlocksProcessed := prevBlocksLeft - blocksLeft
			currentSpeed = int(float64(recentBlocksProcessed) / period.Seconds())
			prevBlocksLeft = blocksLeft
		})

		fmt.Printf("\rBlocks left: %v. Speed: %v blocks/sec. Avg speed: %v blocks/sec. Estimate time left: %v",
			blocksLeft, currentSpeed, avgSpeed, estimateRunTime)
	}
}

func GetAnchorOutput(chainState old_kv.KVStoreReader) *old_iotago.AliasOutput {
	contractState := oldstate.GetContactStateReader(chainState, old_blocklog.Contract.Hname())

	registry := old_collections.NewArrayReadOnly(contractState, old_blocklog.PrefixBlockRegistry)
	if registry.Len() == 0 {
		panic("Block registry is empty")
	}

	blockInfoBytes := registry.GetAt(registry.Len() - 1)

	var blockInfo old_blocklog.BlockInfo
	lo.Must0(blockInfo.Read(bytes.NewReader(blockInfoBytes)))

	return blockInfo.PreviousAliasOutput.GetAliasOutput()
}
