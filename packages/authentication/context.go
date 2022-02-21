package authentication

import (
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type AuthContext struct {
	echo.Context

	Scheme          string
	IsAuthenticated bool
	Claims          JWTAuthClaims
}

func (a *AuthContext) IsPermitted(permission string) bool {
	if a.Scheme == AuthJWT {
		token := a.Get("jwt").(*jwt.Token)
		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			return false
		}

		return claims[permission] == true
	}

	// Only JWT has claims, therefore basic auth and ip whitelisting allows *ANY* call
	return true
}
