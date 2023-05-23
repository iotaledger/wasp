package dto

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type (
	ContractsMap map[isc.Hname]*root.ContractRecord
)

type ChainMetadata struct {
	EVMJsonRPCURL   string `json:"evmJsonRpcUrl" swagger:"desc(The EVM json rpc url),required"`
	EVMWebSocketURL string `json:"evmWebSocketUrl" swagger:"desc(The EVM websocket url)),required"`

	ChainName        string `json:"chainName" swagger:"desc(The name of the chain),required"`
	ChainDescription string `json:"chainDescription" swagger:"desc(The description of the chain.),required"`
	ChainOwnerEmail  string `json:"chainOwnerEmail" swagger:"desc(The email of the chain owner.),required"`
	ChainWebsite     string `json:"chainWebsite" swagger:"desc(The official website of the chain.),required"`
}

type ChainInfo struct {
	IsActive     bool
	ChainID      isc.ChainID
	ChainOwnerID isc.AgentID
	GasFeePolicy *gas.FeePolicy
	GasLimits    *gas.Limits
	PublicURL    string

	Metadata ChainMetadata
}

func MapChainInfo(info *isc.ChainInfo, isActive bool) *ChainInfo {
	chainInfo := &ChainInfo{
		IsActive:     isActive,
		ChainID:      info.ChainID,
		ChainOwnerID: info.ChainOwnerID,
		GasFeePolicy: info.GasFeePolicy,
		GasLimits:    info.GasLimits,
		PublicURL:    info.PublicURL,
		Metadata: ChainMetadata{
			EVMJsonRPCURL:    info.Metadata.EVMJsonRPCURL,
			EVMWebSocketURL:  info.Metadata.EVMWebSocketURL,
			ChainName:        info.Metadata.ChainName,
			ChainDescription: info.Metadata.ChainDescription,
			ChainOwnerEmail:  info.Metadata.ChainOwnerEmail,
			ChainWebsite:     info.Metadata.ChainWebsite,
		},
	}

	return chainInfo
}
