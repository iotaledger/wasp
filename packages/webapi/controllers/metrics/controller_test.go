package metrics_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi"
	apimetrics "github.com/iotaledger/wasp/packages/webapi/controllers/metrics"
	"github.com/iotaledger/wasp/packages/webapi/services"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMetricsNodeHealth(t *testing.T) {
	t.Skip("find a way to initialize services")
	version := "testVersion"
	logger := zap.NewExample().Sugar().Named("WebAPI/v0")
	chainsProvider := func() *chains.Chains {
		return &chains.Chains{}
	}
	chainMetricsProvider := &metrics.ChainMetricsProvider{}
	chainRecordRegistryProvider := &registry.ChainRecordRegistryImpl{}
	vmService := services.NewVMService(chainsProvider, chainRecordRegistryProvider)
	chainService := services.NewChainService(logger, chainsProvider, chainMetricsProvider, chainRecordRegistryProvider, vmService)
	metricsService := services.NewMetricsService(chainsProvider, chainMetricsProvider)
	c := apimetrics.NewMetricsController(chainService, metricsService)
	e := echo.New()
	server := echoswagger.New(e, "/doc", &echoswagger.Info{
		Title:       "Test Wasp API",
		Description: "Test REST API for the Wasp node",
		Version:     version,
	})
	group := server.Group(c.Name(), fmt.Sprintf("/v%d/", 0))
	mocker := webapi.NewMocker()
	c.RegisterPublic(group, mocker)

	req := httptest.NewRequest(http.MethodGet, "/v0/metrics/node/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
