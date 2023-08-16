package authentication

import "github.com/labstack/echo/v4"

type AuthContext struct {
	echo.Context

	scheme string
	claims *WaspClaims
	name   string
}

func (a *AuthContext) Name() string {
	return a.name
}

func (a *AuthContext) Scheme() string {
	return a.scheme
}
