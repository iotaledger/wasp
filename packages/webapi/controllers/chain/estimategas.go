package chain

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/context"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/v2/packages/webapi/common"
	"github.com/iotaledger/wasp/v2/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
)

func (c *Controller) estimateGasOnLedger(e echo.Context) error {
	controllerutils.SetOperation(e, "estimate_gas_onledger")
	ch, err := c.chainService.GetChain()
	if err != nil {
		return err
	}

	var estimateGasRequest models.EstimateGasRequestOnledger
	if err = e.Bind(&estimateGasRequest); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	txBytes, err := hexutil.Decode(estimateGasRequest.TransactionBytes)
	if err != nil {
		return apierrors.InvalidPropertyError("transactionBytes", err)
	}

	txData, err := bcs.Unmarshal[iotago.TransactionData](txBytes)
	if err != nil {
		return apierrors.InvalidPropertyError("transactionBytes", fmt.Errorf("failed to unmarshal tx bytes into TransactionData: %w", err))
	}

	if txData.V1 == nil {
		return apierrors.InvalidPropertyError("transactionBytes", errors.New("TransactionData V1 is supported"))
	}
	if txData.V1.GasData.Price == 0 {
		return apierrors.InvalidPropertyError("transactionBytes", errors.New("transaction must have a non-zero gas price"))
	}

	// Unsetting gas coin objects and gas budget for purpose of gas estimation.
	txData.V1.GasData.Payment = nil
	txData.V1.GasData.Budget = iotaclient.MaxGasBudget

	txBytes, err = bcs.Marshal(&txData)
	if err != nil {
		return fmt.Errorf("failed to marshal tx data into bytes: %w", err)
	}

	callContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dryRunResponse, err := c.l1Client.DryRunTransaction(callContext, txBytes)
	if err != nil {
		return apierrors.NewHTTPError(http.StatusBadRequest, "DryRun error", err)
	}
	if dryRunResponse.Effects.Data.V1.Status.Error != "" {
		return apierrors.NewHTTPError(http.StatusBadRequest, "DryRun status error", fmt.Errorf("%s: %s",
			dryRunResponse.Effects.Data.V1.Status.Status,
			dryRunResponse.Effects.Data.V1.Status.Error,
		))
	}

	req, err := isc.ReconstructOnLedgerRequest(dryRunResponse)
	if err != nil {
		return fmt.Errorf("cant generate fake request: %s", err)
	}

	rec, err := common.EstimateGas(ch, req)
	if err != nil {
		return apierrors.NewHTTPError(http.StatusBadRequest, "VM run error", err)
	}

	res := rec.DeserializedRequest()
	fmt.Printf("RequestBytes: %s\n", hexutil.Encode(rec.Request))
	fmt.Printf("Request data: %v %v", res, res.Message())

	return e.JSON(http.StatusOK, models.OnLedgerEstimationResponse{
		L1: models.MapL1EstimationResult(&dryRunResponse.Effects.Data.V1.GasUsed),
		L2: models.MapReceiptResponse(rec),
	})
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
