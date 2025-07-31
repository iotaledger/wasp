package services

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/evm/jsonrpc"
)

type activityRateLimiter struct {
	canceled     bool
	rateLimiter  *rate.Limiter
	lastActivity time.Time
}

func newActivityRateLimiter(rateLimiter *rate.Limiter) *activityRateLimiter {
	return &activityRateLimiter{
		rateLimiter:  rateLimiter,
		lastActivity: time.Now(),
	}
}

func (a *activityRateLimiter) Allow() bool {
	a.UpdateLastActivity()

	if a.Canceled() {
		return false
	}

	allowed := a.rateLimiter.Allow()
	a.canceled = !allowed

	return allowed
}

func (a *activityRateLimiter) Canceled() bool {
	return a.canceled
}

func (a *activityRateLimiter) LastActivity() time.Time {
	return a.lastActivity
}

func (a *activityRateLimiter) UpdateLastActivity() {
	a.lastActivity = time.Now()
}

func (a *activityRateLimiter) Tokens() float64 {
	return a.rateLimiter.TokensAt(time.Now())
}

type websocketContext struct {
	rateLimiterMutex sync.Mutex
	rateLimiters     map[string]*activityRateLimiter
	logger           log.Logger
	syncPool         *sync.Pool
	jsonRPCParams    *jsonrpc.Parameters
}

func (w *websocketContext) runCleanupTimer(ctx context.Context) {
	t := time.NewTicker(w.jsonRPCParams.WebsocketConnectionCleanupDuration)
	defer t.Stop()

	w.logger.LogInfof("[EVM WS] Cleanup process started")

	for {
		select {
		case <-t.C:
			w.cleanupRateLimiters()
		case <-ctx.Done():
			w.logger.LogInfof("[EVM WS] Cleanup process stopped")
			return
		}
	}
}

func newWebsocketContext(logger log.Logger, jsonrpcParameters *jsonrpc.Parameters) *websocketContext {
	return &websocketContext{
		syncPool:         new(sync.Pool),
		jsonRPCParams:    jsonrpcParameters,
		rateLimiterMutex: sync.Mutex{},
		rateLimiters:     map[string]*activityRateLimiter{},
		logger:           logger,
	}
}

func (w *websocketContext) getRateLimiter(remoteIP string) *activityRateLimiter {
	w.rateLimiterMutex.Lock()
	defer w.rateLimiterMutex.Unlock()

	if w.rateLimiters[remoteIP] != nil {
		return w.rateLimiters[remoteIP]
	}

	limit := rate.Limit(w.jsonRPCParams.WebsocketRateLimitMessagesPerSecond)
	burst := w.jsonRPCParams.WebsocketRateLimitBurst

	if !w.jsonRPCParams.WebsocketRateLimitEnabled {
		limit = rate.Inf
		burst = 0
	}

	limiter := rate.NewLimiter(limit, burst)

	w.rateLimiters[remoteIP] = newActivityRateLimiter(limiter)

	return w.rateLimiters[remoteIP]
}

func (w *websocketContext) cleanupRateLimiters() {
	w.rateLimiterMutex.Lock()
	defer w.rateLimiterMutex.Unlock()

	for ip, rateLimiter := range w.rateLimiters {
		if time.Since(rateLimiter.LastActivity()) > w.jsonRPCParams.WebsocketClientBlockDuration {
			w.logger.LogDebugf("[EVM WS] Removing rate limiter for ip:[%v], lastActivity:[%v], blocked:[%v]\n", ip, rateLimiter.LastActivity().Format(time.RFC822), rateLimiter.Canceled())
			delete(w.rateLimiters, ip)
		} else {
			w.logger.LogDebugf("[EVM WS] Keeping rate limiter for ip:[%v], lastActivity:[%v], blocked:[%v]\n", ip, rateLimiter.LastActivity().Format(time.RFC822), rateLimiter.Canceled())
		}
	}
}

type rateLimitedConn struct {
	net.Conn
	logger  log.Logger
	limiter *activityRateLimiter
	realIP  string
}

func newRateLimitedConn(conn net.Conn, logger log.Logger, r *activityRateLimiter, realIP string) *rateLimitedConn {
	return &rateLimitedConn{
		Conn:    conn,
		logger:  logger,
		limiter: r,
		realIP:  realIP,
	}
}

func (rlc *rateLimitedConn) Read(b []byte) (int, error) {
	if rlc.limiter == nil {
		return 0, rlc.Close()
	}

	if rlc.limiter.Canceled() {
		rlc.logger.LogWarnf("[EVM WS Conn/Read] connections for ip:[%v] canceled, lastActivity:[%v]", rlc.realIP, rlc.limiter.LastActivity())
		return 0, rlc.Close()
	}

	if !rlc.limiter.Allow() {
		rlc.logger.LogWarnf("[EVM WS Conn/Read] rate limit exceeded for ip:[%v], lastActivity:[%v]", rlc.realIP, rlc.limiter.LastActivity())
		return 0, rlc.Close()
	}

	numBytes, err := rlc.Conn.Read(b)

	if rlc.limiter.Canceled() {
		rlc.logger.LogWarnf("[EVM WS Conn/Read] connections for ip:[%v] canceled, lastActivity:[%v]", rlc.realIP, rlc.limiter.LastActivity())
		return 0, rlc.Close()
	}

	return numBytes, err
}

func (rlc *rateLimitedConn) Write(b []byte) (int, error) {
	if rlc.limiter == nil {
		return 0, rlc.Close()
	}

	if rlc.limiter.Canceled() {
		rlc.logger.LogWarnf("[EVM WS Conn/Write] connections for ip:[%v] canceled, lastActivity:[%v]", rlc.realIP, rlc.limiter.LastActivity())
		return 0, rlc.Close()
	}

	numBytes, err := rlc.Conn.Write(b)

	if rlc.limiter.Canceled() {
		rlc.logger.LogWarnf("[EVM WS Conn/Write] connections for ip:[%v] canceled, lastActivity:[%v]", rlc.realIP, rlc.limiter.LastActivity())
		return 0, rlc.Close()
	}

	return numBytes, err
}

type rateLimitedEchoResponse struct {
	*echo.Response
	logger  log.Logger
	limiter *activityRateLimiter
	realIP  string
}

func newRateLimitedEchoResponse(r *echo.Response, logger log.Logger, limiter *activityRateLimiter, realIP string) *rateLimitedEchoResponse {
	return &rateLimitedEchoResponse{r, logger, limiter, realIP}
}

/*
Hijack overrides the original echo.Response:Hijack method and returns a wrapped net.Conn that counts messages to enable the rate limit.

As `go-ethereum`s Websocket Handler does not support custom websocket.Conn implementations, we need to hook a layer deeper into the Net.Conn instead.
(This is not optimal as we can't count whole JSON messages - only buffers.)
We do this by passing a modified echo.Response into the websocket.Upgrader.
This websocket.Upgrader will eventually call `echo.Response:Hijack` which allows echo middlewares to take over the established tcp/http request socket.
Our custom echo.Response overrides the original Hijack method and will return a custom Net.Conn implementation allowing us to hook and count Read/Write calls.
*/
func (r rateLimitedEchoResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, buffer, err := r.Response.Hijack()
	if err != nil {
		return conn, buffer, err
	}

	buffer.Reader = bufio.NewReader(conn)
	buffer.Writer = bufio.NewWriter(conn)

	return newRateLimitedConn(conn, r.logger, r.limiter, r.realIP), buffer, err
}

const (
	readBufferSize     = 4096
	writeBufferSize    = 4096
	wsDefaultReadLimit = 32 * 1024 * 1024
)

func websocketHandler(server *chainServer, wsContext *websocketContext, realIP string) http.Handler {
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
		wsLogger := wsContext.logger

		wsLogger.LogWarnf("Using rate limiter for:[%v], lastActivity:[%v], blocked:[%v]", realIP, rateLimiter.LastActivity(), rateLimiter.Canceled())
		if rateLimiter.Canceled() {
			wsLogger.LogWarnf("[EVM WS Conn] Connection from ip:[%v] dropped (previous rate limit exceeded)\n", realIP)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		echoResponse, ok := w.(*echo.Response)
		if !ok {
			wsLogger.LogWarn("[EVM WS] Could not cast response to echo.Response")
			http.Error(w, "", http.StatusInternalServerError)

			return
		}

		rateLimitedResponseWriter := newRateLimitedEchoResponse(echoResponse, wsLogger, rateLimiter, realIP)
		conn, err := upgrader.Upgrade(rateLimitedResponseWriter, r, nil)
		if err != nil {
			wsLogger.LogInfo(fmt.Sprintf("[EVM WS] %s", err))
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		codec := rpc.NewWebSocketCodec(conn, r.Host, r.Header, wsDefaultReadLimit)
		conn.SetPongHandler(func(appData string) error {
			_ = conn.SetReadDeadline(time.Time{})
			rateLimiter.UpdateLastActivity()
			return nil
		})

		server.rpc.ServeCodec(codec, 0)
	})
}
