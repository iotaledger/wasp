package jsonrpc

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/eth/tracers"
)

type tracerFactory func(cfg json.RawMessage) (tracers.Tracer, error)

var allTracers = map[string]tracerFactory{}

func registerTracer(tracerType string, fn tracerFactory) {
	allTracers[tracerType] = fn
}

func newTracer(tracerType string, cfg json.RawMessage) (tracers.Tracer, error) {
	fn := allTracers[tracerType]
	if fn == nil {
		return nil, fmt.Errorf("unsupported tracer type: %s", tracerType)
	}
	return fn(cfg)
}
