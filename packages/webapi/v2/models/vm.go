package models

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type ReceiptError struct {
	ContractID isc.Hname
	ErrorID    uint16
	ErrorCode  string
	Message    string
	Parameters []interface{}
}

func MapReceiptError(err *isc.VMError) *ReceiptError {
	if err == nil {
		return nil
	}

	return &ReceiptError{
		ContractID: err.Code().ContractID,
		ErrorID:    err.Code().ID,
		ErrorCode:  err.Code().String(),
		Message:    err.Error(),
		Parameters: err.Params(),
	}
}

type Receipt struct {
	Request       []byte           `json:"request"`
	Error         *ReceiptError    `json:"error"`
	GasBudget     uint64           `json:"gasBudget"`
	GasBurned     uint64           `json:"gasBurned"`
	GasFeeCharged uint64           `json:"gasFeeCharged"`
	BlockIndex    uint32           `json:"blockIndex"`
	RequestIndex  uint16           `json:"requestIndex"`
	GasBurnLog    []gas.BurnRecord `json:"gasBurnLog"`
}

func MapReceipt(receipt *isc.Receipt, resolvedError *isc.VMError) *Receipt {
	return &Receipt{
		Request:       receipt.Request,
		Error:         MapReceiptError(resolvedError),
		BlockIndex:    receipt.BlockIndex,
		RequestIndex:  receipt.RequestIndex,
		GasBudget:     receipt.GasBudget,
		GasBurned:     receipt.GasBurned,
		GasFeeCharged: receipt.GasFeeCharged,
		GasBurnLog:    receipt.GasBurnLog.Records,
	}
}
