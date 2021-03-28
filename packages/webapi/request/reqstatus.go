package request

import (
	"fmt"
	"net/http"
	"time"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func AddEndpoints(server echoswagger.ApiRouter) {
	server.GET(routes.RequestStatus(":chainID", ":reqID"), handleRequestStatus).
		SetSummary("Get the processing status of a given request in the node").
		AddParamPath("", "chainID", "ChainID (base58)").
		AddParamPath("", "reqID", "Request ID (base58)").
		AddResponse(http.StatusOK, "Request status", model.RequestStatusResponse{}, nil)

	server.GET(routes.WaitRequestProcessed(":chainID", ":reqID"), handleWaitRequestProcessed).
		SetSummary("Wait until the given request has been processed by the node").
		AddParamPath("", "chainID", "ChainID (base58)").
		AddParamPath("", "reqID", "Request ID (base58)").
		AddParamBody(model.WaitRequestProcessedParams{}, "Params", "Optional parameters", false)
}

func handleRequestStatus(c echo.Context) error {
	ch, reqID, err := parseParams(c)
	if err != nil {
		return err
	}
	var isProcessed bool
	switch ch.GetRequestProcessingStatus(reqID) {
	case chain.RequestProcessingStatusCompleted:
		isProcessed = true
	case chain.RequestProcessingStatusBacklog:
		isProcessed = false
	}
	return c.JSON(http.StatusOK, model.RequestStatusResponse{
		IsProcessed: isProcessed,
	})
}

func handleWaitRequestProcessed(c echo.Context) error {
	ch, reqID, err := parseParams(c)
	if err != nil {
		return err
	}

	req := model.WaitRequestProcessedParams{
		Timeout: model.WaitRequestProcessedDefaultTimeout,
	}
	if c.Request().Header.Get("Content-Type") == "application/json" {
		if err := c.Bind(&req); err != nil {
			return httperrors.BadRequest("Invalid request body")
		}
	}

	if ch.GetRequestProcessingStatus(reqID) == chain.RequestProcessingStatusCompleted {
		// request is already processed, no need to wait
		return nil
	}

	// subscribe to event
	requestProcessed := make(chan bool)
	handler := events.NewClosure(func(rid coretypes.RequestID) {
		if rid == *reqID {
			requestProcessed <- true
		}
	})
	ch.EventRequestProcessed().Attach(handler)
	defer ch.EventRequestProcessed().Detach(handler)

	select {
	case <-requestProcessed:
		return nil
	case <-time.After(req.Timeout):
		// check again, in case event was triggered just before we subscribed
		if ch.GetRequestProcessingStatus(reqID) == chain.RequestProcessingStatusCompleted {
			return nil
		}
		return httperrors.Timeout("Timeout while waiting for request to be processed")
	}
}

func parseParams(c echo.Context) (chain.Chain, *coretypes.RequestID, error) {
	chainID, err := coretypes.ChainIDFromBase58(c.Param("chainID"))
	if err != nil {
		return nil, nil, httperrors.BadRequest(fmt.Sprintf("Invalid theChain ID %+v: %s", c.Param("chainID"), err.Error()))
	}
	theChain := chains.AllChains().Get(chainID)
	if theChain == nil {
		return nil, nil, httperrors.NotFound(fmt.Sprintf("Chain not found: %s", chainID.String()))
	}
	reqID, err := coretypes.RequestIDFromBase58(c.Param("reqID"))
	if err != nil {
		return nil, nil, httperrors.BadRequest(fmt.Sprintf("Invalid request id %+v: %s", c.Param("reqID"), err.Error()))
	}
	return theChain, &reqID, nil
}
