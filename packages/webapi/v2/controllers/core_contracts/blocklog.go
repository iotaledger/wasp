package corecontracts

import (
	"net/http"
	"strconv"
	"time"

	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"

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
	StateAddress     string
	GoverningAddress string
	SinceBlockIndex  uint32
}

func (c *Controller) getControlAddresses(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
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
	BlockIndex                  uint32
	Timestamp                   time.Time
	TotalRequests               uint16
	NumSuccessfulRequests       uint16
	NumOffLedgerRequests        uint16
	PreviousL1CommitmentHash    string
	L1CommitmentHash            string
	AnchorTransactionID         string
	TransactionSubEssenceHash   string
	TotalBaseTokensInL2Accounts uint64
	TotalStorageDeposit         uint64
	GasBurned                   uint64
	GasFeeCharged               uint64
}

func mapBlockInfoResponse(info *blocklog.BlockInfo) *BlockInfoResponse {
	transactionEssenceHash := hexutil.Encode(info.TransactionSubEssenceHash[:])
	commitmentHash := ""

	if info.L1Commitment != nil {
		commitmentHash = info.L1Commitment.BlockHash.String()
	}

	return &BlockInfoResponse{
		BlockIndex:                  info.BlockIndex,
		Timestamp:                   info.Timestamp,
		TotalRequests:               info.TotalRequests,
		NumSuccessfulRequests:       info.NumSuccessfulRequests,
		NumOffLedgerRequests:        info.NumOffLedgerRequests,
		PreviousL1CommitmentHash:    info.PreviousL1Commitment.BlockHash.String(),
		L1CommitmentHash:            commitmentHash,
		AnchorTransactionID:         info.AnchorTransactionID.ToHex(),
		TransactionSubEssenceHash:   transactionEssenceHash,
		TotalBaseTokensInL2Accounts: info.TotalBaseTokensInL2Accounts,
		TotalStorageDeposit:         info.TotalStorageDeposit,
		GasBurned:                   info.GasBurned,
		GasFeeCharged:               info.GasFeeCharged,
	}
}

func (c *Controller) getLatestBlockInfo(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	blockInfo, err := c.blocklog.GetLatestBlockInfo(chainID)
	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	blockInfoResponse := mapBlockInfoResponse(blockInfo)

	return e.JSON(http.StatusOK, blockInfoResponse)
}

func (c *Controller) getBlockInfo(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	blockIndex, err := strconv.ParseUint(e.Param("blockIndex"), 10, 64)
	if err != nil {
		return apierrors.InvalidPropertyError("blockIndex", err)
	}

	blockInfo, err := c.blocklog.GetBlockInfo(chainID, uint32(blockIndex))
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
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	blockIndex, err := strconv.ParseUint(e.Param("blockIndex"), 10, 64)
	if err != nil {
		return apierrors.InvalidPropertyError("blockIndex", err)
	}

	requestIDs, err := c.blocklog.GetRequestIDsForBlock(chainID, uint32(blockIndex))
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
	// TODO request may be big (blobs). Do we want to store it all?
	Request       isc.Request `json:"request"`
	Error         *Error      `json:"error"`
	GasBudget     uint64      `json:"gasBudget"`
	GasBurned     uint64      `json:"gasBurned"`
	GasFeeCharged uint64      `json:"gasFeeCharged"`
	// not persistent
	BlockIndex   uint32       `json:"blockIndex"`
	RequestIndex uint16       `json:"requestIndex"`
	GasBurnLog   *gas.BurnLog `json:"-"`
}

func mapRequestReceiptResponse(vmService interfaces.VMService, chainID *isc.ChainID, receipt *blocklog.RequestReceipt) (*RequestReceiptResponse, error) {
	response := &RequestReceiptResponse{
		Request:       receipt.Request,
		GasBudget:     receipt.GasBudget,
		GasBurned:     receipt.GasBurned,
		GasFeeCharged: receipt.GasFeeCharged,
		BlockIndex:    receipt.BlockIndex,
		RequestIndex:  receipt.RequestIndex,
		GasBurnLog:    receipt.GasBurnLog,
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
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	requestID, err := isc.RequestIDFromString(e.Param("requestID"))
	if err != nil {
		return apierrors.InvalidPropertyError("requestID", err)
	}

	receipt, err := c.blocklog.GetRequestReceipt(chainID, requestID)
	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	mappedReceiptResponse, err := mapRequestReceiptResponse(c.vmService, chainID, receipt)
	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	return e.JSON(http.StatusOK, mappedReceiptResponse)
}
