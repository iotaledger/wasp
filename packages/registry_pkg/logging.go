package registry_pkg

import "github.com/iotaledger/hive.go/logger"

const modulename = "registry"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
