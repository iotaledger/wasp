package metrics

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/cpu"
)

var (
	cpuInfo    prometheus.Gauge
	memInfo    prometheus.Gauge
	goRoutines prometheus.Gauge
)

func registerProcessMetrics() {
	cpuInfo = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "process_cpu_usage",
		Help: "CPU (System) usage",
	})
	memInfo = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "process_mem_usage_bytes",
		Help: "memory usage [bytes]",
	})
	goRoutines = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_routines",
		Help: "Go routines",
	})
	prometheus.MustRegister(cpuInfo)
	prometheus.MustRegister(memInfo)
	prometheus.MustRegister(goRoutines)
}

func updateProcessMetrics() {
	// update cpu usage metrics
	percent, err := cpu.Percent(time.Second, false)
	if err == nil && len(percent) > 0 {
		cpuInfo.Set(percent[0])
	}

	// update memory usage info
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memInfo.Set(float64(m.Alloc))

	// number of go routines metrics
	goRoutines.Set(float64(runtime.NumGoroutine()))
}
