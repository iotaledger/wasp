package corecontracts

import (
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo/v4"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) getControlAddresses(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	controlAddresses, err := c.blocklog.GetControlAddresses(chainID)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	controlAddressesResponse := &models.ControlAddressesResponse{
		GoverningAddress: controlAddresses.GoverningAddress.Bech32(parameters.L1().Protocol.Bech32HRP),
		SinceBlockIndex:  controlAddresses.SinceBlockIndex,
		StateAddress:     controlAddresses.StateAddress.Bech32(parameters.L1().Protocol.Bech32HRP),
	}

	return e.JSON(http.StatusOK, controlAddressesResponse)
}

func (c *Controller) getBlockInfo(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	var blockInfo *blocklog.BlockInfo
	blockIndex := e.Param("blockIndex")

	if blockIndex == "" {
		blockInfo, err = c.blocklog.GetLatestBlockInfo(chainID)
	} else {
		var blockIndexNum uint64
		blockIndexNum, err = strconv.ParseUint(e.Param("blockIndex"), 10, 64)
		if err != nil {
			return apierrors.InvalidPropertyError("blockIndex", err)
		}

		blockInfo, err = c.blocklog.GetBlockInfo(chainID, uint32(blockIndexNum))
	}
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	blockInfoResponse := models.MapBlockInfoResponse(blockInfo)

	return e.JSON(http.StatusOK, blockInfoResponse)
}

func (c *Controller) getRequestIDsForBlock(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	var requestIDs []isc.RequestID
	blockIndex := e.Param("blockIndex")

	if blockIndex == "" {
		requestIDs, err = c.blocklog.GetRequestIDsForLatestBlock(chainID)
	} else {
		var blockIndexNum uint64
		blockIndexNum, err = params.DecodeUInt(e, "blockIndex")
		if err != nil {
			return err
		}

		requestIDs, err = c.blocklog.GetRequestIDsForBlock(chainID, uint32(blockIndexNum))
	}

	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	requestIDsResponse := &models.RequestIDsResponse{
		RequestIDs: make([]string, len(requestIDs)),
	}

	for k, v := range requestIDs {
		requestIDsResponse.RequestIDs[k] = v.String()
	}

	return e.JSON(http.StatusOK, requestIDsResponse)
}

func MapRequestReceiptResponse(vmService interfaces.VMService, chainID isc.ChainID, receipt *blocklog.RequestReceipt) (*models.RequestReceiptResponse, error) {
	response := &models.RequestReceiptResponse{
		BlockIndex:    receipt.BlockIndex,
		GasBudget:     iotago.EncodeUint64(receipt.GasBudget),
		GasBurnLog:    receipt.GasBurnLog,
		GasBurned:     iotago.EncodeUint64(receipt.GasBurned),
		GasFeeCharged: iotago.EncodeUint64(receipt.GasFeeCharged),
		Request:       models.MapRequestDetail(receipt.Request),
		RequestIndex:  receipt.RequestIndex,
	}

	if receipt.Error != nil {
		resolved, err := errors.Resolve(receipt.Error, func(contract string, function string, params dict.Dict) (dict.Dict, error) {
			return vmService.CallViewByChainID(chainID, isc.Hn(contract), isc.Hn(function), params)
		})
		if err != nil {
			return nil, err
		}

		response.Error = &models.BlockReceiptError{
			Hash:         hexutil.EncodeUint64(uint64(resolved.Hash())),
			ErrorMessage: resolved.Error(),
		}
	}

	return response, nil
}

func (c *Controller) getRequestReceipt(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	requestID, err := params.DecodeRequestID(e)
	if err != nil {
		return err
	}

	receipt, err := c.blocklog.GetRequestReceipt(chainID, requestID)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	mappedReceiptResponse, err := MapRequestReceiptResponse(c.vmService, chainID, receipt)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	return e.JSON(http.StatusOK, mappedReceiptResponse)
}

func (c *Controller) getRequestReceiptsForBlock(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	var receipts []*blocklog.RequestReceipt
	blockIndex := e.Param("blockIndex")

	if blockIndex == "" {
		var blockInfo *blocklog.BlockInfo
		blockInfo, err = c.blocklog.GetLatestBlockInfo(chainID)
		if err != nil {
			return c.handleViewCallError(err, chainID)
		}

		receipts, err = c.blocklog.GetRequestReceiptsForBlock(chainID, blockInfo.BlockIndex)
	} else {
		var blockIndexNum uint64
		blockIndexNum, err = params.DecodeUInt(e, "blockIndex")
		if err != nil {
			return err
		}

		receipts, err = c.blocklog.GetRequestReceiptsForBlock(chainID, uint32(blockIndexNum))
	}
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	receiptsResponse := models.BlockReceiptsResponse{
		Receipts: make([]*models.RequestReceiptResponse, len(receipts)),
	}

	for k, v := range receipts {
		receipt, err := MapRequestReceiptResponse(c.vmService, chainID, v)
		if err != nil {
			return apierrors.InvalidPropertyError("receipt", err)
		}

		receiptsResponse.Receipts[k] = receipt
	}

	return e.JSON(http.StatusOK, receiptsResponse)
}

func (c *Controller) getIsRequestProcessed(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	requestID, err := params.DecodeRequestID(e)
	if err != nil {
		return err
	}

	requestProcessed, err := c.blocklog.IsRequestProcessed(chainID, requestID)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	requestProcessedResponse := models.RequestProcessedResponse{
		ChainID:     chainID.String(),
		RequestID:   requestID.String(),
		IsProcessed: requestProcessed,
	}

	return e.JSON(http.StatusOK, requestProcessedResponse)
}

func (c *Controller) getBlockEvents(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	var events []string
	blockIndex := e.Param("blockIndex")

	if blockIndex != "" {
		blockIndexNum, err := params.DecodeUInt(e, "blockIndex")
		if err != nil {
			return err
		}

		events, err = c.blocklog.GetEventsForBlock(chainID, uint32(blockIndexNum))
		if err != nil {
			return c.handleViewCallError(err, chainID)
		}
	} else {
		blockInfo, err := c.blocklog.GetLatestBlockInfo(chainID)
		if err != nil {
			return c.handleViewCallError(err, chainID)
		}

		events, err = c.blocklog.GetEventsForBlock(chainID, blockInfo.BlockIndex)
		if err != nil {
			return c.handleViewCallError(err, chainID)
		}
	}

	eventsResponse := models.EventsResponse{
		Events: events,
	}

	return e.JSON(http.StatusOK, eventsResponse)
}

func (c *Controller) getContractEvents(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	contractHname, err := params.DecodeHNameFromHNameHexString(e, "contractHname")
	if err != nil {
		return err
	}

	events, err := c.blocklog.GetEventsForContract(chainID, contractHname)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	eventsResponse := models.EventsResponse{
		Events: events,
	}

	return e.JSON(http.StatusOK, eventsResponse)
}

func (c *Controller) getRequestEvents(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	requestID, err := params.DecodeRequestID(e)
	if err != nil {
		return err
	}

	events, err := c.blocklog.GetEventsForRequest(chainID, requestID)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	eventsResponse := models.EventsResponse{
		Events: events,
	}

	return e.JSON(http.StatusOK, eventsResponse)
}
