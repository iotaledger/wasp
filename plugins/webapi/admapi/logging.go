package admapi

import "github.com/iotaledger/hive.go/logger"

const modulename = "admapi"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
