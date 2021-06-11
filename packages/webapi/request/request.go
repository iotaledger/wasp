package request

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"golang.org/x/xerrors"
)

type getChainFn func(chainID *coretypes.ChainID) chain.Chain

func AddEndpoints(server echoswagger.ApiRouter, getChain getChainFn) {
	instance := &offLedgerReqAPI{
		getChain: getChain,
	}
	server.POST(routes.NewRequest(":chainID"), instance.handleNewRequest).
		SetSummary("New off-ledger request").
		AddParamPath("", "request", "binary data").
		AddResponse(http.StatusOK, "Request submitted", nil, nil)
}

type offLedgerReqAPI struct {
	getChain getChainFn
}

func (o *offLedgerReqAPI) handleNewRequest(c echo.Context) error {
	chainID, offLedgerReq, err := parseParams(c)
	if err != nil {
		return err
	}

	ch := o.getChain(chainID)
	if ch == nil {
		return xerrors.Errorf("Unknown chain: %s", chainID.Base58())
	}
	ch.ReceiveOffLedgerRequest(offLedgerReq)

	// TODO look into calling a view to verify the request has been processed
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
