package publishernano

import (
	"github.com/iotaledger/hive.go/core/app"
)

type ParametersPublisher struct {
	Enabled bool `default:"true" usage:"whether the publisher plugin is enabled"`
	Port    int  `default:"5550" usage:"the port for the nanomsg event publisher"`
}

var ParamsPublisher = &ParametersPublisher{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"nanomsg": ParamsPublisher,
	},
	Masked: nil,
}
