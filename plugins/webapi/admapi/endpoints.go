package admapi

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/labstack/echo"
)

var log *logger.Logger

func initLogger() {
	log = logger.NewLogger("webapi/adm")
}

func AddEndpoints(server *echo.Group) {
	initLogger()
	addShutdownEndpoint(server)
	addPublicKeyEndpoint(server)
	addBootupEndpoints(server)
	addProgramEndpoints(server)
	addStateEndpoints(server)
}
