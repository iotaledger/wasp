package services

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
)

/**
`go-ethereum` uses Gorilla Websockets internally and is relying on Gorillas `websockets.Conn` struct.
As this is a struct, it is not easy to replace the underlying connection as it's very hard coupled.

This approach goes a layer deeper. The WebsocketHandler was copied from `go-ethereum` and was modified a bit to take over the upgrade behavior.
It wraps the echo.Response struct and overrides the `Hijack` function which is used by Gorillas ws upgrade.
This way we can inject our rate limiting logic inside.
The downside is that it's not possible to limit per JSON message, but it should work regardless.
*/

type websocketContext struct {
	rateLimiterMutex sync.Mutex
	rateLimiter      map[string]*rate.Limiter
	syncPool         *sync.Pool
	jsonRPCParams    *jsonrpc.Parameters
}

func (w *websocketContext) getRateLimiter(remoteIP string) *rate.Limiter {
	w.rateLimiterMutex.Lock()
	defer w.rateLimiterMutex.Unlock()

	if w.rateLimiter[remoteIP] != nil {
		return w.rateLimiter[remoteIP]
	}

	w.rateLimiter[remoteIP] = rate.NewLimiter(rate.Every(time.Minute), w.jsonRPCParams.WebsocketRateLimitMessagesPerMinute)

	return w.rateLimiter[remoteIP]
}

//nolint:unused
func (w *websocketContext) deleteRateLimiter(remoteIP string) {
	w.rateLimiterMutex.Lock()
	defer w.rateLimiterMutex.Unlock()

	if w.rateLimiter[remoteIP] != nil {
		delete(w.rateLimiter, remoteIP)
	}
}

type rateLimitedConn struct {
	net.Conn
	logger  *logger.Logger
	limiter *rate.Limiter
}

func newRateLimitedConn(conn net.Conn, logger *logger.Logger, r *rate.Limiter) *rateLimitedConn {
	return &rateLimitedConn{
		Conn:    conn,
		logger:  logger,
		limiter: r,
	}
}

func (rlc *rateLimitedConn) Read(b []byte) (int, error) {
	n, err := rlc.Conn.Read(b)
	if err != nil {
		return n, err
	}

	if !rlc.limiter.Allow() {
		rlc.logger.Info("[EVM WS Conn] rate limit exceeded")
		return 0, rlc.Conn.Close()
	}

	return n, nil
}

func (rlc *rateLimitedConn) Write(b []byte) (int, error) {
	return rlc.Conn.Write(b)
}

type rateLimitedEchoResponse struct {
	*echo.Response
	logger  *logger.Logger
	limiter *rate.Limiter
}

func newRateLimitedEchoResponse(r *echo.Response, logger *logger.Logger, limiter *rate.Limiter) *rateLimitedEchoResponse {
	return &rateLimitedEchoResponse{r, logger, limiter}
}

func (r rateLimitedEchoResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, buffer, err := r.Response.Hijack()
	if err != nil {
		return conn, buffer, err
	}

	buffer.Reader = bufio.NewReader(conn)
	buffer.Writer = bufio.NewWriter(conn)

	return newRateLimitedConn(conn, r.logger, r.limiter), buffer, err
}

const (
	readBufferSize  = 1024
	writeBufferSize = 1024
)

func websocketHandler(logger *logger.Logger, server *chainServer, wsContext *websocketContext, realIP string) http.Handler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		WriteBufferPool: wsContext.syncPool,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rateLimiter := wsContext.getRateLimiter(realIP)
		if !rateLimiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		echoResponse, ok := w.(*echo.Response)
		if !ok {
			logger.Info("[EVM WS] Could not cast response to echo.Response")
			return
		}

		rateLimitedResponseWriter := newRateLimitedEchoResponse(echoResponse, logger, rateLimiter)
		conn, err := upgrader.Upgrade(rateLimitedResponseWriter, r, nil)
		if err != nil {
			logger.Info(fmt.Sprintf("[EVM WS] %s", err))
			return
		}

		codec := rpc.NewWebSocketCodec(conn, r.Host, r.Header)
		server.rpc.ServeCodec(codec, 0)
	})
}
