package prometheus

import (
	"github.com/iotaledger/hive.go/core/app"
)

// ParametersPrometheus contains the definition of the parameters used by Prometheus.
type ParametersPrometheus struct {
	// Enabled defines whether the prometheus plugin is enabled.
	Enabled bool `default:"true" usage:"whether the prometheus plugin is enabled"`
	// defines the bind address on which the Prometheus exporter listens on.
	BindAddress string `default:"127.0.0.1:2112" usage:"the bind address on which the Prometheus exporter listens on"`

	// NodeMetrics defines whether to include node metrics.
	NodeMetrics bool `default:"true" usage:"whether to include node metrics"`
	// NodeConnMetrics defines whether to include node connection metrics.
	NodeConnMetrics bool `default:"true" usage:"whether to include node connection metrics"`
	// BlockWALMetrics defines whether to include block Write-Ahead Log (WAL) metrics.
	BlockWALMetrics bool `default:"true" usage:"whether to include block Write-Ahead Log (WAL) metrics"`
	// RestAPIMetrics include restAPI metrics.
	RestAPIMetrics bool `default:"true" usage:"whether to include restAPI metrics"`
	// GoMetrics defines whether to include go metrics.
	GoMetrics bool `default:"true" usage:"whether to include go metrics"`
	// ProcessMetrics defines whether to include process metrics.
	ProcessMetrics bool `default:"true" usage:"whether to include process metrics"`
	// PromhttpMetrics defines whether to include promhttp metrics.
	PromhttpMetrics bool `default:"true" usage:"whether to include promhttp metrics"`
}

var ParamsPrometheus = &ParametersPrometheus{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"prometheus": ParamsPrometheus,
	},
	Masked: nil,
}
