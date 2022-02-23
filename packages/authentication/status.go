package authentication

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type StatusWebAPIModel struct {
	config AuthConfiguration
}

func (a *StatusWebAPIModel) handleAuthenticationStatus(c echo.Context) error {
	model := AuthStatusModel{
		Scheme: a.config.Scheme,
	}

	if model.Scheme == AuthJWT {
		model.AuthURL = AuthRoute()
	}

	return c.JSON(http.StatusOK, model)
}

func addAuthenticationStatus(webAPI WebAPI, config AuthConfiguration) {
	c := &StatusWebAPIModel{
		config: config,
	}

	webAPI.GET(AuthStatusRoute(), c.handleAuthenticationStatus)
}
