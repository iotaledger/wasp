package wal

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/wal"
)

const PluginName = "Wal"

var (
	log *logger.Logger
	w   *wal.WAL
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, nil, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	if !parameters.GetBool(parameters.WALEnabled) {
		return
	}
	log = logger.NewLogger(PluginName)
	walDir := parameters.GetString(parameters.WALDirectory)
	w = wal.New(log, walDir)
}

func run(_ *node.Plugin) {
}

func GetWAL() *wal.WAL {
	return w
}
