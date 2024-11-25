package jsonrpc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/vm"
)

type TraceBlock struct {
	Jsonrpc string   `json:"jsonrpc"`
	Result  []*Trace `json:"result"`
	ID      int      `json:"id"`
}

type Trace struct {
	Action      interface{}  `json:"action"`
	BlockHash   *common.Hash `json:"blockHash,omitempty"`
	BlockNumber uint64       `json:"blockNumber,omitempty"`
	Error       string       `json:"error,omitempty"`
	Result      interface{}  `json:"result"`
	// Subtraces is an integer that represents the number of direct nested (child) calls, or "subtraces," within a trace entry.
	Subtraces int `json:"subtraces"`
	// TraceAddress is an array of integers that represents the path from the top-level transaction trace down to the current subtrace.
	// Each integer in the array specifies an index in the sequence of calls, showing how to "navigate" down the call stack to reach this trace.
	TraceAddress        []int        `json:"traceAddress"`
	TransactionHash     *common.Hash `json:"transactionHash,omitempty"`
	TransactionPosition uint64       `json:"transactionPosition"`
	Type                string       `json:"type"`
}

type CallTraceAction struct {
	From     *common.Address `json:"from"`
	CallType string          `json:"callType"`
	Gas      hexutil.Uint64  `json:"gas"`
	Input    hexutil.Bytes   `json:"input"`
	To       *common.Address `json:"to"`
	Value    hexutil.Big     `json:"value"`
}

type CreateTraceAction struct {
	From  *common.Address `json:"from"`
	Gas   hexutil.Uint64  `json:"gas"`
	Init  hexutil.Bytes   `json:"init"`
	Value hexutil.Big     `json:"value"`
}

type SuicideTraceAction struct {
	Address       *common.Address `json:"address"`
	RefundAddress *common.Address `json:"refundAddress"`
	Balance       hexutil.Big     `json:"balance"`
}

type CreateTraceResult struct {
	Address *common.Address `json:"address,omitempty"`
	Code    hexutil.Bytes   `json:"code"`
	GasUsed hexutil.Uint64  `json:"gasUsed"`
}

type TraceResult struct {
	GasUsed hexutil.Uint64 `json:"gasUsed"`
	Output  hexutil.Bytes  `json:"output"`
}

// Trace types
const (
	call         = "call"
	staticCall   = "staticcall"
	delegateCall = "delegatecall"
	create       = "create"
	suicide      = "suicide"
)

func convertToTrace(debugTrace CallFrame, blockHash *common.Hash, blockNumber uint64, txHash *common.Hash, txPosition uint64) []*Trace {
	result := make([]*Trace, 0)
	traces := parseTraceInternal(debugTrace, blockHash, blockNumber, txHash, txPosition, make([]int, 0))
	result = append(result, traces...)

	return result
}

func isPrecompiled(address *common.Address) bool {
	_, ok := vm.PrecompiledContractsPrague[*address]
	return ok
}

//nolint:funlen
func parseTraceInternal(debugTrace CallFrame, blockHash *common.Hash, blockNumber uint64, txHash *common.Hash, txPosition uint64, traceAddress []int) []*Trace {
	traceResult := make([]*Trace, 0)

	traceType := mapTraceType(debugTrace.Type.String())
	traceTypeSimple := mapTraceTypeSimple(debugTrace.Type.String())

	traceEntry := Trace{
		BlockHash:           blockHash,
		BlockNumber:         blockNumber,
		TraceAddress:        traceAddress,
		TransactionHash:     txHash,
		TransactionPosition: txPosition,
		Type:                traceTypeSimple,
		Error:               debugTrace.Error,
	}

	var gasUsed uint64
	var gas uint64
	const baseGasCost = 21_000

	// If this is the root trace, we need to subtract the base gas cost from the gas and gasUsed
	if len(traceAddress) == 0 {
		gas = uint64(debugTrace.Gas) - baseGasCost
		gasUsed = uint64(debugTrace.GasUsed) - baseGasCost
	} else {
		gas = uint64(debugTrace.Gas)
		gasUsed = uint64(debugTrace.GasUsed)
	}

	traceResult = append(traceResult, &traceEntry)

	subCalls := 0
	for _, call := range debugTrace.Calls {
		// Precompiled contracts are not included in parity trace
		if isPrecompiled(call.To) {
			continue
		}

		traceCopy := make([]int, len(traceAddress))
		copy(traceCopy, traceAddress)
		traceCopy = append(traceCopy, subCalls)
		traces := parseTraceInternal(call, blockHash, blockNumber, txHash, txPosition, traceCopy)
		traceResult = append(traceResult, traces...)
		subCalls++
	}

	traceEntry.Subtraces = subCalls

	switch traceTypeSimple {
	case call:
		action := CallTraceAction{}
		action.CallType = traceType
		action.From = &debugTrace.From
		action.Gas = hexutil.Uint64(gas)
		action.Input = debugTrace.Input
		action.To = debugTrace.To
		action.Value = debugTrace.Value

		traceEntry.Action = action

		result := TraceResult{}
		result.GasUsed = hexutil.Uint64(gasUsed)
		result.Output = debugTrace.Output

		traceEntry.Result = result
	case create:
		action := CreateTraceAction{}
		action.From = &debugTrace.From
		action.Gas = hexutil.Uint64(gas)
		action.Init = debugTrace.Input
		action.Value = debugTrace.Value

		traceEntry.Action = action

		result := CreateTraceResult{}
		result.GasUsed = hexutil.Uint64(gasUsed)
		result.Code = debugTrace.Output
		result.Address = debugTrace.To

		traceEntry.Result = result
	case suicide:
		action := SuicideTraceAction{}
		action.Address = &debugTrace.From
		action.RefundAddress = debugTrace.To
		action.Balance = debugTrace.Value

		traceEntry.Action = action
	}

	return traceResult
}

func mapTraceType(traceType string) string {
	switch traceType {
	case "CALL", "CALLCODE":
		return call
	case "STATICCALL":
		return staticCall
	case "DELEGATECALL":
		return delegateCall
	case "CREATE", "CREATE2":
		return create
	case "SELFDESTRUCT":
		return suicide
	}
	return ""
}

func mapTraceTypeSimple(traceType string) string {
	switch traceType {
	case "CALL", "CALLCODE", "STATICCALL", "DELEGATECALL":
		return call
	case "CREATE", "CREATE2":
		return create
	case "SELFDESTRUCT":
		return suicide
	}
	return ""
}
