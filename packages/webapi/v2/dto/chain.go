package dto

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type (
	ContractsMap map[isc.Hname]*root.ContractRecord
)

type ChainInfo struct {
	IsActive        bool
	ChainID         *isc.ChainID
	ChainOwnerID    isc.AgentID
	Description     string
	GasFeePolicy    *gas.GasFeePolicy
	MaxBlobSize     uint32
	MaxEventSize    uint16
	MaxEventsPerReq uint16
}

func MapChainInfo(info *governance.ChainInfo, isActive bool) *ChainInfo {
	return &ChainInfo{
		IsActive:        isActive,
		ChainID:         info.ChainID,
		ChainOwnerID:    info.ChainOwnerID,
		Description:     info.Description,
		GasFeePolicy:    info.GasFeePolicy,
		MaxBlobSize:     info.MaxBlobSize,
		MaxEventSize:    info.MaxEventSize,
		MaxEventsPerReq: info.MaxEventsPerReq,
	}
}

type OffLedgerRequestBody struct {
	Request string `swagger:"desc(Offledger Request (base64))"`
}
