package chain

import (
	"net/http"

	"github.com/labstack/echo/v4"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) handleJSONRPC(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	if !c.chainService.HasChain(chainID) {
		return apierrors.ChainNotFoundError(chainID.String())
	}

	return c.evmService.HandleJSONRPC(chainID, e.Request(), e.Response())
}

func (c *Controller) getRequestID(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	if !c.chainService.HasChain(chainID) {
		return apierrors.ChainNotFoundError(chainID.String())
	}

	txHash := e.Param(params.ParamTxHash)
	requestID, err := c.evmService.GetRequestID(chainID, txHash)
	if err != nil {
		return apierrors.InvalidPropertyError(params.ParamTxHash, err)
	}

	return e.JSON(http.StatusOK, models.RequestIDResponse{
		RequestID: iotago.EncodeHex(requestID.Bytes()),
	})
}
