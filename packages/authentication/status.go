package authentication

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/authentication/shared"

	"github.com/labstack/echo/v4"
)

type StatusWebAPIModel struct {
	config AuthConfiguration
}

func (a *StatusWebAPIModel) handleAuthenticationStatus(c echo.Context) error {
	model := shared.AuthInfoModel{
		Scheme: a.config.Scheme,
	}

	if model.Scheme == AuthJWT {
		model.AuthURL = shared.AuthRoute()
	}

	return c.JSON(http.StatusOK, model)
}

func addAuthenticationStatus(webAPI WebAPI, config AuthConfiguration) {
	c := &StatusWebAPIModel{
		config: config,
	}

	webAPI.GET(shared.AuthInfoRoute(), c.handleAuthenticationStatus)
}
