package services

import (
	"bufio"
	"context"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
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

	// We could either limit per message, or per bytes sent (using `n` instead of 1)
	if err := rlc.limiter.WaitN(context.Background(), 1); err != nil {
		return 0, err
	}

	return n, nil
}

func (rlc *RateLimitedConn) Write(b []byte) (int, error) {
	// We could either limit per message, or per bytes sent (using `len(b)` instead of 1)
	if err := rlc.limiter.WaitN(context.Background(), 1); err != nil {
		return 0, err
	}

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

var wsBufferPool = new(sync.Pool)

func WebsocketHandler(server *chainServer) http.Handler {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		WriteBufferPool: wsBufferPool,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	limiter := rate.NewLimiter(rate.Limit(1), 50) // 10 requests per second with a burst of 5
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		echoResponse, ok := w.(*echo.Response)
		if !ok {
			log.Print("Could not cast response to echo.Response")

			return
		}

		rateLimitedResponseWriter := NewRateLimitedResponseWriter(echoResponse, limiter)
		conn, err := upgrader.Upgrade(rateLimitedResponseWriter, r, nil)

		if err != nil {
			log.Print(err)
			return
		}

		// Replace the WebSocket connection in the context
		codec := rpc.NewWebSocketCodec(conn, r.Host, r.Header)
		server.rpc.ServeCodec(codec, 0)
	})
}
