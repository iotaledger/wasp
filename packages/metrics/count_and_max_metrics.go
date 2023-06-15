package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type countAndMaxMetrics struct {
	count    *prometheus.CounterVec
	max      *prometheus.CounterVec
	maxValue float64
}

func newCountAndMaxMetrics(count, max *prometheus.CounterVec) *countAndMaxMetrics {
	return &countAndMaxMetrics{
		count:    count,
		max:      max,
		maxValue: -1,
	}
}

// init values so they appear in prometheus
func (camm *countAndMaxMetrics) with(labels prometheus.Labels) {
	camm.count.With(labels)
	camm.max.With(labels)
}

func (camm *countAndMaxMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{camm.count, camm.max}
}

func (camm *countAndMaxMetrics) countValue(labels prometheus.Labels, value float64) {
	camm.count.With(labels).Inc()
	if camm.maxValue < 0 {
		camm.max.With(labels).Add(value)
		camm.maxValue = value
	} else {
		diff := value - camm.maxValue
		if diff > 0 {
			camm.max.With(labels).Add(diff)
			camm.maxValue = value
		}
	}
}
