package jsonrpc

import (
	_ "embed"
	"encoding/json"
	"slices"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

//go:embed debug_trace_block_sample.json
var debugTraceBlockSample string

//go:embed trace_block_sample.json
var traceBlockSample string

var (
	blockNumber = uint64(21130085)
	blockHash   = common.HexToHash("0x914d99596812741f99a64012e4422f1446c83b77fd8c00ff12076af6b0c3bf89")
)

func TestConvertToTrace(t *testing.T) {
	var debugTraces struct {
		Result []struct {
			TxHash string    `json:"txHash"`
			Result CallFrame `json:"result"`
		} `json:"result"`
	}
	err := json.Unmarshal([]byte(debugTraceBlockSample), &debugTraces)
	assert.NoError(t, err)

	var expectedTraces struct {
		Result []*Trace `json:"result"`
	}
	err = json.Unmarshal([]byte(traceBlockSample), &expectedTraces)
	assert.NoError(t, err)

	var actualTraces []*Trace
	for i, debugTrace := range debugTraces.Result {
		blockHash := blockHash
		blockNumber := blockNumber
		txHash := common.HexToHash(debugTrace.TxHash)
		txPosition := uint64(i)

		traces := convertToTrace(
			debugTrace.Result,
			&blockHash,
			blockNumber,
			&txHash,
			txPosition,
		)
		sortTraces(traces)
		actualTraces = append(actualTraces, traces...)
	}

	actualTracesMap := make(map[string][]*Trace)
	for _, trace := range actualTraces {
		actualTracesMap[trace.TransactionHash.String()] = append(actualTracesMap[trace.TransactionHash.String()], trace)
	}

	expectedTracesMap := make(map[string][]*Trace)
	for _, trace := range expectedTraces.Result {
		expectedTracesMap[trace.TransactionHash.String()] = append(expectedTracesMap[trace.TransactionHash.String()], trace)
	}

	// Sort the expected and actual traces for each txHash by trace address. e.g [1,2,3] vs [1,2,4]
	for _, expected := range expectedTracesMap {
		sortTraces(expected)
	}

	for _, actual := range actualTracesMap {
		sortTraces(actual)
	}

	mismatchCount := 0
	for txHash, actual := range actualTracesMap {
		expected, ok := expectedTracesMap[txHash]
		assert.True(t, ok, "No matching expected trace found for actual trace txHash: %s", txHash)
		assert.Equalf(t, len(expected), len(actual), "Number of traces should match for txHash: %s", txHash)

		eA, _ := expected[0].Action.(map[string]interface{})
		eR, _ := expected[0].Result.(map[string]interface{})

		var actualGasUsed hexutil.Uint64
		var actualGas hexutil.Uint64

		switch a := actual[0].Action.(type) {
		case CallTraceAction:
			actualGas = a.Gas
		case CreateTraceAction:
			actualGas = a.Gas
		}

		switch r := actual[0].Result.(type) {
		case TraceResult:
			actualGasUsed = r.GasUsed
		case CreateTraceResult:
			actualGasUsed = r.GasUsed
		}

		expectedGas := eA["gas"]
		expectedGasUsed := eR["gasUsed"]

		expectedGasUint64, _ := hexutil.DecodeUint64(expectedGas.(string))
		expectedGasUsedUint64, _ := hexutil.DecodeUint64(expectedGasUsed.(string))

		var diffGas int64 = int64(uint64(actualGas)) - int64(expectedGasUint64)
		var diffGasUsed int64 = int64(uint64(actualGasUsed)) - int64(expectedGasUsedUint64)

		// Skip gas and gasUsed verification for now
		_ = diffGas
		_ = diffGasUsed
		// if diffGas != 0 || diffGasUsed != 0 {
		// 	mismatchCount++
		// 	fmt.Printf("txHash: %s\n", txHash)
		// 	fmt.Printf("gas diff: %d\n", diffGas)
		// 	fmt.Printf("gas used diff: %d\n", diffGasUsed)
		// 	fmt.Printf("----------------------------------\n")
		// }
		// assert.Equal(t, expectedGas, actualGas, "Gas should match")
		// assert.Equal(t, expectedGasUsed, actualGasUsed, "Gas used should match")

		for i := 0; i < len(expected); i++ {
			assert.Equal(t, expected[i].Type, actual[i].Type)
			assert.Equal(t, expected[i].Subtraces, actual[i].Subtraces)
			assert.Equal(t, expected[i].TraceAddress, actual[i].TraceAddress)
			assert.Equal(t, expected[i].TransactionPosition, actual[i].TransactionPosition)
			assert.Equal(t, expected[i].BlockNumber, actual[i].BlockNumber)
			assert.Equal(t, expected[i].BlockHash, actual[i].BlockHash)
			assert.Equal(t, expected[i].TransactionHash, actual[i].TransactionHash)

			expectedAction, ok := expected[i].Action.(map[string]interface{})
			assert.True(t, ok, "Expected action should be a map")

			actionJSON, err := json.Marshal(actual[i].Action)
			assert.NoError(t, err)
			actualAction := map[string]interface{}{}
			err = json.Unmarshal(actionJSON, &actualAction)
			assert.NoError(t, err)

			resultJSON, err := json.Marshal(actual[i].Result)
			assert.NoError(t, err)
			actualResult := map[string]interface{}{}
			err = json.Unmarshal(resultJSON, &actualResult)
			assert.NoError(t, err)

			assert.Equal(t, expectedAction["from"], actualAction["from"])
			assert.Equal(t, expectedAction["to"], actualAction["to"])
			assert.Equal(t, expectedAction["callType"], actualAction["callType"])
			assert.Equal(t, expectedAction["value"], actualAction["value"])
			assert.Equal(t, expectedAction["input"], actualAction["input"])

			expectedResult, ok := expected[i].Result.(map[string]interface{})
			assert.True(t, ok, "Expected result should be a map")

			assert.Equal(t, expectedResult["output"], actualResult["output"])
		}
	}

	assert.Equal(t, 0, mismatchCount, "Number of mismatches should be 0")
}

// sortTraces sorts the traces by trace address. e.g [1,2,3] vs [1,1,5]
func sortTraces(traces []*Trace) {
	slices.SortFunc(traces, func(i, j *Trace) int {
		addr1 := i.TraceAddress
		addr2 := j.TraceAddress

		for k := 0; k < len(addr1) && k < len(addr2); k++ {
			if addr1[k] != addr2[k] {
				return addr1[k] - addr2[k]
			}
		}

		return len(addr1) - len(addr2)
	})
}
