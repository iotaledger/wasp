package corecontracts

import (
	"net/http"
	"strconv"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/labstack/echo/v4"
)

type ErrorMessageFormatResponse struct {
	MessageFormat string
}

func (c *Controller) getMessageFormat(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	contractHname, err := isc.HnameFromString(e.Param("contractHname"))
	if err != nil {
		return apierrors.InvalidPropertyError("contractHname", err)
	}

	errorID, err := strconv.ParseUint(e.Param("errorID"), 10, 64)
	if err != nil {
		return apierrors.InvalidPropertyError("errorID", err)
	}

	messageFormat, err := c.errors.GetMessageFormat(chainID, contractHname, uint16(errorID))

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	errorMessageFormatResponse := &ErrorMessageFormatResponse{
		MessageFormat: messageFormat,
	}

	return e.JSON(http.StatusOK, errorMessageFormatResponse)
}
