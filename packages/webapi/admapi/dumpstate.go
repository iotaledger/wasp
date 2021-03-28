package admapi

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addStateEndpoints(adm echoswagger.ApiGroup) {
	adm.GET(routes.DumpState(":contractID"), handleDumpSCState).
		AddParamPath("", "contractID", "ContractID").
		AddResponse(http.StatusOK, "State dump", model.SCStateDump{}, nil).
		SetSummary("Dump the whole contract state").
		SetDescription("This may be a dangerous operation if the state is too large. Only for testing use!")
}

func handleDumpSCState(c echo.Context) error {
	chainID, err := coretypes.ChainIDFromBase58(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid SC id: %s", c.Param("chainID")))
	}
	contractHname, err := coretypes.HnameFromString(c.Param("contractHname"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid SC id: %s", c.Param("contractHname")))
	}

	virtualState, _, ok, err := state.LoadSolidState(chainID)
	if err != nil {
		return err
	}
	if !ok {
		return httperrors.NotFound(fmt.Sprintf("State not found for chainID %s", chainID.String()))
	}

	vars, err := dict.FromKVStore(subrealm.New(
		virtualState.Variables().DangerouslyDumpToDict(),
		kv.Key(contractHname.Bytes()),
	))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &model.SCStateDump{
		Index:     virtualState.BlockIndex(),
		Variables: vars,
	})
}
