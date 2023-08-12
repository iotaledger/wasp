package models

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
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
		prevAOStr = iotago.EncodeHex(info.PreviousAliasOutput.Bytes())
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

type RequestProcessedResponse struct {
	ChainID     string `json:"chainId" swagger:"required"`
	RequestID   string `json:"requestId" swagger:"required"`
	IsProcessed bool   `json:"isProcessed" swagger:"required"`
}

type EventsResponse struct {
	Events []*isc.EventJSON `json:"events" swagger:"required"`
}
