package admapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
)

type ShutdownFunc func()

type shutdownService struct {
	shutdownFunc ShutdownFunc
}

func addShutdownEndpoint(adm echoswagger.ApiGroup, shutdown ShutdownFunc) {
	s := &shutdownService{shutdown}

	adm.GET(routes.Shutdown(), s.handleShutdown).
		SetDeprecated().
		SetSummary("Shut down the node")
}

// handleShutdown gracefully shuts down the server.
// This endpoint is needed for integration tests, because shutting down via an interrupt
// signal does not work on Windows.
func (s *shutdownService) handleShutdown(c echo.Context) error {
	log.Info("Received a shutdown request from WebAPI.")
	s.shutdownFunc()
	return c.String(http.StatusOK, "Shutting down...")
}
