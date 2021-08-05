package metrics

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/prometheus/client_golang/prometheus"
)

type MempoolMetrics interface {
	CountOffLedgerRequestIn()
	CountOnLedgerRequestIn()
	CountRequestOut()
}

func (c *chainMetrics) CountOffLedgerRequestIn() {
	c.metrics.offLedgerRequestCounter.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetrics) CountOnLedgerRequestIn() {
	c.metrics.onLedgerRequestCounter.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetrics) CountRequestOut() {
	c.metrics.processedRequestCounter.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (m *Metrics) registerMempoolMetrics() {
	m.log.Info("Registering mempool metrics to prometheus")
	m.offLedgerRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_off_ledger_request_counter",
		Help: "Number of off-ledger requests made to chains",
	}, []string{"chain"})
	prometheus.MustRegister(m.offLedgerRequestCounter)

	m.onLedgerRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_on_ledger_request_counter",
		Help: "Number of on-ledger requests made to chain",
	}, []string{"chain"})
	prometheus.MustRegister(m.onLedgerRequestCounter)

	m.processedRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_processed_on_ledger_request_counter",
		Help: "Number of requests processed on ledger",
	}, []string{"chain"})
	prometheus.MustRegister(m.processedRequestCounter)
}

// default mempool metrics with no counters
type defaultMempoolMetrics struct {
	log *logger.Logger
}

func DefaultMempoolMetrics(log *logger.Logger) MempoolMetrics {
	return &defaultMempoolMetrics{log}
}

func (m *defaultMempoolMetrics) CountOffLedgerRequestIn() {
	m.log.Info("No counter configured for off ledger requests.")
}

func (m *defaultMempoolMetrics) CountOnLedgerRequestIn() {
	m.log.Info("No counter configured for on ledger requests.")
}

func (m *defaultMempoolMetrics) CountRequestOut() {
	m.log.Info("No counter configured for processed requests.")
}
