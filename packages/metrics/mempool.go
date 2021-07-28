package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type MempoolMetrics interface {
	NewOffLedgerRequest(chainID string)
	NewOnLedgerRequest(chainID string)
}

var (
	offLedgerRequestCounter *prometheus.CounterVec
	onLedgerRequestCounter  *prometheus.CounterVec
	// Remember to track process requests
)

func (m *Metrics) NewOffLedgerRequest(chainID string) {
	offLedgerRequestCounter.With(prometheus.Labels{"chain": chainID})
}

func (m *Metrics) NewOnLedgerRequest(chainID string) {
	onLedgerRequestCounter.With(prometheus.Labels{"chain": chainID})
}

func registerMempoolMetrics() {
	offLedgerRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_off_ledger_request_counter_per_chain",
		Help: "Number of ledger requests made to chains",
	}, []string{"chain"})
	prometheus.MustRegister(offLedgerRequestCounter)

	onLedgerRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_on_ledger_request_counter_per_chain",
		Help: "Number of on ledger requests made to chain",
	}, []string{"chain"})
	prometheus.MustRegister(onLedgerRequestCounter)
}
