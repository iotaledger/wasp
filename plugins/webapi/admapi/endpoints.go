package admapi

import (
	"net"
	"strings"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/client"
	"github.com/labstack/echo"
)

var log *logger.Logger

func initLogger() {
	log = logger.NewLogger("webapi/adm")
}

func AddEndpoints(e *echo.Echo, adminWhitelist []net.IP) {
	initLogger()

	adm := e.Group("/" + client.AdminRoutePrefix)
	adm.Use(protected(adminWhitelist))

	addShutdownEndpoint(adm)
	addPublicKeyEndpoint(adm)
	addChainRecordEndpoints(adm)
	addProgramEndpoints(adm)
	addChainEndpoints(adm)
	addStateEndpoints(adm)
	addDKShareEndpoints(adm)
}

// allow only if the remote address is private or in whitelist
// TODO this is a very basic/limited form of protection
func protected(whitelist []net.IP) echo.MiddlewareFunc {
	isAllowed := func(ip net.IP) bool {
		if ip.IsLoopback() {
			return true
		}
		for _, whitelistedIP := range whitelist {
			if ip.Equal(whitelistedIP) {
				return true
			}
		}
		return false
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			parts := strings.Split(c.Request().RemoteAddr, ":")
			if len(parts) == 2 {
				ip := net.ParseIP(parts[0])
				if ip != nil && isAllowed(ip) {
					return next(c)
				}
			}
			log.Warnf("Blocking request from %s: %s %s", c.Request().RemoteAddr, c.Request().Method, c.Request().RequestURI)
			return echo.ErrUnauthorized
		}
	}
}
