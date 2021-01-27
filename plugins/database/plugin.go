// Package database is a plugin that manages the badger database (e.g. garbage collection).
package database

import (
	"errors"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/parameters"
	"sync"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
)

const pluginName = "Database"

var (
	log *logger.Logger

	dbProvider *dbprovider.DBProvider
	doOnce     sync.Once
)

// Init is an entry point for the plugin.
func Init() *node.Plugin {
	return node.NewPlugin(pluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(pluginName)

	err := checkDatabaseVersion()
	if errors.Is(err, ErrDBVersionIncompatible) {
		log.Panicf("The database scheme was updated. Please delete the database folder.\n%s", err)
	}
	if err != nil {
		log.Panicf("Failed to check database version: %s", err)
	}

	// we open the database in the configure, so we must also make sure it's closed here
	err = daemon.BackgroundWorker(pluginName, func(shutdownSignal <-chan struct{}) {
		<-shutdownSignal
		log.Infof("syncing database to disk...")
		dbProvider.Close()
		log.Infof("syncing database to disk... done")
	}, parameters.PriorityDatabase)
	if err != nil {
		log.Panicf("failed to start a daemon: %s", err)
	}
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(pluginName+"[GC]", dbProvider.RunGC, parameters.PriorityBadgerGarbageCollection)
	if err != nil {
		log.Errorf("failed to start as daemon: %s", err)
	}
}

func GetInstance() *dbprovider.DBProvider {
	doOnce.Do(createInstance)
	return dbProvider
}

func createInstance() {
	if parameters.GetBool(parameters.DatabaseInMemory) {
		log.Infof("IN MEMORY DATABASE")
		dbProvider = dbprovider.NewInMemoryDBProvider(log)
	} else {
		dbDir := parameters.GetString(parameters.DatabaseDir)
		dbProvider = dbprovider.NewPersistentDBProvider(dbDir, log)
	}
}

// each key in DB is prefixed with `chainID` | `SC index` | `object type byte`
// GetPartition returns a Partition, which is a KVStore prefixed with the chain ID.
func GetPartition(chainID *coretypes.ChainID) kvstore.KVStore {
	return GetInstance().GetPartition(chainID)
}

func GetRegistryPartition() kvstore.KVStore {
	return GetInstance().GetRegistryPartition()
}
