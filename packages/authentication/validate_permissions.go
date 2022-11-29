package authentication

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func ValidatePermissions(permissions []string) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			authContext := e.Get("auth").(*AuthContext)

			if authContext == nil {
				return e.NoContent(http.StatusUnauthorized)
			}

			for _, permission := range permissions {
				if !authContext.claims.HasPermission(permission) {
					return e.NoContent(http.StatusUnauthorized)
				}
			}

			return next(e)
		}
	}
}
