// +build ignore

// access to the solid state of the smart contract
package state

import (
	"fmt"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/model/statequery"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addStateQueryEndpoint(server echoswagger.ApiRouter) {
	server.GET(routes.StateQuery(":chainID"), handleStateQuery).
		SetDeprecated().
		SetSummary("Query the chain state").
		AddParamPath("", "chainID", "ChainID (base58)").
		AddParamBody(statequery.Request{}, "query", "Query parameters", true).
		AddResponse(http.StatusOK, "Query result", statequery.Results{}, nil)
}

func handleStateQuery(c echo.Context) error {
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid chain ID: %+v", c.Param("chainID")))
	}

	var req statequery.Request
	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Failed parsing query request params")
	}

	// TODO serialize access to solid state
	state, exist, err := state.LoadSolidState(&chainID)
	if err != nil {
		return err
	}
	if !exist {
		return httperrors.NotFound(fmt.Sprintf("State not found with address %s", chainID.String()))
	}
	txid := batch.ApprovingOutputID()
	ret := &statequery.Results{
		KeyQueryResults: make([]*statequery.QueryResult, len(req.KeyQueries)),

		StateIndex: state.BlockIndex(),
		Timestamp:  time.Unix(0, state.Timestamp()),
		StateHash:  state.Hash(),
		StateTxId:  model.NewValueTxID(&txid),
		Requests:   make([]iscp.RequestID, len(batch.RequestIDs())),
	}
	copy(ret.Requests, batch.RequestIDs())
	vars := state.KVStore()
	for i, q := range req.KeyQueries {
		result, err := q.Execute(vars)
		if err != nil {
			return err
		}
		ret.KeyQueryResults[i] = result
	}

	return c.JSON(http.StatusOK, ret)
}
