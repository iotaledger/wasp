package jsonrpc

import (
	"time"

	"github.com/iotaledger/wasp/packages/metrics"
)

func withMetrics[T any](
	metrics metrics.IChainMetrics,
	method string,
	f func() (T, error),
) (T, error) {
	started := time.Now()
	ret, err := f()
	metrics.EvmRPCCall(method, err == nil, time.Since(started))
	return ret, err
}
