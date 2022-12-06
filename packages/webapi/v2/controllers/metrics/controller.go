package metrics

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"

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
	adminAPI.GET("metrics/chain/:chainID", c.getChainMetrics, authentication.ValidatePermissions([]string{permissions.MetricsRead})).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all available metrics.", models.ChainMetrics{}, nil).
		SetOperationId("getChainMetrics").
		SetSummary("Get chain specific metrics.")

	adminAPI.GET("metrics/chain/:chainID/workflow", c.getChainWorkflowMetrics, authentication.ValidatePermissions([]string{permissions.MetricsRead})).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all available metrics.", mocker.Get(models.ConsensusWorkflowMetrics{}), nil).
		SetOperationId("getChainWorkflowMetrics").
		SetSummary("Get chain workflow metrics.")

	adminAPI.GET("metrics/chain/:chainID/pipe", c.getChainPipeMetrics, authentication.ValidatePermissions([]string{permissions.MetricsRead})).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all available metrics.", mocker.Get(models.ConsensusPipeMetrics{}), nil).
		SetOperationId("getChainPipeMetrics").
		SetSummary("Get chain pipe event metrics.")
}
