package admapi

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo"
)

func addStateEndpoints(adm *echo.Group) {
	adm.GET("/"+client.DumpSCStateRoute(":scid"), handleDumpSCState)
}

func handleDumpSCState(c echo.Context) error {
	scid, err := coretypes.NewContractIDFromBase58(c.Param("scid"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid SC id: %s", c.Param("scid")))
	}

	chainID := scid.ChainID()
	virtualState, _, ok, err := state.LoadSolidState(&chainID)
	if err != nil {
		return err
	}
	if !ok {
		return httperrors.NotFound(fmt.Sprintf("State not found for contract %s", scid.String()))
	}

	return c.JSON(http.StatusOK, &client.SCStateDump{
		Index:     virtualState.StateIndex(),
		Variables: virtualState.Variables().DangerouslyDumpToDict().ToGoMap(),
	})
}
