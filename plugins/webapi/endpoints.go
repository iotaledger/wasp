package webapi

import (
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/clientapi"
	"github.com/iotaledger/wasp/plugins/webapi/dkgapi"
	"net/http"

	"github.com/labstack/echo"
)

func addEndpoints() {
	Server.GET("/", IndexRequest)
	// dkgapi
	Server.POST("/adm/newdks", dkgapi.HandlerNewDks)
	Server.POST("/adm/aggregatedks", dkgapi.HandlerAggregateDks)
	Server.POST("/adm/commitdks", dkgapi.HandlerCommitDks)
	Server.POST("/adm/signdigest", dkgapi.HandlerSignDigest)
	Server.POST("/adm/getpubkeyinfo", dkgapi.HandlerGetKeyPubInfo)
	// admapi
	Server.POST("/adm/putscdata", admapi.HandlerPutSCData)
	Server.POST("/adm/getscdata", admapi.HandlerGetSCData)
	Server.GET("/adm/getsclist", admapi.HandlerGetSCList)
	Server.GET("/adm/shutdown", admapi.HandlerShutdown)
	// clientapi
	Server.POST("/client/testreq", clientapi.HandlerTestRequestTx)

	log.Infof("added web api endpoints")
}

// IndexRequest returns INDEX
func IndexRequest(c echo.Context) error {
	return c.String(http.StatusOK, "INDEX")
}
