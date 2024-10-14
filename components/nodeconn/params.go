package nodeconn

import (
	"github.com/iotaledger/hive.go/app"
)

type ParametersNodeCon struct {
	WebsocketURL          string `default:"wss://api.iota-rebased-alphanet.iota.cafe/websocket" usage:"the WS address to which to connect to"`
	PackageID             string `default:"" usage:"the identifier of the isc move package"`
	MaxConnectionAttempts uint   `default:"30" usage:"the amount of times the connection to INX will be attempted before it fails (1 attempt per second)"`
	TargetNetworkName     string `default:"" usage:"the network name on which the node should operate on (optional)"`
}

var ParamsWS = &ParametersNodeCon{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"ws": ParamsWS,
	},
	Masked: nil,
}
