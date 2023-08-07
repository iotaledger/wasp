package authentication

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ValidationError struct {
	MissingPermission string `json:"missingPermission" swagger:"required"`
	Error             string `json:"error" swagger:"required"`
}

func ValidatePermissions(permissions []string) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			auth := e.Get("auth")
			if auth == nil {
				return e.JSON(http.StatusUnauthorized, ValidationError{Error: "Invalid token"})
			}

			authContext, ok := auth.(*AuthContext)
			if !ok {
				return e.JSON(http.StatusUnauthorized, ValidationError{Error: "Invalid token"})
			}

			if authContext.scheme == AuthNone {
				return next(e)
			}

			for _, permission := range permissions {
				if !authContext.claims.HasPermission(permission) {
					return e.JSON(http.StatusUnauthorized, ValidationError{MissingPermission: permission, Error: "Missing permission"})
				}
			}

			return next(e)
		}
	}
}
