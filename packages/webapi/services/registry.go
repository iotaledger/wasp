package services

import (
	"github.com/iotaledger/wasp/v2/packages/chains"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
)

type RegistryService struct {
	chainsProvider              chains.Provider
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
}

func NewRegistryService(chainsProvider chains.Provider, chainRecordRegistryProvider registry.ChainRecordRegistryProvider) interfaces.RegistryService {
	return &RegistryService{
		chainsProvider:              chainsProvider,
		chainRecordRegistryProvider: chainRecordRegistryProvider,
	}
}

func (c *RegistryService) GetChainRecordByChainID() (*registry.ChainRecord, error) {
	return c.chainRecordRegistryProvider.ChainRecord(chainID)
}
