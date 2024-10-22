package sbtests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestGetSet(t *testing.T) { run2(t, testGetSet) }
func testGetSet(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParams(sbtestsc.FuncSetInt.Message("ppp", 314), ScName).
		WithGasBudget(100_000)
	_, err := chain.PostRequestSync(req.AddBaseTokens(1), nil)
	require.NoError(t, err)

	ret, err := sbtestsc.FuncGetInt.Call("ppp", func(msg isc.Message) (isc.CallArguments, error) {
		return chain.CallViewWithContract(ScName, msg)
	})
	require.NoError(t, err)
	require.EqualValues(t, 314, ret)
}

func TestCallRecursive(t *testing.T) { run2(t, testCallRecursive) }
func testCallRecursive(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	depth := 27
	t.Logf("originator base tokens: %d", chain.L2BaseTokens(chain.OriginatorAgentID))
	req := solo.NewCallParams(
		sbtestsc.FuncCallOnChain.Message(isc.NewCallArguments(
			codec.Encode(depth),
			codec.Encode(HScName),
			codec.Encode(sbtestsc.FuncRunRecursion.Hname()),
		)),
		ScName,
	).
		WithGasBudget(5_000_000)
	_, err := chain.PostRequestSync(req, nil)
	t.Logf("receipt: %s", chain.LastReceipt())
	require.NoError(t, err)

	r, err := sbtestsc.FuncGetCounter.Call(func(msg isc.Message) (isc.CallArguments, error) {
		return chain.CallViewWithContract(ScName, msg)
	})
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

	ret, err := sbtestsc.FuncGetFibonacci.Call(fiboN, func(msg isc.Message) (isc.CallArguments, error) {
		return chain.CallViewWithContract(ScName, msg)
	})

	require.NoError(t, err)
	require.EqualValues(t, fibonacci(fiboN), ret)
}

func TestCallFibonacciIndirect(t *testing.T) { run2(t, testCallFibonacciIndirect) }
func testCallFibonacciIndirect(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	ret, err := sbtestsc.FuncGetFibonacciIndirect.Call(fiboN, func(msg isc.Message) (isc.CallArguments, error) {
		return chain.CallViewWithContract(ScName, msg)
	})

	require.NoError(t, err)
	require.EqualValues(t, fibonacci(fiboN), ret)
}

func TestIndirectCallFibonacci(t *testing.T) { run2(t, testIndirectCallFibonacci) }
func testIndirectCallFibonacci(t *testing.T) { //nolint:dupl
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncCallOnChain.Name, isc.NewCallArguments(
		codec.Encode(fiboN),
		codec.EncodeSome(HScName),
		codec.EncodeSome(sbtestsc.FuncGetFibonacci.Hname()),
	)).
		WithGasBudget(5_000_000)
	ret, err := chain.PostRequestSync(req.AddBaseTokens(1), nil)
	require.NoError(t, err)
	r, err := isc.ResAt[uint64](ret, 0)
	require.NoError(t, err)
	require.EqualValues(t, fibonacci(fiboN), r)

	ret, err = chain.CallViewEx(ScName, sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)

	r, err = isc.ResAt[uint64](ret, 0)
	require.NoError(t, err)
	require.EqualValues(t, 1, r)
}

func TestIndirectCallFibonacciIndirect(t *testing.T) { run2(t, testIndirectCallFibonacciIndirect) }
func testIndirectCallFibonacciIndirect(t *testing.T) { //nolint:dupl
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncCallOnChain.Name, isc.NewCallArguments(
		codec.Encode(fiboN),
		codec.EncodeSome(HScName),
		codec.EncodeSome(sbtestsc.FuncGetFibonacciIndirect.Hname()),
	)).
		WithGasBudget(5_000_000)
	ret, err := chain.PostRequestSync(req.AddBaseTokens(1), nil)
	require.NoError(t, err)
	r, err := isc.ResAt[uint64](ret, 0)
	require.NoError(t, err)
	require.EqualValues(t, fibonacci(fiboN), r)

	ret, err = chain.CallViewEx(ScName, sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)

	r, err = isc.ResAt[uint64](ret, 0)
	require.NoError(t, err)
	require.EqualValues(t, 1, r)
}
