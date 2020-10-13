package admapi

import (
	"net/http"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/plugins/gracefulshutdown"
	"github.com/labstack/echo"
)

func addShutdownEndpoint(server *echo.Group) {
	server.GET("/"+client.ShutdownRoute, handleShutdown)
}

func handleShutdown(c echo.Context) error {
	log.Info("Received a shutdown request from WebAPI.")
	gracefulshutdown.Shutdown()
	return c.String(http.StatusOK, "Shutting down...")
}
