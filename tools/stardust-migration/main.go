// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strings"
	"time"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_buffered "github.com/nnikolash/wasp-types-exported/packages/kv/buffered"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
	old_trietest "github.com/nnikolash/wasp-types-exported/packages/trie/test"
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
	// For pprof profilings
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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

	if newChainIDStr == "dummy" {
		// just for easier testing
		newChainIDStr = "0x00000000000000000000000000000000000000000000000000000000000000ff"
	}
	oldChainID := old_isc.ChainID(GetAnchorOutput(lo.Must(srcStore.LatestState())).AliasID)
	newChainID := lo.Must(isc.ChainIDFromString(newChainIDStr))

	lo.Must0(os.MkdirAll(destChainDBDir, 0o755))

	entries := lo.Must(os.ReadDir(destChainDBDir))
	if len(entries) > 0 {
		// TODO: Disabled this check now, so you can run the migrator multiple times for testing
		// log.Fatalf("destination directory is not empty: %v", destChainDBDir)
	}

	destKVS := db.Create(destChainDBDir)
	destStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVS))

	//migrateAllBlocks(srcStore, destStore, oldChainID, newChainID)

	srcState := lo.Must(srcStore.LatestState())
	destStateDraft := destStore.NewOriginStateDraft()

	v := migrations.MigrateRootContract(srcState, destStateDraft)
	migrations.MigrateAccountsContract(v, srcState, destStateDraft, oldChainID, newChainID)
	migrations.MigrateBlocklogContract(srcState, destStateDraft, oldChainID, newChainID)
	// migrations.MigrateGovernanceContract(srcState, destStateDraft)
	migrations.MigrateEVMContract(srcState, destStateDraft)

	destKVS.Flush()
}

// migrateAllBlocks calls migration functions for all mutations of each block.
func migrateAllBlocks(srcStore old_indexedstore.IndexedStore, destStore indexedstore.IndexedStore, oldChainID old_isc.ChainID, newChainID isc.ChainID) {
	var prevL1Commitment *state.L1Commitment

	_oldState := old_buffered.NewBufferedKVStore(NoopKVStoreReader[old_kv.Key]{})
	_ = _oldState

	oldStateStore := old_trietest.NewInMemoryKVStore()
	oldStateTrie := lo.Must(old_trie.NewTrieUpdatable(oldStateStore, old_trie.MustInitRoot(oldStateStore)))
	oldState := &old_state.TrieKVAdapter{oldStateTrie.TrieReader}
	oldStateTriePrevRoot := oldStateTrie.Root()

	newState := NewInMemoryKVStore(true)

	lastPrintTime := time.Now()
	blocksProcessed := 0
	oldSetsProcessed, oldDelsProcessed, newSetsProcessed, newDelsProcessed := 0, 0, 0, 0
	rootMutsProcessed, accountMutsProcessed, blocklogMutsProcessed, govMutsProcessed, evmMutsProcessed := 0, 0, 0, 0, 0

	forEachBlock(srcStore, func(blockIndex uint32, blockHash old_trie.Hash, block old_state.Block) {
		oldMuts := block.Mutations()
		for k, v := range oldMuts.Sets {
			oldStateTrie.Update([]byte(k), v)
		}
		for k := range oldMuts.Dels {
			oldStateTrie.Delete([]byte(k))
		}
		oldStateTrieRoot, _ := oldStateTrie.Commit(oldStateStore)
		lo.Must(old_trie.Prune(oldStateStore, oldStateTriePrevRoot))
		oldStateTriePrevRoot = oldStateTrieRoot

		oldStateMutsOnly := old_buffered.NewBufferedKVStoreForMutations(NoopKVStoreReader[old_kv.Key]{}, oldMuts)

		newState.StartMarking()

		v := migrations.MigrateRootContract(oldState, newState)
		rootMuts := newState.MutationsCount()

		migrations.MigrateAccountsContract(v, oldState, newState, oldChainID, newChainID)
		accountsMuts := newState.MutationsCount() - rootMuts

		migrations.MigrateGovernanceContract(oldState, newState, oldChainID, newChainID)
		governanceMuts := newState.MutationsCount() - rootMuts - accountsMuts

		newState.StopMarking()
		newState.DeleteMarkedIfNotSet()

		migrations.MigrateBlocklogContract(oldStateMutsOnly, newState, oldChainID, newChainID)
		blocklogMuts := newState.MutationsCount() - rootMuts - accountsMuts - governanceMuts

		migrations.MigrateEVMContract(oldStateMutsOnly, newState)
		evmMuts := newState.MutationsCount() - rootMuts - accountsMuts - governanceMuts - blocklogMuts

		newMuts := newState.Commit(true)

		// TODO: time??
		var nextStateDraft state.StateDraft
		if prevL1Commitment == nil {
			nextStateDraft = destStore.NewOriginStateDraft()
		} else {
			// TODO: NewStateDraft, which most likely needs SaveNextBlockInfo for Commit
			nextStateDraft = lo.Must(destStore.NewEmptyStateDraft(prevL1Commitment))
		}
		newMuts.ApplyTo(nextStateDraft)

		// TODO: SaveNextBlockInfo?
		newBlock := destStore.Commit(nextStateDraft)
		prevL1Commitment = newBlock.L1Commitment()

		//Ugly stats code
		blocksProcessed++
		oldSetsProcessed += len(oldMuts.Sets)
		oldDelsProcessed += len(oldMuts.Dels)
		newSetsProcessed += len(newMuts.Sets)
		newDelsProcessed += len(newMuts.Dels)
		rootMutsProcessed += rootMuts
		accountMutsProcessed += accountsMuts
		blocklogMutsProcessed += blocklogMuts
		govMutsProcessed += governanceMuts
		evmMutsProcessed += evmMuts

		periodicAction(3*time.Second, &lastPrintTime, func() {
			cli.Logf("Blocks index: %v", blockIndex)
			cli.Logf("Blocks processed: %v", blocksProcessed)
			cli.Logf("State %v size: old = %v, new = %v", blockIndex, len(oldStateStore), newState.CommittedSize())
			cli.Logf("Mutations per state processed (sets/dels): old = %.1f/%.1f, new = %.1f/%.1f",
				float64(oldSetsProcessed)/float64(blocksProcessed), float64(oldDelsProcessed)/float64(blocksProcessed),
				float64(newSetsProcessed)/float64(blocksProcessed), float64(newDelsProcessed)/float64(blocksProcessed),
			)
			cli.Logf("New mutations per block by contracts:\n\tRoot: %.1f\n\tAccounts: %.1f\n\tBlocklog: %.1f\n\tGovernance: %.1f\n\tEVM: %.1f",
				float64(rootMutsProcessed)/float64(blocksProcessed), float64(accountMutsProcessed)/float64(blocksProcessed),
				float64(blocklogMutsProcessed)/float64(blocksProcessed), float64(govMutsProcessed)/float64(blocksProcessed),
				float64(evmMutsProcessed)/float64(blocksProcessed),
			)

			blocksProcessed = 0
			oldSetsProcessed, oldDelsProcessed, newSetsProcessed, newDelsProcessed = 0, 0, 0, 0
			rootMutsProcessed, accountMutsProcessed, blocklogMutsProcessed, govMutsProcessed, evmMutsProcessed = 0, 0, 0, 0, 0
		})
	})
}

// forEachBlock iterates over all blocks.
// It uses index file index.bin if it is present, otherwise it uses indexing on-the-fly with blockindex.BlockIndexer.
// If index file does not have enough entries, it retrieves the rest of the blocks without indexing.
// Index file is created using stardust-block-indexer tool.
func forEachBlock(srcStore old_indexedstore.IndexedStore, f func(blockIndex uint32, blockHash old_trie.Hash, block old_state.Block)) {
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
			f(uint32(i), trieRoot, block)
		}

		cli.Logf("Retrieving next blocks without indexing...")

		for i := uint32(len(blockTrieRoots)); i < totalBlocksCount; i++ {
			printProgress(func() uint32 { return uint32(i) })
			block := lo.Must(srcStore.BlockByIndex(i))
			f(uint32(i), block.TrieRoot(), block)
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
		f(i, trieRoot, block)
	}
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
