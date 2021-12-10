package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/vm/wasmsolo"
	"github.com/stretchr/testify/require"
)

const n = 10

func fibo(n int64) int64 {
	if n == 0 || n == 1 {
		return n
	}
	return fibo(n-1) + fibo(n-2)
}

func TestCallFibonacci(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		if *wasmsolo.TsWasm {
			t.SkipNow()
		}

		ctx := deployTestCore(t, w)

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
		if *wasmsolo.TsWasm {
			t.SkipNow()
		}

		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.CallOnChain(ctx)
		f.Params.IntValue().SetValue(n)
		f.Params.HnameContract().SetValue(testcore.HScName)
		f.Params.HnameEP().SetValue(testcore.HViewFibonacci)
		f.Func.TransferIotas(1).Post()
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
		// TODO need to adjust stack size for Go Wasm for this to succeed
		if *wasmsolo.GoWasm || *wasmsolo.TsWasm {
			t.SkipNow()
		}

		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.CallOnChain(ctx)
		f.Params.IntValue().SetValue(31)
		f.Params.HnameContract().SetValue(testcore.HScName)
		f.Params.HnameEP().SetValue(testcore.HFuncRunRecursion)
		f.Func.TransferIotas(1).Post()
		require.NoError(t, ctx.Err)

		v := testcore.ScFuncs.GetCounter(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
		counter := v.Results.Counter()
		require.True(t, counter.Exists())
		require.EqualValues(t, 32, counter.Value())
	})
}

func TestGetSet(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.SetInt(ctx)
		f.Params.Name().SetValue("ppp")
		f.Params.IntValue().SetValue(314)
		f.Func.TransferIotas(1).Post()
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
