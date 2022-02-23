package authentication

import (
	"github.com/iotaledger/wasp/packages/users"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func AddBasicAuth(webAPI WebAPI, users *[]users.User) {
	webAPI.Use(middleware.BasicAuth(func(user, password string, c echo.Context) (bool, error) {
		authContext := c.Get("auth").(*AuthContext)

		for _, v := range *users {
			if user == v.Username && password == v.Password {
				authContext.isAuthenticated = true
				return true, nil
			}
		}

		return false, nil
	}))
}
