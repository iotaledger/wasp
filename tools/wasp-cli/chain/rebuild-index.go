package chain

import (
	"errors"
	"os"
	"runtime"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	hivedb "github.com/iotaledger/hive.go/db"
	log2 "github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/packages/database"
	"github.com/iotaledger/wasp/v2/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/state/indexedstore"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func initBuildIndex() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build-index <waspdb path> <indexdb destination path>",
		Short: "Builds a new EVM JSONRPC index db",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			logger := log2.NewLogger()

			waspDbPath := args[0]
			_, err := os.Stat(waspDbPath)
			if errors.Is(err, os.ErrNotExist) {
				logger.LogError("DB path does not exist")
			}

			db, err := database.NewDatabase(hivedb.EngineRocksDB, waspDbPath, false, database.CacheSizeDefault)
			waspDbStore := indexedstore.New(lo.Must(state.NewStoreReadonly(db.KVStore())))

			latestIndex, err := waspDbStore.LatestBlockIndex()
			log.Check(err)

			logger.LogInfo("Creating index in parallel mode.")
			logger.LogInfof("Latest block index: %d\n", latestIndex)

			index := jsonrpc.NewIndex(waspDbStore.StateByTrieRoot, hivedb.EngineRocksDB, args[1])

			block, err := waspDbStore.StateByIndex(latestIndex)
			log.Check(err)

			numCPU := runtime.NumCPU() - 2
			logger.LogInfof("Indexing with %d cores.\n", numCPU)

			// Right now this callback just returns one established instance of a database kvstore
			// Technically, we can return multiple instances to improve reading times more.
			// This however adds much more strain to the system, so for now keep it like this.
			storeProvider := func() indexedstore.IndexedStore {
				return waspDbStore
			}

			err = index.IndexAllBlocksInParallel(logger, storeProvider, block.TrieRoot(), numCPU)
			log.Check(err)
		},
	}

	return cmd
}
