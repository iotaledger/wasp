package state

import "github.com/iotaledger/hive.go/logger"

const modulename = "state"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
