package node

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
)

func (c *Controller) getConfiguration(e echo.Context) error {
	configMap := make(map[string]interface{})

	for k, v := range c.config.Koanf().All() {
		if !strings.HasPrefix(k, "users") {
			configMap[k] = v
		}
	}

	return e.JSON(http.StatusOK, configMap)
}

func (c *Controller) getPublicInfo(e echo.Context) error {
	return e.JSON(http.StatusOK, c.waspVersion)
}

func (c *Controller) getInfo(e echo.Context) error {
	identity := c.peeringService.GetIdentity()
	l1Params := parameters.L1()

	return e.JSON(http.StatusOK, &models.InfoResponse{
		Version:   c.waspVersion,
		PublicKey: identity.PublicKey.String(),
		NetID:     identity.NetID,
		L1Params:  models.MapL1Params(l1Params),
	})
}
