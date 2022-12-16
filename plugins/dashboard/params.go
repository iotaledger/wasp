package dashboard

import (
	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/packages/authentication"
)

type ParametersDashboard struct {
	Enabled                   bool                             `default:"true" usage:"whether the dashboard plugin is enabled"`
	BindAddress               string                           `default:"0.0.0.0:7000" usage:"the bind address for the node dashboard"`
	ExploreAddressURL         string                           `default:"" usage:"URL to add as href to addresses in the dashboard"`
	DebugRequestLoggerEnabled bool                             `default:"false" usage:"whether the debug logging for requests should be enabled"`
	Auth                      authentication.AuthConfiguration `usage:"configures the authentication for the dashboard service"`
}

var ParamsDashboard = &ParametersDashboard{
	Auth: authentication.AuthConfiguration{
		Scheme: "basic",
	},
}

var params = &app.ComponentParams{
	Params: map[string]any{
		"dashboard": ParamsDashboard,
	},
	Masked: nil,
}
