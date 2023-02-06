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
	GoverningAddress string `json:"governingAddress" swagger:"required"`
	SinceBlockIndex  uint32 `json:"sinceBlockIndex" swagger:"required,min(1)"`
	StateAddress     string `json:"stateAddress" swagger:"required"`
}

type BlockInfoResponse struct {
	AnchorTransactionID         string    `json:"anchorTransactionId" swagger:"required"`
	BlockIndex                  uint32    `json:"blockIndex" swagger:"required,min(1)"`
	GasBurned                   string    `json:"gasBurned" swagger:"required,desc(The burned gas (uint64 as string))"`
	GasFeeCharged               string    `json:"gasFeeCharged" swagger:"required,desc(The charged gas fee (uint64 as string))"`
	L1CommitmentHash            string    `json:"l1CommitmentHash" swagger:"required"`
	NumOffLedgerRequests        uint16    `json:"numOffLedgerRequests" swagger:"required,min(1)"`
	NumSuccessfulRequests       uint16    `json:"numSuccessfulRequests" swagger:"required,min(1)"`
	PreviousL1CommitmentHash    string    `json:"previousL1CommitmentHash" swagger:"required"`
	Timestamp                   time.Time `json:"timestamp" swagger:"required"`
	TotalBaseTokensInL2Accounts string    `json:"totalBaseTokensInL2Accounts" swagger:"required,desc(The total L2 base tokens (uint64 as string))"`
	TotalRequests               uint16    `json:"totalRequests" swagger:"required,min(1)"`
	TotalStorageDeposit         string    `json:"totalStorageDeposit" swagger:"required,desc(The total storage deposit (uint64 as string))"`
	TransactionSubEssenceHash   string    `json:"transactionSubEssenceHash" swagger:"required"`
}

func MapBlockInfoResponse(info *blocklog.BlockInfo) *BlockInfoResponse {
	transactionEssenceHash := iotago.EncodeHex(info.TransactionSubEssenceHash[:])
	commitmentHash := ""

	if info.L1Commitment != nil {
		commitmentHash = info.L1Commitment.BlockHash().String()
	}

	return &BlockInfoResponse{
		AnchorTransactionID:         info.AnchorTransactionID.ToHex(),
		BlockIndex:                  info.BlockIndex,
		GasBurned:                   iotago.EncodeUint64(info.GasBurned),
		GasFeeCharged:               iotago.EncodeUint64(info.GasFeeCharged),
		L1CommitmentHash:            commitmentHash,
		NumOffLedgerRequests:        info.NumOffLedgerRequests,
		NumSuccessfulRequests:       info.NumSuccessfulRequests,
		PreviousL1CommitmentHash:    info.PreviousL1Commitment.BlockHash().String(),
		Timestamp:                   info.Timestamp,
		TotalBaseTokensInL2Accounts: iotago.EncodeUint64(info.TotalBaseTokensInL2Accounts),
		TotalRequests:               info.TotalRequests,
		TotalStorageDeposit:         iotago.EncodeUint64(info.TotalStorageDeposit),
		TransactionSubEssenceHash:   transactionEssenceHash,
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

type RequestDetail struct {
	Allowance     *Assets          `json:"allowance" swagger:"required"`
	CallTarget    isc.CallTarget   `json:"callTarget" swagger:"required"`
	Assets        *Assets          `json:"fungibleTokens" swagger:"required"`
	GasGudget     string           `json:"gasGudget,string" swagger:"required,desc(The gas budget (uint64 as string))"`
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
		CallTarget:    request.CallTarget(),
		Assets:        mapAssets(request.Assets()),
		GasGudget:     iotago.EncodeUint64(gasBudget),
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
