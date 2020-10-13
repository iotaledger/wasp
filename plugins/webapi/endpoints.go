package webapi

import (
	"net"
	"strings"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/dkgapi"
	"github.com/iotaledger/wasp/plugins/webapi/info"
	"github.com/iotaledger/wasp/plugins/webapi/request"
	"github.com/iotaledger/wasp/plugins/webapi/state"
	"github.com/labstack/echo"
)

func addEndpoints(adminWhitelist []net.IP) {
	info.AddEndpoints(Server)
	request.AddEndpoints(Server)
	state.AddEndpoints(Server)

	{
		adm := Server.Group("/" + client.AdminRoutePrefix)
		adm.Use(protected(adminWhitelist))

		// dkgapi
		adm.POST("/newdks", dkgapi.HandlerNewDks)
		adm.POST("/aggregatedks", dkgapi.HandlerAggregateDks)
		adm.POST("/commitdks", dkgapi.HandlerCommitDks)
		adm.POST("/signdigest", dkgapi.HandlerSignDigest)
		adm.POST("/exportdkshare", dkgapi.HandlerExportDKShare)
		adm.POST("/importdkshare", dkgapi.HandlerImportDKShare)

		admapi.AddEndpoints(adm)
	}

	log.Infof("added web api endpoints")
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
