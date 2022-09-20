package admapi

import (
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

type ShutdownFunc func()

type shutdownService struct {
	shutdown ShutdownFunc
}

func addShutdownEndpoint(adm echoswagger.ApiGroup, shutdown ShutdownFunc) {
	s := &shutdownService{shutdown}

	adm.GET(routes.Shutdown(), s.handleShutdown).
		SetSummary("Shut down the node")
}

// handleShutdown gracefully shuts down the server.
// This endpoint is needed for integration tests, because shutting down via an interrupt
// signal does not work on Windows.
func (s *shutdownService) handleShutdown(c echo.Context) error {
	log.Info("Received a shutdown request from WebAPI.")
	s.shutdown()
	return c.String(http.StatusOK, "Shutting down...")
}
