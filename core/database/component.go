// Package database is a plugin that manages the badger database (e.g. garbage collection).
package database

import (
	"context"
	"path"
	"runtime"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/kvstore/rocksdb"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/database/dbmanager"
	journalpkg "github.com/iotaledger/wasp/packages/journal"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/registry"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:      "Database",
			DepsFunc:  func(cDeps dependencies) { deps = cDeps },
			Params:    params,
			Provide:   provide,
			Configure: configure,
			Run:       run,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies
)

type dependencies struct {
	dig.In

	DatabaseManager *dbmanager.DBManager
	// TODO: remove this in the database refactor PR
	JournalDatabase kvstore.KVStore `name:"journalDatabase"`
}

func provide(c *dig.Container) error {
	type databaseManagerDeps struct {
		dig.In

		ChainRecordRegistryProvider registry.ChainRecordRegistryProvider
	}

	type databaseManagerResult struct {
		dig.Out

		DatabaseManager *dbmanager.DBManager
	}

	if err := c.Provide(func(deps databaseManagerDeps) databaseManagerResult {
		return databaseManagerResult{
			DatabaseManager: dbmanager.NewDBManager(
				CoreComponent.App().NewLogger("dbmanager"),
				ParamsDatabase.InMemory,
				ParamsDatabase.Directory,
				deps.ChainRecordRegistryProvider,
			),
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	type consensusJournalResult struct {
		dig.Out

		ConsensusJournalRegistryProvider journal.Provider

		// TODO: remove this in the database refactor PR
		JournalDatabase kvstore.KVStore `name:"journalDatabase"`
	}

	if err := c.Provide(func(deps databaseManagerDeps) consensusJournalResult {
		// TODO: remove this in the database refactor PR
		newRocksDB := func(path string) (*rocksdb.RocksDB, error) {
			opts := []rocksdb.Option{
				rocksdb.IncreaseParallelism(runtime.NumCPU() - 1),
				rocksdb.Custom([]string{
					"periodic_compaction_seconds=43200",
					"level_compaction_dynamic_level_bytes=true",
					"keep_log_file_num=2",
					"max_log_file_size=50000000", // 50MB per log file
				}),
			}

			return rocksdb.CreateDB(path, opts...)
		}

		rocksDB, err := newRocksDB(path.Join(ParamsDatabase.Directory, "journal"))
		if err != nil {
			CoreComponent.LogPanic(err)
		}
		rocksDBKVStore := rocksdb.New(rocksDB)

		return consensusJournalResult{
			JournalDatabase:                  rocksDBKVStore,
			ConsensusJournalRegistryProvider: journalpkg.NewConsensusJournal(rocksDBKVStore),
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}

func configure() error {
	// we open the database in the configure, so we must also make sure it's closed here
	err := CoreComponent.Daemon().BackgroundWorker(CoreComponent.Name, func(ctx context.Context) {
		<-ctx.Done()
		CoreComponent.LogInfof("syncing database to disk...")
		deps.DatabaseManager.Close()
		// TODO: remove this in the database refactor PR
		deps.JournalDatabase.Flush()
		deps.JournalDatabase.Close()
		CoreComponent.LogInfof("syncing database to disk... done")
	}, parameters.PriorityDatabase)
	if err != nil {
		CoreComponent.LogPanicf("failed to start a daemon: %s", err)
	}

	return err
}

func run() error {
	err := CoreComponent.Daemon().BackgroundWorker(CoreComponent.Name+"[GC]", deps.DatabaseManager.RunGC, parameters.PriorityDBGarbageCollection)
	if err != nil {
		CoreComponent.LogErrorf("failed to start as daemon: %s", err)
	}

	return err
}
