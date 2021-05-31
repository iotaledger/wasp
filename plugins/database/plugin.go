// Package databaseplugin is a plugin that manages the badger database (e.g. garbage collection).
package databaseplugin

import (
	"github.com/iotaledger/wasp/packages/database/dbmanager"
	"github.com/iotaledger/wasp/packages/parameters"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
)

const pluginName = "Database"

var log *logger.Logger

// Init is an entry point for the plugin.
func Init() *node.Plugin {
	return node.NewPlugin(pluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(pluginName)

	// we open the database in the configure, so we must also make sure it's closed here
	err := daemon.BackgroundWorker(pluginName, func(shutdownSignal <-chan struct{}) {
		<-shutdownSignal
		log.Infof("syncing database to disk...")
		dbmanager.Instance().Close()
		log.Infof("syncing database to disk... done")
	}, parameters.PriorityDatabase)
	if err != nil {
		log.Panicf("failed to start a daemon: %s", err)
	}
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(pluginName+"[GC]", dbmanager.Instance().RunGC, parameters.PriorityBadgerGarbageCollection)
	if err != nil {
		log.Errorf("failed to start as daemon: %s", err)
	}
}
