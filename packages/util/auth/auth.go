package auth

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func AddAuthentication(e *echo.Echo, config map[string]string) {
	if len(config) == 0 {
		return
	}
	scheme, ok := config["scheme"]
	if !ok {
		return
	}
	switch scheme {
	case "basic":
		addBasicAuth(e, config["username"], config["password"])
	default:
		panic(fmt.Sprintf("Unknown auth scheme %s", scheme))
	}
}

func addBasicAuth(e *echo.Echo, username string, password string) {
	e.Use(middleware.BasicAuth(func(u, p string, c echo.Context) (bool, error) {
		return u == username && p == password, nil
	}))
}
