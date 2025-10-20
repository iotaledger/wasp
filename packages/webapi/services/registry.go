package services

import (
	"github.com/iotaledger/wasp/v2/packages/chains"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
)

type RegistryService struct {
	chainsProvider      chains.Provider
	chainRecordRegistry registry.ChainRecordRegistry
}

func NewRegistryService(chainsProvider chains.Provider, chainRecordRegistry registry.ChainRecordRegistry) interfaces.RegistryService {
	return &RegistryService{
		chainsProvider:      chainsProvider,
		chainRecordRegistry: chainRecordRegistry,
	}
}

func (c *RegistryService) GetChainRecordByChainID(chainID isc.ChainID) (*registry.ChainRecord, error) {
	return c.chainRecordRegistry.ChainRecord(chainID)
}
