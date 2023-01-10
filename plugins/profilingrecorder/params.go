package profilingrecorder

import "github.com/iotaledger/hive.go/core/app"

// ParametersProfilingRecorder contains the definition of the parameters used by ProfilingRecorder.
type ParametersProfilingRecorder struct {
	// Enabled defines whether the plugin is enabled.
	Enabled bool `default:"false" usage:"whether the ProfilingRecorder plugin is enabled"`
}

var ParamsProfilingRecorder = &ParametersProfilingRecorder{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"profilingRecorder": ParamsProfilingRecorder,
	},
	Masked: nil,
}
