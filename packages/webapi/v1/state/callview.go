package state

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
)

type callViewService struct {
	chains chains.Provider
}

func AddEndpoints(server echoswagger.ApiRouter, allChains chains.Provider) {
	dictExample := dict.Dict{
		"key1": []byte("value1"),
	}.JSONDict()

	s := &callViewService{allChains}

	server.POST(routes.CallViewByName(":chainID", ":contractHname", ":fname"), s.handleCallViewByName).
		SetDeprecated().
		SetSummary("Call a view function on a contract by name").
		AddParamPath("", "chainID", "ChainID").
		AddParamPath("", "contractHname", "Contract Hname").
		AddParamPath("getInfo", "fname", "Function name").
		AddParamBody(dictExample, "params", "Parameters", false).
		AddResponse(http.StatusOK, "Result", dictExample, nil)

	server.GET(routes.CallViewByName(":chainID", ":contractHname", ":fname"), s.handleCallViewByName).
		SetDeprecated().
		SetSummary("Call a view function on a contract by name").
		AddParamPath("", "chainID", "ChainID").
		AddParamPath("", "contractHname", "Contract Hname").
		AddParamPath("getInfo", "fname", "Function name").
		AddParamBody(dictExample, "params", "Parameters", false).
		AddResponse(http.StatusOK, "Result", dictExample, nil)

	server.POST(routes.CallViewByHname(":chainID", ":contractHname", ":functionHname"), s.handleCallViewByHname).
		SetDeprecated().
		SetSummary("Call a view function on a contract by Hname").
		AddParamPath("", "chainID", "ChainID").
		AddParamPath("", "contractHname", "Contract Hname").
		AddParamPath("getInfo", "functionHname", "Function Hname").
		AddParamBody(dictExample, "params", "Parameters", false).
		AddResponse(http.StatusOK, "Result", dictExample, nil)

	server.GET(routes.CallViewByHname(":chainID", ":contractHname", ":functionHname"), s.handleCallViewByHname).
		SetDeprecated().
		SetSummary("Call a view function on a contract by Hname").
		AddParamPath("", "chainID", "ChainID").
		AddParamPath("", "contractHname", "Contract Hname").
		AddParamPath("getInfo", "functionHname", "Function Hname").
		AddParamBody(dictExample, "params", "Parameters", false).
		AddResponse(http.StatusOK, "Result", dictExample, nil)

	server.GET(routes.StateGet(":chainID", ":key"), s.handleStateGet).
		SetDeprecated().
		SetSummary("Fetch the raw value associated with the given key in the chain state").
		AddParamPath("", "chainID", "ChainID").
		AddParamPath("", "key", "Key (hex-encoded)").
		AddResponse(http.StatusOK, "Result", []byte("value"), nil)
}

func (s *callViewService) handleCallView(c echo.Context, functionHname isc.Hname) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid chain ID: %+v", c.Param("chainID")))
	}
	contractHname, err := isc.HnameFromHexString(c.Param("contractHname"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid contract ID: %+v", c.Param("contractHname")))
	}

	var params dict.Dict
	if c.Request().Body != http.NoBody {
		if err2 := json.NewDecoder(c.Request().Body).Decode(&params); err2 != nil {
			return httperrors.BadRequest("Invalid request body")
		}
	}
	theChain := s.chains().Get(chainID)
	if theChain == nil {
		return httperrors.NotFound(fmt.Sprintf("Chain not found: %s", chainID))
	}
	// TODO should blockIndex be an optional parameter of this endpoint?
	latestState, err := theChain.LatestState(chain.ActiveOrCommittedState)
	if err != nil {
		return httperrors.ServerError(fmt.Sprintf("View call failed: %v", err))
	}
	ret, err := chainutil.CallView(latestState, theChain, contractHname, functionHname, params)
	if err != nil {
		return httperrors.ServerError(fmt.Sprintf("View call failed: %v", err))
	}

	return c.JSON(http.StatusOK, ret)
}

func (s *callViewService) handleCallViewByName(c echo.Context) error {
	fname := c.Param("fname")
	return s.handleCallView(c, isc.Hn(fname))
}

func (s *callViewService) handleCallViewByHname(c echo.Context) error {
	functionHname, err := isc.HnameFromHexString(c.Param("functionHname"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid function ID: %+v", c.Param("functionHname")))
	}
	return s.handleCallView(c, functionHname)
}

func (s *callViewService) handleStateGet(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid chain ID: %+v", c.Param("chainID")))
	}

	key, err := iotago.DecodeHex(c.Param("key"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("cannot parse hex-encoded key: %+v", c.Param("key")))
	}

	theChain := s.chains().Get(chainID)
	if theChain == nil {
		return httperrors.NotFound(fmt.Sprintf("Chain not found: %s", chainID))
	}

	state, err := theChain.LatestState(chain.ActiveOrCommittedState)
	if err != nil {
		reason := fmt.Sprintf("View call failed: %v", err)
		return httperrors.ServerError(reason)
	}

	ret, err := state.Get(kv.Key(key))
	if err != nil {
		reason := fmt.Sprintf("View call failed: %v", err)
		return httperrors.ServerError(reason)
	}

	return c.JSON(http.StatusOK, ret)
}
