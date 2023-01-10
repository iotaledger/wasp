package webapi

import (
	"time"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/packages/authentication"
)

type ParametersWebAPI struct {
	Enabled                   bool                             `default:"true" usage:"whether the web api plugin is enabled"`
	NodeOwnerAddresses        []string                         `default:"" usage:"defines a list of node owner addresses (bech32)"`
	BindAddress               string                           `default:"0.0.0.0:9090" usage:"the bind address for the node web api"`
	DebugRequestLoggerEnabled bool                             `default:"false" usage:"whether the debug logging for requests should be enabled"`
	Auth                      authentication.AuthConfiguration `usage:"configures the authentication for the dashboard service"`
	ReadTimeout               time.Duration                    `default:"20s" usage:"ReadTimeout for go http.Server"`
	WriteTimeout              time.Duration                    `default:"30s" usage:"WriteTimeout for go http.Server"`
}

var ParamsWebAPI = &ParametersWebAPI{
	Auth: authentication.AuthConfiguration{
		Scheme: "jwt",
	},
}

var params = &app.ComponentParams{
	Params: map[string]any{
		"webapi": ParamsWebAPI,
	},
	Masked: nil,
}
