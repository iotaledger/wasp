// Code on this file adapted from
// https://github.com/ethereum/go-ethereum/blob/master/eth/tracers/native/call.go

package jsonrpc

import (
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/wasp/packages/evm/evmutil"
)

func init() {
	registerTracer("callTracer", newCallTracer)
}

// CallFrame contains the result of a trace with "callTracer".
// Code is 100% copied from go-ethereum (since the type is unexported there)

type CallLog struct {
	Address common.Address `json:"address"`
	Topics  []common.Hash  `json:"topics"`
	Data    hexutil.Bytes  `json:"data"`
	// Position of the log relative to subcalls within the same trace
	// See https://github.com/ethereum/go-ethereum/pull/28389 for details
	Position hexutil.Uint `json:"position"`
}

type OpCodeJSON struct {
	vm.OpCode
}

func NewOpCodeJSON(code vm.OpCode) OpCodeJSON {
	return OpCodeJSON{OpCode: code}
}

func (o OpCodeJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.ToUpper(o.String()))
}

func (o *OpCodeJSON) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	o.OpCode = vm.StringToOp(strings.ToUpper(s))
	return nil
}

type CallFrame struct {
	Type         OpCodeJSON      `json:"type"`
	From         common.Address  `json:"from"`
	Gas          hexutil.Uint64  `json:"gas"`
	GasUsed      hexutil.Uint64  `json:"gasUsed"`
	To           *common.Address `json:"to,omitempty" rlp:"optional"`
	Input        hexutil.Bytes   `json:"input" rlp:"optional"`
	Output       hexutil.Bytes   `json:"output,omitempty" rlp:"optional"`
	Error        string          `json:"error,omitempty" rlp:"optional"`
	RevertReason string          `json:"revertReason,omitempty"`
	Calls        []CallFrame     `json:"calls,omitempty" rlp:"optional"`
	Logs         []CallLog       `json:"logs,omitempty" rlp:"optional"`
	// Placed at end on purpose. The RLP will be decoded to 0 instead of
	// nil if there are non-empty elements after in the struct.
	Value            hexutil.Big `json:"value,omitempty" rlp:"optional"`
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
	if f.Type.OpCode == vm.CREATE || f.Type.OpCode == vm.CREATE2 {
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
	TxHash common.Hash     `json:"txHash"`           // transaction hash
	Result json.RawMessage `json:"result,omitempty"` // Trace results produced by the tracer
	Error  string          `json:"error,omitempty"`  // Trace failure produced by the tracer
}

type callTracer struct {
	txToStack  map[common.Hash][]CallFrame
	currentTx  common.Hash
	config     callTracerConfig
	gasLimit   uint64
	depth      int
	interrupt  atomic.Bool // Atomic flag to signal execution interruption
	reason     error       // Textual reason for the interruption
	traceBlock bool
}

type callTracerConfig struct {
	OnlyTopCall bool `json:"onlyTopCall"` // If true, call tracer won't collect any subcalls
	WithLog     bool `json:"withLog"`     // If true, call tracer will collect event logs
}

// newCallTracer returns a native go tracer which tracks
// call frames of a tx, and implements vm.EVMLogger.
func newCallTracer(ctx *tracers.Context, cfg json.RawMessage, traceBlock bool) (*Tracer, error) {
	t, err := newCallTracerObject(ctx, cfg, traceBlock)
	if err != nil {
		return nil, err
	}
	return &Tracer{
		Tracer: &tracers.Tracer{
			Hooks: &tracing.Hooks{
				OnTxStart: t.OnTxStart,
				OnTxEnd:   t.OnTxEnd,
				OnEnter:   t.OnEnter,
				OnExit:    t.OnExit,
				OnLog:     t.OnLog,
			},
			GetResult: t.GetResult,
			Stop:      t.Stop,
		},
		TraceFakeTx: t.TraceFakeTx,
	}, nil
}

func newCallTracerObject(_ *tracers.Context, cfg json.RawMessage, traceBlock bool) (*callTracer, error) {
	var config callTracerConfig
	if cfg != nil {
		if err := json.Unmarshal(cfg, &config); err != nil {
			return nil, err
		}
	}
	// First callframe contains tx context info
	// and is populated on start and end.
	return &callTracer{txToStack: make(map[common.Hash][]CallFrame), currentTx: common.Hash{}, config: config, traceBlock: traceBlock}, nil
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
		Type:  NewOpCodeJSON(vm.OpCode(typ)),
		From:  from,
		To:    &toCopy,
		Input: common.CopyBytes(input),
		Gas:   hexutil.Uint64(gas),
		Value: hexutil.Big(*value),
	}
	if depth == 0 {
		call.Gas = hexutil.Uint64(t.gasLimit)
	}
	t.txToStack[t.currentTx] = append(t.txToStack[t.currentTx], call)
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

	size := len(t.txToStack[t.currentTx])
	if size <= 1 {
		return
	}
	// Pop call.
	call := t.txToStack[t.currentTx][size-1]
	t.txToStack[t.currentTx] = t.txToStack[t.currentTx][:size-1]
	size--

	call.GasUsed = hexutil.Uint64(gasUsed)
	call.processOutput(output, err, reverted)
	// Nest call into parent.
	t.txToStack[t.currentTx][size-1].Calls = append(t.txToStack[t.currentTx][size-1].Calls, call)
}

func (t *callTracer) captureEnd(output []byte, _ uint64, err error, reverted bool) {
	if len(t.txToStack[t.currentTx]) != 1 {
		return
	}
	t.txToStack[t.currentTx][0].processOutput(output, err, reverted)
}

func (t *callTracer) OnTxStart(env *tracing.VMContext, tx *types.Transaction, from common.Address) {
	t.gasLimit = tx.Gas()
	t.currentTx = tx.Hash()
	t.txToStack[t.currentTx] = make([]CallFrame, 0, 1)
}

func (t *callTracer) OnTxEnd(receipt *types.Receipt, err error) {
	// Error happened during tx validation.
	if err != nil {
		return
	}
	t.txToStack[t.currentTx][0].GasUsed = hexutil.Uint64(receipt.GasUsed)
	if t.config.WithLog {
		// Logs are not emitted when the call fails
		clearFailedLogs(&t.txToStack[t.currentTx][0], false)
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
	l := CallLog{
		Address:  log.Address,
		Topics:   log.Topics,
		Data:     log.Data,
		Position: hexutil.Uint(len(t.txToStack[t.currentTx][len(t.txToStack[t.currentTx])-1].Calls)),
	}
	t.txToStack[t.currentTx][len(t.txToStack[t.currentTx])-1].Logs = append(t.txToStack[t.currentTx][len(t.txToStack[t.currentTx])-1].Logs, l)
}

var ErrIncorrectTopLevelCalls = errors.New("incorrect number of top-level calls")

// GetResult returns the json-encoded nested list of call traces, and any
// error arising from the encoding or forceful termination (via `Stop`).
func (t *callTracer) GetResult() (json.RawMessage, error) {
	if t.traceBlock {

		// otherwise return all call frames
		results := make([]TxTraceResult, 0, len(t.txToStack))
		for txHash, stack := range t.txToStack {
			csJSON, err := json.Marshal(stack[0])
			if err != nil {
				return nil, err
			}
			results = append(results, TxTraceResult{
				TxHash: txHash,
				Result: csJSON,
			})
		}

		res, err := json.Marshal(results)
		if err != nil {
			return nil, err
		}
		return res, t.reason
	} else {
		if len(t.txToStack) != 1 {
			return nil, ErrIncorrectTopLevelCalls
		}

		res, err := json.Marshal(t.txToStack[t.currentTx][0])
		if err != nil {
			return nil, err
		}

		return res, t.reason
	}
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

func (t *callTracer) TraceFakeTx(tx *types.Transaction) (json.RawMessage, error) {
	return json.Marshal(CallFrame{
		Type:    NewOpCodeJSON(vm.CALL),
		From:    evmutil.MustGetSenderIfTxSigned(tx),
		Gas:     hexutil.Uint64(tx.Gas()),
		GasUsed: hexutil.Uint64(tx.Gas()),
		To:      tx.To(),
		Input:   []byte{},
		Output:  []byte{},
		Value:   hexutil.Big(*tx.Value()),
	})
}
