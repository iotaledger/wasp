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

	chainsProvider   chains.Provider
	registryProvider registry.Provider
}

func NewRegistryService(log *logger.Logger, chainsProvider chains.Provider, registryProvider registry.Provider) interfaces.Registry {
	return &RegistryService{
		logger: log,

		chainsProvider:   chainsProvider,
		registryProvider: registryProvider,
	}
}

func (c *RegistryService) GetChainRecordByChainID(chainID *isc.ChainID) (*registry.ChainRecord, error) {
	chainInfo, err := c.registryProvider().GetChainRecordByChainID(chainID)

	return chainInfo, err
}
