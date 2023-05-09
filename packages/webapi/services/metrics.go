package services

import (
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/webapi/dto"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/models"
)

type MetricsService struct {
	chainProvider        chains.Provider
	chainMetricsProvider *metrics.ChainMetricsProvider
}

func NewMetricsService(chainProvider chains.Provider, chainMetricsProvider *metrics.ChainMetricsProvider) interfaces.MetricsService {
	return &MetricsService{
		chainProvider:        chainProvider,
		chainMetricsProvider: chainMetricsProvider,
	}
}

func (c *MetricsService) GetNodeMessageMetrics() *dto.NodeMessageMetrics {
	return &dto.NodeMessageMetrics{
		RegisteredChainIDs: c.chainMetricsProvider.RegisteredChains(),

		InMilestone:        dto.MapMetricItem(c.chainMetricsProvider.InMilestone()),
		InStateOutput:      dto.MapMetricItem(c.chainMetricsProvider.InStateOutput()),
		InAliasOutput:      dto.MapMetricItem(c.chainMetricsProvider.InAliasOutput()),
		InOutput:           dto.MapMetricItem(c.chainMetricsProvider.InOutput()),
		InOnLedgerRequest:  dto.MapMetricItem(c.chainMetricsProvider.InOnLedgerRequest()),
		InTxInclusionState: dto.MapMetricItem(c.chainMetricsProvider.InTxInclusionState()),

		OutPublishStateTransaction:      dto.MapMetricItem(c.chainMetricsProvider.OutPublishStateTransaction()),
		OutPublishGovernanceTransaction: dto.MapMetricItem(c.chainMetricsProvider.OutPublishGovernanceTransaction()),
		OutPullLatestOutput:             dto.MapMetricItem(c.chainMetricsProvider.OutPullLatestOutput()),
		OutPullTxInclusionState:         dto.MapMetricItem(c.chainMetricsProvider.OutPullTxInclusionState()),
		OutPullOutputByID:               dto.MapMetricItem(c.chainMetricsProvider.OutPullOutputByID()),
	}
}

func (c *MetricsService) GetChainMessageMetrics(chainID isc.ChainID) *dto.ChainMessageMetrics {
	chain, err := c.chainProvider().Get(chainID)
	if err != nil {
		return nil
	}

	chainMetrics := chain.GetChainMetrics()

	return &dto.ChainMessageMetrics{
		InStateOutput:      dto.MapMetricItem(chainMetrics.InStateOutput()),
		InAliasOutput:      dto.MapMetricItem(chainMetrics.InAliasOutput()),
		InOutput:           dto.MapMetricItem(chainMetrics.InOutput()),
		InOnLedgerRequest:  dto.MapMetricItem(chainMetrics.InOnLedgerRequest()),
		InTxInclusionState: dto.MapMetricItem(chainMetrics.InTxInclusionState()),

		OutPublishStateTransaction:      dto.MapMetricItem(chainMetrics.OutPublishStateTransaction()),
		OutPublishGovernanceTransaction: dto.MapMetricItem(chainMetrics.OutPublishGovernanceTransaction()),
		OutPullLatestOutput:             dto.MapMetricItem(chainMetrics.OutPullLatestOutput()),
		OutPullTxInclusionState:         dto.MapMetricItem(chainMetrics.OutPullTxInclusionState()),
		OutPullOutputByID:               dto.MapMetricItem(chainMetrics.OutPullOutputByID()),
	}
}

func (c *MetricsService) GetChainConsensusWorkflowMetrics(chainID isc.ChainID) *models.ConsensusWorkflowMetrics {
	chain, err := c.chainProvider().Get(chainID)
	if err != nil {
		return nil
	}

	metrics := chain.GetConsensusWorkflowStatus()
	if metrics == nil {
		return nil
	}

	return models.MapConsensusWorkflowStatus(metrics)
}

func (c *MetricsService) GetChainConsensusPipeMetrics(chainID isc.ChainID) *models.ConsensusPipeMetrics {
	chain, err := c.chainProvider().Get(chainID)
	if err != nil {
		return nil
	}

	metrics := chain.GetConsensusPipeMetrics()
	if metrics == nil {
		return nil
	}

	return models.MapConsensusPipeMetrics(metrics)
}

func (c *MetricsService) GetMaxChainConfirmedStateLag() uint32 {
	return c.chainMetricsProvider.MaxChainConfirmedStateLag()
}
