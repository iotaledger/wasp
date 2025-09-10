package metrics

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

type countAndMaxMetrics struct {
	count *prometheus.CounterVec
	max   *prometheus.CounterVec

	// Cannot use float64 directly, because it creates data race.
	// And unfortunately atomic does not support float64 directly.
	// We could use sync.Mutex, but atomic would avoid blocking, so less influence on execution.
	maxValue atomic.Value
}

func newCountAndMaxMetrics(count, maximum *prometheus.CounterVec) *countAndMaxMetrics {
	c := &countAndMaxMetrics{
		count: count,
		max:   maximum,
	}
	c.maxValue.Store(float64(-1))

	return c
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
	maxValue := camm.maxValue.Load().(float64) // There is a race, but we don't care
	if maxValue < 0 {
		camm.max.With(labels).Add(value)
		camm.maxValue.Store(value)
	} else {
		diff := value - maxValue
		if diff > 0 {
			camm.max.With(labels).Add(diff)
			camm.maxValue.Store(value)
		}
	}
}
