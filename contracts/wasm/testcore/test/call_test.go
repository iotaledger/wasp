package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

//  N   Fib(N)   Calls
//  0        1       1
//  1        1       1
//  2        2       3
//  3        3       5
//  4        5       9
//  5        8      15
//  6       13      25
//  7       21      41
//  8       34      67
//  9       55     109
// 10       89     177

// Note: we will need enough gas to go N deep
// Fib(N) requires 2*Fib(N)-1 calls
// Rust and Go burn about 70K gas per call
// Typescript burns about 600K gas per call
// We have a hard-coded budget of 5M gas for a call

// Turns out N=8 stays just within budget for Rust and Go
const fiboN = int64(8)

func fibo(n int64) int64 {
	if n == 0 || n == 1 {
		return n
	}
	return fibo(n-1) + fibo(n-2)
}

func TestCallFibonacci(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		n := fiboN
		if *wasmsolo.TsWasm {
			// Typescript burns 600K gas per call
			n = 7
		}

		f := testcore.ScFuncs.Fibonacci(ctx)
		f.Params.IntValue().SetValue(n)
		f.Func.Call()
		require.NoError(t, ctx.Err)
		result := f.Results.IntValue()
		require.True(t, result.Exists())
		require.EqualValues(t, fibo(n), result.Value())
	})
}

func TestCallFibonacciIndirect(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		n := fiboN
		if *wasmsolo.TsWasm {
			n = 7
		}

		f := testcore.ScFuncs.CallOnChain(ctx)
		f.Params.IntValue().SetValue(n)
		f.Params.HnameContract().SetValue(testcore.HScName)
		f.Params.HnameEP().SetValue(testcore.HViewFibonacci)
		f.Func.Post()
		require.NoError(t, ctx.Err)
		result := f.Results.IntValue()
		require.True(t, result.Exists())
		require.EqualValues(t, fibo(n), result.Value())

		v := testcore.ScFuncs.GetCounter(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
		counter := v.Results.Counter()
		require.True(t, counter.Exists())
		require.EqualValues(t, 1, counter.Value())
	})
}

func TestCallRecursive(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		depth := int64(27)

		f := testcore.ScFuncs.CallOnChain(ctx)
		f.Params.IntValue().SetValue(depth)
		f.Params.HnameContract().SetValue(testcore.HScName)
		f.Params.HnameEP().SetValue(testcore.HFuncRunRecursion)
		f.Func.Post()
		require.NoError(t, ctx.Err)

		v := testcore.ScFuncs.GetCounter(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
		counter := v.Results.Counter()
		require.True(t, counter.Exists())
		require.EqualValues(t, depth+1, counter.Value())
	})
}

func TestGetSet(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.SetInt(ctx)
		f.Params.Name().SetValue("ppp")
		f.Params.IntValue().SetValue(314)
		f.Func.Post()
		require.NoError(t, ctx.Err)

		v := testcore.ScFuncs.GetInt(ctx)
		v.Params.Name().SetValue("ppp")
		v.Func.Call()
		require.NoError(t, ctx.Err)
		value := v.Results.Values().GetInt64("ppp")
		require.True(t, value.Exists())
		require.EqualValues(t, 314, value.Value())
	})
}
