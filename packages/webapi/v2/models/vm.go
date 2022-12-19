package models

import (
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type ReceiptError struct {
	ContractID    isc.Hname     `json:"contractId"`
	ErrorID       uint16        `json:"errorId"`
	ErrorCode     string        `json:"errorCode"`
	Message       string        `json:"message"`
	MessageFormat string        `json:"messageFormat"`
	Parameters    []interface{} `json:"parameters"`
}

func MapReceiptError(err *isc.VMError) *ReceiptError {
	if err == nil {
		return nil
	}

	return &ReceiptError{
		ContractID:    err.Code().ContractID,
		ErrorID:       err.Code().ID,
		ErrorCode:     err.Code().String(),
		Message:       err.Error(),
		MessageFormat: err.MessageFormat(),
		Parameters:    err.Params(),
	}
}

type ReceiptResponse struct {
	Request       string           `json:"request"`
	Error         *ReceiptError    `json:"error"`
	GasBudget     uint64           `json:"gasBudget"`
	GasBurned     uint64           `json:"gasBurned"`
	GasFeeCharged uint64           `json:"gasFeeCharged"`
	BlockIndex    uint32           `json:"blockIndex"`
	RequestIndex  uint16           `json:"requestIndex"`
	GasBurnLog    []gas.BurnRecord `json:"gasBurnLog"`
}

func MapReceiptResponse(receipt *isc.Receipt, resolvedError *isc.VMError) *ReceiptResponse {
	burnRecords := make([]gas.BurnRecord, 0)

	if receipt.GasBurnLog != nil {
		burnRecords = append(burnRecords, receipt.GasBurnLog.Records...)
	}

	return &ReceiptResponse{
		Request:       hexutil.Encode(receipt.Request),
		Error:         MapReceiptError(resolvedError),
		BlockIndex:    receipt.BlockIndex,
		RequestIndex:  receipt.RequestIndex,
		GasBudget:     receipt.GasBudget,
		GasBurned:     receipt.GasBurned,
		GasFeeCharged: receipt.GasFeeCharged,
		GasBurnLog:    burnRecords,
	}
}
