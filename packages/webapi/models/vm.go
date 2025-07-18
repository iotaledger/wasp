// Package models implements VM models for the webapi
package models

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
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

type OnLedgerEstimationResponse struct {
	L1 *L1EstimationResult `json:"l1" swagger:"required"`
	L2 *ReceiptResponse    `json:"l2" swagger:"required"`
}

func MapL1EstimationResult(gasSummary *iotajsonrpc.GasCostSummary) *L1EstimationResult {
	// Total L1 gas = computation cost + storage cost - storage rebate
	var totalGas big.Int
	totalGas.Add(&totalGas, gasSummary.ComputationCost.Int)
	totalGas.Add(&totalGas, gasSummary.StorageCost.Int)
	totalGas.Sub(&totalGas, gasSummary.StorageRebate.Int)

	gasBudget := totalGas
	if gasBudget.Cmp(gasSummary.ComputationCost.Int) < 0 {
		// L1 gas budget must be >= max(computation cost, total cost)
		// See: https://docs.iota.org/about-iota/tokenomics/gas-in-iota#gas-budgets
		gasBudget.Set(gasSummary.ComputationCost.Int)
	}
	if gasBudget.Cmp(big.NewInt(iotaclient.MinGasBudget)) < 0 {
		// L1 gas budget must be at least 1,000,000
		gasBudget.SetInt64(iotaclient.MinGasBudget)
	}

	return &L1EstimationResult{
		ComputationFee: gasSummary.ComputationCost.String(),
		StorageFee:     gasSummary.StorageCost.String(),
		StorageRebate:  gasSummary.StorageRebate.String(),
		GasFeeCharged:  totalGas.String(),
		GasBudget:      gasBudget.String(),
	}
}

type L1EstimationResult struct {
	ComputationFee string `json:"computationFee,omitempty" swagger:"required,desc(Gas cost for computation (uint64 as string))"`
	StorageFee     string `json:"storageFee,omitempty" swagger:"required,desc(Gas cost for storage (uint64 as string))"`
	StorageRebate  string `json:"storageRebate,omitempty" swagger:"required,desc(Gas rebate for storage (uint64 as string))"`
	GasFeeCharged  string `json:"gasFeeCharged,omitempty" swagger:"required,desc(Total gas fee charged: computation fee + storage fee - storage rebate (uint64 as string))"`
	GasBudget      string `json:"gasBudget,omitempty" swagger:"required,desc(Gas budget required for processing of transaction: max(computation fee, total fee) (uint64 as string))"`
}

type UnresolvedVMErrorJSON struct {
	ErrorCode string   `json:"code"`
	Params    []string `json:"params"`
}

// ToUnresolvedVMErrorJSON produces the params as humanly readable json, and the uints as strings
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
