package metrics

import (
	"net/http"

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
		AddResponse(http.StatusOK, "A list of all available metrics.", nil, nil).
		SetOperationId("getChainMetrics").
		SetResponseContentType("application/json").
		SetSummary("Get all available chain contracts.")

	adminAPI.GET("metrics/chain/:chainID/workflow", c.getChainWorkflowMetrics).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all available metrics.", nil, nil).
		SetOperationId("getChainWorkflowMetrics").
		SetResponseContentType("application/json").
		SetSummary("Get all available chain contracts.")

	adminAPI.GET("metrics/chain/:chainID/pipe", c.getChainPipeMetrics).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all available metrics.", nil, nil).
		SetOperationId("getChainPipeMetrics").
		SetResponseContentType("application/json").
		SetSummary("Get all available chain contracts.")
}
