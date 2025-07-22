package dto

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

type (
	ContractsMap []lo.Tuple2[*isc.Hname, *root.ContractRecord]
)

type PublicChainMetadata struct {
	EVMJsonRPCURL   string `json:"evmJsonRpcUrl" swagger:"desc(The EVM json rpc url),required"`
	EVMWebSocketURL string `json:"evmWebSocketUrl" swagger:"desc(The EVM websocket url)),required"`

	Name        string `json:"name" swagger:"desc(The name of the chain),required"`
	Description string `json:"description" swagger:"desc(The description of the chain.),required"`
	Website     string `json:"website" swagger:"desc(The official website of the chain.),required"`
}

type ChainInfo struct {
	IsActive     bool
	ChainID      isc.ChainID
	ChainAdmin   isc.AgentID
	GasFeePolicy *gas.FeePolicy
	GasLimits    *gas.Limits
	PublicURL    string

	Metadata PublicChainMetadata
}

func MapChainInfo(info *isc.ChainInfo, isActive bool) *ChainInfo {
	chainInfo := &ChainInfo{
		IsActive:     isActive,
		ChainID:      info.ChainID,
		ChainAdmin:   info.ChainAdmin,
		GasFeePolicy: info.GasFeePolicy,
		GasLimits:    info.GasLimits,
		PublicURL:    info.PublicURL,
		Metadata: PublicChainMetadata{
			EVMJsonRPCURL:   info.Metadata.EVMJsonRPCURL,
			EVMWebSocketURL: info.Metadata.EVMWebSocketURL,
			Name:            info.Metadata.Name,
			Description:     info.Metadata.Description,
			Website:         info.Metadata.Website,
		},
	}

	return chainInfo
}
