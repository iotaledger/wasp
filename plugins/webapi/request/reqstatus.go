package request

import (
	"fmt"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo"
)

func AddEndpoints(server *echo.Echo) {
	server.GET("/"+client.RequestStatusRoute(":chainId", ":reqId"), handleRequestStatus)
	server.GET("/"+client.WaitRequestProcessedRoute(":chainId", ":reqId"), handleWaitRequestProcessed)
}

func handleRequestStatus(c echo.Context) error {
	ch, reqId, err := parseParams(c)
	if err != nil {
		return err
	}
	var isProcessed bool
	switch ch.GetRequestProcessingStatus(reqId) {
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
	ch, reqId, err := parseParams(c)
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

	start := time.Now()
	for time.Since(start) < req.Timeout {
		if ch.GetRequestProcessingStatus(reqId) == chain.RequestProcessingStatusCompleted {
			return nil
		}
		// TODO subscribe to EventRequestProcessed event instead of polling
		time.Sleep(1 * time.Second)
	}
	return httperrors.Timeout("Timeout while waiting for request to be processed")
}

func parseParams(c echo.Context) (chain.Chain, *coretypes.RequestID, error) {
	chainId, err := coretypes.NewChainIDFromBase58(c.Param("chainId"))
	if err != nil {
		return nil, nil, httperrors.BadRequest(fmt.Sprintf("Invalid chain ID %+v: %s", c.Param("chainId"), err.Error()))
	}
	chain := chains.GetChain(chainId)
	if chain == nil {
		return nil, nil, httperrors.NotFound(fmt.Sprintf("Chain not found: %+v", chainId.String()))
	}
	reqId, err := coretypes.NewRequestIDFromBase58(c.Param("reqId"))
	if err != nil {
		return nil, nil, httperrors.BadRequest(fmt.Sprintf("Invalid request id %+v: %s", c.Param("reqId"), err.Error()))
	}
	return chain, &reqId, nil
}
