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
	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
)

type reqstatusWebAPI struct {
	getChain               func(chainID isc.ChainID) chain.Chain
	getReceiptFromBlocklog func(ch chain.Chain, reqID isc.RequestID) (*blocklog.RequestReceipt, error)
	resolveReceipt         func(c echo.Context, ch chain.Chain, rec *blocklog.RequestReceipt) error
}

// TODO  add examples for receipt json
func AddEndpoints(server echoswagger.ApiRouter, getChain chains.ChainProvider) {
	r := &reqstatusWebAPI{
		getChain:               getChain,
		getReceiptFromBlocklog: getReceiptFromBlocklog,
		resolveReceipt:         resolveReceipt,
	}

	server.GET(routes.RequestReceipt(":chainID", ":reqID"), r.handleRequestReceipt).
		SetDeprecated().
		SetSummary("Get the processing status of a given request in the node").
		AddParamPath("", "chainID", "ChainID (bech32)").
		AddParamPath("", "reqID", "Request ID").
		AddResponse(http.StatusOK, "Request Receipt", model.RequestReceiptResponse{}, nil)

	server.GET(routes.WaitRequestProcessed(":chainID", ":reqID"), r.handleWaitRequestProcessed).
		SetDeprecated().
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
	rec, err := r.getReceiptFromBlocklog(ch, reqID)
	if err != nil {
		return err
	}
	return r.resolveReceipt(c, ch, rec)
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

	delay := time.NewTimer(req.Timeout)
	select {
	case rec := <-ch.AwaitRequestProcessed(c.Request().Context(), reqID, true):
		if !delay.Stop() {
			// empty the channel to avoid leak
			<-delay.C
		}
		return r.resolveReceipt(c, ch, rec)
	case <-delay.C:
		return httperrors.Timeout("Timeout")
	}
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

func getReceiptFromBlocklog(ch chain.Chain, reqID isc.RequestID) (*blocklog.RequestReceipt, error) {
	latestState, err := ch.LatestState(chain.ActiveOrCommittedState)
	if err != nil {
		return nil, httperrors.ServerError("error getting latest chain state")
	}
	ret, err := chainutil.CallView(latestState, ch, blocklog.Contract.Hname(), blocklog.ViewGetRequestReceipt.Hname(),
		dict.Dict{
			blocklog.ParamRequestID: reqID.Bytes(),
		})
	if err != nil {
		return nil, httperrors.ServerError("error calling get receipt view")
	}
	if ret == nil {
		return nil, nil // not processed yet
	}
	binRec, err := ret.Get(blocklog.ParamRequestRecord)
	if err != nil {
		return nil, httperrors.ServerError("error parsing getReceipt view call result")
	}
	rec, err := blocklog.RequestReceiptFromBytes(binRec)
	if err != nil {
		return nil, httperrors.ServerError("error decoding receipt from getReceipt view call result")
	}
	return rec, nil
}

func resolveReceipt(c echo.Context, ch chain.Chain, rec *blocklog.RequestReceipt) error {
	if rec == nil {
		return httperrors.NotFound("receipt not found")
	}
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
