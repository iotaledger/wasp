package committee

import "github.com/iotaledger/hive.go/logger"

const modulename = "wasp/commtypes"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
