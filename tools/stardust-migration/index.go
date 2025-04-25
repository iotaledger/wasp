package main

import (
	"fmt"
	"os"

	cmd "github.com/urfave/cli/v2"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/db"
)

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

	rebasedDB := db.ConnectNew(c.Args().Get(0))
	rebasedDBStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(rebasedDB))

	indexerDBPath := c.Args().Get(1)

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
