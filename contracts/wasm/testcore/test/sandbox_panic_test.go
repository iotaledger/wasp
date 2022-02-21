package test

import (
	"strings"
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func verifyReceipts(t *testing.T, w bool, ctx *wasmsolo.SoloContext, requests, panics int) {
	recStr := ctx.Chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)
	str := strings.Join(recStr, "\n")
	t.Logf("\n%s", str)
	extra := 0
	if w {
		extra = 1
	}
	require.EqualValues(t, requests+extra, strings.Count(str, "OnLedgerRequestData::"))
	require.EqualValues(t, panics, strings.Count(str, "panic in VM"))
}

func TestPanicFull(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.TestPanicFullEP(ctx)
		f.Func.Post()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), sbtestsc.MsgFullPanic)

		verifyReceipts(t, w, ctx, 4, 1)
	})
}

func TestPanicViewCall(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.TestPanicViewEP(ctx)
		f.Func.Call()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), sbtestsc.MsgViewPanic)

		verifyReceipts(t, w, ctx, 3, 0)
	})
}

func TestCallPanicFull(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.TestCallPanicFullEP(ctx)
		f.Func.Post()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), sbtestsc.MsgFullPanic)

		verifyReceipts(t, w, ctx, 4, 1)
	})
}

func TestCallPanicViewFromFull(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.TestCallPanicViewEPFromFull(ctx)
		f.Func.Post()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), sbtestsc.MsgViewPanic)

		verifyReceipts(t, w, ctx, 4, 1)
	})
}

func TestCallPanicViewFromView(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.TestCallPanicViewEPFromView(ctx)
		f.Func.Call()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), sbtestsc.MsgViewPanic)

		verifyReceipts(t, w, ctx, 3, 0)
	})
}
