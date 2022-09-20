package services

import (
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/wal"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type RegistryService struct {
	logger *logger.Logger

	chainsProvider   chains.Provider
	metrics          *metrics.Metrics
	registryProvider registry.Provider
	wal              *wal.WAL
}

func NewRegistryService(logger *logger.Logger, chainsProvider chains.Provider, metrics *metrics.Metrics, registryProvider registry.Provider, wal *wal.WAL) interfaces.Registry {
	return &RegistryService{
		logger: logger,

		chainsProvider:   chainsProvider,
		metrics:          metrics,
		registryProvider: registryProvider,
		wal:              wal,
	}
}

func (c *RegistryService) GetChainRecordByChainID(chainID *isc.ChainID) (*registry.ChainRecord, error) {
	chainInfo, err := c.registryProvider().GetChainRecordByChainID(chainID)

	return chainInfo, err
}
