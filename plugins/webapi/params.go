package webapi

import (
	"time"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/packages/authentication"
)

type ParametersWebAPI struct {
	Enabled            bool                             `default:"true" usage:"whether the web api plugin is enabled"`
	BindAddress        string                           `default:"0.0.0.0:9090" usage:"the bind address for the node web api"`
	NodeOwnerAddresses []string                         `default:"" usage:"defines a list of node owner addresses (bech32)"`
	Auth               authentication.AuthConfiguration `usage:"configures the authentication for the API service"`

	Limits struct {
		Timeout                        time.Duration `default:"30s" usage:"the timeout after which a long running operation will be canceled"`
		ReadTimeout                    time.Duration `default:"10s" usage:"the read timeout for the HTTP request body"`
		WriteTimeout                   time.Duration `default:"10s" usage:"the write timeout for the HTTP response body"`
		MaxBodyLength                  string        `default:"2M" usage:"the maximum number of characters that the body of an API call may contain"`
		MaxTopicSubscriptionsPerClient int           `default:"0" usage:"defines the max amount of subscriptions per client. 0 = deactivated (default)"`
	}

	DebugRequestLoggerEnabled bool `default:"false" usage:"whether the debug logging for requests should be enabled"`
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
