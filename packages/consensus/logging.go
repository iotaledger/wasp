package consensus

import "github.com/iotaledger/hive.go/logger"

const modulename = "wasp/consensus"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
