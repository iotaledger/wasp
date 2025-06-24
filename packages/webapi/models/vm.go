// Package models implements VM models for the webapi
package models

import (
	"fmt"
	"reflect"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type ReceiptResponse struct {
	Request       RequestJSON            `json:"request" swagger:"required"`
	RawError      *UnresolvedVMErrorJSON `json:"rawError,omitempty"`
	ErrorMessage  string                 `json:"errorMessage,omitempty"`
	GasBudget     string                 `json:"gasBudget" swagger:"required,desc(The gas budget (uint64 as string))"`
	GasBurned     string                 `json:"gasBurned" swagger:"required,desc(The burned gas (uint64 as string))"`
	GasFeeCharged string                 `json:"gasFeeCharged" swagger:"required,desc(The charged gas fee (uint64 as string))"`
	SDCharged     string                 `json:"storageDepositCharged" swagger:"required,desc(Storage deposit charged (uint64 as string))"`
	BlockIndex    uint32                 `json:"blockIndex" swagger:"required,min(1)"`
	RequestIndex  uint16                 `json:"requestIndex" swagger:"required,min(1)"`
	GasBurnLog    []gas.BurnRecord       `json:"gasBurnLog" swagger:"required"`
}

func MapReceiptResponse(receipt *isc.Receipt) *ReceiptResponse {
	burnRecords := make([]gas.BurnRecord, 0)

	if receipt.GasBurnLog != nil {
		burnRecords = append(burnRecords, receipt.GasBurnLog.Records...)
	}

	req, err := isc.RequestFromBytes(receipt.Request)
	if err != nil {
		panic(err)
	}

	return &ReceiptResponse{
		Request:       RequestToJSONObject(req),
		RawError:      ToUnresolvedVMErrorJSON(receipt.Error),
		ErrorMessage:  receipt.ResolvedError,
		BlockIndex:    receipt.BlockIndex,
		RequestIndex:  receipt.RequestIndex,
		GasBudget:     fmt.Sprint(receipt.GasBudget),
		GasBurned:     fmt.Sprint(receipt.GasBurned),
		GasFeeCharged: receipt.GasFeeCharged.String(),
		SDCharged:     receipt.SDCharged.String(),
		GasBurnLog:    burnRecords,
	}
}

type UnresolvedVMErrorJSON struct {
	ErrorCode string   `json:"code"`
	Params    []string `json:"params"`
}

// ToJSONStruct produces the params as humanly readable json, and the uints as strings
func ToUnresolvedVMErrorJSON(e *isc.UnresolvedVMError) *UnresolvedVMErrorJSON {
	if e == nil {
		return &UnresolvedVMErrorJSON{
			Params:    []string{},
			ErrorCode: "",
		}
	}
	return &UnresolvedVMErrorJSON{
		Params:    humanlyReadableParams(e.Params),
		ErrorCode: e.ErrorCode.String(),
	}
}

func humanlyReadableParams(params []isc.VMErrorParam) []string {
	res := make([]string, len(params))
	for i, param := range params {
		res[i] = fmt.Sprintf("%v:%s", param, reflect.TypeOf(param).String())
	}
	return res
}
