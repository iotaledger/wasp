// scmeta package runs integration tests by calling WebAPi to itself for SC meta data
package nodeconntest

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/plugins/testplugins"
)

// PluginName is the name of the database plugin.
const PluginName = "TestingNodeConn"

var (
	// Plugin is the plugin instance of the database plugin.
	Plugin = node.NewPlugin(PluginName, testplugins.Status(PluginName), configure, run)
	log    *logger.Logger
)

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	orig1, _ := testplugins.CreateOriginData(testplugins.SC1, nil)
	orig2, _ := testplugins.CreateOriginData(testplugins.SC2, nil)
	orig3, _ := testplugins.CreateOriginData(testplugins.SC3, nil)
	log.Infof("origin transaction SC1 ID = %s", orig1.ID().String())
	log.Infof("origin transaction SC2 ID = %s", orig2.ID().String())
	log.Infof("origin transaction SC3 ID = %s", orig3.ID().String())
}
