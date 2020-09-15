package webapi

import (
	"net/http"

	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/dkgapi"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"

	"github.com/labstack/echo"
)

func addEndpoints() {
	Server.GET("/", IndexRequest)
	// sc api
	Server.POST("/sc/state/query", stateapi.HandlerQueryState)
	Server.POST("/sc/state/request", stateapi.HandlerQueryRequestState)
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
	Server.POST("/adm/sc/:scaddress/activate", admapi.HandlerActivateSC)
	Server.POST("/adm/sc/:scaddress/deactivate", admapi.HandlerDeactivateSC)
	Server.GET("/adm/sc/:scaddress/dumpstate", admapi.HandlerDumpSCState)

	Server.POST("/adm/putprogrammetadata", admapi.HandlerPutProgramMetaData)
	Server.POST("/adm/getprogrammetadata", admapi.HandlerGetProgramMetadata)

	log.Infof("added web api endpoints")
}

// IndexRequest returns INDEX
func IndexRequest(c echo.Context) error {
	return c.String(http.StatusOK, "INDEX")
}
