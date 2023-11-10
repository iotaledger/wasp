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

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
)

// Used to check how long an IP has to be inactive to be removed from the limiting map
// It is also used to block an IP if it has reached the rate limit.
const clientWaitTime = 1 * time.Minute

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

func (w *websocketContext) runCleanupTimer(ctx context.Context) {
	go func() {
		t := time.NewTicker(clientWaitTime)
		defer t.Stop()

		w.log.Infof("[EVM WS] Cleanup process started")

		for {
			select {
			case <-t.C:
				w.cleanupRateLimiters()
			case <-ctx.Done():
				w.log.Infof("[EVM WS] Cleanup process stopped")
				return
			}
		}
	}()
}

func newWebsocketContext(log *logger.Logger, jsonrpcParameters *jsonrpc.Parameters) *websocketContext {
	return &websocketContext{
		syncPool:         new(sync.Pool),
		jsonRPCParams:    jsonrpcParameters,
		rateLimiterMutex: sync.Mutex{},
		rateLimiters:     map[string]*activityRateLimiter{},
		log:              log,
	}
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
			w.log.Debugf("[EVM WS] Removing rate limiter for ip:[%v], lastActivity:[%v], blocked:[%v]\n", ip, rateLimiter.LastActivity().Format(time.RFC822), rateLimiter.Canceled())
			delete(w.rateLimiters, ip)
		} else {
			w.log.Debugf("[EVM WS] Keeping rate limiter for ip:[%v], lastActivity:[%v], blocked:[%v]\n", ip, rateLimiter.LastActivity().Format(time.RFC822), rateLimiter.Canceled())
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
	if rlc.limiter == nil {
		return 0, rlc.Conn.Close()
	}

	if rlc.limiter.Canceled() {
		rlc.logger.Warnf("[EVM WS Conn/Read] connections for ip:[%v] canceled, lastActivity:[%v]", rlc.realIP, rlc.limiter.LastActivity())
		return 0, rlc.Conn.Close()
	}

	if !rlc.limiter.Allow() {
		rlc.logger.Warnf("[EVM WS Conn/Read] rate limit exceeded for ip:[%v], lastActivity:[%v]", rlc.realIP, rlc.limiter.LastActivity())
		return 0, rlc.Conn.Close()
	}

	n, err := rlc.Conn.Read(b)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (rlc *rateLimitedConn) Write(b []byte) (int, error) {
	if rlc.limiter == nil {
		return 0, rlc.Conn.Close()
	}

	if rlc.limiter.Canceled() {
		rlc.logger.Warnf("[EVM WS Conn/Write] connections for ip:[%v] canceled", rlc.realIP)
		return 0, rlc.Conn.Close()
	}

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

/*
Hijack overrides the original echo.Response:Hijack method and returns a wrapped net.Conn that counts messages to enable the rate limit.

We want to rate limit a Websocket connection (websocket.Conn). `go-ethereum`s Websocket Handler does not support custom websocket.Conn implementations.
This makes it impossible to hook between Read/WriteJSON to count the messages going in and out.
An alternative path is to go a level deeper and to hook into a net.Conn instead.
The websocket Upgrader is a middleware that requires the echo.Response to be passed. This Response allows the takeover of the established http socket by calling `echoResponse.Hijack`
Once the websocket.Upgrader is called, it will call Hijack internally eventually to get access to the underlying socket.

The rateLimitedEchoResponse wraps the echo.Response with its Hijack method.
Once it's called, it will call the original echoResponse.Hijack method, get the socket, wrap the socket and return the wrapped socket to the upgrader.
This allows us to hook directly into the Stream and to count the Read/Write per connection.
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
		rateLimiter := wsContext.getRateLimiter(realIP)
		logger.Warnf("Using rate limiter for:[%v], lastActivity:[%v], blocked:[%v]", realIP, rateLimiter.LastActivity(), rateLimiter.Canceled())
		if rateLimiter.Canceled() {
			logger.Warnf("[EVM WS Conn] Connection from ip:[%v] dropped (previous rate limit exceeded)\n", realIP)
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
