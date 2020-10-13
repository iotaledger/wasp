package admapi

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/labstack/echo"
)

var log *logger.Logger

func initLogger() {
	log = logger.NewLogger("webapi/adm")
}

func AddEndpoints(adm *echo.Group) {
	initLogger()
	addShutdownEndpoint(adm)
	addPublicKeyEndpoint(adm)
	addBootupEndpoints(adm)
	addProgramEndpoints(adm)
	addSCEndpoints(adm)
	addStateEndpoints(adm)
}
