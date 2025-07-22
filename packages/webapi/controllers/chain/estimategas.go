package chain

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/v2/packages/webapi/common"
	"github.com/iotaledger/wasp/v2/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
)

func (c *Controller) estimateGasOnLedger(e echo.Context) error {
	panic("TODO")
	// controllerutils.SetOperation(e, "estimate_gas_onledger")
	// ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	// if err != nil {
	// 	return err
	// }

	// var estimateGasRequest models.EstimateGasRequestOnledger
	// if err = e.Bind(&estimateGasRequest); err != nil {
	// 	return apierrors.InvalidPropertyError("body", err)
	// }

	// outputBytes, err := cryptolib.DecodeHex(estimateGasRequest.Output)
	// if err != nil {
	// 	return apierrors.InvalidPropertyError("Request", err)
	// }
	// output, err := util.OutputFromBytes(outputBytes)
	// if err != nil {
	// 	return apierrors.InvalidPropertyError("Output", err)
	// }

	// req, err := isc.OnLedgerFromUTXO(
	// 	output,
	// 	iotago.OutputID{}, // empty outputID for estimation
	// )
	// if err != nil {
	// 	return apierrors.InvalidPropertyError("Output", err)
	// }
	// if !req.TargetAddress().Equals(chainID.AsAddress()) {
	// 	return apierrors.InvalidPropertyError("Request", errors.New("wrong chainID"))
	// }

	// rec, err := common.EstimateGas(ch, req)
	// if err != nil {
	// 	return apierrors.NewHTTPError(http.StatusBadRequest, "VM run error", err)
	// }
	// return e.JSON(http.StatusOK, models.MapReceiptResponse(rec))
}

func (c *Controller) estimateGasOffLedger(e echo.Context) error {
	controllerutils.SetOperation(e, "estimate_gas_offledger")
	ch, err := c.chainService.GetChain()
	if err != nil {
		return err
	}

	var estimateGasRequest models.EstimateGasRequestOffledger
	if err = e.Bind(&estimateGasRequest); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	if estimateGasRequest.FromAddress == "" {
		return apierrors.InvalidPropertyError("fromAddress", err)
	}

	requestFrom, err := cryptolib.NewAddressFromHexString(estimateGasRequest.FromAddress)
	if err != nil {
		return apierrors.InvalidPropertyError("fromAddress", err)
	}

	requestBytes, err := cryptolib.DecodeHex(estimateGasRequest.Request)
	if err != nil {
		return apierrors.InvalidPropertyError("requestBytes", err)
	}

	req, err := c.offLedgerService.ParseRequest(requestBytes)
	if err != nil {
		return apierrors.InvalidPropertyError("requestBytes", fmt.Errorf("not an offledger request"))
	}

	offLedgerRequest, ok := req.(isc.OffLedgerRequest)
	if !ok {
		return apierrors.InvalidPropertyError("requestBytes", fmt.Errorf("not an offledger request"))
	}

	offLedgerRequestData, ok := offLedgerRequest.(*isc.OffLedgerRequestData)
	if !ok {
		return apierrors.InvalidPropertyError("requestBytes", fmt.Errorf("not an unsigned offledger request"))
	}

	impRequest := isc.NewImpersonatedOffLedgerRequest(&offLedgerRequestData.OffLedgerRequestDataEssence).
		WithSenderAddress(requestFrom)

	rec, err := common.EstimateGas(ch, impRequest)
	if err != nil {
		return apierrors.NewHTTPError(http.StatusBadRequest, "VM run error", err)
	}

	res := rec.DeserializedRequest()
	fmt.Printf("RequestBytes: %s\n", hexutil.Encode(rec.Request))
	fmt.Printf("Request data: %v %v", res, res.Message())
	return e.JSON(http.StatusOK, models.MapReceiptResponse(rec))
}
