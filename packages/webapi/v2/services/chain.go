package services

import (
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/wal"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type ChainService struct {
	logger *logger.Logger

	chainsProvider chains.Provider
	metrics *metrics.Metrics
	registryProvider registry.Provider
	wal *wal.WAL
}

func NewChainService(logger *logger.Logger, chainsProvider chains.Provider, metrics *metrics.Metrics, registryProvider registry.Provider, wal *wal.WAL) interfaces.Chain {
	return &ChainService{
		logger:           logger,

		chainsProvider:   chainsProvider,
		metrics: metrics,
		registryProvider: registryProvider,
		wal:              wal,
	}
}

func (c *ChainService) ActivateChain(chainID *isc.ChainID) error {
	chainRecord, err := c.registryProvider().ActivateChainRecord(chainID)

	if err != nil {
		return err
	}

	c.logger.Debugw("calling Chains.Activate", "chainID", chainID.String())

	err = c.chainsProvider().
		Activate(chainRecord, c.registryProvider, c.metrics, c.wal)

	return err
}

func (c *ChainService) DeactivateChain(chainID *isc.ChainID) error {
	chainRecord, err := c.registryProvider().DeactivateChainRecord(chainID)

	if err != nil {
		return err
	}

	c.logger.Debugw("calling Chains.Activate", "chainID", chainID.String())

	err = c.chainsProvider().Deactivate(chainRecord)

	return err
}

func (c*ChainService) GetChainByID(chainID *isc.ChainID) chain.Chain {
	chain := c.chainsProvider().Get(chainID, true)

	return chain
}

