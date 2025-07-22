// Package metrics provides methods for getting chain metrics
package metrics

import (
	"net/http"

	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/v2/packages/authentication"
	"github.com/iotaledger/wasp/v2/packages/authentication/shared/permissions"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
)

type Controller struct {
	chainService   interfaces.ChainService
	metricsService interfaces.MetricsService
}

func NewMetricsController(chainService interfaces.ChainService, metricsService interfaces.MetricsService) interfaces.APIController {
	return &Controller{
		chainService:   chainService,
		metricsService: metricsService,
	}
}

func (c *Controller) Name() string {
	return "metrics"
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET("metrics/chain/messages", c.getChainMessageMetrics, authentication.ValidatePermissions([]string{permissions.Read})).
		AddResponse(http.StatusNotFound, "Chain not found", nil, nil).
		AddResponse(http.StatusOK, "A list of all available metrics.", models.ChainMessageMetrics{}, nil).
		SetOperationId("getChainMessageMetrics").
		SetSummary("Get chain specific message metrics.")

	adminAPI.GET("metrics/chain/workflow", c.getChainWorkflowMetrics, authentication.ValidatePermissions([]string{permissions.Read})).
		AddResponse(http.StatusNotFound, "Chain not found", nil, nil).
		AddResponse(http.StatusOK, "A list of all available metrics.", mocker.Get(models.ConsensusWorkflowMetrics{}), nil).
		SetOperationId("getChainWorkflowMetrics").
		SetSummary("Get chain workflow metrics.")

	adminAPI.GET("metrics/chain/pipe", c.getChainPipeMetrics, authentication.ValidatePermissions([]string{permissions.Read})).
		AddResponse(http.StatusNotFound, "Chain not found", nil, nil).
		AddResponse(http.StatusOK, "A list of all available metrics.", mocker.Get(models.ConsensusPipeMetrics{}), nil).
		SetOperationId("getChainPipeMetrics").
		SetSummary("Get chain pipe event metrics.")
}
