package services

import (
	"github.com/iotaledger/wasp/v2/packages/chainrunner"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
)

type RegistryService struct {
	chainRunner                 *chainrunner.ChainRunner
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
}

func NewRegistryService(chainRunner *chainrunner.ChainRunner, chainRecordRegistryProvider registry.ChainRecordRegistryProvider) interfaces.RegistryService {
	return &RegistryService{
		chainRunner:                 chainRunner,
		chainRecordRegistryProvider: chainRecordRegistryProvider,
	}
}

func (c *RegistryService) GetChainRecord() (*registry.ChainRecord, error) {
	return c.chainRecordRegistryProvider.ChainRecord()
}
