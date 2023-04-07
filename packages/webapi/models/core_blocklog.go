package models

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type ControlAddressesResponse struct {
	GoverningAddress string `json:"governingAddress" swagger:"required,desc(The governing address (Bech32))"`
	SinceBlockIndex  uint32 `json:"sinceBlockIndex" swagger:"required,min(1),desc(The block index (uint32)"`
	StateAddress     string `json:"stateAddress" swagger:"required,desc(The state address (Bech32))"`
}

type BlockInfoResponse struct {
	BlockIndex            uint32    `json:"blockIndex" swagger:"required,min(1)"`
	Timestamp             time.Time `json:"timestamp" swagger:"required"`
	TotalRequests         uint16    `json:"totalRequests" swagger:"required,min(1)"`
	NumSuccessfulRequests uint16    `json:"numSuccessfulRequests" swagger:"required,min(1)"`
	NumOffLedgerRequests  uint16    `json:"numOffLedgerRequests" swagger:"required,min(1)"`
	PreviousAliasOutput   string    `json:"previousAliasOutput" swagger:"required,min(1)"`
	GasBurned             string    `json:"gasBurned" swagger:"required,desc(The burned gas (uint64 as string))"`
	GasFeeCharged         string    `json:"gasFeeCharged" swagger:"required,desc(The charged gas fee (uint64 as string))"`
}

func MapBlockInfoResponse(info *blocklog.BlockInfo) *BlockInfoResponse {
	blockindex := uint32(0)
	prevAOStr := ""
	if info.PreviousAliasOutput != nil {
		blockindex = info.PreviousAliasOutput.GetAliasOutput().StateIndex + 1
		prevAOStr = string(info.PreviousAliasOutput.Bytes())
	}
	return &BlockInfoResponse{
		BlockIndex:            blockindex,
		PreviousAliasOutput:   prevAOStr,
		Timestamp:             info.Timestamp,
		TotalRequests:         info.TotalRequests,
		NumSuccessfulRequests: info.NumSuccessfulRequests,
		NumOffLedgerRequests:  info.NumOffLedgerRequests,
		GasBurned:             iotago.EncodeUint64(info.GasBurned),
		GasFeeCharged:         iotago.EncodeUint64(info.GasFeeCharged),
	}
}

type RequestIDsResponse struct {
	RequestIDs []string `json:"requestIds" swagger:"required"`
}

type BlockReceiptError struct {
	Hash         string `json:"hash" swagger:"required"`
	ErrorMessage string `json:"errorMessage" swagger:"required"`
}

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

func MapRequestDetail(request isc.Request) *RequestDetail {
	gasBudget, isEVM := request.GasBudget()

	return &RequestDetail{
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

type RequestReceiptResponse struct {
	BlockIndex    uint32             `json:"blockIndex" swagger:"required,min(1)"`
	Error         *BlockReceiptError `json:"error" swagger:""`
	GasBudget     string             `json:"gasBudget" swagger:"required,desc(The gas budget (uint64 as string))"`
	GasBurnLog    *gas.BurnLog       `json:"gasBurnLog" swagger:"required"`
	GasBurned     string             `json:"gasBurned" swagger:"required,desc(The burned gas (uint64 as string))"`
	GasFeeCharged string             `json:"gasFeeCharged" swagger:"required,desc(The charged gas fee (uint64 as string))"`
	Request       *RequestDetail     `json:"request" swagger:"required"`
	RequestIndex  uint16             `json:"requestIndex" swagger:"required,min(1)"`
}

type BlockReceiptsResponse struct {
	Receipts []*RequestReceiptResponse `json:"receipts" swagger:"required"`
}

type RequestProcessedResponse struct {
	ChainID     string `json:"chainId" swagger:"required"`
	RequestID   string `json:"requestId" swagger:"required"`
	IsProcessed bool   `json:"isProcessed" swagger:"required"`
}

type EventsResponse struct {
	Events []string `json:"events" swagger:"required"`
}
