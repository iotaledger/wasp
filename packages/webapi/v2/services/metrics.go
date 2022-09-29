package services

import (
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
)

type MetricsService struct {
	log *logger.Logger

	chainProvider chains.Provider
}

func NewMetricsService(log *logger.Logger, chainProvider chains.Provider) *MetricsService {
	return &MetricsService{
		log: log,

		chainProvider: chainProvider,
	}
}

func (c *MetricsService) GetConnectionMetrics(chainID *isc.ChainID) {
	chain := c.chainProvider().Get(chainID)

	if chain == nil {
		// err
		return
	}

	metrics := chain.GetNodeConnectionMetrics()
	metrics.GetInTxInclusionState().GetLastMessage()
}
