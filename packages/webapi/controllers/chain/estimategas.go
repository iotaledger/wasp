package chain

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/models"
)

func (c *Controller) estimateGasOnLedger(e echo.Context) error {
	controllerutils.SetOperation(e, "estimate_gas_onledger")
	chainID, err := controllerutils.ChainIDFromParams(e, c.chainService)
	if err != nil {
		return err
	}

	var estimateGasRequest models.EstimateGasRequestOnledger
	if err = e.Bind(&estimateGasRequest); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	mu := marshalutil.New(estimateGasRequest.Output)
	output, err := util.OutputFromMarshalUtil(mu)
	if err != nil {
		return apierrors.InvalidPropertyError("Output", err)
	}

	req, err := isc.OnLedgerFromUTXO(
		output,
		iotago.OutputID{}, // empty outputID for estimation
	)
	if err != nil {
		return apierrors.InvalidPropertyError("Output", err)
	}

	rec, err := c.vmService.EstimateGas(chainID, req)
	if err != nil {
		return apierrors.NewHTTPError(http.StatusBadRequest, "VM run error", err)
	}
	return e.JSON(http.StatusOK, models.MapReceiptResponse(rec))
}

func (c *Controller) estimateGasOffLedger(e echo.Context) error {
	controllerutils.SetOperation(e, "estimate_gas_offledger")
	chainID, err := controllerutils.ChainIDFromParams(e, c.chainService)
	if err != nil {
		return err
	}

	var estimateGasRequest models.EstimateGasRequestOffledger
	if err = e.Bind(&estimateGasRequest); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	requestBytes, err := iotago.DecodeHex(estimateGasRequest.Request)
	if err != nil {
		return apierrors.InvalidPropertyError("Request", err)
	}

	req, err := c.offLedgerService.ParseRequest(requestBytes)
	if err != nil {
		return apierrors.InvalidPropertyError("Request", err)
	}

	rec, err := c.vmService.EstimateGas(chainID, req)
	if err != nil {
		return apierrors.NewHTTPError(http.StatusBadRequest, "VM run error", err)
	}
	return e.JSON(http.StatusOK, models.MapReceiptResponse(rec))
}
