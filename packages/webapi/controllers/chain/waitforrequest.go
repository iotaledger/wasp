// Package chain defines the methods evm chain in the webapi
package chain

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/v2/packages/vm/gas"
	"github.com/iotaledger/wasp/v2/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
	"github.com/iotaledger/wasp/v2/packages/webapi/params"
)

func (c *Controller) waitForRequestToFinish(e echo.Context) error {
	controllerutils.SetOperation(e, "wait_request")
	const maximumTimeoutSeconds = 60
	const defaultTimeoutSeconds = 30

	requestID, err := params.DecodeRequestID(e)
	if err != nil {
		return err
	}

	timeout := defaultTimeoutSeconds * time.Second

	timeoutInSeconds := e.QueryParam("timeoutSeconds")
	if timeoutInSeconds != "" {
		parsedTimeout, innerErr := strconv.Atoi(timeoutInSeconds)
		if innerErr != nil {
			return innerErr
		}
		if parsedTimeout < 0 {
			parsedTimeout = 0
		}
		if parsedTimeout > maximumTimeoutSeconds {
			parsedTimeout = maximumTimeoutSeconds
		}
		timeout = time.Duration(parsedTimeout) * time.Second
	}
	var waitForL1Confirmation bool
	echo.QueryParamsBinder(e).Bool("waitForL1Confirmation", &waitForL1Confirmation)

	receipt, err := c.chainService.WaitForRequestProcessed(e.Request().Context(), requestID, waitForL1Confirmation, timeout)
	if err != nil {
		return err
	}

	if receipt == nil {
		return e.JSON(http.StatusOK, models.ReceiptResponse{
			RawError:      &models.UnresolvedVMErrorJSON{},
			ErrorMessage:  "",
			GasBudget:     "",
			GasBurned:     "",
			GasFeeCharged: "",
			BlockIndex:    0,
			RequestIndex:  0,
			GasBurnLog:    []gas.BurnRecord{},
		})
	}

	mappedReceipt := models.MapReceiptResponse(receipt)

	return e.JSON(http.StatusOK, mappedReceipt)
}
