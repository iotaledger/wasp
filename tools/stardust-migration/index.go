package main

import (
	"fmt"
	"os"
	"runtime"

	cmd "github.com/urfave/cli/v2"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/db"
)

func startIndexerSequential(rebasedDBPath string, indexerDBPath string) error {
	rebasedDB := db.ConnectNew(rebasedDBPath)
	rebasedDBStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(rebasedDB))

	latestIndex, err := rebasedDBStore.LatestBlockIndex()
	if err != nil {
		return err
	}

	fmt.Printf("Latest block index: %d\n", latestIndex)

	index := jsonrpc.NewIndex(func(trieRoot trie.Hash) (state.State, error) {
		return rebasedDBStore.StateByTrieRoot(trieRoot)
	}, hivedb.EngineRocksDB, indexerDBPath)

	block, err := rebasedDBStore.StateByIndex(latestIndex)
	if err != nil {
		return err
	}

	err = index.IndexBlock(block.TrieRoot())
	if err != nil {
		return err
	}

	return nil
}

func startIndexerParallelized(rebasedDBPath string, indexerDBPath string) error {
	rebasedDBP := db.ConnectNew(rebasedDBPath)
	rebasedDBStoreP := indexedstore.New(state.NewStoreWithUniqueWriteMutex(rebasedDBP))

	storeProvider := func() indexedstore.IndexedStore {
		return rebasedDBStoreP
	}

	latestIndex, err := rebasedDBStoreP.LatestBlockIndex()
	if err != nil {
		return err
	}

	fmt.Printf("Latest block index: %d\n", latestIndex)

	index := jsonrpc.NewIndex(rebasedDBStoreP.StateByTrieRoot, hivedb.EngineRocksDB, indexerDBPath)

	block, err := rebasedDBStoreP.StateByIndex(latestIndex)
	if err != nil {
		return err
	}

	// Taking your cores and moving them somewhere else.
	// Good luck, may god have mercy on your soul
	numCPU := runtime.NumCPU()
	err = index.IndexBlockParallel(storeProvider, block.TrieRoot(), numCPU)
	if err != nil {
		return err
	}

	return nil
}

func createIndex(c *cmd.Context) error {
	go func() {
		<-c.Done()
		cli.Logf("Interrupted")
		os.Exit(1)
	}()

	cli.DebugLoggingEnabled = true

	defer func() {
		if err := recover(); err != nil {
			cli.Logf("Validation panicked")
			panic(err)
		}
	}()

	return startIndexerParallelized(c.Args().Get(0), c.Args().Get(1))
}
