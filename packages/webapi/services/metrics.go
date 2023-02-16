package services

import (
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/dto"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/models"
)

type MetricsService struct {
	chainProvider chains.Provider
}

func NewMetricsService(chainProvider chains.Provider) interfaces.MetricsService {
	return &MetricsService{
		chainProvider: chainProvider,
	}
}

func (c *MetricsService) GetAllChainsMetrics() *dto.ChainMetrics {
	chain := c.chainProvider()
	if chain == nil {
		return nil
	}

	nodeConnMetrics := chain.GetNodeConnectionMetrics()
	registered := nodeConnMetrics.GetRegistered()

	return &dto.ChainMetrics{
		InAliasOutput:      dto.MapMetricItem(nodeConnMetrics.GetInAliasOutput()),
		InOnLedgerRequest:  dto.MapMetricItem(nodeConnMetrics.GetInOnLedgerRequest()),
		InOutput:           dto.MapMetricItem(nodeConnMetrics.GetInOutput()),
		InStateOutput:      dto.MapMetricItem(nodeConnMetrics.GetInStateOutput()),
		InTxInclusionState: dto.MapMetricItem(nodeConnMetrics.GetInTxInclusionState()),
		InMilestone:        dto.MapMetricItem(nodeConnMetrics.GetInMilestone()),

		OutPublishGovernanceTransaction: dto.MapMetricItem(nodeConnMetrics.GetOutPublishGovernanceTransaction()),
		OutPullLatestOutput:             dto.MapMetricItem(nodeConnMetrics.GetOutPullLatestOutput()),
		OutPullOutputByID:               dto.MapMetricItem(nodeConnMetrics.GetOutPullOutputByID()),
		OutPullTxInclusionState:         dto.MapMetricItem(nodeConnMetrics.GetOutPullTxInclusionState()),
		OutPublisherStateTransaction:    dto.MapMetricItem(nodeConnMetrics.GetOutPublishStateTransaction()),

		RegisteredChainIDs: registered,
	}
}

func (c *MetricsService) GetChainMetrics(chainID isc.ChainID) *dto.ChainMetrics {
	chain := c.chainProvider().Get(chainID)
	if chain == nil {
		return nil
	}

	nodeConnMetrics := chain.GetNodeConnectionMetrics()
	registered := nodeConnMetrics.GetRegistered()

	return &dto.ChainMetrics{
		InAliasOutput:                   dto.MapMetricItem(nodeConnMetrics.GetInAliasOutput()),
		InOnLedgerRequest:               dto.MapMetricItem(nodeConnMetrics.GetInOnLedgerRequest()),
		InOutput:                        dto.MapMetricItem(nodeConnMetrics.GetInOutput()),
		InStateOutput:                   dto.MapMetricItem(nodeConnMetrics.GetInStateOutput()),
		InTxInclusionState:              dto.MapMetricItem(nodeConnMetrics.GetInTxInclusionState()),
		InMilestone:                     dto.MapMetricItem(nodeConnMetrics.GetInMilestone()),
		OutPublishGovernanceTransaction: dto.MapMetricItem(nodeConnMetrics.GetOutPublishGovernanceTransaction()),

		OutPullLatestOutput:          dto.MapMetricItem(nodeConnMetrics.GetOutPullLatestOutput()),
		OutPullOutputByID:            dto.MapMetricItem(nodeConnMetrics.GetOutPullOutputByID()),
		OutPullTxInclusionState:      dto.MapMetricItem(nodeConnMetrics.GetOutPullTxInclusionState()),
		OutPublisherStateTransaction: dto.MapMetricItem(nodeConnMetrics.GetOutPublishStateTransaction()),

		RegisteredChainIDs: registered,
	}
}

func (c *MetricsService) GetChainConsensusWorkflowMetrics(chainID isc.ChainID) *models.ConsensusWorkflowMetrics {
	chain := c.chainProvider().Get(chainID)
	if chain == nil {
		return nil
	}

	metrics := chain.GetConsensusWorkflowStatus()
	if metrics == nil {
		return nil
	}

	return models.MapConsensusWorkflowStatus(metrics)
}

func (c *MetricsService) GetChainConsensusPipeMetrics(chainID isc.ChainID) *models.ConsensusPipeMetrics {
	chain := c.chainProvider().Get(chainID)
	if chain == nil {
		return nil
	}

	metrics := chain.GetConsensusPipeMetrics()
	if metrics == nil {
		return nil
	}

	return models.MapConsensusPipeMetrics(metrics)
}
