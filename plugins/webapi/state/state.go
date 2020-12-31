// access to the solid state of the smart contract
package state

import (
	"fmt"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/jsonable"
	"github.com/iotaledger/wasp/client/statequery"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addStateQueryEndpoint(server echoswagger.ApiRouter) {
	server.GET("/"+client.StateQueryRoute(":chainID"), handleStateQuery).
		SetDeprecated().
		SetSummary("Query the chain state").
		AddParamBody(statequery.Request{}, "query", "Query parameters", true).
		AddResponse(http.StatusOK, "Query result", statequery.Results{}, nil)
}

func handleStateQuery(c echo.Context) error {
	chainID, err := coretypes.NewChainIDFromBase58(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid chain ID: %+v", c.Param("chainID")))
	}

	var req statequery.Request
	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Failed parsing query request params")
	}

	// TODO serialize access to solid state
	state, batch, exist, err := state.LoadSolidState(&chainID)
	if err != nil {
		return err
	}
	if !exist {
		return httperrors.NotFound(fmt.Sprintf("State not found with address %s", chainID.String()))
	}
	txid := batch.StateTransactionID()
	ret := &statequery.Results{
		KeyQueryResults: make([]*statequery.QueryResult, len(req.KeyQueries)),

		StateIndex: state.BlockIndex(),
		Timestamp:  time.Unix(0, state.Timestamp()),
		StateHash:  state.Hash(),
		StateTxId:  jsonable.NewValueTxID(&txid),
		Requests:   make([]*coretypes.RequestID, len(batch.RequestIDs())),
	}
	copy(ret.Requests, batch.RequestIDs())
	vars := state.Variables()
	for i, q := range req.KeyQueries {
		result, err := q.Execute(vars)
		if err != nil {
			return err
		}
		ret.KeyQueryResults[i] = result
	}

	return c.JSON(http.StatusOK, ret)
}
