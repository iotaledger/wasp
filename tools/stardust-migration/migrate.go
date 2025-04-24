package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	"github.com/slack-go/slack"
	cmd "github.com/urfave/cli/v2"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_buffered "github.com/nnikolash/wasp-types-exported/packages/kv/buffered"
	old_dict "github.com/nnikolash/wasp-types-exported/packages/kv/dict"
	old_parameters "github.com/nnikolash/wasp-types-exported/packages/parameters"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
	old_errors "github.com/nnikolash/wasp-types-exported/packages/vm/core/errors"
	old_evm "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	old_evmimpl "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/evmimpl"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	old_iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters/parameterstest"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	isc_migration "github.com/iotaledger/wasp/packages/migration"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/tools/stardust-migration/blockindex"
	"github.com/iotaledger/wasp/tools/stardust-migration/bot"
	"github.com/iotaledger/wasp/tools/stardust-migration/migrations"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/db"
)

const (
	inMemoryStatesSavingPeriodBlocks = 20000
)

type migrationOptions struct {
	EnableRefCountCache bool
	ContinueMigration   bool
	DryRun              bool
	ChainOwner          *cryptolib.Address
}

func initMigration(srcChainDBDir, destChainDBDir string, o *migrationOptions) (
	old_indexedstore.IndexedStore,
	indexedstore.IndexedStore,
	old_isc.ChainID,
	*transaction.StateMetadata,
	func(),
) {

	if srcChainDBDir == "" {
		log.Fatalf("source chain database directories must be specified")
	}

	srcChainDBDir = lo.Must(filepath.Abs(srcChainDBDir))

	if !o.DryRun {
		if destChainDBDir == "" {
			log.Fatalf("destination chain database directories must be specified")
		}

		destChainDBDir = lo.Must(filepath.Abs(destChainDBDir))

		if strings.HasPrefix(destChainDBDir, srcChainDBDir) {
			log.Fatalf("destination database cannot reside inside source database folder")
		}

		if o.ContinueMigration {
			if _, err := os.Stat(destChainDBDir); os.IsNotExist(err) {
				log.Fatalf("destination directory does not exist - cannot continue migration: %v", destChainDBDir)
			}

			entries := lo.Must(os.ReadDir(destChainDBDir))
			if len(entries) == 0 {
				log.Fatalf("destination directory is empty - cannot continue migration: %v", destChainDBDir)
			}
		} else {
			lo.Must0(os.MkdirAll(destChainDBDir, 0o755))
			entries := lo.Must(os.ReadDir(destChainDBDir))

			if len(entries) > 0 {
				os.RemoveAll(destChainDBDir)
				//log.Fatalf("destination directory is not empty - cannot create new db: %v", destChainDBDir)
			}
		}
	} else if o.ContinueMigration {
		log.Fatalf("cannot continue migration in dry-run mode")
	}

	srcKVS := db.ConnectOld(srcChainDBDir)
	srcStore := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(srcKVS))
	oldChainID := old_isc.ChainID(GetAnchorOutput(lo.Must(srcStore.LatestState())).AliasID)

	var destStore indexedstore.IndexedStore
	var close func()
	if o.DryRun {
		destStore = indexedstore.New(state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()))
		close = func() {}
	} else if !o.EnableRefCountCache {
		destKVS := db.Create(destChainDBDir)
		destStore = indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVS))
		close = func() { destKVS.Close() }
	} else {
		destKVS := db.Create(destChainDBDir)

		refCountsStore := mapdb.NewMapDB()
		hybridStore := utils.NewHybridKVStore(destKVS, map[string]kvstore.KVStore{
			string([]byte{1, 2}): refCountsStore,
			string([]byte{1, 3}): refCountsStore,
		})

		cli.Logf("Loading all refCount data into memory...")
		n := lo.Must(hybridStore.LoadAllFromDefault())
		cli.Logf("Loaded %v refCount entries", n)

		destStore = indexedstore.New(state.NewStoreWithUniqueWriteMutex(hybridStore))

		close = func() {
			cli.Logf("Flushing all refCount data from memory into db...")
			n := lo.Must(hybridStore.CopyAllToDefault())
			cli.Logf("Flushed %v refCount entries", n)
			destKVS.Close()
		}
	}

	old_parameters.InitL1(&old_parameters.L1Params{
		Protocol: &old_iotago.ProtocolParameters{
			Bech32HRP: old_iotago.PrefixMainnet,
		},
		BaseToken: &old_parameters.BaseToken{
			Decimals: 6,
		},
	})

	var stateMetadata *transaction.StateMetadata
	if o.ContinueMigration {
		stateMetadata = getStateMetadataByIndex(destStore, lo.Must(destStore.LatestBlockIndex()))
	} else {
		stateMetadata = initializeMigrateChainState(destStore, o.ChainOwner, iotago.ObjectID{})
	}

	return srcStore, destStore, oldChainID, stateMetadata, close
}

func initializeMigrateChainState(store indexedstore.IndexedStore, stateController *cryptolib.Address, gasCoinObject iotago.ObjectID) *transaction.StateMetadata {
	initParams := origin.DefaultInitParams(isc.NewAddressAgentID(stateController)).Encode()
	_, stateMetadata := origin.InitChain(allmigrations.SchemaVersionMigratedRebased, store, initParams, gasCoinObject, isc.GasCoinTargetValue, parameterstest.L1Mock)
	return stateMetadata
}

func getStateMetadataByIndex(store indexedstore.IndexedStore, stateIndex uint32) *transaction.StateMetadata {
	state := lo.Must(store.StateByIndex(stateIndex))
	block := lo.Must(store.BlockByIndex(stateIndex))

	stateMetaBytes := GetStateAnchor(state).GetStateMetadata()
	stateMetadata := lo.Must(transaction.StateMetadataFromBytes(stateMetaBytes))
	stateMetadata.L1Commitment = block.L1Commitment()

	return stateMetadata
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

type debugOptions struct {
	DestKeyMustContain    string
	DestValueMustContain  string
	StackTraceMustContain string
}

func dumpMuts(m *old_buffered.Mutations) {
	return
	fmt.Println("")
	for d, _ := range m.Dels {
		if strings.Contains(string(d), emulator.KeyTransactionsByBlockNumber) || strings.Contains(string(d), emulator.KeyReceiptsByBlockNumber) || strings.Contains(string(d), emulator.KeyBlockIndexByTxHash) {
			fmt.Printf("DEL: %s: %v - %v\n", d, d.Hex(), []byte(d))
		}
	}

	for d, _ := range m.Sets {
		if strings.Contains(string(d), emulator.KeyTransactionsByBlockNumber) || strings.Contains(string(d), emulator.KeyReceiptsByBlockNumber) || strings.Contains(string(d), emulator.KeyBlockIndexByTxHash) {
			fmt.Printf("SET: %s: %v - %v\n", d, d.Hex(), []byte(d))
		}
	}
}

// migrateAllBlocks calls migration functions for all mutations of each block.
func migrateAllStates(c *cmd.Context) error {
	srcChainDBDir := c.Args().Get(0)
	destChainDBDir := c.Args().Get(1)
	startBlockIndex := uint32(c.Uint64("from-index"))
	endBlockIndex := uint32(c.Uint64("to-index"))
	skipLoad := c.Bool("skip-load")
	continueMigration := c.Bool("continue")
	disableStateCache := c.Bool("no-state-cache")
	periodStateSave := c.Bool("periodic-state-save")
	enableRefcountCache := c.Bool("refcount-cache")
	useDummyChainOwner := c.Bool("dummy-chain-owner")
	dryRun := c.Bool("dry-run")
	printBlockIdx := c.Bool("print-block-idx")
	debugOpts := debugOptions{
		DestKeyMustContain:    c.String("debug-dest-key"),
		DestValueMustContain:  c.String("debug-dest-value"),
		StackTraceMustContain: c.String("debug-filter-trace"),
	}

	var chainOwner *cryptolib.Address

	if useDummyChainOwner {
		chainOwner = cryptolib.NewEmptyAddress()
	} else {
		chainOwnerStr := c.String("chain-owner")
		if chainOwnerStr == "" {
			panic("--chain-owner is unset! Either set a chain owner or use --dummy-chain-owner")
		}

		chainOwner = lo.Must(cryptolib.NewAddressFromHexString(chainOwnerStr))
	}

	cli.Logf("Setting %s as the chain owner of the whole chain!", chainOwner.String())

	if continueMigration {
		if startBlockIndex != 0 {
			log.Fatalf("cannot continue migration from block index other than what is in the destination database")
		}
		if skipLoad {
			log.Fatalf("cannot skip loading source state when continuing migration")
		}
	}

	srcStore, destStore, oldChainID, stateMetadata, flush := initMigration(srcChainDBDir, destChainDBDir, &migrationOptions{
		ContinueMigration:   continueMigration,
		DryRun:              dryRun,
		EnableRefCountCache: enableRefcountCache,
		ChainOwner:          chainOwner,
	})
	defer flush()

	bot.Get().PostMessage(":running: *Executing All-States Migration*", slack.MsgOptionIconEmoji(":running:"))

	oldStateStore, oldState, newState, startBlockIndex := initInMemoryStates(c.Context, srcStore, destStore, srcChainDBDir, inMemoryStatesOptions{
		StartBlockIndex:   startBlockIndex,
		SkipLoad:          skipLoad,
		ContinueMigration: continueMigration,
		DisableCache:      disableStateCache,
		Debug:             debugOpts,
	})
	if c.Err() != nil {
		cli.Logf("Interrupted before migration started")
		return nil
	}

	var latestDestBlockIndex uint32
	if continueMigration {
		latestDestBlockIndex = lo.Must(destStore.LatestBlockIndex())
		if startBlockIndex <= latestDestBlockIndex {
			cli.Logf("Pre-loaded state index is lower, then actual latest state index in destination db (%v <= %v) - not committing first %v blocks",
				startBlockIndex, latestDestBlockIndex, latestDestBlockIndex-startBlockIndex)
			stateMetadata = getStateMetadataByIndex(destStore, startBlockIndex-1)
		}
	}

	lastPrintTime := time.Now()
	lastProcessedBlockIndex := uint32(0)
	recentlyBlocksProcessed := 0
	oldSetsProcessed, oldDelsProcessed, newSetsProcessed, newDelsProcessed := 0, 0, 0, 0

	forEachBlock(srcStore, startBlockIndex, endBlockIndex, func(blockIndex uint32, blockHash old_trie.Hash, block old_state.Block) bool {
		defer func() {
			if err := recover(); err != nil {
				cli.Logf("Error at block index %v", blockIndex)
				bot.Get().PostMessage(fmt.Sprintf("Error at block index %v", blockIndex))
				utils.PrintLastDBOperations(oldState, newState)
				panic(err)
			}
		}()

		if printBlockIdx {
			cli.Logf("Block Index: %d\n", blockIndex)
		}

		oldMuts := block.Mutations()
		valuesBeforeMutation := oldState.W.ApplyMutations(oldMuts)

		var oldStateMutsOnly old_kv.KVStoreReader
		migrateFullState := startBlockIndex != 0 && blockIndex == startBlockIndex && !continueMigration
		if migrateFullState {
			cli.Logf("Migrating entire state. Reason: startBlockIndex = %v, blockIndex = %v, continueMigration = %v", startBlockIndex, blockIndex, continueMigration)
			oldStateMutsOnly = oldState
		} else {
			oldStateMutsOnly = utils.DictKvFromMuts(oldMuts)
		}

		if blockIndex >= 9999 {
			dumpMuts(oldMuts)
		}

		compositeOldState := struct {
			old_kv.KVReader
			old_kv.KVIterator
		}{
			KVReader:   oldState,
			KVIterator: oldStateMutsOnly,
		}

		migratedBlockTimestamp := migrateBlock(compositeOldState, utils.OnlyReader(valuesBeforeMutation), newState, oldChainID, stateMetadata, chainOwner, migrateFullState)
		newMuts, _ := newState.W.Commit(true)

		if !dryRun {
			nextStateDraft := lo.Must(destStore.NewStateDraft(migratedBlockTimestamp, stateMetadata.L1Commitment))
			newMuts.ApplyTo(nextStateDraft)
			if !continueMigration || blockIndex > latestDestBlockIndex {
				newBlock := destStore.Commit(nextStateDraft)
				lo.Must0(destStore.SetLatest(newBlock.TrieRoot()))
				stateMetadata.L1Commitment = newBlock.L1Commitment()
			} else {
				newBlock := destStore.ExtractBlock(nextStateDraft)
				stateMetadata.L1Commitment = newBlock.L1Commitment()
			}
		}

		lastProcessedBlockIndex = blockIndex

		//Ugly stats code
		recentlyBlocksProcessed++
		oldSetsProcessed += len(oldMuts.Sets)
		oldDelsProcessed += len(oldMuts.Dels)
		newSetsProcessed += len(newMuts.Sets)
		newDelsProcessed += len(newMuts.Dels)

		if blockIndex%10000 == 0 && !dryRun {
			cli.Logf("Block Index: %d\n", blockIndex)
			writeMigrationResult(stateMetadata, blockIndex)
		}
		if periodStateSave && blockIndex != 0 && blockIndex%inMemoryStatesSavingPeriodBlocks == 0 {
			cli.Logf("Pre-saving in-memory states at block index %v", blockIndex)
			saveInMemoryStates(oldStateStore, newState.W, blockIndex, srcChainDBDir)
			deleteInMemoryStateFiles(srcChainDBDir, (blockIndex - inMemoryStatesSavingPeriodBlocks))
		}

		utils.PeriodicAction(3*time.Second, &lastPrintTime, func() {
			cli.Logf("Blocks index: %v", blockIndex)
			cli.Logf("Blocks processed from last print: %v", recentlyBlocksProcessed)
			cli.Logf("State %v size: old = %v, new = %v", blockIndex, len(oldStateStore), newState.W.CommittedSize())
			cli.Logf("Mutations per state processed (sets/dels): old = %.1f/%.1f, new = %.1f/%.1f",
				float64(oldSetsProcessed)/float64(recentlyBlocksProcessed), float64(oldDelsProcessed)/float64(recentlyBlocksProcessed),
				float64(newSetsProcessed)/float64(recentlyBlocksProcessed), float64(newDelsProcessed)/float64(recentlyBlocksProcessed),
			)

			recentlyBlocksProcessed = 0
			oldSetsProcessed, oldDelsProcessed, newSetsProcessed, newDelsProcessed = 0, 0, 0, 0
		})

		if c.Err() != nil {
			cli.Logf("Interrupted after block %v", blockIndex)
		}

		return c.Err() == nil
	})

	cli.Logf("Finished at Index: %d\n", lastProcessedBlockIndex)
	if !dryRun {
		writeMigrationResult(stateMetadata, lastProcessedBlockIndex)
	}

	bot.Get().PostMessage(fmt.Sprintf("All-States migration succeeded at index %d", lastProcessedBlockIndex))

	saveInMemoryStates(oldStateStore, newState.W, lastProcessedBlockIndex, srcChainDBDir)

	return nil
}

func migrateBlock(oldState, valuesBeforeMutation old_kv.KVStoreReader, newState kv.KVStore, oldChainID old_isc.ChainID, stateMetadata *transaction.StateMetadata, chainOwner *cryptolib.Address, skipLoad bool) time.Time {
	v := migrations.MigrateRootContract(oldState, newState)
	blockKeepAmount := migrations.MigrateGovernanceContract(oldState, newState, oldChainID, chainOwner)
	migrations.MigrateErrorsContract(oldState, newState)
	migrations.MigrateAccountsContract(v, valuesBeforeMutation, oldState, newState, oldChainID)
	migratedBlock := migrations.MigrateBlocklogContract(oldState, newState, oldChainID, stateMetadata, chainOwner, blockKeepAmount, !skipLoad)
	migrations.MigrateEVMContract(oldState, newState)

	return migratedBlock.Timestamp
}

type inMemoryStatesOptions struct {
	StartBlockIndex   uint32
	SkipLoad          bool
	ContinueMigration bool
	DisableCache      bool
	Debug             debugOptions
}

func initInMemoryStates(
	ctx context.Context,
	srcStore old_indexedstore.IndexedStore,
	destStore indexedstore.IndexedStore,
	srcChainDBDir string,
	o inMemoryStatesOptions,
) (
	old_dict.Dict,
	*utils.RecordingKVStore[old_kv.Key, *utils.PrefixKVStore, *utils.PrefixKVStore],
	*utils.RecordingKVStore[kv.Key, *utils.InMemoryKVStore, *utils.InMemoryKVStore],
	uint32,
) {
	defer cli.ClearStatusBar()

	if o.ContinueMigration {
		o.StartBlockIndex = lo.Must(destStore.LatestBlockIndex()) + 1
		cli.Logf("Continuing migration from block index %v", o.StartBlockIndex)
	}

	if o.StartBlockIndex != 0 && !o.DisableCache {
		savedSrcStateStore, saverSrcState, savedDestState, loaded := tryLoadInMemoryStates(srcChainDBDir, o.StartBlockIndex-1, o.ContinueMigration)
		if loaded {
			cli.Logf("Loaded in-memory states from disk: blockIndex = %v", o.StartBlockIndex-1)
			setDestStateKeyValidator(savedDestState, o.Debug)
			return savedSrcStateStore, utils.NewRecordingKVStore(saverSrcState), utils.NewRecordingKVStore(savedDestState), o.StartBlockIndex
		}

		cli.Logf("In-memory states not found on disk for block %v", o.StartBlockIndex-1)

		closestAutoSavedBlockIndex := (o.StartBlockIndex - 2) - (o.StartBlockIndex-2)%inMemoryStatesSavingPeriodBlocks
		cli.Logf("Trying to load auto-saved in-memory states from disk for block %v", closestAutoSavedBlockIndex)
		savedSrcStateStore, saverSrcState, savedDestState, loaded = tryLoadInMemoryStates(srcChainDBDir, closestAutoSavedBlockIndex, o.ContinueMigration)
		if loaded {
			cli.Logf("Loaded auto-saved in-memory states from disk: blockIndex = %v", closestAutoSavedBlockIndex)
			setDestStateKeyValidator(savedDestState, o.Debug)
			return savedSrcStateStore, utils.NewRecordingKVStore(saverSrcState), utils.NewRecordingKVStore(savedDestState), closestAutoSavedBlockIndex + 1
		}
	}

	// // Trie-based state
	// oldStateStore := old_trietest.NewInMemoryKVStore()
	// oldStateTrie := lo.Must(old_trie.NewTrieUpdatable(oldStateStore, old_trie.MustInitRoot(oldStateStore)))
	// oldState := &old_state.TrieKVAdapter{oldStateTrie.TrieReader}
	// oldStateTriePrevRoot := oldStateTrie.Root()

	// // Dict-based state
	//oldState := old_dict.New()

	// // Hybrid-KV-based state
	oldStateStore := old_dict.New()
	oldState := initSrcState(oldStateStore, o.StartBlockIndex != 0)
	newState := utils.NewInMemoryKVStore(false)
	setDestStateKeyValidator(newState, o.Debug)

	if o.StartBlockIndex != 0 {
		cli.Logf("Real from-index: %d", o.StartBlockIndex)

		var preloadStateIdx uint32
		if o.SkipLoad {
			cli.Logf("Loading of source state is SKIPPED - resulting database will be INVALID")
			// Still preloading at least block 0, because it has old initial state.
		} else {
			preloadStateIdx = o.StartBlockIndex - 1
		}

		cli.Logf("Loading source state at block index %v", preloadStateIdx)
		count := 0

		s := lo.Must(srcStore.StateByIndex(preloadStateIdx))
		s.Iterate("", func(k old_kv.Key, v []byte) bool {
			//oldStateTrie.Update([]byte(k), v)
			oldState.Set(k, v)
			count++
			cli.UpdateStatusBarf("Loading entries: %v loaded", count)
			return ctx.Err() == nil
		})

		cli.Logf("Loaded %v entries into in-memory source state", count)
	}

	if o.ContinueMigration {
		cli.Logf("Loading destination state at block index %v", o.StartBlockIndex-1)
		count := 0

		latestDestState := lo.Must(destStore.LatestState())
		latestDestState.Iterate("", func(k kv.Key, v []byte) bool {
			newState.Set(k, v)
			count++
			cli.UpdateStatusBarf("Loading entries: %v loaded", count)
			return ctx.Err() == nil
		})

		_, _ = newState.Commit(false)

		cli.Logf("Loaded %v entries into in-memory destination state", count)
	}

	if o.StartBlockIndex != 0 {
		cli.Logf("Saving in-memory states to disk to avoid loading them next time")
		saveInMemoryStates(oldStateStore, newState, o.StartBlockIndex-1, srcChainDBDir)
	}

	return oldStateStore, utils.NewRecordingKVStore(oldState), utils.NewRecordingKVStore(newState), o.StartBlockIndex
}

func saveInMemoryStates(oldState old_dict.Dict, newState *utils.InMemoryKVStore, lastProcessedBlockIndex uint32, srcChainDBDir string) {
	cli.Logf("Saving in-memory states to disk: blockIndex = %v, srcChainDBDir = %v", lastProcessedBlockIndex, srcChainDBDir)
	oldStateFilePath, newStateFilePath := getInMemoryStateFilePaths(srcChainDBDir, lastProcessedBlockIndex)

	// Doing this in separate steps to avoid too high memory usage (maybe its too naive, but won't hurt)
	cli.Logf("Marshaling old in-memory state: size = %v", len(oldState))
	oldStateBytes := bcs.MustMarshal(&oldState)

	cli.Logf("Writing old in-memory state to disk: path = %v, size = %v", oldStateFilePath, len(oldStateBytes))
	lo.Must0(os.WriteFile(oldStateFilePath, oldStateBytes, os.ModePerm))
	oldStateBytes = nil

	if newState == nil || newState.Len() == 0 {
		cli.Logf("Not saving new in-memory state to disk")
	} else {
		cli.Logf("Marshaling new in-memory state: size = %v", newState.Len())
		committedState, committedMarks := newState.CommittedState()
		s := lo.T2(committedState, committedMarks)
		newStateBytes := bcs.MustMarshal(&s)

		cli.Logf("Writing new in-memory state to disk: path = %v, size = %v", newStateFilePath, len(newStateBytes))
		lo.Must0(os.WriteFile(newStateFilePath, newStateBytes, os.ModePerm))
	}
}

func tryLoadInMemoryStates(srcChainDBDir string, blockIndex uint32, loadDestState bool) (old_dict.Dict, *utils.PrefixKVStore, *utils.InMemoryKVStore, bool) {
	cli.Logf("Trying to load in-memory states from disk: blockIndex = %v, srcChainDBDir = %v", blockIndex, srcChainDBDir)
	oldStateFilePath, newStateFilePath := getInMemoryStateFilePaths(srcChainDBDir, blockIndex)

	cli.Logf("Read old in-memory state from disk: path = %v", oldStateFilePath)
	oldStateBytes, err := os.ReadFile(oldStateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			cli.Logf("Old in-memory state file not found: path = %v", oldStateFilePath)
			return nil, nil, nil, false
		}
		panic(err)
	}

	cli.Logf("Unmarshaling old in-memory state: size = %v", len(oldStateBytes))
	oldStateStore := old_dict.New()
	bcs.MustUnmarshalInto(oldStateBytes, &oldStateStore)
	oldStateBytes = nil
	cli.Logf("Old in-memory state loaded: size = %v", len(oldStateStore))

	cli.Logf("Indexing old state prefixes...")
	oldState := initSrcState(oldStateStore, true)
	oldState.IndexRecords()

	newState := utils.NewInMemoryKVStore(false)

	if !loadDestState {
		cli.Logf("Not loading destination state")
		return oldStateStore, oldState, newState, true
	}

	cli.Logf("Read new in-memory state from disk: path = %v", newStateFilePath)
	newStateBytes, err := os.ReadFile(newStateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			cli.Logf("New in-memory state file not found: path = %v", newStateFilePath)
			return nil, nil, nil, false
		}
		panic(err)
	}

	cli.Logf("Unmarshaling new in-memory state: size = %v", len(newStateBytes))

	var committedState map[kv.Key][]byte
	var committedMarks map[kv.Key]struct{}
	r := bytes.NewReader(newStateBytes)
	bcs.MustUnmarshalStreamInto(r, &committedState)
	bcs.MustUnmarshalStreamInto(r, &committedMarks)

	newState.SetCommittedState(committedState, committedMarks)
	cli.Logf("New in-memory state loaded: size = %v", newState.Len())

	return oldStateStore, oldState, newState, true
}

func deleteInMemoryStateFiles(srcChainDBDir string, blockIndex uint32) {
	cli.Logf("Deleting obsolete files of in-memory states: blockIndex = %v, srcChainDBDir = %v", blockIndex, srcChainDBDir)
	oldStateFilePath, newStateFilePath := getInMemoryStateFilePaths(srcChainDBDir, blockIndex)
	if err := os.Remove(oldStateFilePath); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	} else {
		cli.Logf("Deleted in-memory old state file: path = %v", oldStateFilePath)
	}

	if err := os.Remove(newStateFilePath); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	} else {
		cli.Logf("Deleted in-memory new state file: path = %v", newStateFilePath)
	}
}

func initSrcState(store old_dict.Dict, willBePrefilled bool) *utils.PrefixKVStore {
	state := utils.NewPrefixKVStore(store, func(key old_kv.Key) [][]byte {
		return utils.GetMapElemPrefixes([]byte(key))
	})

	state.RegisterPrefix(old_accounts.PrefixBaseTokens, old_accounts.Contract.Hname())
	state.RegisterPrefix(old_accounts.PrefixNativeTokens, old_accounts.Contract.Hname())
	state.RegisterPrefix(old_accounts.PrefixFoundries, old_accounts.Contract.Hname())
	state.RegisterPrefix(old_errors.PrefixErrorTemplateMap, old_errors.Contract.Hname())

	if willBePrefilled {
		// These are needed only when initial state is non-empty and only on that first block (when full state is used instead of mutations).
		state.RegisterPrefix(old_evmimpl.PrefixPrivileged, old_evm.Contract.Hname(), old_evm.KeyISCMagic)
		state.RegisterPrefix(old_evmimpl.PrefixAllowance, old_evm.Contract.Hname(), old_evm.KeyISCMagic)
		state.RegisterPrefix(old_evmimpl.PrefixERC20ExternalNativeTokens, old_evm.Contract.Hname(), old_evm.KeyISCMagic)
		state.RegisterPrefix("", old_evm.Contract.Hname(), old_evm.KeyEmulatorState)
	}

	return state
}

func setDestStateKeyValidator(s *utils.InMemoryKVStore, o debugOptions) {
	if o.DestKeyMustContain == "" && o.DestValueMustContain == "" {
		return
	}

	var keyBytes []byte
	if o.DestKeyMustContain != "" {
		if !strings.HasPrefix(o.DestKeyMustContain, "0x") {
			o.DestKeyMustContain = "0x" + o.DestKeyMustContain
		}
		keyBytes = lo.Must(hexutil.Decode(o.DestKeyMustContain))
	}

	var valueBytes []byte
	if o.DestValueMustContain != "" {
		if !strings.HasPrefix(o.DestValueMustContain, "0x") {
			o.DestValueMustContain = "0x" + o.DestValueMustContain
		}
		valueBytes = lo.Must(hexutil.Decode(o.DestValueMustContain))
	}

	s.SetKeyValidator(func(k kv.Key, v []byte) {
		keyContains := keyBytes == nil || bytes.Contains([]byte(k), keyBytes)
		valueContains := valueBytes == nil || bytes.Contains(v, valueBytes)

		if !keyContains || !valueContains {
			return
		}

		stack := string(debug.Stack())
		if o.StackTraceMustContain != "" && !strings.Contains(stack, o.StackTraceMustContain) {
			return
		}

		cli.Logf("Record key and/or value contains specified bytes: searchedKeyBytes = %v, searchedValueBytes = %v, key = %x / %v, value = %x / %v:\n%v",
			o.DestKeyMustContain, o.DestValueMustContain, k, string(k), v, string(v), string(debug.Stack()))
	})
}

func getInMemoryStateFilePaths(srcChainDBDir string, blockIndex uint32) (string, string) {
	srcChainDBDir = lo.Must(filepath.Abs(srcChainDBDir))
	fileNameHash := hashing.HashStrings(srcChainDBDir)
	oldStateFilePath := fmt.Sprintf("%v/stardust_migration_block_%v_old_state_%v.bin", os.TempDir(), blockIndex, fileNameHash.Hex())
	newStateFilePath := fmt.Sprintf("%v/stardust_migration_block_%v_new_state_%v.bin", os.TempDir(), blockIndex, fileNameHash.Hex())

	return oldStateFilePath, newStateFilePath
}

// forEachBlock iterates over all blocks.
// It uses index file index.bin if it is present, otherwise it uses indexing on-the-fly with blockindex.BlockIndexer.
// If index file does not have enough entries, it retrieves the rest of the blocks without indexing.
// Index file is created using stardust-block-indexer tool.
func forEachBlock(srcStore old_indexedstore.IndexedStore, startIndex, endIndex uint32, f func(blockIndex uint32, blockHash old_trie.Hash, block old_state.Block) bool) {
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

	blockIndex := blockindex.New(srcStore)
	bot.Get().PostMessage(fmt.Sprintf("Migrating from: *%d*, to: *%d*", startIndex, endIndex))

	for i := startIndex; i <= endIndex; i++ {
		printProgress()
		block, trieRoot := blockIndex.BlockByIndex(i)
		if !f(i, trieRoot, block) {
			return
		}
	}
}
