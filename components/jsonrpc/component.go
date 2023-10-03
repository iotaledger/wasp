package jsonrpc

import (
	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
)

var Component = &app.Component{
	Name: "JSONRPC",
	Params: &app.ComponentParams{
		Params: map[string]any{
			"jsonrpc": jsonrpc.Params,
		},
	},
}
