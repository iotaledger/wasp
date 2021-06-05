package state

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/kv/optimism"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func AddEndpoints(server echoswagger.ApiRouter) {
	dictExample := dict.Dict{
		kv.Key("key1"): []byte("value1"),
	}.JSONDict()

	server.GET(routes.CallView(":chainID", ":contractHname", ":fname"), handleCallView).
		SetSummary("Call a view function on a contract").
		AddParamPath("", "chainID", "ChainID (base58-encoded)").
		AddParamPath("", "contractHname", "Contract Hname").
		AddParamPath("getInfo", "fname", "Function name").
		AddParamBody(dictExample, "params", "Parameters", false).
		AddResponse(http.StatusOK, "Result", dictExample, nil)
}

func handleCallView(c echo.Context) error {
	chainID, err := coretypes.ChainIDFromBase58(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid theChain ID: %+v", c.Param("chainID")))
	}
	contractHname, err := coretypes.HnameFromString(c.Param("contractHname"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid contract ID: %+v", c.Param("contractHname")))
	}

	fname := c.Param("fname")

	var params dict.Dict
	if c.Request().Body != nil {
		if err := json.NewDecoder(c.Request().Body).Decode(&params); err != nil {
			return httperrors.BadRequest("Invalid request body")
		}
	}
	theChain := chains.AllChains().Get(chainID)
	if theChain == nil {
		return httperrors.NotFound(fmt.Sprintf("Chain not found: %s", chainID))
	}
	vctx := viewcontext.NewFromChain(theChain)
	var ret dict.Dict
	_ = optimism.RepeatIfUnlucky(func() error {
		ret, err = vctx.CallView(contractHname, coretypes.Hn(fname), params)
		return err
	})
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("View call failed: %v", err))
	}

	return c.JSON(http.StatusOK, ret)
}
