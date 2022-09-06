// Package database is a plugin that manages the badger database (e.g. garbage collection).
package database

import (
	"context"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/wasp/packages/database/dbmanager"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/registry"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:      "Database",
			DepsFunc:  func(cDeps dependencies) { deps = cDeps },
			Params:    params,
			Configure: configure,
			Run:       run,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies

	dbm *dbmanager.DBManager
)

type dependencies struct {
	dig.In
	RegistryConfig *registry.Config
}

func configure() error {
	dbm = dbmanager.NewDBManager(CoreComponent.Logger().Named("dbmanager"), ParamsDatabase.InMemory, ParamsDatabase.Directory, deps.RegistryConfig)

	// we open the database in the configure, so we must also make sure it's closed here
	err := CoreComponent.Daemon().BackgroundWorker(CoreComponent.Name, func(ctx context.Context) {
		<-ctx.Done()
		CoreComponent.LogInfof("syncing database to disk...")
		dbm.Close()
		CoreComponent.LogInfof("syncing database to disk... done")
	}, parameters.PriorityDatabase)
	if err != nil {
		CoreComponent.LogPanicf("failed to start a daemon: %s", err)
	}

	return err
}

func run() error {
	err := CoreComponent.Daemon().BackgroundWorker(CoreComponent.Name+"[GC]", dbm.RunGC, parameters.PriorityDBGarbageCollection)
	if err != nil {
		CoreComponent.LogErrorf("failed to start as daemon: %s", err)
	}

	return err
}

func GetRegistryKVStore() kvstore.KVStore {
	return dbm.GetRegistryKVStore()
}

func GetOrCreateKVStore(chainID *isc.ChainID) kvstore.KVStore {
	return dbm.GetOrCreateKVStore(chainID)
}

func GetKVStore(chainID *isc.ChainID) kvstore.KVStore {
	return dbm.GetKVStore(chainID)
}
