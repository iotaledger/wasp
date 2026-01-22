package webapi_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	slogzap "github.com/samber/slog-zap/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/components/webapi"
	"github.com/iotaledger/wasp/v2/packages/authentication"
)

func TestInternalServerErrors(t *testing.T) {
	// start a webserver with a test log
	logCore, logObserver := observer.New(zapcore.DebugLevel)
	zapLogger := zap.New(logCore)
	logger := slogzap.Option{Level: slog.LevelDebug, Logger: zapLogger}.NewZapHandler()

	e := webapi.NewEcho(&webapi.ParametersWebAPI{
		Enabled:     true,
		BindAddress: ":9999",
		Auth:        authentication.AuthConfiguration{},
		Limits: webapi.ParametersWebAPILimits{
			Timeout:                        time.Minute,
			ReadTimeout:                    time.Minute,
			WriteTimeout:                   time.Minute,
			MaxBodyLength:                  "1M",
			MaxTopicSubscriptionsPerClient: 0,
			ConfirmedStateLagThreshold:     2,
			Jsonrpc:                        webapi.ParametersJSONRPC{},
		},
		DebugRequestLoggerEnabled: true,
	},
		nil,
		log.NewLogger(log.WithHandler(logger)),
	)

	// Add an endpoint that just panics with "foobar" and start the server
	exceptionText := "foobar"
	e.GET("/test", func(c echo.Context) error { panic(exceptionText) })

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	e.Listener = l

	go func() {
		err := e.Start("")
		require.ErrorIs(t, err, http.ErrServerClosed)
	}()
	defer e.Shutdown(context.Background())

	// query the endpoint
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/test", e.Listener.Addr().String()), http.NoBody)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	res.Body.Close()

	// assert the exception is not present in the response (prevent leaking errors)
	require.Equal(t, http.StatusInternalServerError, res.StatusCode)
	require.NotContains(t, string(resBody), exceptionText)

	// assert the exception is logged
	logEntries := logObserver.All()
	require.Len(t, logEntries, 1)
	require.Contains(t, logEntries[0].Message, exceptionText)
}
