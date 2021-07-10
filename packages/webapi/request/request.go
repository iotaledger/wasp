package request

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

type getChainFn func(chainID *chainid.ChainID) chain.ChainCore

func AddEndpoints(server echoswagger.ApiRouter, getChain getChainFn) {
	instance := &offLedgerReqAPI{
		getChain: getChain,
	}
	server.POST(routes.NewRequest(":chainID"), instance.handleNewRequest).
		SetSummary("New off-ledger request").
		AddParamPath("", "chainID", "chainID represented in base58").
		AddParamBody(
			model.OffLedgerRequestBody{Request: "base64 string"},
			"Request",
			"Offledger Request encoded in base64. Optinally, the body can be the binary representation of the offledger request, but mime-type must be specified to \"application/octet-stream\"", //nolint:misspell
			false).
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
		return httperrors.NotFound(fmt.Sprintf("Unknown chain: %s", chainID.Base58()))
	}
	ch.ReceiveOffLedgerRequest(offLedgerReq, "")

	return c.NoContent(http.StatusAccepted)
}

func parseParams(c echo.Context) (chainID *chainid.ChainID, req *request.RequestOffLedger, err error) {
	chainID, err = chainid.ChainIDFromBase58(c.Param("chainID"))
	if err != nil {
		return nil, nil, httperrors.BadRequest(fmt.Sprintf("Invalid Chain ID %+v: %s", c.Param("chainID"), err.Error()))
	}

	contentType := c.Request().Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "json") {
		r := new(model.OffLedgerRequestBody)
		if err = c.Bind(r); err != nil {
			return nil, nil, httperrors.BadRequest("Error parsing request from payload")
		}
		req, err = request.OffLedgerFromBytes(r.Request.Bytes())
		if err != nil {
			return nil, nil, httperrors.BadRequest(fmt.Sprintf("Error constructing off-ledger request from base64 string: \"%s\"", r.Request))
		}
		return chainID, req, err
	}

	// binary format
	reqBytes, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return nil, nil, httperrors.BadRequest("Error parsing request from payload")
	}
	req, err = request.OffLedgerFromBytes(reqBytes)
	if err != nil {
		return nil, nil, httperrors.BadRequest("Error parsing request from payload")
	}
	return chainID, req, err
}
