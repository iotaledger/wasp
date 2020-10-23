// access to the solid state of the smart contract
package state

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"net/http"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/jsonable"
	"github.com/iotaledger/wasp/client/statequery"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo"
)

func AddEndpoints(server *echo.Echo) {
	server.GET("/"+client.StateQueryRoute(":scAddr"), handleStateQuery)
}

func handleStateQuery(c echo.Context) error {
	scAddr, err := address.FromBase58(c.Param("scAddr"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid SC address: %+v", c.Param("scAddr")))
	}

	var req statequery.Request
	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Failed parsing query request params")
	}

	// TODO serialize access to solid state
	state, batch, exist, err := state.LoadSolidState(&scAddr)
	if err != nil {
		return err
	}
	if !exist {
		return httperrors.NotFound(fmt.Sprintf("State not found with address %s", scAddr.String()))
	}
	txid := batch.StateTransactionId()
	ret := &statequery.Results{
		KeyQueryResults: make([]*statequery.QueryResult, len(req.KeyQueries)),

		StateIndex: state.StateIndex(),
		Timestamp:  time.Unix(0, state.Timestamp()),
		StateHash:  state.Hash(),
		StateTxId:  jsonable.NewValueTxID(&txid),
		Requests:   make([]*coretypes.RequestID, len(batch.RequestIds())),
	}
	copy(ret.Requests, batch.RequestIds())
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
