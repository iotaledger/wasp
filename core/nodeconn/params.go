package nodeconn

import (
	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/inx-app/core/inx"
)

var ParamsINX = &inx.ParametersINX{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"inx": ParamsINX,
	},
	Masked: nil,
}
