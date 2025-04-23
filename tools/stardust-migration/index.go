package main

import (
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

	destKVS := db.ConnectNew(c.Args().Get(0))
	destStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVS))

	latestIndex, err := destStore.LatestBlockIndex()
	if err != nil {
		return err
	}

	index := jsonrpc.NewIndex(func(trieRoot trie.Hash) (state.State, error) {
		return destStore.StateByTrieRoot(trieRoot)
	}, hivedb.EngineRocksDB, c.Args().Get(1))

	for i := uint32(0); i < latestIndex; i += 10000 {
		block, err := destStore.StateByIndex(latestIndex - i)
		if err != nil {
			return err
		}

		err = index.IndexBlock(block.TrieRoot())
		if err != nil {
			return err
		}
	}

	return nil
}
