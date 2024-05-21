package sbtests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestGetSet(t *testing.T) { run2(t, testGetSet) }
func testGetSet(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncSetInt.Name,
		sbtestsc.ParamIntParamName, "ppp",
		sbtestsc.ParamIntParamValue, 314).
		WithGasBudget(100_000)
	_, err := chain.PostRequestSync(req.AddBaseTokens(1), nil)
	require.NoError(t, err)

	ret, err := chain.CallViewEx(ScName, sbtestsc.FuncGetInt.Name,
		sbtestsc.ParamIntParamName, "ppp")
	require.NoError(t, err)

	retInt, err := codec.Int64.Decode(ret.Get("ppp"))
	require.NoError(t, err)
	require.EqualValues(t, 314, retInt)
}

func TestCallRecursive(t *testing.T) { run2(t, testCallRecursive) }
func testCallRecursive(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	depth := 27
	t.Logf("originator base tokens: %d", chain.L2BaseTokens(chain.OriginatorAgentID))
	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncCallOnChain.Name,
		sbtestsc.ParamN, depth,
		sbtestsc.ParamHnameContract, HScName,
		sbtestsc.ParamHnameEP, sbtestsc.FuncRunRecursion.Hname()).
		WithGasBudget(5_000_000)
	_, err := chain.PostRequestSync(req, nil)
	t.Logf("receipt: %s", chain.LastReceipt())
	require.NoError(t, err)

	ret, err := chain.CallViewEx(ScName, sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)

	r, err := codec.Int64.Decode(ret.Get(sbtestsc.VarCounter))
	require.NoError(t, err)
	require.EqualValues(t, depth+1, r)
}

const fiboN = 8

func fibonacci(n int64) int64 {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func TestCallFibonacci(t *testing.T) { run2(t, testCallFibonacci) }
func testCallFibonacci(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	ret, err := chain.CallViewEx(ScName, sbtestsc.FuncGetFibonacci.Name,
		sbtestsc.ParamN, fiboN,
	)
	require.NoError(t, err)
	val, err := codec.Uint64.Decode(ret.Get(sbtestsc.ParamN))
	require.NoError(t, err)
	require.EqualValues(t, fibonacci(fiboN), val)
}

func TestCallFibonacciIndirect(t *testing.T) { run2(t, testCallFibonacciIndirect) }
func testCallFibonacciIndirect(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	ret, err := chain.CallViewEx(ScName, sbtestsc.FuncGetFibonacciIndirect.Name,
		sbtestsc.ParamN, fiboN,
	)
	require.NoError(t, err)
	val, err := codec.Uint64.Decode(ret.Get(sbtestsc.ParamN))
	require.NoError(t, err)
	require.EqualValues(t, fibonacci(fiboN), val)
}

func TestIndirectCallFibonacci(t *testing.T) { run2(t, testIndirectCallFibonacci) }
func testIndirectCallFibonacci(t *testing.T) { //nolint:dupl
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncCallOnChain.Name,
		sbtestsc.ParamN, fiboN,
		sbtestsc.ParamHnameContract, HScName,
		sbtestsc.ParamHnameEP, sbtestsc.FuncGetFibonacci.Hname()).
		WithGasBudget(5_000_000)
	ret, err := chain.PostRequestSync(req.AddBaseTokens(1), nil)
	require.NoError(t, err)
	r, err := codec.Uint64.Decode(ret.Get(sbtestsc.ParamN))
	require.NoError(t, err)
	require.EqualValues(t, fibonacci(fiboN), r)

	ret, err = chain.CallViewEx(ScName, sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)

	r, err = codec.Uint64.Decode(ret.Get(sbtestsc.VarCounter))
	require.NoError(t, err)
	require.EqualValues(t, 1, r)
}

func TestIndirectCallFibonacciIndirect(t *testing.T) { run2(t, testIndirectCallFibonacciIndirect) }
func testIndirectCallFibonacciIndirect(t *testing.T) { //nolint:dupl
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncCallOnChain.Name,
		sbtestsc.ParamN, fiboN,
		sbtestsc.ParamHnameContract, HScName,
		sbtestsc.ParamHnameEP, sbtestsc.FuncGetFibonacciIndirect.Hname()).
		WithGasBudget(5_000_000)
	ret, err := chain.PostRequestSync(req.AddBaseTokens(1), nil)
	require.NoError(t, err)
	r, err := codec.Uint64.Decode(ret.Get(sbtestsc.ParamN))
	require.NoError(t, err)
	require.EqualValues(t, fibonacci(fiboN), r)

	ret, err = chain.CallViewEx(ScName, sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)

	r, err = codec.Uint64.Decode(ret.Get(sbtestsc.VarCounter))
	require.NoError(t, err)
	require.EqualValues(t, 1, r)
}
