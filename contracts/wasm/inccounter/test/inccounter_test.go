// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/wasp/contracts/wasm/inccounter/go/inccounter"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/iotaledger/wasp/packages/vm/wasmsolo"
	"github.com/stretchr/testify/require"
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
	increment.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	checkStateCounter(t, ctx, 1)
}

func TestIncrementTwice(t *testing.T) {
	ctx := setupTest(t)

	increment := inccounter.ScFuncs.Increment(ctx)
	increment.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	increment = inccounter.ScFuncs.Increment(ctx)
	increment.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	checkStateCounter(t, ctx, 2)
}

func TestIncrementRepeatThrice(t *testing.T) {
	ctx := setupTest(t)

	repeatMany := inccounter.ScFuncs.RepeatMany(ctx)
	repeatMany.Params.NumRepeats().SetValue(3)
	repeatMany.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	require.True(t, ctx.WaitForPendingRequests(3))

	checkStateCounter(t, ctx, 4)
}

func TestIncrementCallIncrement(t *testing.T) {
	ctx := setupTest(t)

	callIncrement := inccounter.ScFuncs.CallIncrement(ctx)
	callIncrement.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	checkStateCounter(t, ctx, 2)
}

func TestIncrementCallIncrementRecurse5x(t *testing.T) {
	ctx := setupTest(t)

	callIncrementRecurse5x := inccounter.ScFuncs.CallIncrementRecurse5x(ctx)
	callIncrementRecurse5x.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	checkStateCounter(t, ctx, 6)
}

func TestIncrementPostIncrement(t *testing.T) {
	ctx := setupTest(t)

	postIncrement := inccounter.ScFuncs.PostIncrement(ctx)
	postIncrement.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	require.True(t, ctx.WaitForPendingRequests(1))

	checkStateCounter(t, ctx, 2)
}

func TestIncrementLocalStateInternalCall(t *testing.T) {
	ctx := setupTest(t)

	localStateInternalCall := inccounter.ScFuncs.LocalStateInternalCall(ctx)
	localStateInternalCall.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	checkStateCounter(t, ctx, 2)
}

func TestIncrementLocalStateSandboxCall(t *testing.T) {
	ctx := setupTest(t)

	localStateSandboxCall := inccounter.ScFuncs.LocalStateSandboxCall(ctx)
	localStateSandboxCall.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	if *wasmsolo.GoDebug {
		// when using WasmGoVM the 3 posts are run only after
		// the LocalStateMustIncrement has been set to true
		checkStateCounter(t, ctx, 2)
		return
	}

	// global var in wasm execution has no effect
	checkStateCounter(t, ctx, nil)
}

func TestIncrementLocalStatePost(t *testing.T) {
	ctx := setupTest(t)

	localStatePost := inccounter.ScFuncs.LocalStatePost(ctx)
	localStatePost.Func.TransferIotas(3).Post()
	require.NoError(t, ctx.Err)

	require.True(t, ctx.WaitForPendingRequests(3))

	if *wasmsolo.GoDebug {
		// when using WasmGoVM the 3 posts are run only after
		// the LocalStateMustIncrement has been set to true
		checkStateCounter(t, ctx, 3)
		return
	}

	// global var in wasm execution has no effect
	checkStateCounter(t, ctx, nil)
}

func TestLeb128(t *testing.T) {
	wasmhost.DisableWasmTimeout = true
	ctx := setupTest(t)
	wasmhost.DisableWasmTimeout = false

	testLeb128 := inccounter.ScFuncs.TestLeb128(ctx)
	testLeb128.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)
}

func TestVli(t *testing.T) {
	*wasmsolo.TsWasm = true
	wasmhost.DisableWasmTimeout = true
	ctx := setupTest(t)
	wasmhost.DisableWasmTimeout = false

	for i := int64(0); i < 200; i++ {
		vli := inccounter.ScFuncs.GetVli(ctx)
		vli.Params.N().SetValue(i)
		vli.Func.Call()
		require.NoError(t, ctx.Err)
		fmt.Printf("Bytes: %s\n", vli.Results.Str().Value())
		require.Equal(t, i, vli.Results.N().Value())
		require.Equal(t, i, vli.Results.X().Value())
	}
}

func TestLoop(t *testing.T) {
	if *wasmsolo.GoDebug || *wasmsolo.GoWasmEdge {
		// no timeout possible because goroutines cannot be killed
		// or because there is no way to interrupt the Wasm code
		t.SkipNow()
	}

	ctx := setupTest(t)

	save := wasmhost.DisableWasmTimeout
	wasmhost.DisableWasmTimeout = false
	wasmhost.WasmTimeout = 1 * time.Second
	endlessLoop := inccounter.ScFuncs.EndlessLoop(ctx)
	endlessLoop.Func.TransferIotas(1).Post()
	require.Error(t, ctx.Err)
	require.Contains(t, ctx.Err.Error(), "interrupt")
	wasmhost.DisableWasmTimeout = save

	inccounter.ScFuncs.Increment(ctx).Func.TransferIotas(1).Post()
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
