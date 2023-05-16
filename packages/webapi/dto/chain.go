package dto

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type (
	ContractsMap map[isc.Hname]*root.ContractRecord
)

type ChainInfo struct {
	IsActive                bool
	ChainID                 isc.ChainID
	ChainOwnerID            isc.AgentID
	GasFeePolicy            *gas.FeePolicy
	GasLimits               *gas.Limits
	PublicURL               string
	MetadataEVMJsonRPCURL   string
	MetadataEVMWebSocketURL string
}

func MapChainInfo(info *isc.ChainInfo, isActive bool) *ChainInfo {
	chainInfo := &ChainInfo{
		IsActive:                isActive,
		ChainID:                 info.ChainID,
		ChainOwnerID:            info.ChainOwnerID,
		GasFeePolicy:            info.GasFeePolicy,
		GasLimits:               info.GasLimits,
		PublicURL:               info.PublicURL,
		MetadataEVMJsonRPCURL:   info.MetadataEVMJsonRPCURL,
		MetadataEVMWebSocketURL: info.MetadataEVMWebSocketURL,
	}

	return chainInfo
}
