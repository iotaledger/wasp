package clientapi

import "github.com/iotaledger/hive.go/logger"

const modulename = "clientapi"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
