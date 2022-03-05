package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func TestInitSuccess(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		init := testcore.ScFuncs.Init(nil)

		require.False(t, init.Params.Fail().Exists(), "init must be successful")
		ctx := deployTestCoreOnChain(t, w, nil, nil, init.Func)
		require.NoError(t, ctx.Err)
	})
}

func TestInitFail(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		init := testcore.ScFuncs.Init(nil)
		init.Params.Fail().SetValue(1)

		require.True(t, init.Params.Fail().Exists(), "init must panic")
		ctx := deployTestCoreOnChain(t, w, nil, nil, init.Func)
		require.Error(t, ctx.Err)
	})
}

func TestInitFailThenInitSuccess(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		wasmsolo.StartChain(t, "chain1")
		init := testcore.ScFuncs.Init(nil)
		init.Params.Fail().SetValue(1)

		require.True(t, init.Params.Fail().Exists(), "init must panic")
		ctx := deployTestCoreOnChain(t, w, nil, nil, init.Func)
		require.Error(t, ctx.Err)

		_, _, rec := ctx.Chain.GetInfo()
		require.EqualValues(t, len(corecontracts.All), len(rec))

		init = testcore.ScFuncs.Init(nil)

		require.False(t, init.Params.Fail().Exists(), "init must be successful")
		ctxRetry := deployTestCoreOnChain(t, w, ctx.Chain, nil, init.Func)
		require.NoError(t, ctxRetry.Err)

		_, _, rec = ctxRetry.Chain.GetInfo()
		require.EqualValues(t, len(corecontracts.All)+1, len(rec))
	})
}

// This test weeds out a problem where TestInitFailRepeat is causing the next
// test to fail when GoWasmVM version is used last. By adding this dummy test
// we prevent the failing test to happen in an unrelated file
func TestInitSuccessAfterRunningTestInitFailThenInitSuccess(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		init := testcore.ScFuncs.Init(nil)

		require.False(t, init.Params.Fail().Exists(), "init must be successful")
		ctx := deployTestCoreOnChain(t, w, nil, nil, init.Func)
		require.NoError(t, ctx.Err)
	})
}
