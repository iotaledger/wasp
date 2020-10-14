package dkg

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/labstack/echo"
)

var log *logger.Logger

func initLogger() {
	log = logger.NewLogger("webapi/dkg")
}

func AddEndpoints(adm *echo.Group) {
	initLogger()
	addExportEndpoints(adm)
}
