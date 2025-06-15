package jsonrpc

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth/tracers"
)

type tracerFactory func(traceCtx *tracers.Context, cfg json.RawMessage) (*tracers.Tracer, error)

var allTracers = map[string]tracerFactory{}

func registerTracer(tracerType string, fn tracerFactory) {
	allTracers[tracerType] = fn
}

func newTracer(
	ctx *tracers.Context,
	config *tracers.TraceConfig,
) (*tracers.Tracer, error) {
	tracerType := "callTracer"
	if config.Tracer != nil {
		tracerType = *config.Tracer
	}

	fn := allTracers[tracerType]
	if fn == nil {
		return nil, fmt.Errorf("unsupported tracer type: %s", tracerType)
	}
	return fn(ctx, config.TracerConfig)
}

type TxTraceResult struct {
	TxHash common.Hash     `json:"txHash"`           // transaction hash
	Result json.RawMessage `json:"result,omitempty"` // Trace results produced by the tracer
	Error  string          `json:"error,omitempty"`  // Trace failure produced by the tracer
}
