package metrics

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
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
	currentStateIndex       *prometheus.GaugeVec
	requestProcessingTime   *prometheus.GaugeVec
	vmRunTime               *prometheus.GaugeVec
	vmRunCounter            *prometheus.CounterVec
	blocksPerChain          *prometheus.CounterVec
	blockSizes              *prometheus.GaugeVec
	nodeconnMetrics         nodeconnmetrics.NodeConnectionMetrics
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
	return &Metrics{
		log:             log,
		nodeconnMetrics: nodeconnmetrics.New(log),
	}
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
	m.nodeconnMetrics.RegisterMetrics()
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

	m.currentStateIndex = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "current_state_index",
		Help: "The current chain state index.",
	}, []string{"chain"})
	prometheus.MustRegister(m.currentStateIndex)

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

	m.vmRunCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_vm_run_counter",
		Help: "Time it takes to run the vm",
	}, []string{"chain"})
	prometheus.MustRegister(m.vmRunCounter)

	m.blocksPerChain = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_block_counter",
		Help: "Number of blocks per chain",
	}, []string{"chain"})
	prometheus.MustRegister(m.blocksPerChain)

	m.blockSizes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "wasp_block_size",
		Help: "Block sizes",
	}, []string{"block_index", "chain"})
	prometheus.MustRegister(m.blockSizes)
}

func (m *Metrics) GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMetrics {
	if m == nil {
		return nodeconnmetrics.NewEmptyNodeConnectionMetrics()
	}
	return m.nodeconnMetrics
}
