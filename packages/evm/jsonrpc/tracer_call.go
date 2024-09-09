package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"
)

func init() {
	registerTracer("callTracer", newCallTracer)
}

// CallFrame contains the result of a trace with "callTracer".
// Code is 100% copied from go-ethereum (since the type is unexported there)

type callLog struct {
	Address common.Address `json:"address"`
	Topics  []common.Hash  `json:"topics"`
	Data    hexutil.Bytes  `json:"data"`
	// Position of the log relative to subcalls within the same trace
	// See https://github.com/ethereum/go-ethereum/pull/28389 for details
	Position hexutil.Uint `json:"position"`
}

type CallFrame struct {
	Type         vm.OpCode       `json:"-"`
	From         common.Address  `json:"from"`
	Gas          uint64          `json:"gas"`
	GasUsed      uint64          `json:"gasUsed"`
	To           *common.Address `json:"to,omitempty" rlp:"optional"`
	Input        []byte          `json:"input" rlp:"optional"`
	Output       []byte          `json:"output,omitempty" rlp:"optional"`
	Error        string          `json:"error,omitempty" rlp:"optional"`
	RevertReason string          `json:"revertReason,omitempty"`
	Calls        []CallFrame     `json:"calls,omitempty" rlp:"optional"`
	Logs         []callLog       `json:"logs,omitempty" rlp:"optional"`
	// Placed at end on purpose. The RLP will be decoded to 0 instead of
	// nil if there are non-empty elements after in the struct.
	Value            *big.Int `json:"value,omitempty" rlp:"optional"`
	revertedSnapshot bool
}

func (f CallFrame) TypeString() string {
	return f.Type.String()
}

func (f CallFrame) failed() bool {
	return f.Error != "" && f.revertedSnapshot
}

func (f *CallFrame) processOutput(output []byte, err error, reverted bool) {
	output = common.CopyBytes(output)
	// Clear error if tx wasn't reverted. This happened
	// for pre-homestead contract storage OOG.
	if err != nil && !reverted {
		err = nil
	}
	if err == nil {
		f.Output = output
		return
	}
	f.Error = err.Error()
	f.revertedSnapshot = reverted
	if f.Type == vm.CREATE || f.Type == vm.CREATE2 {
		f.To = nil
	}
	if !errors.Is(err, vm.ErrExecutionReverted) || len(output) == 0 {
		return
	}
	f.Output = output
	if len(output) < 4 {
		return
	}
	if unpacked, err := abi.UnpackRevert(output); err == nil {
		f.RevertReason = unpacked
	}
}

type TxTraceResult struct {
	TxHash common.Hash `json:"txHash"`           // transaction hash
	Result CallFrame   `json:"result,omitempty"` // Trace results produced by the tracer
	Error  string      `json:"error,omitempty"`  // Trace failure produced by the tracer
}

type callTracer struct {
	callstack []CallFrame
	config    callTracerConfig
	gasLimit  uint64
	depth     int
	interrupt atomic.Bool        // Atomic flag to signal execution interruption
	reason    error              // Textual reason for the interruption
	blockTxs  types.Transactions // for block tracing we need this to get ordered tx hashes
}

type callTracerConfig struct {
	OnlyTopCall bool `json:"onlyTopCall"` // If true, call tracer won't collect any subcalls
	WithLog     bool `json:"withLog"`     // If true, call tracer will collect event logs
}

// newCallTracer returns a native go tracer which tracks
// call frames of a tx, and implements vm.EVMLogger.
func newCallTracer(ctx *tracers.Context, cfg json.RawMessage, initValue any) (*tracers.Tracer, error) {
	var txs types.Transactions
	if initValue != nil {

		var ok bool
		txs, ok = initValue.(types.Transactions)
		if !ok {
			return nil, fmt.Errorf("invalid init value type for tracer: %T", initValue)
		}
	}

	t, err := newCallTracerObject(ctx, cfg, txs)
	if err != nil {
		return nil, err
	}
	return &tracers.Tracer{
		Hooks: &tracing.Hooks{
			OnTxStart: t.OnTxStart,
			OnTxEnd:   t.OnTxEnd,
			OnEnter:   t.OnEnter,
			OnExit:    t.OnExit,
			OnLog:     t.OnLog,
		},
		GetResult: t.GetResult,
		Stop:      t.Stop,
	}, nil
}

func newCallTracerObject(_ *tracers.Context, cfg json.RawMessage, blockTxs types.Transactions) (*callTracer, error) {
	var config callTracerConfig
	if cfg != nil {
		if err := json.Unmarshal(cfg, &config); err != nil {
			return nil, err
		}
	}
	// First callframe contains tx context info
	// and is populated on start and end.
	return &callTracer{callstack: make([]CallFrame, 0, 1), config: config, blockTxs: blockTxs}, nil
}

// OnEnter is called when EVM enters a new scope (via call, create or selfdestruct).
func (t *callTracer) OnEnter(depth int, typ byte, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
	t.depth = depth
	if t.config.OnlyTopCall && depth > 0 {
		return
	}
	// Skip if tracing was interrupted
	if t.interrupt.Load() {
		return
	}

	toCopy := to
	call := CallFrame{
		Type:  vm.OpCode(typ),
		From:  from,
		To:    &toCopy,
		Input: common.CopyBytes(input),
		Gas:   gas,
		Value: value,
	}
	if depth == 0 {
		call.Gas = t.gasLimit
	}
	t.callstack = append(t.callstack, call)
}

// OnExit is called when EVM exits a scope, even if the scope didn't
// execute any code.
func (t *callTracer) OnExit(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
	if depth == 0 {
		t.captureEnd(output, gasUsed, err, reverted)
		return
	}

	t.depth = depth - 1
	if t.config.OnlyTopCall {
		return
	}

	size := len(t.callstack)
	if size <= 1 {
		return
	}
	// Pop call.
	call := t.callstack[size-1]
	t.callstack = t.callstack[:size-1]
	size--

	call.GasUsed = gasUsed
	call.processOutput(output, err, reverted)
	// Nest call into parent.
	t.callstack[size-1].Calls = append(t.callstack[size-1].Calls, call)
}

func (t *callTracer) captureEnd(output []byte, _ uint64, err error, reverted bool) {
	if len(t.callstack) != 1 {
		return
	}
	t.callstack[0].processOutput(output, err, reverted)
}

func (t *callTracer) OnTxStart(env *tracing.VMContext, tx *types.Transaction, from common.Address) {
	t.gasLimit = tx.Gas()
}

func (t *callTracer) OnTxEnd(receipt *types.Receipt, err error) {
	// Error happened during tx validation.
	if err != nil {
		return
	}
	t.callstack[0].GasUsed = receipt.GasUsed
	if t.config.WithLog {
		// Logs are not emitted when the call fails
		clearFailedLogs(&t.callstack[0], false)
	}
}

func (t *callTracer) OnLog(log *types.Log) {
	// Only logs need to be captured via opcode processing
	if !t.config.WithLog {
		return
	}
	// Avoid processing nested calls when only caring about top call
	if t.config.OnlyTopCall && t.depth > 0 {
		return
	}
	// Skip if tracing was interrupted
	if t.interrupt.Load() {
		return
	}
	l := callLog{
		Address:  log.Address,
		Topics:   log.Topics,
		Data:     log.Data,
		Position: hexutil.Uint(len(t.callstack[len(t.callstack)-1].Calls)),
	}
	t.callstack[len(t.callstack)-1].Logs = append(t.callstack[len(t.callstack)-1].Logs, l)
}

// GetResult returns the json-encoded nested list of call traces, and any
// error arising from the encoding or forceful termination (via `Stop`).
func (t *callTracer) GetResult() (json.RawMessage, error) {
	if len(t.callstack) == 1 {
		res, err := json.Marshal(t.callstack[0])
		if err != nil {
			return nil, err
		}
		return res, t.reason
	}

	// otherwise return all call frames
	results := make([]TxTraceResult, 0, len(t.callstack))
	for i, cs := range t.callstack {
		results = append(results, TxTraceResult{
			TxHash: t.blockTxs[i].Hash(),
			Result: cs,
		})
	}

	resJson, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}

	return resJson, t.reason
}

// Stop terminates execution of the tracer at the first opportune moment.
func (t *callTracer) Stop(err error) {
	t.reason = err
	t.interrupt.Store(true)
}

// clearFailedLogs clears the logs of a callframe and all its children
// in case of execution failure.
func clearFailedLogs(cf *CallFrame, parentFailed bool) {
	failed := cf.failed() || parentFailed
	// Clear own logs
	if failed {
		cf.Logs = nil
	}
	for i := range cf.Calls {
		clearFailedLogs(&cf.Calls[i], failed)
	}
}
