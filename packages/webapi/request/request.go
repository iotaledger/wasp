package request

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"golang.org/x/xerrors"
)

type getChainFn func(chainID *chainid.ChainID) chain.ChainCore

func AddEndpoints(server echoswagger.ApiRouter, getChain getChainFn) {
	instance := &offLedgerReqAPI{
		getChain: getChain,
	}
	server.POST(routes.NewRequest(":chainID"), instance.handleNewRequest).
		SetSummary("New off-ledger request").
		AddParamPath("", "request", "request in base64").
		AddResponse(http.StatusAccepted, "Request submitted", nil, nil)
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

	// TODO consider about calling a view to verify the request has been processed
	return c.NoContent(http.StatusAccepted)
}

type OffLedgerRequestBody struct {
	Request string `json:"request"`
}

func parseParams(c echo.Context) (chainID *chainid.ChainID, req *request.RequestOffLedger, err error) {
	chainID, err = chainid.ChainIDFromBase58(c.Param("chainID"))
	if err != nil {
		return nil, nil, httperrors.BadRequest(fmt.Sprintf("Invalid Chain ID %+v: %s", c.Param("chainID"), err.Error()))
	}
	// reqBase64 := c.FormValue("request")
	// req, err = request.NewRequestOffLedgerFromBase64(reqBase64)

	r := new(OffLedgerRequestBody)
	if err = c.Bind(r); err != nil {
		return nil, nil, httperrors.BadRequest("wtf")
	}
	req, err = request.NewRequestOffLedgerFromBase64(r.Request)
	if err != nil {
		return nil, nil, httperrors.BadRequest(fmt.Sprintf("Error constructing off-ledger request from base64 string: \"%s\"", r.Request))
	}
	return chainID, req, err
}
