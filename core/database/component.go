// Package database is a plugin that manages the badger database (e.g. garbage collection).
package database

import (
	"context"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/packages/database/dbmanager"
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
}

func provide(c *dig.Container) error {
	type databaseManagerDeps struct {
		dig.In

		RegistryConfig *registry.Config
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
				deps.RegistryConfig,
			),
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
