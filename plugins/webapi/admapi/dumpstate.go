package admapi

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo"
)

func addStateEndpoints(adm *echo.Group) {
	adm.GET("/"+client.DumpSCStateRoute(":address"), handleDumpSCState)
}

func handleDumpSCState(c echo.Context) error {
	scAddress, err := address.FromBase58(c.Param("address"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid SC address: %s", c.Param("scAddr")))
	}

	virtualState, _, ok, err := state.LoadSolidState(&scAddress)
	if err != nil {
		return err
	}
	if !ok {
		return httperrors.NotFound(fmt.Sprintf("State not found with address %s", scAddress.String()))
	}

	return c.JSON(http.StatusOK, &client.SCStateDump{
		Index:     virtualState.StateIndex(),
		Variables: virtualState.Variables().DangerouslyDumpToMap().ToGoMap(),
	})
}
