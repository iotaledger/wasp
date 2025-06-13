// Package models implements VM models for the webapi
package models

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type ReceiptResponse struct {
	Request       isc.RequestJSON            `json:"request" swagger:"required"`
	RawError      *isc.UnresolvedVMErrorJSON `json:"rawError,omitempty"`
	ErrorMessage  string                     `json:"errorMessage,omitempty"`
	GasBudget     string                     `json:"gasBudget" swagger:"required,desc(The gas budget (uint64 as string))"`
	GasBurned     string                     `json:"gasBurned" swagger:"required,desc(The burned gas (uint64 as string))"`
	GasFeeCharged string                     `json:"gasFeeCharged" swagger:"required,desc(The charged gas fee (uint64 as string))"`
	SDCharged     string                     `json:"storageDepositCharged" swagger:"required,desc(Storage deposit charged (uint64 as string))"`
	BlockIndex    uint32                     `json:"blockIndex" swagger:"required,min(1)"`
	RequestIndex  uint16                     `json:"requestIndex" swagger:"required,min(1)"`
	GasBurnLog    []gas.BurnRecord           `json:"gasBurnLog" swagger:"required"`
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
		Request:       isc.RequestToJSONObject(req),
		RawError:      receipt.Error.ToJSONStruct(),
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

type OnLedgerEstimationResponse struct {
	L1 *L1EstimationResult `json:"l1" swagger:"required"`
	L2 *ReceiptResponse    `json:"l2" swagger:"required"`
}

func MapL1EstimationResult(gasFeeCharged, minGasBudget *big.Int) *L1EstimationResult {
	return &L1EstimationResult{
		GasFeeCharged: gasFeeCharged.String(),
		GasBudget:     minGasBudget.String(),
	}
}

type L1EstimationResult struct {
	GasFeeCharged string `json:"gasFeeCharged,omitempty" swagger:"required,desc(Total gas fee charged (uint64 as string))"`
	GasBudget     string `json:"gasBudget,omitempty" swagger:"required,desc(Gas budget required for processing of transaction (uint64 as string))"`
}
