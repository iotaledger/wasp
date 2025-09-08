package services

import (
	"github.com/iotaledger/wasp/v2/packages/chainrunner"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/webapi/dto"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
)

type MetricsService struct {
	chainRunner          *chainrunner.ChainRunner
	chainMetricsProvider *metrics.ChainMetricsProvider
}

func NewMetricsService(chainRunner *chainrunner.ChainRunner, chainMetricsProvider *metrics.ChainMetricsProvider) interfaces.MetricsService {
	return &MetricsService{
		chainRunner:          chainRunner,
		chainMetricsProvider: chainMetricsProvider,
	}
}

func (c *MetricsService) GetNodeMessageMetrics() *dto.NodeMessageMetrics {
	return &dto.NodeMessageMetrics{
		RegisteredChainIDs:         c.chainMetricsProvider.RegisteredChains(),
		InAnchor:                   dto.MapMetricItem(c.chainMetricsProvider.Message.InAnchor()),
		InOnLedgerRequest:          dto.MapMetricItem(c.chainMetricsProvider.Message.InOnLedgerRequest()),
		OutPublishStateTransaction: dto.MapMetricItem(c.chainMetricsProvider.Message.OutPublishStateTransaction()),
	}
}

func (c *MetricsService) GetChainMessageMetrics() *dto.ChainMessageMetrics {
	chain, err := c.chainRunner.Chain()
	if err != nil {
		return nil
	}

	chainMetrics := chain.GetChainMetrics()

	return &dto.ChainMessageMetrics{
		InAnchor:                   dto.MapMetricItem(chainMetrics.Message.InAnchor()),
		InOnLedgerRequest:          dto.MapMetricItem(chainMetrics.Message.InOnLedgerRequest()),
		OutPublishStateTransaction: dto.MapMetricItem(chainMetrics.Message.OutPublishStateTransaction()),
	}
}

func (c *MetricsService) GetChainConsensusWorkflowMetrics() *models.ConsensusWorkflowMetrics {
	chain, err := c.chainRunner.Chain()
	if err != nil {
		return nil
	}

	metrics := chain.GetConsensusWorkflowStatus()
	if metrics == nil {
		return nil
	}

	return models.MapConsensusWorkflowStatus(metrics)
}

func (c *MetricsService) GetChainConsensusPipeMetrics() *models.ConsensusPipeMetrics {
	chain, err := c.chainRunner.Chain()
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
	return c.chainMetricsProvider.StateManager.MaxChainConfirmedStateLag()
}
