package chain

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/context"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/common"
	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/models"
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

	msgBytes, err := hexutil.Decode(estimateGasRequest.RequestBytes)
	if err != nil {
		return err
	}

	var msg iscmove.Request
	msg, err = bcs.Unmarshal[iscmove.Request](msgBytes)
	if err != nil {
		return apierrors.InvalidPropertyError("requestBytes", err)
	}

	dryRunResBytes, err := hexutil.Decode(estimateGasRequest.TransactionBytes)
	if err != nil {
		return apierrors.InvalidPropertyError("transactionBytes", err)
	}

	callContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dryRunResponse, err := c.l1Client.DryRunTransaction(callContext, dryRunResBytes)

	req, err := isc.FakeEstimateOnLedger(dryRunResponse, msg)
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
	return e.JSON(http.StatusOK, models.MapReceiptResponse(rec))
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
