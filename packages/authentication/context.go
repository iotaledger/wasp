package authentication

import (
	"github.com/labstack/echo/v4"
)

type ClaimValidator func(claims *WaspClaims) bool
type AccessValidator func(validator ClaimValidator) bool

type AuthContext struct {
	echo.Context

	scheme          string
	isAuthenticated bool
	claims          *WaspClaims
}

func (a *AuthContext) IsAuthenticated() bool {
	return a.isAuthenticated
}

func (a *AuthContext) Scheme() string {
	return a.scheme
}

func (a *AuthContext) IsAllowedTo(validator ClaimValidator) bool {
	if !a.isAuthenticated {
		return false
	}

	if a.scheme == AuthJWT {
		return validator(a.claims)
	}

	// IP Whitelist and Basic Auth will always give access to everything!
	return true
}
