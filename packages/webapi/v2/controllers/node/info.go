package node

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/wasp"

	"github.com/labstack/echo/v4"
)

func (c *Controller) getConfiguration(e echo.Context) error {
	return e.JSON(http.StatusOK, c.config.Koanf().All())
}

func (c *Controller) getInfo(e echo.Context) error {
	return e.JSON(http.StatusOK, wasp.Version)
}
