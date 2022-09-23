// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/wasm/inccounter/go/inccounter"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
)

func setupTest(t *testing.T) *wasmsolo.SoloContext {
	return wasmsolo.NewSoloContext(t, inccounter.ScName, inccounter.OnLoad)
}

func TestDeploy(t *testing.T) {
	ctx := setupTest(t)
	require.NoError(t, ctx.ContractExists(inccounter.ScName))
}

func TestStateAfterDeploy(t *testing.T) {
	ctx := setupTest(t)

	checkStateCounter(t, ctx, nil)
}

func TestIncrementOnce(t *testing.T) {
	ctx := setupTest(t)

	increment := inccounter.ScFuncs.Increment(ctx)
	increment.Func.Post()
	require.NoError(t, ctx.Err)

	checkStateCounter(t, ctx, 1)
}

func TestIncrementTwice(t *testing.T) {
	ctx := setupTest(t)

	increment := inccounter.ScFuncs.Increment(ctx)
	increment.Func.Post()
	require.NoError(t, ctx.Err)

	increment = inccounter.ScFuncs.Increment(ctx)
	increment.Func.Post()
	require.NoError(t, ctx.Err)

	checkStateCounter(t, ctx, 2)
}

func TestIncrementRepeatThrice(t *testing.T) {
	ctx := setupTest(t)

	repeatMany := inccounter.ScFuncs.RepeatMany(ctx)
	repeatMany.Params.NumRepeats().SetValue(3)
	repeatMany.Func.Post()
	require.NoError(t, ctx.Err)

	require.True(t, ctx.WaitForPendingRequests(3))

	checkStateCounter(t, ctx, 4)
}

func TestIncrementCallIncrement(t *testing.T) {
	ctx := setupTest(t)

	callIncrement := inccounter.ScFuncs.CallIncrement(ctx)
	callIncrement.Func.Post()
	require.NoError(t, ctx.Err)

	checkStateCounter(t, ctx, 2)
}

func TestIncrementCallIncrementRecurse5x(t *testing.T) {
	ctx := setupTest(t)

	callIncrementRecurse5x := inccounter.ScFuncs.CallIncrementRecurse5x(ctx)
	callIncrementRecurse5x.Func.Post()
	require.NoError(t, ctx.Err)

	checkStateCounter(t, ctx, 6)
}

func TestIncrementPostIncrement(t *testing.T) {
	ctx := setupTest(t)

	postIncrement := inccounter.ScFuncs.PostIncrement(ctx)
	postIncrement.Func.Post()
	require.NoError(t, ctx.Err)

	require.True(t, ctx.WaitForPendingRequests(1))

	checkStateCounter(t, ctx, 2)
}

func TestIncrementLocalStateInternalCall(t *testing.T) {
	ctx := setupTest(t)

	localStateInternalCall := inccounter.ScFuncs.LocalStateInternalCall(ctx)
	localStateInternalCall.Func.Post()
	require.NoError(t, ctx.Err)

	checkStateCounter(t, ctx, 2)
}

func TestIncrementLocalStateSandboxCall(t *testing.T) {
	ctx := setupTest(t)

	localStateSandboxCall := inccounter.ScFuncs.LocalStateSandboxCall(ctx)
	localStateSandboxCall.Func.Post()
	require.NoError(t, ctx.Err)

	if ctx.IsWasm {
		// global var in wasm execution has no effect
		checkStateCounter(t, ctx, nil)
		return
	}

	// when using WasmGoVM the 3 posts are run only after
	// the LocalStateMustIncrement has been set to true
	checkStateCounter(t, ctx, 2)
}

func TestIncrementLocalStatePost(t *testing.T) {
	ctx := setupTest(t)

	localStatePost := inccounter.ScFuncs.LocalStatePost(ctx)
	localStatePost.Func.Post()
	require.NoError(t, ctx.Err)

	require.True(t, ctx.WaitForPendingRequests(3))

	if ctx.IsWasm {
		// global var in wasm execution has no effect
		checkStateCounter(t, ctx, nil)
		return
	}

	// when using WasmGoVM the 3 posts are run only after
	// the LocalStateMustIncrement has been set to true
	checkStateCounter(t, ctx, 3)
}

func TestVliCodec(t *testing.T) {
	ctx := setupTest(t)

	f := inccounter.ScFuncs.TestVliCodec(ctx)
	save := wasmhost.DisableWasmTimeout
	wasmhost.DisableWasmTimeout = false
	f.Func.Post()
	wasmhost.DisableWasmTimeout = save
	require.NoError(t, ctx.Err)
}

func TestVluCodec(t *testing.T) {
	ctx := setupTest(t)

	f := inccounter.ScFuncs.TestVluCodec(ctx)
	save := wasmhost.DisableWasmTimeout
	wasmhost.DisableWasmTimeout = false
	f.Func.Post()
	wasmhost.DisableWasmTimeout = save
	require.NoError(t, ctx.Err)
}

func TestVli(t *testing.T) {
	ctx := setupTest(t)

	for i := int64(-200); i < 200; i++ {
		vli := inccounter.ScFuncs.GetVli(ctx)
		vli.Params.Ni64().SetValue(i)
		vli.Func.Call()
		require.NoError(t, ctx.Err)
		fmt.Printf("Bytes: %s\n", vli.Results.Str().Value())
		require.Equal(t, i, vli.Results.Ni64().Value())
		require.Equal(t, i, vli.Results.Xi64().Value())
	}
}

func TestVlu(t *testing.T) {
	ctx := setupTest(t)

	for i := uint64(0); i < 400; i++ {
		vli := inccounter.ScFuncs.GetVlu(ctx)
		vli.Params.Nu64().SetValue(i)
		vli.Func.Call()
		require.NoError(t, ctx.Err)
		fmt.Printf("Bytes: %s\n", vli.Results.Str().Value())
		require.Equal(t, i, vli.Results.Nu64().Value())
		require.Equal(t, i, vli.Results.Xu64().Value())
	}
}

func TestLoop(t *testing.T) {
	ctx := setupTest(t)

	if !ctx.IsWasm || *wasmsolo.UseWasmEdge {
		// no timeout possible because goroutines cannot be killed
		// or because there is no way to interrupt the Wasm code
		t.SkipNow()
	}

	save := wasmhost.DisableWasmTimeout
	wasmhost.DisableWasmTimeout = false
	wasmhost.WasmTimeout = 1 * time.Second
	endlessLoop := inccounter.ScFuncs.EndlessLoop(ctx)
	endlessLoop.Func.Post()
	require.Error(t, ctx.Err)
	require.Contains(t, ctx.Err.Error(), "gas budget exceeded")
	wasmhost.DisableWasmTimeout = save

	inccounter.ScFuncs.Increment(ctx).Func.Post()
	require.NoError(t, ctx.Err)

	checkStateCounter(t, ctx, 1)
}

func checkStateCounter(t *testing.T, ctx *wasmsolo.SoloContext, expected interface{}) {
	getCounter := inccounter.ScFuncs.GetCounter(ctx)
	getCounter.Func.Call()
	require.NoError(t, ctx.Err)
	counter := getCounter.Results.Counter()
	if expected == nil {
		require.False(t, counter.Exists())
		return
	}
	require.True(t, counter.Exists())
	require.EqualValues(t, expected, counter.Value())
}
