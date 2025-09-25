package chain

import (
	"runtime"

	"fortio.org/safecast"
	"github.com/spf13/cobra"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/wasp/v2/packages/database"
	"github.com/iotaledger/wasp/v2/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/state/indexedstore"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func initBuildIndex() *cobra.Command {
	var workers uint8

	cmd := &cobra.Command{
		Use:   "build-index <waspdb path> <indexdb destination path>",
		Short: "Builds a new EVM JSONRPC index db",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.HiveLogger()

			waspDBPath := args[0]
			db, err := database.NewReadOnlyDatabase(waspDBPath)
			if err != nil {
				return err
			}

			storeRO, err := state.NewStoreReadonly(db.KVStore())
			if err != nil {
				return err
			}
			waspDBStore := indexedstore.New(storeRO)

			latestIndex, err := waspDBStore.LatestBlockIndex()
			if err != nil {
				return err
			}

			logger.LogInfo("Creating index in parallel mode.")
			logger.LogInfof("Latest block index: %d\n", latestIndex)

			index := jsonrpc.NewIndex(waspDBStore.StateByTrieRoot, hivedb.EngineRocksDB, args[1])

			block, err := waspDBStore.StateByIndex(latestIndex)
			if err != nil {
				return err
			}

			logger.LogInfof("Indexing with %d cores.\n", workers)

			// Right now this callback just returns one established instance of a database kvstore
			// Technically, we can return multiple instances to improve reading times more.
			// This however adds much more strain to the system, so for now keep it like this.
			storeProvider := func() indexedstore.IndexedStore {
				return waspDBStore
			}

			if err := index.IndexAllBlocksInParallel(logger, storeProvider, block.TrieRoot(), workers); err != nil {
				return err
			}
			return nil
		},
	}

	defaultWorkers := (runtime.NumCPU() + 1) / 2
	if defaultWorkers < 2 {
		defaultWorkers = 2
	}

	numWorkers, err := safecast.Convert[uint8](defaultWorkers)
	if err != nil {
		numWorkers = 2
	}

	cmd.Flags().Uint8Var(&workers, "workers", numWorkers, "the amount of parallel block read workers")

	return cmd
}
