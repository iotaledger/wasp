package dkgapi

import "github.com/iotaledger/hive.go/logger"

const modulename = "dkgapi"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
