package prometheus

import (
	"github.com/iotaledger/hive.go/app"
)

// ParametersPrometheus contains the definition of the parameters used by Prometheus.
type ParametersPrometheus struct {
	Enabled     bool   `default:"true" usage:"whether the prometheus plugin is enabled"`
	BindAddress string `default:"0.0.0.0:2112" usage:"the bind address on which the Prometheus exporter listens on"`
}

var ParamsPrometheus = &ParametersPrometheus{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"prometheus": ParamsPrometheus,
	},
	Masked: nil,
}
