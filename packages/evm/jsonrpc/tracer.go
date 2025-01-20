package jsonrpc

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/wasp/packages/evm/evmutil"
)

type Tracer struct {
	*tracers.Tracer
	TraceFakeTx func(tx *types.Transaction) (json.RawMessage, error)
}

type tracerFactory func(traceCtx *tracers.Context, cfg json.RawMessage, traceBlock bool, initValue any) (*Tracer, error)

var allTracers = map[string]tracerFactory{}

func registerTracer(tracerType string, fn tracerFactory) {
	allTracers[tracerType] = fn
}

func newTracer(tracerType string, ctx *tracers.Context, cfg json.RawMessage, traceBlock bool, initValue any) (*Tracer, error) {
	fn := allTracers[tracerType]
	if fn == nil {
		return nil, fmt.Errorf("unsupported tracer type: %s", tracerType)
	}
	return fn(ctx, cfg, traceBlock, initValue)
}

func GetTraceResults(
	blockTxs []*types.Transaction,
	traceBlock bool,
	getFakeTxTrace func(tx *types.Transaction) (json.RawMessage, error),
	getTxTrace func(tx *types.Transaction) (json.RawMessage, error),
	getSingleTxTrace func() (json.RawMessage, error),
	reason error,
) (json.RawMessage, error) {
	var traceResult []byte
	var err error
	if traceBlock {
		results := make([]TxTraceResult, 0, len(blockTxs))
		var jsResult json.RawMessage
		for _, tx := range blockTxs {
			if evmutil.IsFakeTransaction(tx) {
				jsResult, err = getFakeTxTrace(tx)
				if err != nil {
					return nil, err
				}
			} else {
				jsResult, err = getTxTrace(tx)
				if err != nil {
					return nil, err
				}
			}

			results = append(results, TxTraceResult{TxHash: tx.Hash(), Result: jsResult})
		}

		traceResult, err = json.Marshal(results)
		if err != nil {
			return nil, err
		}
	} else {
		traceResult, err = getSingleTxTrace()
		if err != nil {
			return nil, err
		}
	}

	return traceResult, reason
}
