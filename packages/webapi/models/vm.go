package models

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type Assets struct {
	BaseTokens   string         `json:"baseTokens" swagger:"required,desc(The base tokens (uint64 as string))"`
	NativeTokens []*NativeToken `json:"nativeTokens" swagger:"required"`
	NFTs         []string       `json:"nfts" swagger:"required"`
}

func mapAssets(assets *isc.Assets) *Assets {
	if assets == nil {
		return nil
	}

	ret := &Assets{
		BaseTokens:   iotago.EncodeUint64(assets.BaseTokens),
		NativeTokens: MapNativeTokens(assets.NativeTokens),
		NFTs:         make([]string, len(assets.NFTs)),
	}

	for k, v := range assets.NFTs {
		ret.NFTs[k] = v.ToHex()
	}
	return ret
}

type CallTarget struct {
	ContractHName string `json:"contractHName" swagger:"desc(The contract name as HName (Hex)),required"`
	FunctionHName string `json:"functionHName" swagger:"desc(The function name as HName (Hex)),required"`
}

func MapCallTarget(target isc.CallTarget) CallTarget {
	return CallTarget{
		ContractHName: target.Contract.String(),
		FunctionHName: target.EntryPoint.String(),
	}
}

type RequestDetail struct {
	Allowance     *Assets          `json:"allowance" swagger:"required"`
	CallTarget    CallTarget       `json:"callTarget" swagger:"required"`
	Assets        *Assets          `json:"fungibleTokens" swagger:"required"`
	GasBudget     string           `json:"gasBudget,string" swagger:"required,desc(The gas budget (uint64 as string))"`
	IsEVM         bool             `json:"isEVM" swagger:"required"`
	IsOffLedger   bool             `json:"isOffLedger" swagger:"required"`
	NFT           *NFTDataResponse `json:"nft" swagger:"required"`
	Params        dict.JSONDict    `json:"params" swagger:"required"`
	RequestID     string           `json:"requestId" swagger:"required"`
	SenderAccount string           `json:"senderAccount" swagger:"required"`
	TargetAddress string           `json:"targetAddress" swagger:"required"`
}

func mapRequestDetail(request isc.Request) RequestDetail {
	gasBudget, isEVM := request.GasBudget()

	return RequestDetail{
		Allowance:     mapAssets(request.Allowance()),
		CallTarget:    MapCallTarget(request.CallTarget()),
		Assets:        mapAssets(request.Assets()),
		GasBudget:     iotago.EncodeUint64(gasBudget),
		IsEVM:         isEVM,
		IsOffLedger:   request.IsOffLedger(),
		NFT:           MapNFTDataResponse(request.NFT()),
		Params:        request.Params().JSONDict(),
		RequestID:     request.ID().String(),
		SenderAccount: request.SenderAccount().String(),
		TargetAddress: request.TargetAddress().Bech32(parameters.L1().Protocol.Bech32HRP),
	}
}

type ReceiptResponse struct {
	Request       RequestDetail              `json:"request" swagger:"required"`
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
		Request:       mapRequestDetail(req),
		RawError:      receipt.Error.ToJSONStruct(),
		ErrorMessage:  receipt.ResolvedError,
		BlockIndex:    receipt.BlockIndex,
		RequestIndex:  receipt.RequestIndex,
		GasBudget:     iotago.EncodeUint64(receipt.GasBudget),
		GasBurned:     iotago.EncodeUint64(receipt.GasBurned),
		GasFeeCharged: iotago.EncodeUint64(receipt.GasFeeCharged),
		SDCharged:     iotago.EncodeUint64(receipt.SDCharged),
		GasBurnLog:    burnRecords,
	}
}
