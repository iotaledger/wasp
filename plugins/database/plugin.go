// Package database is a plugin that manages the badger database (e.g. garbage collection).
package database

import (
	"context"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/database/dbmanager"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/registry"
)

const pluginName = "Database"

var log *logger.Logger

var dbm *dbmanager.DBManager

// Init is an entry point for the plugin.
func Init() *node.Plugin {
	return node.NewPlugin(pluginName, nil, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(pluginName)
	dbm = dbmanager.NewDBManager(logger.NewLogger("dbmanager"), parameters.GetBool(parameters.DatabaseInMemory), registryConfig())

	// we open the database in the configure, so we must also make sure it's closed here
	err := daemon.BackgroundWorker(pluginName, func(ctx context.Context) {
		<-ctx.Done()
		log.Infof("syncing database to disk...")
		dbm.Close()
		log.Infof("syncing database to disk... done")
	}, parameters.PriorityDatabase)
	if err != nil {
		log.Panicf("failed to start a daemon: %s", err)
	}
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(pluginName+"[GC]", dbm.RunGC, parameters.PriorityDBGarbageCollection)
	if err != nil {
		log.Errorf("failed to start as daemon: %s", err)
	}
}

func registryConfig() *registry.Config {
	useText := parameters.GetBool(parameters.RegistryUseText)
	filename := parameters.GetString(parameters.RegistryFile)
	return &registry.Config{UseText: useText, Filename: filename}
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
