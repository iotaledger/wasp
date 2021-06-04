package gossip

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/offledger"
)

const pluginName = "Gossip"

var log *logger.Logger

var gossip *offledger.Gossip

// Init is an entry point for the plugin.
func Init() *node.Plugin {
	return node.NewPlugin(pluginName, node.Enabled, configure)
}

func configure(_ *node.Plugin) {
	gossip = offledger.NewGossip()
}

func Gossip() *offledger.Gossip {
	return gossip
}
