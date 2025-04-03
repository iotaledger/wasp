package node

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/models"
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
	return e.JSON(http.StatusOK, models.VersionResponse{Version: c.waspVersion})
}

func (c *Controller) getInfo(e echo.Context) error {
	identity := c.peeringService.GetIdentity()
	l1Params, err := c.nodeService.L1Params(e.Request().Context())
	if err != nil {
		return apierrors.NewHTTPError(http.StatusInternalServerError, err.Error(), nil)
	}

	return e.JSON(http.StatusOK, &models.InfoResponse{
		Version:    c.waspVersion,
		PublicKey:  identity.PublicKey.String(),
		PeeringURL: identity.PeeringURL,
		L1Params:   l1Params,
	})
}
