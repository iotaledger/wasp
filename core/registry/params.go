package registry

import (
	"github.com/iotaledger/hive.go/core/app"
)

type ParametersRegistry struct {
	UseText  bool   `default:"false" usage:"enable text key/value store for registry db."`
	FileName string `default:"chain-registry.json" usage:"registry filename. Ignored if registry.useText is false."`
}

var ParamsRegistry = &ParametersRegistry{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"registry": ParamsRegistry,
	},
	Masked: nil,
}
