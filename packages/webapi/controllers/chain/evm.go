package chain

import (
	"net/http"

	"github.com/labstack/echo/v4"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) handleJSONRPC(e echo.Context) error {
	controllerutils.SetOperation(e, "evm_json_rpc")
	chainID, err := controllerutils.ChainIDFromParams(e, c.chainService)
	if err != nil {
		return err
	}

	return c.evmService.HandleJSONRPC(chainID, e.Request(), e.Response())
}

func (c *Controller) handleWebsocket(e echo.Context) error {
	controllerutils.SetOperation(e, "evm_websocket")
	chainID, err := controllerutils.ChainIDFromParams(e, c.chainService)
	if err != nil {
		return err
	}

	return c.evmService.HandleWebsocket(chainID, e.Request(), e.Response())
}

func (c *Controller) getRequestID(e echo.Context) error {
	controllerutils.SetOperation(e, "evm_get_request_id")
	chainID, err := controllerutils.ChainIDFromParams(e, c.chainService)
	if err != nil {
		return err
	}

	txHash := e.Param(params.ParamTxHash)
	requestID, err := c.evmService.GetRequestID(chainID, txHash)
	if err != nil {
		return apierrors.NoRecordFoundError(err)
	}

	return e.JSON(http.StatusOK, models.RequestIDResponse{
		RequestID: iotago.EncodeHex(requestID.Bytes()),
	})
}
