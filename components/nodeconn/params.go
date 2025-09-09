package nodeconn

import (
	"github.com/iotaledger/hive.go/app"
)

type ParametersNodeCon struct {
	WebsocketURL          string `default:"ws://localhost:9000" usage:"the WS address to which to connect to"`
	HTTPURL               string `default:"http://localhost:9000" usage:"the HTTP address to which to connect to"`
	MaxConnectionAttempts uint   `default:"30" usage:"the amount of times the connection to INX will be attempted before it fails (1 attempt per second)"`
	TargetNetworkName     string `default:"" usage:"the network name on which the node should operate on (optional)"`
}

var ParamsL1 = &ParametersNodeCon{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"l1": ParamsL1,
	},
	Masked: nil,
}
