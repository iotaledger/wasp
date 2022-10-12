package node

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (c *Controller) ShutdownNode(e echo.Context) error {

	return e.NoContent(http.StatusOK)
}
