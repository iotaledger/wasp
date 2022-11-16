package metrics

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/dto"

	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type Controller struct {
	log *loggerpkg.Logger

	metricsService interfaces.MetricsService
}

func NewMetricsController(log *loggerpkg.Logger, metricsService interfaces.MetricsService) interfaces.APIController {
	return &Controller{
		log:            log,
		metricsService: metricsService,
	}
}

func (c *Controller) Name() string {
	return "metrics"
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET("metrics/chain/:chainID", c.getChainMetrics).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all available metrics.", mocker.Get(dto.ChainMetrics{}), nil).
		SetOperationId("getChainMetrics").
		SetSummary("Get chain specific metrics.")

	adminAPI.GET("metrics/chain/:chainID/workflow", c.getChainWorkflowMetrics).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all available metrics.", mocker.Get(dto.ConsensusWorkflowMetrics{}), nil).
		SetOperationId("getChainWorkflowMetrics").
		SetSummary("Get chain workflow metrics.")

	adminAPI.GET("metrics/chain/:chainID/pipe", c.getChainPipeMetrics).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all available metrics.", mocker.Get(dto.ConsensusPipeMetrics{}), nil).
		SetOperationId("getChainPipeMetrics").
		SetSummary("Get chain pipe event metrics.")
}
