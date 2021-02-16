// +build ignore

package dashboard

import (
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

func handleWwalletJson(c echo.Context) error {
	settings := viper.AllSettings()
	delete(settings, "wallet")
	return c.JSONPretty(200, settings, " ")
}
