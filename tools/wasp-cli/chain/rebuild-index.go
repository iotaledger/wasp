package chain

import (
	"runtime"

	"github.com/samber/lo"
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
		Run: func(cmd *cobra.Command, args []string) {
			logger := log.HiveLogger()

			waspDBPath := args[0]
			db, err := database.NewDatabase(hivedb.EngineRocksDB, waspDBPath, false, database.CacheSizeDefault)
			log.Check(err)

			waspDBStore := indexedstore.New(lo.Must(state.NewStoreReadonly(db.KVStore())))

			latestIndex, err := waspDBStore.LatestBlockIndex()
			log.Check(err)

			logger.LogInfo("Creating index in parallel mode.")
			logger.LogInfof("Latest block index: %d\n", latestIndex)

			index := jsonrpc.NewIndex(waspDBStore.StateByTrieRoot, hivedb.EngineRocksDB, args[1])

			block, err := waspDBStore.StateByIndex(latestIndex)
			log.Check(err)

			logger.LogInfof("Indexing with %d cores.\n", workers)

			// Right now this callback just returns one established instance of a database kvstore
			// Technically, we can return multiple instances to improve reading times more.
			// This however adds much more strain to the system, so for now keep it like this.
			storeProvider := func() indexedstore.IndexedStore {
				return waspDBStore
			}

			err = index.IndexAllBlocksInParallel(logger, storeProvider, block.TrieRoot(), workers)
			log.Check(err)
		},
	}
	cmd.Flags().Uint8Var(&workers, "workers", uint8(runtime.NumCPU()/2), "the amount of parallel block read workers")

	return cmd
}
