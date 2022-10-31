package node

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (c *Controller) ShutdownNode(e echo.Context) error {
	c.nodeService.ShutdownNode()
	return e.NoContent(http.StatusOK)
}
