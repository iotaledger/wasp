package models

import (
	"fmt"
	"reflect"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type ReceiptError struct {
	ContractHName string           `json:"contractHName" swagger:"required,desc(The contract hname (Hex))"`
	ErrorID       uint16           `json:"errorId" swagger:"required,min(1)"`
	ErrorCode     string           `json:"errorCode" swagger:"required"`
	Message       string           `json:"message" swagger:"required"`
	MessageFormat string           `json:"messageFormat" swagger:"required"`
	Parameters    []ErrorParameter `json:"parameters" swagger:""`
}

type ErrorParameter struct {
	Value string `json:"value" swagger:"required"`
	Type  string `json:"type" swagger:"required"`
}

func MapErrorParameter(param interface{}) ErrorParameter {
	result := ErrorParameter{
		Value: fmt.Sprintf("%v", param),
	}

	paramType := reflect.TypeOf(param)
	if paramType != nil {
		result.Type = paramType.String()
	}

	return result
}

func MapReceiptError(vmError *isc.VMError) *ReceiptError {
	if vmError == nil {
		return nil
	}

	receiptError := &ReceiptError{
		ContractHName: vmError.Code().ContractID.String(),
		ErrorID:       vmError.Code().ID,
		ErrorCode:     vmError.Code().String(),
		Message:       vmError.Error(),
		MessageFormat: vmError.MessageFormat(),
		Parameters:    make([]ErrorParameter, len(vmError.Params())),
	}

	for k, v := range vmError.Params() {
		receiptError.Parameters[k] = MapErrorParameter(v)
	}

	return receiptError
}

type ReceiptResponse struct {
	Request       string           `json:"request" swagger:"required"`
	Error         *ReceiptError    `json:"error"`
	GasBudget     string           `json:"gasBudget" swagger:"required,desc(The gas budget (uint64 as string))"`
	GasBurned     string           `json:"gasBurned" swagger:"required,desc(The burned gas (uint64 as string))"`
	GasFeeCharged string           `json:"gasFeeCharged" swagger:"required,desc(The charged gas fee (uint64 as string))"`
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
		GasBudget:     iotago.EncodeUint64(receipt.GasBudget),
		GasBurned:     iotago.EncodeUint64(receipt.GasBurned),
		GasFeeCharged: iotago.EncodeUint64(receipt.GasFeeCharged),
		GasBurnLog:    burnRecords,
	}
}
