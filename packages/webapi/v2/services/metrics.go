package services

import (
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/dto"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type MetricsService struct {
	log *logger.Logger

	chainProvider chains.Provider
}

func NewMetricsService(log *logger.Logger, chainProvider chains.Provider) interfaces.MetricsService {
	return &MetricsService{
		log: log,

		chainProvider: chainProvider,
	}
}

func (c *MetricsService) GetAllChainsMetrics() *dto.ChainMetricsReport {
	chain := c.chainProvider()

	if chain == nil {
		return nil
	}

	metrics := chain.GetNodeConnectionMetrics()

	return &dto.ChainMetricsReport{
		InAliasOutput:                   dto.MapMetricItem(metrics.GetInAliasOutput()),
		InOnLedgerRequest:               dto.MapMetricItem(metrics.GetInOnLedgerRequest()),
		InOutput:                        dto.MapMetricItem(metrics.GetInOutput()),
		InStateOutput:                   dto.MapMetricItem(metrics.GetInStateOutput()),
		InTxInclusionState:              dto.MapMetricItem(metrics.GetInTxInclusionState()),
		OutPublishGovernanceTransaction: dto.MapMetricItem(metrics.GetOutPublishGovernanceTransaction()),
		OutPullLatestOutput:             dto.MapMetricItem(metrics.GetOutPullLatestOutput()),
		OutPullOutputByID:               dto.MapMetricItem(metrics.GetOutPullOutputByID()),
		OutPullTxInclusionState:         dto.MapMetricItem(metrics.GetOutPullTxInclusionState()),
	}
}

func (c *MetricsService) GetChainMetrics(chainID *isc.ChainID) *dto.ChainMetricsReport {
	chain := c.chainProvider().Get(chainID)

	if chain == nil {
		return nil
	}

	metrics := chain.GetNodeConnectionMetrics()

	return &dto.ChainMetricsReport{
		InAliasOutput:                   dto.MapMetricItem(metrics.GetInAliasOutput()),
		InOnLedgerRequest:               dto.MapMetricItem(metrics.GetInOnLedgerRequest()),
		InOutput:                        dto.MapMetricItem(metrics.GetInOutput()),
		InStateOutput:                   dto.MapMetricItem(metrics.GetInStateOutput()),
		InTxInclusionState:              dto.MapMetricItem(metrics.GetInTxInclusionState()),
		OutPublishGovernanceTransaction: dto.MapMetricItem(metrics.GetOutPublishGovernanceTransaction()),
		OutPullLatestOutput:             dto.MapMetricItem(metrics.GetOutPullLatestOutput()),
		OutPullOutputByID:               dto.MapMetricItem(metrics.GetOutPullOutputByID()),
		OutPullTxInclusionState:         dto.MapMetricItem(metrics.GetOutPullTxInclusionState()),
	}
}

func (c *MetricsService) GetChainConsensusWorkflowMetrics(chainID *isc.ChainID) *dto.ConsensusWorkflowMetrics {
	chain := c.chainProvider().Get(chainID)

	if chain == nil {
		return nil
	}

	metrics := chain.GetConsensusWorkflowStatus()

	if metrics == nil {
		return nil
	}

	return dto.NewConsensusWorkflowStatus(metrics)
}

func (c *MetricsService) GetChainConsensusPipeMetrics(chainID *isc.ChainID) *dto.ConsensusPipeMetrics {
	chain := c.chainProvider().Get(chainID)

	if chain == nil {
		return nil
	}

	metrics := chain.GetConsensusPipeMetrics()

	if metrics == nil {
		return nil
	}

	return dto.NewConsensusPipeMetrics(metrics)
}
