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
	messagesReceived        *prometheus.CounterVec
	requestAckMessages      *prometheus.CounterVec
	requestProcessingTime   *prometheus.GaugeVec
	vmRunTime               *prometheus.GaugeVec
}

func (m *Metrics) NewChainMetrics(chainID *iscp.ChainID) ChainMetrics {
	if m == nil {
		return DefaultChainMetrics()
	}

	return &chainMetricsObj{
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
		m.log.Infof("Prometheus metrics accessible at: %s", addr)
		m.server = &http.Server{Addr: addr, Handler: e}
		m.registerMetrics()
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

func (m *Metrics) registerMetrics() {
	m.log.Info("Registering mempool metrics to prometheus")
	m.offLedgerRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_off_ledger_request_counter",
		Help: "Number of off-ledger requests made to chain",
	}, []string{"chain"})
	prometheus.MustRegister(m.offLedgerRequestCounter)

	m.onLedgerRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_on_ledger_request_counter",
		Help: "Number of on-ledger requests made to the chain",
	}, []string{"chain"})
	prometheus.MustRegister(m.onLedgerRequestCounter)

	m.processedRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_processed_request_counter",
		Help: "Number of requests processed",
	}, []string{"chain"})
	prometheus.MustRegister(m.processedRequestCounter)

	m.messagesReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "messages_received_per_chain",
		Help: "Number of messages received",
	}, []string{"chain"})
	prometheus.MustRegister(m.messagesReceived)

	m.requestAckMessages = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "receive_requests_acknowledgement_message",
		Help: "Receive request acknowledgement messages per chain",
	}, []string{"chain"})
	prometheus.MustRegister(m.requestAckMessages)

	m.requestProcessingTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "request_processing_time",
		Help: "Time to process request",
	}, []string{"chain", "request"})
	prometheus.MustRegister(m.requestProcessingTime)

	m.vmRunTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vm_run_time",
		Help: "Time it takes to run the vm",
	}, []string{"chain"})
	prometheus.MustRegister(m.vmRunTime)
}
