package admapi

import (
	"net/http"

	"github.com/iotaledger/wasp/plugins/gracefulshutdown"
	"github.com/labstack/echo"
)

func HandlerShutdown(c echo.Context) error {
	log.Info("Received a shutdown request from WebAPI.")
	gracefulshutdown.Shutdown()
	return c.String(http.StatusOK, "Shutting down...")
}
