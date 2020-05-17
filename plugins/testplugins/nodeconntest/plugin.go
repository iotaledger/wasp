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
	if !ownerAddressOk() {
		log.Errorf("wrong test data. Can't continue")
	} else {
		log.Errorf("test data OK")
	}
	originTx := createOriginTx()
	log.Infof("origin transaction ID = %s", originTx.ID().String())
}
