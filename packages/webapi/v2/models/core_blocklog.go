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
	GoverningAddress string `json:"governingAddress"`
	SinceBlockIndex  uint32 `json:"sinceBlockIndex"`
	StateAddress     string `json:"stateAddress"`
}

type BlockInfoResponse struct {
	AnchorTransactionID         string    `json:"anchorTransactionId"`
	BlockIndex                  uint32    `json:"blockIndex"`
	GasBurned                   uint64    `json:"gasBurned"`
	GasFeeCharged               uint64    `json:"gasFeeCharged"`
	L1CommitmentHash            string    `json:"l1CommitmentHash"`
	NumOffLedgerRequests        uint16    `json:"numOffLedgerRequests"`
	NumSuccessfulRequests       uint16    `json:"numSuccessfulRequests"`
	PreviousL1CommitmentHash    string    `json:"previousL1CommitmentHash"`
	Timestamp                   time.Time `json:"timestamp"`
	TotalBaseTokensInL2Accounts uint64    `json:"totalBaseTokensInL2Accounts"`
	TotalRequests               uint16    `json:"totalRequests"`
	TotalStorageDeposit         uint64    `json:"totalStorageDeposit"`
	TransactionSubEssenceHash   string    `json:"transactionSubEssenceHash"`
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
	RequestIDs []string `json:"requestIds"`
}

type BlockReceiptError struct {
	Hash         string `json:"hash"`
	ErrorMessage string `json:"errorMessage"`
}

type FungibleTokens struct {
	BaseTokens   uint64         `json:"baseTokens"`
	NativeTokens []*NativeToken `json:"nativeTokens"`
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
	FungibleTokens *FungibleTokens `json:"fungibleTokens"`
	NFTs           []string        `json:"nfts"`
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
	Allowance      *Allowance       `json:"allowance"`
	CallTarget     isc.CallTarget   `json:"callTarget"`
	FungibleTokens *FungibleTokens  `json:"fungibleTokens"`
	GasGudget      uint64           `json:"gasGudget"`
	IsEVM          bool             `json:"isEVM"`
	IsOffLedger    bool             `json:"isOffLedger"`
	NFT            *NFTDataResponse `json:"nft"`
	Params         dict.JSONDict    `json:"params"`
	RequestID      string           `json:"requestId"`
	SenderAccount  string           `json:"senderAccount"`
	TargetAddress  string           `json:"targetAddress"`
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
	BlockIndex    uint32             `json:"blockIndex"`
	Error         *BlockReceiptError `json:"error"`
	GasBudget     uint64             `json:"gasBudget"`
	GasBurnLog    *gas.BurnLog       `json:"gasBurnLog"`
	GasBurned     uint64             `json:"gasBurned"`
	GasFeeCharged uint64             `json:"gasFeeCharged"`
	Request       *RequestDetail     `json:"request"`
	RequestIndex  uint16             `json:"requestIndex"`
}

type BlockReceiptsResponse struct {
	Receipts []*RequestReceiptResponse `json:"receipts"`
}

type RequestProcessedResponse struct {
	ChainID     string `json:"chainId"`
	RequestID   string `json:"requestId"`
	IsProcessed bool   `json:"isProcessed"`
}

type EventsResponse struct {
	Events []string `json:"events"`
}
