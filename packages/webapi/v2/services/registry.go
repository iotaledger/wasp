package services

import (
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type RegistryService struct {
	logger *logger.Logger

	chainsProvider              chains.Provider
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
}

func NewRegistryService(log *logger.Logger, chainsProvider chains.Provider, chainRecordRegistryProvider registry.ChainRecordRegistryProvider) interfaces.RegistryService {
	return &RegistryService{
		logger:                      log,
		chainsProvider:              chainsProvider,
		chainRecordRegistryProvider: chainRecordRegistryProvider,
	}
}

func (c *RegistryService) GetChainRecordByChainID(chainID isc.ChainID) (*registry.ChainRecord, error) {
	return c.chainRecordRegistryProvider.ChainRecord(chainID)
}
