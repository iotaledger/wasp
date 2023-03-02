package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/hive.go/app"
)

var appInfo *prometheus.GaugeVec

func configureNode(registry *prometheus.Registry, info *app.Info) {
	appInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "node",
			Name:      "app_info",
			Help:      "Node software name and version.",
		},
		[]string{"name", "version"},
	)

	appInfo.WithLabelValues(info.Name, info.Version).Set(1)
	registry.MustRegister(appInfo)
}
