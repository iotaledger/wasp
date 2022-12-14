package models

import (
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type ControlAddressesResponse struct {
	GoverningAddress string
	SinceBlockIndex  uint32
	StateAddress     string
}

type BlockInfoResponse struct {
	AnchorTransactionID         string
	BlockIndex                  uint32
	GasBurned                   uint64
	GasFeeCharged               uint64
	L1CommitmentHash            string
	NumOffLedgerRequests        uint16
	NumSuccessfulRequests       uint16
	PreviousL1CommitmentHash    string
	Timestamp                   time.Time
	TotalBaseTokensInL2Accounts uint64
	TotalRequests               uint16
	TotalStorageDeposit         uint64
	TransactionSubEssenceHash   string
}

func MapBlockInfoResponse(info *blocklog.BlockInfo) *BlockInfoResponse {
	transactionEssenceHash := hexutil.Encode(info.TransactionSubEssenceHash[:])
	commitmentHash := ""

	if info.L1Commitment != nil {
		commitmentHash = info.L1Commitment.BlockHash.String()
	}

	return &BlockInfoResponse{
		AnchorTransactionID:         info.AnchorTransactionID.ToHex(),
		BlockIndex:                  info.BlockIndex,
		GasBurned:                   info.GasBurned,
		GasFeeCharged:               info.GasFeeCharged,
		L1CommitmentHash:            commitmentHash,
		NumOffLedgerRequests:        info.NumOffLedgerRequests,
		NumSuccessfulRequests:       info.NumSuccessfulRequests,
		PreviousL1CommitmentHash:    info.PreviousL1Commitment.BlockHash.String(),
		Timestamp:                   info.Timestamp,
		TotalBaseTokensInL2Accounts: info.TotalBaseTokensInL2Accounts,
		TotalRequests:               info.TotalRequests,
		TotalStorageDeposit:         info.TotalStorageDeposit,
		TransactionSubEssenceHash:   transactionEssenceHash,
	}
}

type RequestIDsResponse struct {
	RequestIDs []string
}

type BlockReceiptError struct {
	Hash         string
	ErrorMessage string
}

type FungibleTokens struct {
	BaseTokens uint64
	Tokens     []*NativeToken
}

func MapFungibleTokens(tokens *isc.FungibleTokens) *FungibleTokens {
	if tokens == nil {
		return nil
	}

	return &FungibleTokens{
		BaseTokens: tokens.BaseTokens,
		Tokens:     MapNativeTokens(tokens.Tokens),
	}
}

type Allowance struct {
	FungibleTokens *FungibleTokens
	NFTs           []string
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
	Allowance      *Allowance
	CallTarget     isc.CallTarget
	FungibleTokens *FungibleTokens
	GasGudget      uint64
	IsEVM          bool
	IsOffLedger    bool
	NFT            *NFTDataResponse
	Params         dict.Dict
	RequestID      string
	SenderAccount  string
	TargetAddress  string
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
		Params:         request.Params(),
		RequestID:      request.ID().String(),
		SenderAccount:  request.SenderAccount().String(),
		TargetAddress:  request.TargetAddress().Bech32(parameters.L1().Protocol.Bech32HRP),
	}
}

type RequestReceiptResponse struct {
	BlockIndex    uint32
	Error         *BlockReceiptError
	GasBudget     uint64
	GasBurnLog    *gas.BurnLog
	GasBurned     uint64
	GasFeeCharged uint64
	Request       *RequestDetail
	RequestIndex  uint16
}

type BlockReceiptsResponse struct {
	Receipts []*RequestReceiptResponse
}

type RequestProcessedResponse struct {
	ChainID     string
	RequestID   string
	IsProcessed bool
}

type EventsResponse struct {
	Events []string
}
