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

// Used to check how long an IP has to be inactive to be removed from the limiting map
// It is also used to block an IP if it has reached the rate limit.
const clientWaitTime = 5 * time.Minute

type activityRateLimiter struct {
	sync.Mutex
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
	a.Lock()
	defer a.Unlock()
	a.lastActivity = time.Now()
}

func (a *activityRateLimiter) Tokens() float64 {
	return a.rateLimiter.TokensAt(time.Now())
}

type websocketContext struct {
	rateLimiterMutex sync.Mutex
	rateLimiters     map[string]*activityRateLimiter
	log              *logger.Logger
	syncPool         *sync.Pool
	jsonRPCParams    *jsonrpc.Parameters
}

func (w *websocketContext) getRateLimiter(remoteIP string) *activityRateLimiter {
	w.rateLimiterMutex.Lock()
	defer w.rateLimiterMutex.Unlock()

	if w.rateLimiters[remoteIP] != nil {
		return w.rateLimiters[remoteIP]
	}

	limiter := rate.NewLimiter(rate.Limit(w.jsonRPCParams.WebsocketRateLimitMessagesPerSecond), w.jsonRPCParams.WebsocketRateLimitBurst)
	w.rateLimiters[remoteIP] = newActivityRateLimiter(limiter)

	return w.rateLimiters[remoteIP]
}

func (w *websocketContext) cleanupRateLimiters() {
	w.rateLimiterMutex.Lock()
	defer w.rateLimiterMutex.Unlock()

	for ip, rateLimiter := range w.rateLimiters {
		if time.Since(rateLimiter.LastActivity()) > clientWaitTime {
			w.log.Debugf("Removing rate limiter for ip:[%v], lastActivity:[%v]\n", ip, rateLimiter.LastActivity().Format(time.RFC822))
			delete(w.rateLimiters, ip)
		} else {
			w.log.Debugf("Keeping rate limiter for ip:[%v], lastActivity:[%v]\n", ip, rateLimiter.LastActivity().Format(time.RFC822))
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

	if rlc.limiter.Canceled() {
		rlc.logger.Infof("[EVM WS Conn] connections for ip:[%v] canceled", rlc.realIP)
		return 0, rlc.Conn.Close()
	}

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
		if rateLimiter.Canceled() {
			logger.Infof("[EVM WS Conn] Connection from ip:[%v] dropped (previous rate limit exceeded)\n", realIP)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		echoResponse, ok := w.(*echo.Response)
		if !ok {
			logger.Warn("[EVM WS] Could not cast response to echo.Response")
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
