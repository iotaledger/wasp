// Package metrics provides functionality for collecting and exposing metrics about the system.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/peering"
)

type PeeringMetricsProvider struct {
	peerCount    prometheus.Gauge
	sendQueueLen prometheus.Gauge
	sendMsgSizes prometheus.Histogram
	recvQueueLen prometheus.Gauge
	recvMsgSizes prometheus.Histogram
}

var _ peering.Metrics = &PeeringMetricsProvider{}

func NewPeeringMetricsProvider() *PeeringMetricsProvider {
	msgCountBuckets := prometheus.ExponentialBucketsRange(1, 100_000, 17)
	return &PeeringMetricsProvider{
		peerCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "peering",
			Name:      "peer_count",
			Help:      "Number active of peers.",
		}),
		sendQueueLen: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "peering",
			Name:      "send_queue_len",
			Help:      "Size of the send queue.",
		}),
		sendMsgSizes: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "peering",
			Name:      "send_msg_sizes",
			Help:      "Sizes of the sent messages.",
			Buckets:   msgCountBuckets,
		}),
		recvQueueLen: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "peering",
			Name:      "recv_queue_len",
			Help:      "Size of the recv queue.",
		}),
		recvMsgSizes: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "peering",
			Name:      "recv_msg_sizes",
			Help:      "Sizes of the received messages.",
			Buckets:   msgCountBuckets,
		}),
	}
}

func (m *PeeringMetricsProvider) Register(reg prometheus.Registerer) {
	reg.MustRegister(
		m.peerCount,
		m.recvQueueLen,
		m.recvMsgSizes,
		m.sendQueueLen,
		m.sendMsgSizes,
	)
}

func (m *PeeringMetricsProvider) PeerCount(peerCount int) {
	m.peerCount.Set(float64(peerCount))
}

func (m *PeeringMetricsProvider) RecvEnqueued(messageSize, newPipeSize int) {
	m.recvQueueLen.Set(float64(newPipeSize))
	m.recvMsgSizes.Observe(float64(messageSize))
}

func (m *PeeringMetricsProvider) RecvDequeued(messageSize, newPipeSize int) {
	m.recvQueueLen.Set(float64(newPipeSize))
	m.recvMsgSizes.Observe(float64(messageSize))
}

func (m *PeeringMetricsProvider) SendEnqueued(messageSize, newPipeSize int) {
	m.sendQueueLen.Set(float64(newPipeSize))
	m.sendMsgSizes.Observe(float64(messageSize))
}

func (m *PeeringMetricsProvider) SendDequeued(messageSize, newPipeSize int) {
	m.sendQueueLen.Set(float64(newPipeSize))
	m.sendMsgSizes.Observe(float64(messageSize))
}
