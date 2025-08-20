package webapi

import (
	"time"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/wasp/v2/packages/authentication"
)

type ParametersWebAPI struct {
	Enabled                   bool                             `default:"true" usage:"whether the web api plugin is enabled"`
	BindAddress               string                           `default:"0.0.0.0:9090" usage:"the bind address for the node web api"`
	Auth                      authentication.AuthConfiguration `usage:"configures the authentication for the API service"`
	IndexDBPath               string                           `default:"waspdb/chains/index" usage:"directory for storing indexes of historical data (only archive nodes will create/use them)"`
	AccountDumpsPath          string                           `default:"waspdb/account_dumps" usage:"directory where account dumps will be stored"`
	Limits                    ParametersWebAPILimits
	DebugRequestLoggerEnabled bool `default:"false" usage:"whether the debug logging for requests should be enabled"`
}

type ParametersWebAPILimits struct {
	Timeout                        time.Duration `default:"30s" usage:"the timeout after which a long running operation will be canceled"`
	ReadTimeout                    time.Duration `default:"10s" usage:"the read timeout for the HTTP request body"`
	WriteTimeout                   time.Duration `default:"60s" usage:"the write timeout for the HTTP response body"`
	MaxBodyLength                  string        `default:"2M" usage:"the maximum number of characters that the body of an API call may contain"`
	MaxTopicSubscriptionsPerClient int           `default:"0" usage:"defines the max amount of subscriptions per client. 0 = deactivated (default)"`
	ConfirmedStateLagThreshold     uint32        `default:"2" usage:"the threshold that define a chain is unsynchronized"`
	Jsonrpc                        ParametersJSONRPC
}

type ParametersJSONRPC struct {
	MaxBlocksInLogsFilterRange int `default:"1000" usage:"maximum amount of blocks in eth_getLogs filter range"`
	MaxLogsInResult            int `default:"10000" usage:"maximum amount of logs in eth_getLogs result"`

	WebsocketRateLimitMessagesPerSecond int           `default:"20" usage:"the websocket rate limit (messages per second)"`
	WebsocketRateLimitBurst             int           `default:"5" usage:"the websocket burst limit"`
	WebsocketRateLimitEnabled           bool          `default:"true" usage:"enable rate limiting on the websocket"`
	WebsocketConnectionCleanupDuration  time.Duration `default:"5m" usage:"defines in which interval stale connections will be cleaned up"`
	WebsocketClientBlockDuration        time.Duration `default:"5m" usage:"the duration a misbehaving client will be blocked"`
}

var ParamsWebAPI = &ParametersWebAPI{
	Auth: authentication.AuthConfiguration{
		Scheme: "jwt",
		JWTConfig: authentication.JWTAuthConfiguration{
			Duration: 24 * time.Hour,
		},
	},
}

var params = &app.ComponentParams{
	Params: map[string]any{
		"webapi": ParamsWebAPI,
	},
	Masked: nil,
}
