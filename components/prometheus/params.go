package prometheus

import (
	"github.com/iotaledger/hive.go/app"
)

// ParametersPrometheus contains the definition of the parameters used by Prometheus.
type ParametersPrometheus struct {
	Enabled                  bool   `default:"true" usage:"whether the prometheus plugin is enabled"`
	BindAddress              string `default:"0.0.0.0:2112" usage:"the bind address on which the Prometheus exporter listens on"`
	NodeMetrics              bool   `default:"true" usage:"whether to include node metrics"`
	BlockWALMetrics          bool   `default:"true" usage:"whether to include block Write-Ahead Log (WAL) metrics"`
	ConsensusMetrics         bool   `default:"true" usage:"whether to include consensus metrics"`
	MempoolMetrics           bool   `default:"true" usage:"whether to include mempool metrics"`
	ChainMessagesMetrics     bool   `default:"true" usage:"whether to include chain messages metrics"`
	ChainStateMetrics        bool   `default:"true" usage:"whether to include chain state metrics"`
	ChainStateManagerMetrics bool   `default:"true" usage:"whether to include chain state manager metrics"`
	ChainNodeConnMetrics     bool   `default:"true" usage:"whether to include chain node conn metrics"`
	RestAPIMetrics           bool   `default:"true" usage:"whether to include restAPI metrics"`
	GoMetrics                bool   `default:"true" usage:"whether to include go metrics"`
	ProcessMetrics           bool   `default:"true" usage:"whether to include process metrics"`
	PromhttpMetrics          bool   `default:"true" usage:"whether to include promhttp metrics"`
	WebAPIMetrics            bool   `default:"true" usage:"whether to include webapi metrics"`
}

var ParamsPrometheus = &ParametersPrometheus{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"prometheus": ParamsPrometheus,
	},
	Masked: nil,
}
