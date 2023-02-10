package requests

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) waitForRequestToFinish(e echo.Context) error {
	const maximumTimeoutSeconds = 60
	const defaultTimeoutSeconds = 30

	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	requestID, err := params.DecodeRequestID(e)
	if err != nil {
		return err
	}

	timeout := defaultTimeoutSeconds * time.Second

	timeoutInSeconds := e.QueryParam("timeoutSeconds")
	if len(timeoutInSeconds) > 0 {
		parsedTimeout, _ := strconv.Atoi(timeoutInSeconds)

		if err != nil {
			if parsedTimeout > maximumTimeoutSeconds {
				parsedTimeout = maximumTimeoutSeconds
			}

			timeout = time.Duration(parsedTimeout) * time.Second
		}
	}

	receipt, vmError, err := c.chainService.WaitForRequestProcessed(e.Request().Context(), chainID, requestID, timeout)
	if err != nil {
		return err
	}

	mappedReceipt := models.MapReceiptResponse(receipt, vmError)

	return e.JSON(http.StatusOK, mappedReceipt)
}
