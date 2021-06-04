package request

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/plugins/gossip"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func AddEndpoints(server echoswagger.ApiRouter) {
	server.POST(routes.NewRequest(":chainID"), handleNewRequest).
		SetSummary("New off-ledger request").
		AddParamPath("", "request", "binary data").
		AddResponse(http.StatusOK, "Request submitted", nil, nil)
}

func handleNewRequest(c echo.Context) error {
	chainID, offLedgerReq, err := parseParams(c)
	if err != nil {
		return err
	}
	gossip.Gossip().ProcessOffLedgerRequest(chainID, offLedgerReq)
	return nil
}

func parseParams(c echo.Context) (chainID *coretypes.ChainID, req *request.RequestOffLedger, err error) {
	chainID, err = coretypes.ChainIDFromBase58(c.Param("chainID"))
	if err != nil {
		return nil, nil, httperrors.BadRequest(fmt.Sprintf("Invalid Chain ID %+v: %s", c.Param("chainID"), err.Error()))
	}
	data, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return nil, nil, httperrors.BadRequest("Error reading request body")
	}
	req, err = request.NewRequestOffLedgerFromBytes(data)
	if err != nil {
		return nil, nil, httperrors.BadRequest("Error constructing off-ledger request from binary data")
	}
	return chainID, req, err
}
