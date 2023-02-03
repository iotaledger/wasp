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
	GasBurned                   uint64    `json:"gasBurned" swagger:"required"`
	GasFeeCharged               uint64    `json:"gasFeeCharged" swagger:"required"`
	L1CommitmentHash            string    `json:"l1CommitmentHash" swagger:"required"`
	NumOffLedgerRequests        uint16    `json:"numOffLedgerRequests" swagger:"required,min(1)"`
	NumSuccessfulRequests       uint16    `json:"numSuccessfulRequests" swagger:"required,min(1)"`
	PreviousL1CommitmentHash    string    `json:"previousL1CommitmentHash" swagger:"required"`
	Timestamp                   time.Time `json:"timestamp" swagger:"required"`
	TotalBaseTokensInL2Accounts uint64    `json:"totalBaseTokensInL2Accounts" swagger:"required"`
	TotalRequests               uint16    `json:"totalRequests" swagger:"required,min(1)"`
	TotalStorageDeposit         uint64    `json:"totalStorageDeposit" swagger:"required"`
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
		GasBurned:                   info.GasBurned,
		GasFeeCharged:               info.GasFeeCharged,
		L1CommitmentHash:            commitmentHash,
		NumOffLedgerRequests:        info.NumOffLedgerRequests,
		NumSuccessfulRequests:       info.NumSuccessfulRequests,
		PreviousL1CommitmentHash:    info.PreviousL1Commitment.BlockHash().String(),
		Timestamp:                   info.Timestamp,
		TotalBaseTokensInL2Accounts: info.TotalBaseTokensInL2Accounts,
		TotalRequests:               info.TotalRequests,
		TotalStorageDeposit:         info.TotalStorageDeposit,
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

type FungibleTokens struct {
	BaseTokens   uint64         `json:"baseTokens" swagger:"required"`
	NativeTokens []*NativeToken `json:"nativeTokens" swagger:"required"`
}

func MapFungibleTokens(tokens *isc.FungibleTokens) *FungibleTokens {
	if tokens == nil {
		return nil
	}

	return &FungibleTokens{
		BaseTokens:   tokens.BaseTokens,
		NativeTokens: MapNativeTokens(tokens.NativeTokens),
	}
}

type Allowance struct {
	FungibleTokens *FungibleTokens `json:"fungibleTokens" swagger:"required"`
	NFTs           []string        `json:"nfts" swagger:"required"`
}

func MapAllowance(allowance *isc.Allowance) *Allowance {
	if allowance == nil {
		return nil
	}

	allowanceResult := Allowance{
		FungibleTokens: MapFungibleTokens(allowance.Assets),
		NFTs:           make([]string, len(allowance.NFTs)),
	}

	for k, v := range allowance.NFTs {
		allowanceResult.NFTs[k] = v.ToHex()
	}

	return &allowanceResult
}

type RequestDetail struct {
	Allowance      *Allowance       `json:"allowance" swagger:"required"`
	CallTarget     isc.CallTarget   `json:"callTarget" swagger:"required"`
	FungibleTokens *FungibleTokens  `json:"fungibleTokens" swagger:"required"`
	GasGudget      uint64           `json:"gasGudget" swagger:"required"`
	IsEVM          bool             `json:"isEVM" swagger:"required"`
	IsOffLedger    bool             `json:"isOffLedger" swagger:"required"`
	NFT            *NFTDataResponse `json:"nft" swagger:"required"`
	Params         dict.JSONDict    `json:"params" swagger:"required"`
	RequestID      string           `json:"requestId" swagger:"required"`
	SenderAccount  string           `json:"senderAccount" swagger:"required"`
	TargetAddress  string           `json:"targetAddress" swagger:"required"`
}

func MapRequestDetail(request isc.Request) *RequestDetail {
	gasBudget, isEVM := request.GasBudget()

	return &RequestDetail{
		Allowance:      MapAllowance(request.Allowance()),
		CallTarget:     request.CallTarget(),
		FungibleTokens: MapFungibleTokens(request.FungibleTokens()),
		GasGudget:      gasBudget,
		IsEVM:          isEVM,
		IsOffLedger:    request.IsOffLedger(),
		NFT:            MapNFTDataResponse(request.NFT()),
		Params:         request.Params().JSONDict(),
		RequestID:      request.ID().String(),
		SenderAccount:  request.SenderAccount().String(),
		TargetAddress:  request.TargetAddress().Bech32(parameters.L1().Protocol.Bech32HRP),
	}
}

type RequestReceiptResponse struct {
	BlockIndex    uint32             `json:"blockIndex" swagger:"required,min(1)"`
	Error         *BlockReceiptError `json:"error" swagger:""`
	GasBudget     uint64             `json:"gasBudget" swagger:"required"`
	GasBurnLog    *gas.BurnLog       `json:"gasBurnLog" swagger:"required"`
	GasBurned     uint64             `json:"gasBurned" swagger:"required"`
	GasFeeCharged uint64             `json:"gasFeeCharged" swagger:"required"`
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
