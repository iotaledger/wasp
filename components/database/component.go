package database

import (
	"context"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	hivedb "github.com/iotaledger/hive.go/db"

	"github.com/iotaledger/wasp/v2/packages/chain"
	"github.com/iotaledger/wasp/v2/packages/daemon"
	"github.com/iotaledger/wasp/v2/packages/database"
	"github.com/iotaledger/wasp/v2/packages/readonly"
	"github.com/iotaledger/wasp/v2/packages/registry"
)

func init() {
	Component = &app.Component{
		Name:             "Database",
		DepsFunc:         func(cDeps dependencies) { deps = cDeps },
		Params:           params,
		InitConfigParams: initConfigParams,
		Provide:          provide,
		Configure:        configure,
	}
}

var (
	Component *app.Component
	deps      dependencies
)

type dependencies struct {
	dig.In

	ChainStateDatabaseManager *database.ChainStateDatabaseManager
}

func initConfigParams(c *dig.Container) error {
	type cfgResult struct {
		dig.Out
		DatabaseEngine hivedb.Engine `name:"databaseEngine"`
		ReadOnlyDBPath string
	}

	if err := c.Provide(func() cfgResult {
		dbEngine, err := hivedb.EngineFromStringAllowed(ParamsDatabase.Engine, database.AllowedEngines)
		if err != nil {
			Component.LogPanic(err.Error())
		}

		return cfgResult{
			DatabaseEngine: dbEngine,
			ReadOnlyDBPath: ParamsDatabase.ReadOnlyFilePath,
		}
	}); err != nil {
		Component.LogPanic(err.Error())
	}

	return nil
}

func provide(c *dig.Container) error {
	type databaseManagerDeps struct {
		dig.In

		ChainRecordRegistryProvider registry.ChainRecordRegistryProvider
		DatabaseEngine              hivedb.Engine `name:"databaseEngine"`

		// NodeConnection is essential, even if it doesn't seem to be used.
		// If we don't have that as a dependency, the L1 parameters would be unknown,
		// but those are required in "NewManager"
		NodeConnection chain.NodeConnection
	}

	type chainStateDatabaseManagerResult struct {
		dig.Out

		ChainStateDatabaseManager *database.ChainStateDatabaseManager
	}

	if err := c.Provide(func(deps databaseManagerDeps) chainStateDatabaseManagerResult {
		path := ParamsDatabase.ChainState.Path
		if ParamsDatabase.ReadOnlyFilePath != "" {
			path = readonly.DataDir(ParamsDatabase.ReadOnlyFilePath)
		}
		manager, err := database.NewChainStateDatabaseManager(
			deps.ChainRecordRegistryProvider,
			database.WithEngine(deps.DatabaseEngine),
			database.WithPath(path),
			database.WithCacheSize(ParamsDatabase.ChainState.CacheSize),
		)
		if err != nil {
			Component.LogPanic(err.Error())
		}

		return chainStateDatabaseManagerResult{
			ChainStateDatabaseManager: manager,
		}
	}); err != nil {
		Component.LogPanic(err.Error())
	}

	return nil
}

func configure() error {
	// Create a background worker that marks the database as corrupted at clean startup.
	// This has to be done in a background worker, because the Daemon could receive
	// a shutdown signal during startup. If that is the case, the BackgroundWorker will never be started
	// and the database will never be marked as corrupted.
	if err := Component.Daemon().BackgroundWorker("Database Health", func(_ context.Context) {
		if err := deps.ChainStateDatabaseManager.MarkStoresCorrupted(); err != nil {
			Component.LogPanic(err.Error())
		}
	}, daemon.PriorityDatabaseHealth); err != nil {
		Component.LogPanicf("failed to start worker: %s", err)
	}

	storesCorrupted, err := deps.ChainStateDatabaseManager.AreStoresCorrupted()
	if err != nil {
		Component.LogPanic(err.Error())
	}

	if storesCorrupted && !ParamsDatabase.DebugSkipHealthCheck {
		Component.LogPanic(`
WASP was not shut down properly, the database may be corrupted.
You need to resolve this situation manually.
`)
	}

	correctStoresVersion, err := deps.ChainStateDatabaseManager.CheckCorrectStoresVersion()
	if err != nil {
		Component.LogPanic(err.Error())
	}

	if !correctStoresVersion {
		storesVersionUpdated, err2 := deps.ChainStateDatabaseManager.UpdateStoresVersion()
		if err2 != nil {
			Component.LogPanic(err2.Error())
		}

		if !storesVersionUpdated {
			Component.LogPanic("WASP database version mismatch. The database scheme was updated.")
		}
	}

	if err = Component.Daemon().BackgroundWorker("Close database", func(ctx context.Context) {
		<-ctx.Done()

		if err = deps.ChainStateDatabaseManager.MarkStoresHealthy(); err != nil {
			Component.LogPanic(err.Error())
		}

		Component.LogInfo("Syncing databases to disk ...")
		if err = deps.ChainStateDatabaseManager.FlushAndCloseStores(); err != nil {
			Component.LogPanicf("Syncing databases to disk ... failed: %s", err)
		}
		Component.LogInfo("Syncing databases to disk ... done")
	}, daemon.PriorityCloseDatabase); err != nil {
		Component.LogPanicf("failed to start worker: %s", err)
	}

	return nil
}
