package metrics

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/prometheus/client_golang/prometheus"
)

type MempoolMetrics interface {
	NewOffLedgerRequest(chainID string)
	NewOnLedgerRequest(chainID string)
	ProcessRequest(chainID string)
}

var (
	offLedgerRequestCounter *prometheus.CounterVec
	onLedgerRequestCounter  *prometheus.CounterVec
	processedRequestCounter *prometheus.CounterVec
)

func (m *Metrics) NewOffLedgerRequest(chainID string) {
	offLedgerRequestCounter.With(prometheus.Labels{"chain": chainID}).Inc()
}

func (m *Metrics) NewOnLedgerRequest(chainID string) {
	onLedgerRequestCounter.With(prometheus.Labels{"chain": chainID}).Inc()
}

func (m *Metrics) ProcessRequest(chainID string) {
	processedRequestCounter.With(prometheus.Labels{"chain": chainID}).Inc()
}

func registerMempoolMetrics(log *logger.Logger) {
	log.Info("Registering mempool metrics to prometheus")
	offLedgerRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_off_ledger_request_counter",
		Help: "Number of ledger requests made to chains",
	}, []string{"chain"})
	prometheus.MustRegister(offLedgerRequestCounter)

	onLedgerRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_on_ledger_request_counter",
		Help: "Number of on ledger requests made to chain",
	}, []string{"chain"})
	prometheus.MustRegister(onLedgerRequestCounter)

	processedRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_processed_on_ledger_request_counter",
		Help: "Number of requests processed on ledger",
	}, []string{"chain"})
	prometheus.MustRegister(processedRequestCounter)
}
