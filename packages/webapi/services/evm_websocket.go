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

type activityRateLimiter struct {
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
	a.lastActivity = time.Now()
	allowed := a.rateLimiter.Allow()
	fmt.Printf("Rate token deducted. Remaining tokens:[%v]\n", a.Tokens())
	return allowed
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
	rateLimiter      map[string]*activityRateLimiter

	syncPool      *sync.Pool
	jsonRPCParams *jsonrpc.Parameters
}

func (w *websocketContext) getRateLimiter(remoteIP string) *activityRateLimiter {
	w.rateLimiterMutex.Lock()
	defer w.rateLimiterMutex.Unlock()

	if w.rateLimiter[remoteIP] != nil {
		return w.rateLimiter[remoteIP]
	}

	limiter := rate.NewLimiter(rate.Limit(w.jsonRPCParams.WebsocketRateLimitMessagesPerMinute), 1)
	w.rateLimiter[remoteIP] = newActivityRateLimiter(limiter)

	return w.rateLimiter[remoteIP]
}

func (w *websocketContext) cleanupRateLimiters() {
	w.rateLimiterMutex.Lock()
	defer w.rateLimiterMutex.Unlock()

	for ip, rateLimiter := range w.rateLimiter {
		fmt.Printf("Found rate limiter for ip:[%v], lastActivity:[%v]\n", ip, rateLimiter.LastActivity().Format(time.RFC822))

		if time.Since(rateLimiter.LastActivity()) > 30*time.Minute {
			fmt.Printf("Removing rate limiter for ip:[%v], lastActivity:[%v]\n", ip, rateLimiter.LastActivity().Format(time.RFC822))
			delete(w.rateLimiter, ip)
		} else {
			fmt.Printf("Keeping rate limiter for ip:[%v], lastActivity:[%v]\n", ip, rateLimiter.LastActivity().Format(time.RFC822))
		}
	}
}

type rateLimitedConn struct {
	net.Conn
	logger  *logger.Logger
	limiter *activityRateLimiter
	realIP  string
}

func newRateLimitedConn(conn net.Conn, logger *logger.Logger, r *activityRateLimiter, realIP string) *rateLimitedConn {
	return &rateLimitedConn{
		Conn:    conn,
		logger:  logger,
		limiter: r,
		realIP:  realIP,
	}
}

func (rlc *rateLimitedConn) Read(b []byte) (int, error) {
	n, err := rlc.Conn.Read(b)
	if err != nil {
		return n, err
	}

	fmt.Printf("READ MSG: %v\n", string(b))

	if !rlc.limiter.Allow() {
		rlc.logger.Infof("[EVM WS Conn] rate limit exceeded for ip:[%v]", rlc.realIP)
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
	limiter *activityRateLimiter
	realIP  string
}

func newRateLimitedEchoResponse(r *echo.Response, logger *logger.Logger, limiter *activityRateLimiter, realIP string) *rateLimitedEchoResponse {
	return &rateLimitedEchoResponse{r, logger, limiter, realIP}
}

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
	readBufferSize  = 4096
	writeBufferSize = 4096
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
		wsContext.cleanupRateLimiters()

		rateLimiter := wsContext.getRateLimiter(realIP)
		if !rateLimiter.Allow() {
			logger.Info("[EVM WS Conn] Connection from ip:[%v] dropped (previous rate limit exceeded) current tokens:[%v]\n", realIP, rateLimiter.Tokens())
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		echoResponse, ok := w.(*echo.Response)
		if !ok {
			logger.Info("[EVM WS] Could not cast response to echo.Response")
			return
		}

		rateLimitedResponseWriter := newRateLimitedEchoResponse(echoResponse, logger, rateLimiter, realIP)
		conn, err := upgrader.Upgrade(rateLimitedResponseWriter, r, nil)
		if err != nil {
			logger.Info(fmt.Sprintf("[EVM WS] %s", err))
			return
		}

		codec := rpc.NewWebSocketCodec(conn, r.Host, r.Header)
		conn.SetPongHandler(func(appData string) error {
			_ = conn.SetReadDeadline(time.Time{})
			rateLimiter.UpdateLastActivity()
			return nil
		})

		server.rpc.ServeCodec(codec, 0)
	})
}
