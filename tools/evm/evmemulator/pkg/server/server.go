package server

import (
	"bytes"
	"io"
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/iotaledger/wasp/v2/tools/evm/evmemulator/pkg/log"
)

func StartServer(jsonRPCServer *rpc.Server, l1addr string) *http.Server {
	mux := http.NewServeMux()

	// WebSocket handler
	mux.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		jsonRPCServer.WebsocketHandler([]string{"*"}).ServeHTTP(w, req)
	})

	// JSON-RPC handler
	mux.Handle("/", logBodyMiddleware(jsonRPCServer))

	return &http.Server{
		Addr:    l1addr,
		Handler: mux,
	}
}

// logBodyMiddleware logs request and response bodies
func logBodyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request body
		var reqBody []byte
		if r.Body != nil {
			reqBody, _ = io.ReadAll(r.Body)
			log.Printf("Incoming Req Body: %s %s\n%s\n", r.Method, r.URL.Path, string(reqBody))
			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// Capture response body
		resBody := new(bytes.Buffer)
		rec := &bodyWriter{ResponseWriter: w, body: resBody, statusCode: http.StatusOK}

		next.ServeHTTP(rec, r)

		// Log response body and status
		log.Printf("Outgoing Resp Body: %s %s [%d]\n%s\n", r.Method, r.URL.Path, rec.statusCode, resBody.String())
	})
}

type bodyWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
