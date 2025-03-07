package main

import (
	"encoding/json"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	cmd "github.com/urfave/cli/v2"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	isc_migration "github.com/iotaledger/wasp/packages/migration"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/tools/stardust-migration/blockindex"
	"github.com/iotaledger/wasp/tools/stardust-migration/migrations"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/db"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_buffered "github.com/nnikolash/wasp-types-exported/packages/kv/buffered"
	old_dict "github.com/nnikolash/wasp-types-exported/packages/kv/dict"
	old_kvdict "github.com/nnikolash/wasp-types-exported/packages/kv/dict"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
	old_errors "github.com/nnikolash/wasp-types-exported/packages/vm/core/errors"
	old_evm "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	old_evmimpl "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/evmimpl"
)

func initMigration(srcChainDBDir, destChainDBDir, overrideNewChainID string, dryRun bool) (
	old_indexedstore.IndexedStore,
	indexedstore.IndexedStore,
	old_isc.ChainID,
	isc.ChainID,
	*isc_migration.PrepareConfiguration,
	state.Block,
	*transaction.StateMetadata,
	func(),
) {
	if srcChainDBDir == "" || destChainDBDir == "" {
		log.Fatalf("source and destination chain database directories must be specified")
	}

	srcChainDBDir = lo.Must(filepath.Abs(srcChainDBDir))
	destChainDBDir = lo.Must(filepath.Abs(destChainDBDir))

	if strings.HasPrefix(destChainDBDir, srcChainDBDir) {
		log.Fatalf("destination database cannot reside inside source database folder")
	}

	srcKVS := db.ConnectOld(srcChainDBDir)
	srcStore := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(srcKVS))

	oldChainID := old_isc.ChainID(GetAnchorOutput(lo.Must(srcStore.LatestState())).AliasID)

	lo.Must0(os.MkdirAll(destChainDBDir, 0o755))

	entries := lo.Must(os.ReadDir(destChainDBDir))
	if len(entries) > 0 {
		// TODO: Disabled this check now, so you can run the migrator multiple times for testing
		// log.Fatalf("destination directory is not empty: %v", destChainDBDir)
	}

	var destStore indexedstore.IndexedStore
	var close func()
	if dryRun {
		destStore = indexedstore.New(state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()))
		close = func() {}
	} else {
		destKVS := db.Create(destChainDBDir)
		destStore = indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVS))
		close = func() { destKVS.Close() }
	}

	migrationConfig := readMigrationConfiguration()
	newChainID := isc.ChainIDFromObjectID(*migrationConfig.AnchorID)

	firstBlock, stateMetadata := initializeMigrateChainState(destStore, migrationConfig.ChainOwner, *migrationConfig.GasCoinID)

	return srcStore, destStore, oldChainID, newChainID, migrationConfig, firstBlock, stateMetadata, close
}

func initializeMigrateChainState(store indexedstore.IndexedStore, stateController *cryptolib.Address, gasCoinObject iotago.ObjectID) (state.Block, *transaction.StateMetadata) {
	initParams := origin.DefaultInitParams(isc.NewAddressAgentID(stateController)).Encode()
	block, stateMetadata := origin.InitChain(allmigrations.LatestSchemaVersion, store, initParams, gasCoinObject, isc.GasCoinTargetValue, isc.BaseTokenCoinInfo)
	return block, stateMetadata
}

func readMigrationConfiguration() *isc_migration.PrepareConfiguration {
	// The wasp-cli migration will have two stages, in which it will write a configuration file once. This needs to be loaded here.
	// For testing, this is not of much relevance but for the real deployment we need real values.
	// So for now return a more or less random configuration

	const debug = false
	if debug {
		// This comes from the default InitChain init params.
		committeeAddress := lo.Must(cryptolib.AddressFromHex("0x92caa380e78d6c4c5229d0be5c1d55d086a56961b83eaf736d8bd16456e1c6d8"))
		chainOwnerAddress := lo.Must(cryptolib.AddressFromHex("0x55d7503847b5484b318e113f98905e4a1b4da50931f96d5b93223e4bae710175"))

		// ChainID == AnchorID (This ID is an existing object on Alphanet)
		chainID := lo.Must(iotago.ObjectIDFromHex("0x64702b66ade80586f6994ab5f3b573ea5977aeac0f1a292fb99ac5ee8a8fbcb1"))
		assetsBagID := lo.Must(iotago.ObjectIDFromHex("0x34dfb08ea4e730bba0e925aef3f53b209b52eb044a4971b2fe27b62984be8c95"))
		gasCoinObjectID := lo.Must(iotago.ObjectIDFromHex("0x0824b5cd76fe0c08ac25d42b875363011e6df0805a76444b933886af26299870"))

		return &isc_migration.PrepareConfiguration{
			ChainOwner:          chainOwnerAddress,
			DKGCommitteeAddress: committeeAddress,
			AnchorID:            chainID,
			GasCoinID:           gasCoinObjectID,
			AssetsBagID:         assetsBagID,
		}
	}

	configBytes, err := os.ReadFile("migration_preparation.json")
	if err != nil {
		panic(fmt.Errorf("error reading migration_preparation.json: %v", err))
	}

	var prepareConfig isc_migration.PrepareConfiguration
	if err := json.Unmarshal(configBytes, &prepareConfig); err != nil {
		panic(fmt.Errorf("error parsing migration_preparation.json: %v", err))
	}

	return &prepareConfig
}

func writeMigrationResult(metadata *transaction.StateMetadata, stateIndex uint32) error {
	result := isc_migration.MigrationResult{
		StateMetadata:    metadata,
		StateMetadataHex: hexutil.Encode(metadata.Bytes()),
		StateIndex:       stateIndex,
	}

	resultJson := lo.Must(json.MarshalIndent(result, "", "  "))

	cli.DebugLogf("Result written:\n%s\n", string(resultJson))

	return os.WriteFile("migration_result.json", resultJson, os.ModePerm)
}

func migrateSingleState(c *cmd.Context) error {
	srcChainDBDir := c.Args().Get(0)
	destChainDBDir := c.Args().Get(1)
	blockIndex, blockIndexSpecified := c.Uint64("index"), c.IsSet("index")
	overrideNewChainID := c.String("new-chain-id")
	dryRun := c.Bool("dry-run")

	srcStore, destStore, oldChainID, newChainID, _, _, stateMetadata, flush := initMigration(srcChainDBDir, destChainDBDir, overrideNewChainID, dryRun)
	defer flush()

	var srcState old_kv.KVStoreReader
	if blockIndexSpecified {
		cli.Logf("Migrating state #%v", blockIndex)
		srcState = lo.Must(srcStore.StateByIndex(uint32(blockIndex)))
	} else {
		cli.Log("Migrating latest state")
		srcState = lo.Must(srcStore.LatestState())
	}

	stateDraft, err := destStore.NewStateDraft(time.Now(), stateMetadata.L1Commitment)
	if err != nil {
		panic(err)
	}

	cli.DebugLoggingEnabled = true

	v := migrations.MigrateRootContract(srcState, stateDraft)

	migrations.MigrateAccountsContract(v, srcState, stateDraft, oldChainID, newChainID)
	migrations.MigrateBlocklogContract(srcState, stateDraft, oldChainID, newChainID, stateMetadata)
	migrations.MigrateGovernanceContract(srcState, stateDraft, oldChainID, newChainID)
	migrations.MigrateEVMContract(srcState, stateDraft)

	newBlock := destStore.Commit(stateDraft)
	destStore.SetLatest(newBlock.TrieRoot())

	return nil
}

// migrateAllBlocks calls migration functions for all mutations of each block.
func migrateAllStates(c *cmd.Context) error {
	srcChainDBDir := c.Args().Get(0)
	destChainDBDir := c.Args().Get(1)
	startBlockIndex := uint32(c.Uint64("from-index"))
	endBlockIndex := uint32(c.Uint64("to-index"))
	overrideNewChainID := c.String("new-chain-id")
	dryRun := c.Bool("dry-run")

	srcStore, destStore, oldChainID, newChainID, _, _, stateMetadata, flush := initMigration(srcChainDBDir, destChainDBDir, overrideNewChainID, dryRun)
	defer flush()

	// // Trie-based state
	// oldStateStore := old_trietest.NewInMemoryKVStore()
	// oldStateTrie := lo.Must(old_trie.NewTrieUpdatable(oldStateStore, old_trie.MustInitRoot(oldStateStore)))
	// oldState := &old_state.TrieKVAdapter{oldStateTrie.TrieReader}
	// oldStateTriePrevRoot := oldStateTrie.Root()

	// // Dict-based state
	//oldState := old_dict.New()

	// // Hybrid-KV-based state
	oldStateStore := old_dict.New()
	oldState := NewPrefixKVStore(oldStateStore, func(key old_kv.Key) [][]byte {
		return utils.GetMapElemPrefixes([]byte(key))
	})

	oldState.RegisterPrefix(old_accounts.PrefixFoundries, old_accounts.Contract.Hname())
	if startBlockIndex != 0 {
		// these are needed only when initial state is non-empty and only on that first block
		oldState.RegisterPrefix(old_errors.PrefixErrorTemplateMap, old_errors.Contract.Hname())
		oldState.RegisterPrefix(old_evmimpl.PrefixPrivileged, old_evm.Contract.Hname(), old_evm.KeyISCMagic)
		oldState.RegisterPrefix(old_evmimpl.PrefixAllowance, old_evm.Contract.Hname(), old_evm.KeyISCMagic)
		oldState.RegisterPrefix(old_evmimpl.PrefixERC20ExternalNativeTokens, old_evm.Contract.Hname(), old_evm.KeyISCMagic)
		oldState.RegisterPrefix("", old_evm.Contract.Hname(), old_evm.KeyEmulatorState)
	}

	cli.Logf("Real from-index: %d", startBlockIndex)

	if startBlockIndex != 0 {
		cli.Logf("Loading state at block index %v", startBlockIndex-1)
		count := 0

		s := lo.Must(srcStore.StateByIndex(startBlockIndex - 1))
		s.Iterate("", func(k old_kv.Key, v []byte) bool {
			//oldStateTrie.Update([]byte(k), v)
			oldState.Set(k, v)
			count++
			cli.UpdateStatusBarf("Loading entries: %v loaded", count)
			return true
		})

		cli.Logf("Loaded %v entries into initial state", count)
	}

	newState := NewInMemoryKVStore(true)

	lastPrintTime := time.Now()
	blocksProcessed := 0
	oldSetsProcessed, oldDelsProcessed, newSetsProcessed, newDelsProcessed := 0, 0, 0, 0
	rootMutsProcessed, accountMutsProcessed, blocklogMutsProcessed, govMutsProcessed, evmMutsProcessed := 0, 0, 0, 0, 0

	var newBlock state.Block
	forEachBlock(srcStore, startBlockIndex, endBlockIndex, func(blockIndex uint32, blockHash old_trie.Hash, block old_state.Block) {
		oldMuts := block.Mutations()
		// for k, v := range oldMuts.Sets {
		// 	oldStateTrie.Update([]byte(k), v)
		// }
		// for k := range oldMuts.Dels {
		// 	oldStateTrie.Delete([]byte(k))
		// }
		// oldStateTrieRoot, _ := oldStateTrie.Commit(oldStateStore)
		// lo.Must(old_trie.Prune(oldStateStore, oldStateTriePrevRoot))
		// oldStateTriePrevRoot = oldStateTrieRoot
		oldMuts.ApplyTo(oldState)

		var oldStateMutsOnly old_kv.KVStoreReader
		if blockIndex == startBlockIndex && startBlockIndex != 0 {
			oldStateMutsOnly = oldState
		} else {
			oldStateMutsOnly = dictKvFromMuts(oldMuts)
		}

		newState.StartMarking()

		v := migrations.MigrateRootContract(oldState, newState)
		rootMuts := newState.MutationsCount()

		migrations.MigrateAccountsContract(v, oldState, newState, oldChainID, newChainID)
		accountsMuts := newState.MutationsCount() - rootMuts

		migrations.MigrateGovernanceContract(oldState, newState, oldChainID, newChainID)
		governanceMuts := newState.MutationsCount() - rootMuts - accountsMuts

		newState.StopMarking()
		newState.DeleteMarkedIfNotSet()

		migratedBlock := migrations.MigrateBlocklogContract(oldStateMutsOnly, newState, oldChainID, newChainID, stateMetadata)
		blocklogMuts := newState.MutationsCount() - rootMuts - accountsMuts - governanceMuts

		migrations.MigrateEVMContract(oldStateMutsOnly, newState)
		evmMuts := newState.MutationsCount() - rootMuts - accountsMuts - governanceMuts - blocklogMuts

		newMuts := newState.Commit(true)

		var nextStateDraft state.StateDraft
		if stateMetadata.L1Commitment == nil || stateMetadata.L1Commitment.IsZero() {
			nextStateDraft = destStore.NewOriginStateDraft()
		} else {
			nextStateDraft = lo.Must(destStore.NewStateDraft(migratedBlock.Timestamp, stateMetadata.L1Commitment))
		}

		newMuts.ApplyTo(nextStateDraft)
		newBlock := destStore.Commit(nextStateDraft)
		destStore.SetLatest(newBlock.TrieRoot())
		stateMetadata.L1Commitment = newBlock.L1Commitment()

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

		if newBlock.StateIndex()%10000 == 0 {
			cli.Logf("Block Index: %d\n", newBlock.StateIndex())
			writeMigrationResult(stateMetadata, newBlock.StateIndex())
		}

		utils.PeriodicAction(3*time.Second, &lastPrintTime, func() {
			cli.Logf("Blocks index: %v", blockIndex)
			cli.Logf("Blocks processed: %v", blocksProcessed)
			//cli.Logf("State %v size: old = %v, new = %v", blockIndex, len(oldStateStore), newState.CommittedSize())
			//cli.Logf("State %v size: old = %v, new = %v", blockIndex, len(oldState), newState.CommittedSize())
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

	cli.Logf("Finished at Index: %d\n", newBlock.StateIndex())
	writeMigrationResult(stateMetadata, newBlock.StateIndex())

	return nil
}

// forEachBlock iterates over all blocks.
// It uses index file index.bin if it is present, otherwise it uses indexing on-the-fly with blockindex.BlockIndexer.
// If index file does not have enough entries, it retrieves the rest of the blocks without indexing.
// Index file is created using stardust-block-indexer tool.
func forEachBlock(srcStore old_indexedstore.IndexedStore, startIndex, endIndex uint32, f func(blockIndex uint32, blockHash old_trie.Hash, block old_state.Block)) {
	latestBlockIndex := lo.Must(srcStore.LatestBlockIndex())

	if startIndex > latestBlockIndex {
		log.Fatalf("start block index %v is greater than the latest block index %v", startIndex, latestBlockIndex)
	}

	if endIndex == 0 {
		endIndex = latestBlockIndex
	} else if endIndex > latestBlockIndex {
		log.Fatalf("end block index %v is greater than the latest block index %v", endIndex, latestBlockIndex)
	}

	totalBlocksCount := (endIndex - startIndex) + 1
	printProgress, clearProgress := cli.NewProgressPrinter("blocks", totalBlocksCount)
	defer clearProgress()

	const indexFilePath = "index.bin"
	cli.Logf("Trying to read index from %v", indexFilePath)

	blockTrieRoots, indexFileFound := blockindex.ReadIndexFromFile(indexFilePath)
	if indexFileFound {
		if len(blockTrieRoots) != int(latestBlockIndex+1) {
			log.Fatalf("index file was created for other database: block in db = %v, index entries = %v", len(blockTrieRoots), latestBlockIndex+1)
		}

		for i := startIndex; i <= endIndex; i++ {
			trieRoot := blockTrieRoots[i]
			printProgress()
			block := lo.Must(srcStore.BlockByTrieRoot(trieRoot))
			f(uint32(i), trieRoot, block)
		}

		return
	}

	cli.Logf("Index file NOT found at %v, using on-the-fly indexing", indexFilePath)

	// Index file is not available - using on-the-fly indexer
	indexer := blockindex.LoadOrCreate(srcStore)
	printIndexerStats(indexer, srcStore)

	for i := startIndex; i <= endIndex; i++ {
		printProgress()
		block, trieRoot := indexer.BlockByIndex(i)
		f(i, trieRoot, block)
	}

}

func printIndexerStats(indexer *blockindex.BlockIndexer, s old_state.Store) {
	latestBlockIndex := lo.Must(s.LatestBlockIndex())
	utils.MeasureTimeAndPrint("Time for retrieving block 0", func() { indexer.BlockByIndex(0) })
	utils.MeasureTimeAndPrint("Time for retrieving block 100", func() { indexer.BlockByIndex(100) })
	utils.MeasureTimeAndPrint("Time for retrieving block 10000", func() { indexer.BlockByIndex(10000) })
	utils.MeasureTimeAndPrint("Time for retrieving block 1000000", func() { indexer.BlockByIndex(1000000) })
	utils.MeasureTimeAndPrint(fmt.Sprintf("Time for retrieving block %v", latestBlockIndex-1000), func() { indexer.BlockByIndex(latestBlockIndex - 1000) })
	utils.MeasureTimeAndPrint(fmt.Sprintf("Time for retrieving block %v", latestBlockIndex), func() { indexer.BlockByIndex(latestBlockIndex) })
}

// Returns KVStoreReader, which will iterate by both Sets and Dels of mutations. For Dels, value will be nil.
func dictKvFromMuts(muts *old_buffered.Mutations) old_kv.KVStoreReader {
	d := old_kvdict.New()
	for k, v := range muts.Sets {
		d[k] = v
	}
	for k := range muts.Dels {
		d[k] = nil
	}

	return d
}
