package metrics

import (
	"github.com/iotaledger/hive.go/core/app"
)

type ParametersMetrics struct {
	Enabled     bool   `default:"true" usage:"whether the metrics plugin is enabled"`
	BindAddress string `default:"127.0.0.1:2112" usage:"the bind address for the node dashboard"`
}

var ParamsMetrics = &ParametersMetrics{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"metrics": ParamsMetrics,
	},
	Masked: nil,
}
