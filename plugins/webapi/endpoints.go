package webapi

import (
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/dkgapi"
	"github.com/iotaledger/wasp/plugins/webapi/redirect"
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
	Server.POST("/adm/exportdkshare", dkgapi.HandlerExportDKShare)
	Server.POST("/adm/importdkshare", dkgapi.HandlerImportDKShare)
	// admapi
	Server.POST("/adm/putscdata", admapi.HandlerPutSCData)
	Server.POST("/adm/getscdata", admapi.HandlerGetSCData)
	Server.GET("/adm/getsclist", admapi.HandlerGetSCList)
	Server.GET("/adm/shutdown", admapi.HandlerShutdown)
	Server.POST("/adm/activatesc", admapi.HandlerActivateSC)
	Server.GET("/adm/dumpscstate/:scaddress", admapi.HandlerDumpSCState)
	Server.POST("/adm/putprogrammetadata", admapi.HandlerPutProgramMetaData)
	Server.POST("/adm/getprogrammetadata", admapi.HandlerGetProgramMetadata)
	// redirect to goshimmer
	Server.GET("/utxodb/outputs/:address", redirect.HandleRedirectGetAddressOutputs)
	Server.POST("/utxodb/tx", redirect.HandleRedirectPostTransaction)

	log.Infof("added web api endpoints")
}

// IndexRequest returns INDEX
func IndexRequest(c echo.Context) error {
	return c.String(http.StatusOK, "INDEX")
}
