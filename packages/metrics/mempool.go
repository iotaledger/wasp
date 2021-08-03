package metrics

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	offLedgerRequestCounter *prometheus.CounterVec
	onLedgerRequestCounter  *prometheus.CounterVec
	processedRequestCounter *prometheus.CounterVec
)

func (m *chainMetrics) NewOffLedgerRequest() {
	offLedgerRequestCounter.With(prometheus.Labels{"chain": m.chainID.String()}).Inc()
}

func (m *chainMetrics) NewOnLedgerRequest() {
	onLedgerRequestCounter.With(prometheus.Labels{"chain": m.chainID.String()}).Inc()
}

func (m *chainMetrics) ProcessRequest() {
	processedRequestCounter.With(prometheus.Labels{"chain": m.chainID.String()}).Inc()
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
