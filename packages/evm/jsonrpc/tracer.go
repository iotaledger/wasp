package jsonrpc

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"
)

type Tracer struct {
	*tracers.Tracer
	TraceFakeTx func(tx *types.Transaction) (json.RawMessage, error)
}

type tracerFactory func(*tracers.Context, json.RawMessage) (*Tracer, error)

var allTracers = map[string]tracerFactory{}

func registerTracer(tracerType string, fn tracerFactory) {
	allTracers[tracerType] = fn
}

func newTracer(tracerType string, ctx *tracers.Context, cfg json.RawMessage) (*Tracer, error) {
	fn := allTracers[tracerType]
	if fn == nil {
		return nil, fmt.Errorf("unsupported tracer type: %s", tracerType)
	}
	return fn(ctx, cfg)
}
