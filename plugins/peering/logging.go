package peering

import "github.com/iotaledger/hive.go/logger"

const modulename = "wasp/peering"

var log *logger.Logger

func initLogger() {
	log = logger.NewLogger(modulename)
}
