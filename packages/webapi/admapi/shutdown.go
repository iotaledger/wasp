package admapi

import (
	"net/http"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/plugins/gracefulshutdown"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addShutdownEndpoint(adm echoswagger.ApiGroup) {
	adm.GET("/"+client.ShutdownRoute, handleShutdown).
		SetSummary("Shut down the node")
}

func handleShutdown(c echo.Context) error {
	log.Info("Received a shutdown request from WebAPI.")
	gracefulshutdown.Shutdown()
	return c.String(http.StatusOK, "Shutting down...")
}
