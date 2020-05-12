package dispatcher

import "github.com/iotaledger/hive.go/logger"

const modulename = "wasp/dispatcher"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
