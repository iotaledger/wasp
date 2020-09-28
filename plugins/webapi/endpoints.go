package webapi

import (
	"bytes"
	"net"
	"strings"

	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/dkgapi"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"
	"github.com/labstack/echo"
)

func addEndpoints(adminWhitelist []net.IP) {
	{
		sc := Server.Group("/sc")

		sc.POST("/state/query", stateapi.HandlerQueryState)
		sc.POST("/state/request", stateapi.HandlerQueryRequestState)
	}

	{
		adm := Server.Group("/adm")
		adm.Use(protected(adminWhitelist))

		// dkgapi
		adm.POST("/newdks", dkgapi.HandlerNewDks)
		adm.POST("/aggregatedks", dkgapi.HandlerAggregateDks)
		adm.POST("/commitdks", dkgapi.HandlerCommitDks)
		adm.POST("/signdigest", dkgapi.HandlerSignDigest)
		adm.POST("/getpubkeyinfo", dkgapi.HandlerGetKeyPubInfo)
		adm.POST("/exportdkshare", dkgapi.HandlerExportDKShare)
		adm.POST("/importdkshare", dkgapi.HandlerImportDKShare)

		adm.POST("/putscdata", admapi.HandlerPutSCData)
		adm.POST("/getscdata", admapi.HandlerGetSCData)
		adm.GET("/getsclist", admapi.HandlerGetSCList)
		adm.GET("/shutdown", admapi.HandlerShutdown)
		adm.POST("/sc/:scaddress/activate", admapi.HandlerActivateSC)
		adm.POST("/sc/:scaddress/deactivate", admapi.HandlerDeactivateSC)
		adm.GET("/sc/:scaddress/dumpstate", admapi.HandlerDumpSCState)

		adm.POST("/putprogrammetadata", admapi.HandlerPutProgramMetaData)
		adm.POST("/getprogrammetadata", admapi.HandlerGetProgramMetadata)
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
			if bytes.Equal(ip, whitelistedIP) {
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
