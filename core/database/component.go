package database

import (
	"context"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	hivedb "github.com/iotaledger/hive.go/core/database"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/daemon"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/registry"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:           "Database",
			DepsFunc:       func(cDeps dependencies) { deps = cDeps },
			Params:         params,
			InitConfigPars: initConfigPars,
			Provide:        provide,
			Configure:      configure,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies
)

type dependencies struct {
	dig.In

	DatabaseManager *database.Manager
}

func initConfigPars(c *dig.Container) error {
	type cfgResult struct {
		dig.Out
		DatabaseEngine hivedb.Engine `name:"databaseEngine"`
	}

	if err := c.Provide(func() cfgResult {
		dbEngine, err := hivedb.EngineFromStringAllowed(ParamsDatabase.Engine, database.AllowedEnginesDefault...)
		if err != nil {
			CoreComponent.LogPanic(err)
		}

		return cfgResult{
			DatabaseEngine: dbEngine,
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}

func provide(c *dig.Container) error {
	type databaseManagerDeps struct {
		dig.In

		ChainRecordRegistryProvider registry.ChainRecordRegistryProvider
		DatabaseEngine              hivedb.Engine `name:"databaseEngine"`
	}

	type databaseManagerResult struct {
		dig.Out

		DatabaseManager *database.Manager
	}

	if err := c.Provide(func(deps databaseManagerDeps) databaseManagerResult {
		dbManager, err := database.NewManager(
			deps.ChainRecordRegistryProvider,
			database.WithEngine(deps.DatabaseEngine),
			database.WithDatabasePathConsensusState(ParamsDatabase.ConsensusState.Path),
			database.WithDatabasesPathChainState(ParamsDatabase.ChainState.Path),
		)
		if err != nil {
			CoreComponent.LogPanic(err)
		}

		return databaseManagerResult{
			DatabaseManager: dbManager,
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	type consensusStateDeps struct {
		dig.In

		DatabaseManager *database.Manager
	}

	type consensusStateResult struct {
		dig.Out

		ConsensusStateRegistryProvider cmtLog.Store
	}

	if err := c.Provide(func(deps consensusStateDeps) consensusStateResult {
		return consensusStateResult{
			ConsensusStateRegistryProvider: database.NewConsensusState(deps.DatabaseManager.ConsensusStateKVStore()),
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}

func configure() error {
	// Create a background worker that marks the database as corrupted at clean startup.
	// This has to be done in a background worker, because the Daemon could receive
	// a shutdown signal during startup. If that is the case, the BackgroundWorker will never be started
	// and the database will never be marked as corrupted.
	if err := CoreComponent.Daemon().BackgroundWorker("Database Health", func(_ context.Context) {
		if err := deps.DatabaseManager.MarkStoresCorrupted(); err != nil {
			CoreComponent.LogPanic(err)
		}
	}, daemon.PriorityDatabaseHealth); err != nil {
		CoreComponent.LogPanicf("failed to start worker: %s", err)
	}

	storesCorrupted, err := deps.DatabaseManager.AreStoresCorrupted()
	if err != nil {
		CoreComponent.LogPanic(err)
	}

	if storesCorrupted && !ParamsDatabase.DebugSkipHealthCheck {
		CoreComponent.LogPanic(`
WASP was not shut down properly, the database may be corrupted.
You need to resolve this situation manually.
`)
	}

	correctStoresVersion, err := deps.DatabaseManager.CheckCorrectStoresVersion()
	if err != nil {
		CoreComponent.LogPanic(err)
	}

	if !correctStoresVersion {
		storesVersionUpdated, err := deps.DatabaseManager.UpdateStoresVersion()
		if err != nil {
			CoreComponent.LogPanic(err)
		}

		if !storesVersionUpdated {
			CoreComponent.LogPanic("WASP database version mismatch. The database scheme was updated.")
		}
	}

	if err = CoreComponent.Daemon().BackgroundWorker("Close database", func(ctx context.Context) {
		<-ctx.Done()

		if err = deps.DatabaseManager.MarkStoresHealthy(); err != nil {
			CoreComponent.LogPanic(err)
		}

		CoreComponent.LogInfo("Syncing databases to disk ...")
		if err = deps.DatabaseManager.FlushAndCloseStores(); err != nil {
			CoreComponent.LogPanicf("Syncing databases to disk ... failed: %s", err)
		}
		CoreComponent.LogInfo("Syncing databases to disk ... done")
	}, daemon.PriorityCloseDatabase); err != nil {
		CoreComponent.LogPanicf("failed to start worker: %s", err)
	}

	return nil
}
