package services

import (
	"github.com/iotaledger/wasp/v2/packages/chainrunner"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
)

type RegistryService struct {
	chainsProvider              chainrunner.Provider
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
}

func NewRegistryService(chainsProvider chainrunner.Provider, chainRecordRegistryProvider registry.ChainRecordRegistryProvider) interfaces.RegistryService {
	return &RegistryService{
		chainsProvider:              chainsProvider,
		chainRecordRegistryProvider: chainRecordRegistryProvider,
	}
}

func (c *RegistryService) GetChainRecord() (*registry.ChainRecord, error) {
	return c.chainRecordRegistryProvider.ChainRecord()
}
