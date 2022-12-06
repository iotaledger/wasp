package reqstatus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

type reqstatusWebAPI struct {
	getChain func(chainID *isc.ChainID) chain.Chain
}

// TODO  add examples for receipt json
func AddEndpoints(server echoswagger.ApiRouter, getChain chains.ChainProvider) {
	r := &reqstatusWebAPI{func(chainID *isc.ChainID) chain.Chain {
		return getChain(chainID)
	}}

	server.GET(routes.RequestReceipt(":chainID", ":reqID"), r.handleRequestReceipt).
		SetSummary("Get the processing status of a given request in the node").
		AddParamPath("", "chainID", "ChainID (bech32)").
		AddParamPath("", "reqID", "Request ID").
		AddResponse(http.StatusOK, "Request Receipt", model.RequestReceiptResponse{}, nil)

	server.GET(routes.WaitRequestProcessed(":chainID", ":reqID"), r.handleWaitRequestProcessed).
		SetSummary("Wait until the given request has been processed by the node").
		AddParamPath("", "chainID", "ChainID (bech32)").
		AddParamPath("", "reqID", "Request ID").
		AddParamBody(model.WaitRequestProcessedParams{}, "Params", "Optional parameters", false).
		AddResponse(http.StatusOK, "Request Receipt", model.RequestReceiptResponse{}, nil)
}

func (r *reqstatusWebAPI) handleRequestReceipt(c echo.Context) error {
	ch, reqID, err := r.parseParams(c)
	if err != nil {
		return err
	}

	blockIndex, err := ch.GetStateReader().LatestBlockIndex()
	if err != nil {
		return httperrors.ServerError("error getting latest chain block index")
	}
	ret, err := chainutil.CallView(blockIndex, ch, blocklog.Contract.Hname(), blocklog.ViewGetRequestReceipt.Hname(),
		dict.Dict{
			blocklog.ParamRequestID: reqID.Bytes(),
		})
	if err != nil {
		return httperrors.ServerError("error calling get receipt view")
	}
	binRec, err := ret.Get(blocklog.ParamRequestRecord)
	if err != nil {
		return httperrors.ServerError("error parsing getReceipt view call result")
	}
	rec, err := blocklog.RequestReceiptFromBytes(binRec)
	if err != nil {
		return httperrors.ServerError("error decoding receipt from getReceipt view call result")
	}
	return resolveReceipt(c, ch, rec)
}

const waitRequestProcessedDefaultTimeout = 30 * time.Second

func (r *reqstatusWebAPI) handleWaitRequestProcessed(c echo.Context) error {
	ch, reqID, err := r.parseParams(c)
	if err != nil {
		return err
	}

	req := model.WaitRequestProcessedParams{
		Timeout: waitRequestProcessedDefaultTimeout,
	}
	if c.Request().Header.Get("Content-Type") == "application/json" {
		if err := c.Bind(&req); err != nil {
			return httperrors.BadRequest("Invalid request body")
		}
	}

	rec := <-ch.AwaitRequestProcessed(c.Request().Context(), reqID)
	return resolveReceipt(c, ch, rec)
}

func (r *reqstatusWebAPI) parseParams(c echo.Context) (chain.Chain, isc.RequestID, error) {
	chainID, err := isc.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return nil, isc.RequestID{}, httperrors.BadRequest(fmt.Sprintf("Invalid Chain ID %+v: %s", c.Param("chainID"), err.Error()))
	}
	theChain := r.getChain(chainID)
	if theChain == nil {
		return nil, isc.RequestID{}, httperrors.NotFound(fmt.Sprintf("Chain not found: %s", chainID.String()))
	}
	reqID, err := isc.RequestIDFromString(c.Param("reqID"))
	if err != nil {
		return nil, isc.RequestID{}, httperrors.BadRequest(fmt.Sprintf("Invalid request id %+v: %s", c.Param("reqID"), err.Error()))
	}
	return theChain, reqID, nil
}

func resolveReceipt(c echo.Context, ch chain.ChainRequests, rec *blocklog.RequestReceipt) error {
	resolvedReceiptErr, err := chainutil.ResolveError(ch, rec.Error)
	if err != nil {
		return httperrors.ServerError("error resolving receipt error")
	}
	iscReceipt := rec.ToISCReceipt(resolvedReceiptErr)
	receiptJSON, err := json.Marshal(iscReceipt)
	if err != nil {
		return httperrors.ServerError("error marshaling receipt into JSON")
	}
	return c.JSON(http.StatusOK,
		&model.RequestReceiptResponse{
			Receipt: string(receiptJSON),
		},
	)
}
