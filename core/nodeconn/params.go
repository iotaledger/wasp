package nodeconn

import (
	"github.com/iotaledger/hive.go/core/app"
)

type ParametersINX struct {
	Address               string `default:"localhost:9029" usage:"the INX address to which to connect to"`
	MaxConnectionAttempts uint   `default:"30" usage:"the amount of times the connection to INX will be attempted before it fails (1 attempt per second)"`
}

var ParamsINX = &ParametersINX{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"inx": ParamsINX,
	},
	Masked: nil,
}
