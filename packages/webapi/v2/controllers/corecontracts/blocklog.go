package corecontracts

import (
	"net/http"
	"strconv"
	"time"

	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/v2/params"

	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/errors"

	"github.com/iotaledger/wasp/packages/vm/gas"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/labstack/echo/v4"
)

type ControlAddressesResponse struct {
	GoverningAddress string
	SinceBlockIndex  uint32
	StateAddress     string
}

func (c *Controller) getControlAddresses(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	controlAddresses, err := c.blocklog.GetControlAddresses(chainID)
	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	controlAddressesResponse := &ControlAddressesResponse{
		GoverningAddress: controlAddresses.GoverningAddress.String(),
		SinceBlockIndex:  controlAddresses.SinceBlockIndex,
		StateAddress:     controlAddresses.StateAddress.String(),
	}

	return e.JSON(http.StatusOK, controlAddressesResponse)
}

type BlockInfoResponse struct {
	AnchorTransactionID         string
	BlockIndex                  uint32
	GasBurned                   uint64
	GasFeeCharged               uint64
	L1CommitmentHash            string
	NumOffLedgerRequests        uint16
	NumSuccessfulRequests       uint16
	PreviousL1CommitmentHash    string
	Timestamp                   time.Time
	TotalBaseTokensInL2Accounts uint64
	TotalRequests               uint16
	TotalStorageDeposit         uint64
	TransactionSubEssenceHash   string
}

func mapBlockInfoResponse(info *blocklog.BlockInfo) *BlockInfoResponse {
	transactionEssenceHash := hexutil.Encode(info.TransactionSubEssenceHash[:])
	commitmentHash := ""

	if info.L1Commitment != nil {
		commitmentHash = info.L1Commitment.BlockHash.String()
	}

	return &BlockInfoResponse{
		AnchorTransactionID:         info.AnchorTransactionID.ToHex(),
		BlockIndex:                  info.BlockIndex,
		GasBurned:                   info.GasBurned,
		GasFeeCharged:               info.GasFeeCharged,
		L1CommitmentHash:            commitmentHash,
		NumOffLedgerRequests:        info.NumOffLedgerRequests,
		NumSuccessfulRequests:       info.NumSuccessfulRequests,
		PreviousL1CommitmentHash:    info.PreviousL1Commitment.BlockHash.String(),
		Timestamp:                   info.Timestamp,
		TotalBaseTokensInL2Accounts: info.TotalBaseTokensInL2Accounts,
		TotalRequests:               info.TotalRequests,
		TotalStorageDeposit:         info.TotalStorageDeposit,
		TransactionSubEssenceHash:   transactionEssenceHash,
	}
}

func (c *Controller) getBlockInfo(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	var blockInfo *blocklog.BlockInfo
	blockIndex := e.Param("blockIndex")

	if blockIndex == "latest" {
		blockInfo, err = c.blocklog.GetLatestBlockInfo(chainID)
	} else {
		blockIndexNum, err := strconv.ParseUint(e.Param("blockIndex"), 10, 64)
		if err != nil {
			return apierrors.InvalidPropertyError("blockIndex", err)
		}

		blockInfo, err = c.blocklog.GetBlockInfo(chainID, uint32(blockIndexNum))
	}

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	blockInfoResponse := mapBlockInfoResponse(blockInfo)

	return e.JSON(http.StatusOK, blockInfoResponse)
}

type RequestIDsResponse struct {
	RequestIDs []string
}

func (c *Controller) getRequestIDsForBlock(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	var requestIDs []isc.RequestID
	blockIndex := e.Param("blockIndex")

	if blockIndex == "latest" {
		requestIDs, err = c.blocklog.GetRequestIDsForLatestBlock(chainID)
	} else {
		blockIndexNum, err := params.DecodeUInt(e, "blockIndex")
		if err != nil {
			return err
		}

		requestIDs, err = c.blocklog.GetRequestIDsForBlock(chainID, uint32(blockIndexNum))
	}

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	requestIDsResponse := &RequestIDsResponse{
		RequestIDs: make([]string, len(requestIDs)),
	}

	for k, v := range requestIDs {
		requestIDsResponse.RequestIDs[k] = v.String()
	}

	return e.JSON(http.StatusOK, requestIDsResponse)
}

type Error struct {
	Hash         string
	ErrorMessage string
}

type RequestReceiptResponse struct {
	BlockIndex    uint32
	Error         *Error
	GasBudget     uint64
	GasBurnLog    *gas.BurnLog
	GasBurned     uint64
	GasFeeCharged uint64
	Request       isc.Request
	RequestIndex  uint16
}

func mapRequestReceiptResponse(vmService interfaces.VMService, chainID *isc.ChainID, receipt *blocklog.RequestReceipt) (*RequestReceiptResponse, error) {
	response := &RequestReceiptResponse{
		BlockIndex:    receipt.BlockIndex,
		GasBudget:     receipt.GasBudget,
		GasBurnLog:    receipt.GasBurnLog,
		GasBurned:     receipt.GasBurned,
		GasFeeCharged: receipt.GasFeeCharged,
		Request:       receipt.Request,
		RequestIndex:  receipt.RequestIndex,
	}

	if receipt.Error != nil {
		resolved, err := errors.Resolve(receipt.Error, func(contract string, function string, params dict.Dict) (dict.Dict, error) {
			return vmService.CallViewByChainID(chainID, isc.Hn(contract), isc.Hn(function), params)
		})

		if err != nil {
			return nil, err
		}

		response.Error = &Error{
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

	receipt, err := c.blocklog.GetRequestReceipt(chainID, *requestID)
	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	mappedReceiptResponse, err := mapRequestReceiptResponse(c.vmService, chainID, receipt)
	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	return e.JSON(http.StatusOK, mappedReceiptResponse)
}

type BlockReceiptsResponse struct {
	Receipts []*RequestReceiptResponse
}

func (c *Controller) getRequestReceiptsForBlock(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	var receipts []*blocklog.RequestReceipt
	blockIndex := e.Param("blockIndex")

	if blockIndex == "latest" {
		blockInfo, err := c.blocklog.GetLatestBlockInfo(chainID)
		if err != nil {
			return apierrors.ContractExecutionError(err)
		}

		receipts, err = c.blocklog.GetRequestReceiptsForBlock(chainID, blockInfo.BlockIndex)
	} else {
		blockIndexNum, err := params.DecodeUInt(e, "blockIndex")
		if err != nil {
			return err
		}

		receipts, err = c.blocklog.GetRequestReceiptsForBlock(chainID, uint32(blockIndexNum))
	}

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	receiptsResponse := BlockReceiptsResponse{
		Receipts: make([]*RequestReceiptResponse, 0, len(receipts)),
	}

	for k, v := range receipts {
		receipt, err := mapRequestReceiptResponse(c.vmService, chainID, v)
		if err != nil {
			return apierrors.InvalidPropertyError("receipt", err)
		}

		receiptsResponse.Receipts[k] = receipt
	}

	return e.JSON(http.StatusOK, receiptsResponse)
}

type RequestProcessedResponse struct {
	ChainID     string
	RequestID   string
	IsProcessed bool
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

	requestProcessed, err := c.blocklog.IsRequestProcessed(chainID, *requestID)
	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	requestProcessedResponse := RequestProcessedResponse{
		ChainID:     chainID.String(),
		RequestID:   requestID.String(),
		IsProcessed: requestProcessed,
	}

	return e.JSON(http.StatusOK, requestProcessedResponse)
}

type EventsResponse struct {
	Events []string
}

func (c *Controller) getBlockEvents(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	var events []string
	blockIndex := e.Param("blockIndex")

	if blockIndex == "latest" {
		blockInfo, err := c.blocklog.GetLatestBlockInfo(chainID)
		if err != nil {
			return apierrors.ContractExecutionError(err)
		}

		events, err = c.blocklog.GetEventsForBlock(chainID, blockInfo.BlockIndex)
	} else {
		blockIndexNum, err := params.DecodeUInt(e, "blockIndex")
		if err != nil {
			return err
		}

		events, err = c.blocklog.GetEventsForBlock(chainID, uint32(blockIndexNum))
	}

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	eventsResponse := EventsResponse{
		Events: events,
	}

	return e.JSON(http.StatusOK, eventsResponse)
}

func (c *Controller) getContractEvents(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	contractHname, err := params.DecodeHNameFromHNameString(e, "contractHname")
	if err != nil {
		return err
	}

	events, err := c.blocklog.GetEventsForContract(chainID, contractHname)

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	eventsResponse := EventsResponse{
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

	events, err := c.blocklog.GetEventsForRequest(chainID, *requestID)

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	eventsResponse := EventsResponse{
		Events: events,
	}

	return e.JSON(http.StatusOK, eventsResponse)
}
