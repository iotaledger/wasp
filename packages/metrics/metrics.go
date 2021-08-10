package metrics

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	server                  *http.Server
	log                     *logger.Logger
	offLedgerRequestCounter *prometheus.CounterVec
	onLedgerRequestCounter  *prometheus.CounterVec
	processedRequestCounter *prometheus.CounterVec
}

type chainMetrics struct {
	metrics *Metrics
	chainID *iscp.ChainID
}

func (m *Metrics) NewChainMetrics(chainID *iscp.ChainID) ChainMetrics {
	return &chainMetrics{
		metrics: m,
		chainID: chainID,
	}
}

func New(log *logger.Logger) *Metrics {
	return &Metrics{log: log}
}

var once sync.Once

func (m *Metrics) Start(addr string) {
	once.Do(func() {
		e := echo.New()
		e.HideBanner = true
		e.Use(middleware.Recover())
		e.GET("/metrics", func(c echo.Context) error {
			handler := promhttp.Handler()
			handler.ServeHTTP(c.Response(), c.Request())
			return nil
		})
		m.registerMempoolMetrics()
		m.log.Infof("Prometheus metrics accessible at: %s", addr)
		m.server = &http.Server{Addr: addr, Handler: e}
		if err := m.server.ListenAndServe(); err != nil {
			m.log.Error("Failed to start metrics server", err)
		}
	})
}

func (m *Metrics) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.server.Shutdown(ctx)
}
