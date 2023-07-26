package corecontracts

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/common"
	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/corecontracts"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) getControlAddresses(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}
	controlAddresses, err := corecontracts.GetControlAddresses(ch)
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
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}
	var blockInfo *blocklog.BlockInfo
	blockIndex := e.Param(params.ParamBlockIndex)

	if blockIndex == "" {
		blockInfo, err = corecontracts.GetLatestBlockInfo(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	} else {
		var blockIndexNum uint64
		blockIndexNum, err = strconv.ParseUint(e.Param(params.ParamBlockIndex), 10, 64)
		if err != nil {
			return apierrors.InvalidPropertyError(params.ParamBlockIndex, err)
		}

		blockInfo, err = corecontracts.GetBlockInfo(ch, uint32(blockIndexNum), e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	}
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	blockInfoResponse := models.MapBlockInfoResponse(blockInfo)

	return e.JSON(http.StatusOK, blockInfoResponse)
}

func (c *Controller) getRequestIDsForBlock(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}
	var requestIDs []isc.RequestID
	blockIndex := e.Param(params.ParamBlockIndex)

	if blockIndex == "" {
		requestIDs, err = corecontracts.GetRequestIDsForLatestBlock(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	} else {
		var blockIndexNum uint64
		blockIndexNum, err = params.DecodeUInt(e, params.ParamBlockIndex)
		if err != nil {
			return err
		}

		requestIDs, err = corecontracts.GetRequestIDsForBlock(ch, uint32(blockIndexNum), e.QueryParam(params.ParamBlockIndexOrTrieRoot))
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

func GetRequestReceipt(e echo.Context, c interfaces.ChainService) error {
	ch, _, err := controllerutils.ChainFromParams(e, c)
	if err != nil {
		return err
	}
	requestID, err := params.DecodeRequestID(e)
	if err != nil {
		return err
	}

	receipt, err := corecontracts.GetRequestReceipt(ch, requestID, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		panic(err)
	}
	if receipt == nil {
		return apierrors.NoRecordFoundError(errors.New("no receipt"))
	}

	resolvedReceipt, err := common.ParseReceipt(ch, receipt)
	if err != nil {
		panic(err)
	}

	return e.JSON(http.StatusOK, models.MapReceiptResponse(resolvedReceipt))
}

func (c *Controller) getRequestReceipt(e echo.Context) error {
	return GetRequestReceipt(e, c.chainService)
}

func (c *Controller) getRequestReceiptsForBlock(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}
	var blocklogReceipts []*blocklog.RequestReceipt
	blockIndex := e.Param(params.ParamBlockIndex)

	if blockIndex == "" {
		var blockInfo *blocklog.BlockInfo
		blockInfo, err = corecontracts.GetLatestBlockInfo(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
		if err != nil {
			return c.handleViewCallError(err, chainID)
		}

		blocklogReceipts, err = corecontracts.GetRequestReceiptsForBlock(ch, blockInfo.BlockIndex(), e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	} else {
		var blockIndexNum uint64
		blockIndexNum, err = params.DecodeUInt(e, params.ParamBlockIndex)
		if err != nil {
			return err
		}

		blocklogReceipts, err = corecontracts.GetRequestReceiptsForBlock(ch, uint32(blockIndexNum), e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	}
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	receiptsResponse := make([]*models.ReceiptResponse, len(blocklogReceipts))

	for i, blocklogReceipt := range blocklogReceipts {
		parsedReceipt, err := common.ParseReceipt(ch, blocklogReceipt)
		if err != nil {
			panic(err)
		}
		receiptResp := models.MapReceiptResponse(parsedReceipt)
		receiptsResponse[i] = receiptResp
	}

	return e.JSON(http.StatusOK, receiptsResponse)
}

func (c *Controller) getIsRequestProcessed(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}
	requestID, err := params.DecodeRequestID(e)
	if err != nil {
		return err
	}

	requestProcessed, err := corecontracts.IsRequestProcessed(ch, requestID, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
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

func eventsResponse(e echo.Context, events []*isc.Event) error {
	eventsJSON := make([]*isc.EventJSON, len(events))
	for i, ev := range events {
		eventsJSON[i] = ev.ToJSONStruct()
	}
	return e.JSON(http.StatusOK, models.EventsResponse{
		Events: eventsJSON,
	})
}

func (c *Controller) getBlockEvents(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}
	var events []*isc.Event
	blockIndex := e.Param(params.ParamBlockIndex)

	if blockIndex != "" {
		blockIndexNum, err := params.DecodeUInt(e, params.ParamBlockIndex)
		if err != nil {
			return err
		}

		events, err = corecontracts.GetEventsForBlock(ch, uint32(blockIndexNum), e.QueryParam(params.ParamBlockIndexOrTrieRoot))
		if err != nil {
			return c.handleViewCallError(err, chainID)
		}
	} else {
		blockInfo, err := corecontracts.GetLatestBlockInfo(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
		if err != nil {
			return c.handleViewCallError(err, chainID)
		}

		events, err = corecontracts.GetEventsForBlock(ch, blockInfo.BlockIndex(), e.QueryParam(params.ParamBlockIndexOrTrieRoot))
		if err != nil {
			return c.handleViewCallError(err, chainID)
		}
	}
	return eventsResponse(e, events)
}

func (c *Controller) getContractEvents(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}
	contractHname, err := params.DecodeHNameFromHNameHexString(e, "contractHname")
	if err != nil {
		return err
	}

	events, err := corecontracts.GetEventsForContract(ch, contractHname, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}
	return eventsResponse(e, events)
}

func (c *Controller) getRequestEvents(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}
	requestID, err := params.DecodeRequestID(e)
	if err != nil {
		return err
	}

	events, err := corecontracts.GetEventsForRequest(ch, requestID, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}
	return eventsResponse(e, events)
}
