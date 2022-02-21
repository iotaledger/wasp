package authentication

import (
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"net/http"
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

type AuthTestModel struct {
	claims    jwt.Claims
	method    jwt.SigningMethod
	header    map[string]interface{}
	signature string
	valid     bool
}

func (a *StatusWebAPIModel) handleAuthenticatedTest(c echo.Context) error {
	if a.config.Scheme == AuthJWT {
		token := c.Get("jwt").(*jwt.Token)

		if token == nil {
			return c.JSON(http.StatusUnauthorized, jwt.Token{})
		}

		model := AuthTestModel{
			claims:    token.Claims,
			method:    token.Method,
			header:    token.Header,
			signature: token.Signature,
			valid:     token.Valid,
		}

		return c.JSON(http.StatusOK, model)
	}

	// IP and Basic auth should have caught this call beforehand.
	return c.NoContent(http.StatusOK)
}

type LoginModel struct {
	Username string `json:"username" form:"Username"`
	Password string `json:"password" form:"Password"`
}

func addAuthenticationStatus(webAPI WebAPI, config AuthConfiguration) {
	c := &StatusWebAPIModel{
		config: config,
	}

	webAPI.GET(AuthStatusRoute(), c.handleAuthenticationStatus)
	webAPI.GET(AuthTestRoute(), c.handleAuthenticatedTest)
}
