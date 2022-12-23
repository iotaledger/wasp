package authentication

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/iotaledger/hive.go/core/basicauth"
	"github.com/iotaledger/wasp/packages/users"
)

func AddBasicAuth(webAPI WebAPI, userManager *users.UserManager) {
	webAPI.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		authContext := c.Get("auth").(*AuthContext)

		user, err := userManager.User(username)
		if err != nil {
			return false, err
		}

		valid, err := basicauth.VerifyPassword([]byte(password), user.PasswordSalt, user.PasswordHash)
		if err != nil {
			return false, fmt.Errorf("failed to verify password: %w", err)
		}

		if !valid {
			return false, nil
		}

		authContext.isAuthenticated = true
		return true, nil
	}))
}
