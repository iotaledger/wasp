package corecontracts

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/corecontracts"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

type ErrorMessageFormatResponse struct {
	MessageFormat string `json:"messageFormat" swagger:"required"`
}

func (c *Controller) getErrorMessageFormat(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}
	contractHname, err := params.DecodeHNameFromHNameHexString(e, "contractHname")
	if err != nil {
		return err
	}

	errorID, err := params.DecodeUInt(e, "errorID")
	if err != nil {
		return err
	}

	messageFormat, err := corecontracts.ErrorMessageFormat(ch, contractHname, uint16(errorID), e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	errorMessageFormatResponse := &ErrorMessageFormatResponse{
		MessageFormat: messageFormat,
	}

	return e.JSON(http.StatusOK, errorMessageFormatResponse)
}
