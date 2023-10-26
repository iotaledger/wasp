package services

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"

	"github.com/iotaledger/hive.go/logger"
)

/**
`go-ethereum` uses Gorilla Websockets internally and is relying on Gorillas `websockets.Conn` struct.
As this is a struct, it is not easy to replace the underlying connection as it's very hard coupled.

This approach goes a layer deeper. The WebsocketHandler was copied from `go-ethereum` and was modified a bit to take over the upgrade behavior.
It wraps the echo.Response struct and overrides the `Hijack` function which is used by Gorillas ws upgrade.
This way we can inject our rate limiting logic inside.
The downside is that it's not possible to limit per JSON message, but it should work regardless.
*/

type RateLimitedConn struct {
	net.Conn
	limiter *rate.Limiter
}

func NewRateLimitedConn(conn net.Conn, r *rate.Limiter) *RateLimitedConn {
	return &RateLimitedConn{
		Conn:    conn,
		limiter: r,
	}
}

func (rlc *RateLimitedConn) Read(b []byte) (int, error) {
	n, err := rlc.Conn.Read(b)
	if err != nil {
		return n, err
	}

	if !rlc.limiter.Allow() {
		log.Print("rate limit exceeded")
		return 0, rlc.Conn.Close()
	}

	return n, nil
}

func (rlc *RateLimitedConn) Write(b []byte) (int, error) {
	return rlc.Conn.Write(b)
}

type RateLimitedResponseWriter struct {
	*echo.Response

	limiter *rate.Limiter
}

func NewRateLimitedResponseWriter(r *echo.Response, limiter *rate.Limiter) *RateLimitedResponseWriter {
	return &RateLimitedResponseWriter{r, limiter}
}

func (r RateLimitedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, buffer, err := r.Response.Hijack()
	if err != nil {
		return conn, buffer, err
	}

	buffer.Reader = bufio.NewReader(conn)
	buffer.Writer = bufio.NewWriter(conn)

	return NewRateLimitedConn(conn, r.limiter), buffer, err
}

const (
	readBufferSize  = 1024
	writeBufferSize = 1024
)

func websocketHandler(logger *logger.Logger, server *chainServer, wsContext *websocketContext) http.Handler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		WriteBufferPool: wsContext.syncPool,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !wsContext.rateLimiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		echoResponse, ok := w.(*echo.Response)
		if !ok {
			logger.Info("[EVM WS] Could not cast response to echo.Response")
			return
		}

		rateLimitedResponseWriter := NewRateLimitedResponseWriter(echoResponse, wsContext.rateLimiter)
		conn, err := upgrader.Upgrade(rateLimitedResponseWriter, r, nil)
		if err != nil {
			logger.Info(fmt.Sprintf("[EVM WS] %s", err))
			return
		}

		// Replace the WebSocket connection in the context
		codec := rpc.NewWebSocketCodec(conn, r.Host, r.Header)
		server.rpc.ServeCodec(codec, 0)
	})
}
