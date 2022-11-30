package node

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/iotaledger/wasp/packages/wasp"

	"github.com/labstack/echo/v4"
)

func (c *Controller) getConfiguration(e echo.Context) error {
	return e.JSON(http.StatusOK, c.config.Koanf().All())
}

func (c *Controller) getPublicInfo(e echo.Context) error {
	return e.JSON(http.StatusOK, wasp.Version)
}

func (c *Controller) getInfo(e echo.Context) error {
	identity := c.peeringService.GetIdentity()
	version := wasp.Version

	return e.JSON(http.StatusOK, &models.InfoResponse{
		Version:   version,
		PublicKey: identity.PublicKey.String(),
		NetID:     identity.NetID,
	})
}
