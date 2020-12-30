package request

import (
	"fmt"
	"net/http"
	"time"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func AddEndpoints(server echoswagger.ApiRouter) {
	server.GET("/"+client.RequestStatusRoute(":chainID", ":reqID"), handleRequestStatus).
		SetSummary("Get the processing status of a given request in the node").
		AddParamPath("", "chainID", "ChainID (base58)").
		AddParamPath("", "reqID", "Request ID (base58)").
		AddResponse(http.StatusOK, "Request status", client.RequestStatusResponse{}, nil)

	server.GET("/"+client.WaitRequestProcessedRoute(":chainID", ":reqID"), handleWaitRequestProcessed).
		SetSummary("Wait until the given request has been processed by the node").
		AddParamPath("", "chainID", "ChainID (base58)").
		AddParamPath("", "reqID", "Request ID (base58)").
		AddParamBody(client.WaitRequestProcessedParams{}, "Params", "Optional parameters", false)
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
	return c.JSON(http.StatusOK, client.RequestStatusResponse{
		IsProcessed: isProcessed,
	})
}

func handleWaitRequestProcessed(c echo.Context) error {
	ch, reqID, err := parseParams(c)
	if err != nil {
		return err
	}

	req := client.WaitRequestProcessedParams{
		Timeout: client.WaitRequestProcessedDefaultTimeout,
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
	chainID, err := coretypes.NewChainIDFromBase58(c.Param("chainID"))
	if err != nil {
		return nil, nil, httperrors.BadRequest(fmt.Sprintf("Invalid chain ID %+v: %s", c.Param("chainID"), err.Error()))
	}
	chain := chains.GetChain(chainID)
	if chain == nil {
		return nil, nil, httperrors.NotFound(fmt.Sprintf("Chain not found: %+v", chainID.String()))
	}
	reqID, err := coretypes.NewRequestIDFromBase58(c.Param("reqID"))
	if err != nil {
		return nil, nil, httperrors.BadRequest(fmt.Sprintf("Invalid request id %+v: %s", c.Param("reqID"), err.Error()))
	}
	return chain, &reqID, nil
}
