package models

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type ReceiptError struct {
	ContractID    isc.Hname     `json:"contractId" swagger:"required"`
	ErrorID       uint16        `json:"errorId" swagger:"required,min(1)"`
	ErrorCode     string        `json:"errorCode" swagger:"required"`
	Message       string        `json:"message" swagger:"required"`
	MessageFormat string        `json:"messageFormat" swagger:"required"`
	Parameters    []interface{} `json:"parameters" swagger:"required"`
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
	Request       string           `json:"request" swagger:"required"`
	Error         *ReceiptError    `json:"error"`
	GasBudget     uint64           `json:"gasBudget" swagger:"required"`
	GasBurned     uint64           `json:"gasBurned" swagger:"required"`
	GasFeeCharged uint64           `json:"gasFeeCharged" swagger:"required"`
	BlockIndex    uint32           `json:"blockIndex" swagger:"required,min(1)"`
	RequestIndex  uint16           `json:"requestIndex" swagger:"required,min(1)"`
	GasBurnLog    []gas.BurnRecord `json:"gasBurnLog" swagger:"required"`
}

func MapReceiptResponse(receipt *isc.Receipt, resolvedError *isc.VMError) *ReceiptResponse {
	burnRecords := make([]gas.BurnRecord, 0)

	if receipt.GasBurnLog != nil {
		burnRecords = append(burnRecords, receipt.GasBurnLog.Records...)
	}

	return &ReceiptResponse{
		Request:       iotago.EncodeHex(receipt.Request),
		Error:         MapReceiptError(resolvedError),
		BlockIndex:    receipt.BlockIndex,
		RequestIndex:  receipt.RequestIndex,
		GasBudget:     receipt.GasBudget,
		GasBurned:     receipt.GasBurned,
		GasFeeCharged: receipt.GasFeeCharged,
		GasBurnLog:    burnRecords,
	}
}
