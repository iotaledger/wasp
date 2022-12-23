package peering

import (
	"github.com/iotaledger/hive.go/core/app"
)

type ParametersPeering struct {
	NetID string `default:"0.0.0.0:4000" usage:"node host address as it is recognized by other peers"`
	Port  int    `default:"4000" usage:"port for Wasp committee connection/peering"`
}

var ParamsPeering = &ParametersPeering{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"peering": ParamsPeering,
	},
	Masked: nil,
}
