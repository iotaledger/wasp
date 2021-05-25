// Package database is a plugin that manages the badger database (e.g. garbage collection).
package database

import (
	"sync"

	"github.com/iotaledger/wasp/packages/dbmanager"
	"github.com/iotaledger/wasp/packages/parameters"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
)

const pluginName = "Database"

var (
	log       *logger.Logger
	doOnce    sync.Once
	DBManager *dbmanager.DBManager
)

// Init is an entry point for the plugin.
func Init() *node.Plugin {
	return node.NewPlugin(pluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	doOnce.Do(
		func() {
			log = logger.NewLogger(pluginName)
			DBManager = dbmanager.NewDBManager(log)

			// we open the database in the configure, so we must also make sure it's closed here
			err := daemon.BackgroundWorker(pluginName, func(shutdownSignal <-chan struct{}) {
				<-shutdownSignal
				log.Infof("syncing database to disk...")
				DBManager.Close()
				log.Infof("syncing database to disk... done")
			}, parameters.PriorityDatabase)
			if err != nil {
				log.Panicf("failed to start a daemon: %s", err)
			}
		})
}

func run(_ *node.Plugin) {
	// DBManager.GC()
	err := daemon.BackgroundWorker(pluginName+"[GC]", DBManager.RunGC, parameters.PriorityBadgerGarbageCollection)
	if err != nil {
		log.Errorf("failed to start as daemon: %s", err)
	}
}

// func GetInstance(chainID *coretypes.ChainID) *dbprovider.DBProvider {
// 	if chainID == nil {
// 		chainID = &coretypes.ChainID{} // registry instance
// 	}
// 	if dbInstances[*chainID] != nil {
// 		return dbInstances[*chainID]
// 	}
// 	dbInstances[*chainID] = createInstance(chainID)
// 	return dbInstances[*chainID]
// }

// func createInstance(chainID *coretypes.ChainID) *dbprovider.DBProvider {
// 	if parameters.GetBool(parameters.DatabaseInMemory) {
// 		log.Infof("IN MEMORY DATABASE")
// 		return dbprovider.NewInMemoryDBProvider(log)
// 	} else {
// 		dbDir := parameters.GetString(parameters.DatabaseDir)
// 		instanceDir := dbDir
// 		if chainRecord.SeparateDbFile { // TODO how to get chain record ? maybe from GetRegistryPartition ?
// 			instanceDir = fmt.Sprintf("%s/%s", dbDir, chainID.Base58())
// 		}
// 		return dbprovider.NewPersistentDBProvider(instanceDir, log)
// 	}
// }

// // each key in DB is prefixed with `chainID` | `SC index` | `object type byte`
// // GetPartition returns a Partition, which is a KVStore prefixed with the chain ID.
// func GetPartition(chainID *coretypes.ChainID) kvstore.KVStore {
// 	return GetInstance(chainID).GetPartition(chainID)
// }

// func GetRegistryPartition() kvstore.KVStore {
// 	return GetInstance(nil).GetRegistryPartition()
// }
