package authentication

import (
	"crypto/subtle"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/iotaledger/wasp/packages/users"
)

func AddBasicAuth(webAPI WebAPI, userMap map[string]*users.UserData) {
	webAPI.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		authContext := c.Get("auth").(*AuthContext)

		userDetail := userMap[username]

		if userDetail == nil {
			return false, nil
		}

		if subtle.ConstantTimeCompare([]byte(userDetail.Password), []byte(password)) != 0 {
			authContext.isAuthenticated = true
			return true, nil
		}

		return false, nil
	}))
}
