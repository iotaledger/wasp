package corecontracts

import (
	"net/http"

	"fortio.org/safecast"
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/corecontracts"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

type ErrorMessageFormatResponse struct {
	MessageFormat string `json:"messageFormat" swagger:"required"`
}

func (c *Controller) getErrorMessageFormat(e echo.Context) error {
	ch, err := c.chainService.GetChain()
	if err != nil {
		return c.handleViewCallError(err)
	}
	contractHname, err := params.DecodeHNameFromHNameHexString(e, "contractHname")
	if err != nil {
		return err
	}

	errorID, err := params.DecodeUInt(e, "errorID")
	if err != nil {
		return err
	}

	errorIDUint16, err := safecast.Convert[uint16](errorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Error ID out of range for uint16")
	}

	messageFormat, err := corecontracts.ErrorMessageFormat(ch, contractHname, errorIDUint16, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err)
	}

	errorMessageFormatResponse := &ErrorMessageFormatResponse{
		MessageFormat: messageFormat,
	}

	return e.JSON(http.StatusOK, errorMessageFormatResponse)
}
