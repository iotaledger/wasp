package services

import (
	"testing"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/stretchr/testify/require"
)

// TestWebsocketContext_RateLimitDisabled ensures no rate limiting occurs when
// WebsocketRateLimitEnabled is set to false.
func TestWebsocketContext_RateLimitDisabled(t *testing.T) {
	testLogger := log.NewLogger(log.WithName("TestWebsocketContext_RateLimitDisabled"))
	params := jsonrpc.ParametersDefault()
	params.WebsocketRateLimitEnabled = false

	wsCtx := newWebsocketContext(testLogger, params)

	remoteIP := "127.0.0.1"
	rateLimiter := wsCtx.getRateLimiter(remoteIP)

	// Because the rate limiter is disabled, every call to .Allow() should return true.
	for i := 0; i < 50; i++ {
		require.True(t, rateLimiter.Allow(), "disabled rate limiter should always allow")
	}
}
