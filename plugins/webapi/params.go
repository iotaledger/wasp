package webapi

import (
	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/packages/authentication"
)

type ParametersWebAPI struct {
	Enabled            bool                             `default:"true" usage:"whether the web api plugin is enabled"`
	NodeOwnerAddresses []string                         `default:"" usage:"defines a list of node owner addresses (bech32)"`
	BindAddress        string                           `default:"127.0.0.1:9090" usage:"the bind address for the node web api"`
	Auth               authentication.AuthConfiguration `usage:"configures the authentication for the dashboard service"`
}

var ParamsWebAPI = &ParametersWebAPI{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"webapi": ParamsWebAPI,
	},
	Masked: nil,
}
