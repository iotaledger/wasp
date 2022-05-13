package test

import (
	"strings"
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func verifyErrorInReceipts(t *testing.T, ctx *wasmsolo.SoloContext, msg string) {
	receipts := ctx.Chain.GetRequestReceiptsForBlockRange(0, 0)
	errorCount := 0
	for _, a := range receipts {
		receiptError := ctx.Chain.ResolveVMError(a.Error)
		if receiptError != nil {
			errorCount++
			if msg != "" {
				require.Contains(t, receiptError.Error(), msg)
			}
		}
	}
	expectedCount := 0
	if msg != "" {
		expectedCount = 1
	}
	require.EqualValues(t, expectedCount, errorCount)

	if wasmsolo.SoloDebug {
		recStr := ctx.Chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)
		str := strings.Join(recStr, "\n")
		t.Logf("\n%s", str)
	}
}

func TestPanicFull(t *testing.T) {
	t.SkipNow()
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.TestPanicFullEP(ctx)
		f.Func.Post()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), sbtestsc.MsgFullPanic)

		verifyErrorInReceipts(t, ctx, sbtestsc.MsgFullPanic)
	})
}

func TestPanicViewCall(t *testing.T) {
	t.SkipNow()
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.TestPanicViewEP(ctx)
		f.Func.Call()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), sbtestsc.MsgViewPanic)

		verifyErrorInReceipts(t, ctx, "")
	})
}

func TestCallPanicFull(t *testing.T) {
	t.SkipNow()
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.TestCallPanicFullEP(ctx)
		f.Func.Post()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), sbtestsc.MsgFullPanic)

		verifyErrorInReceipts(t, ctx, sbtestsc.MsgFullPanic)
	})
}

func TestCallPanicViewFromFull(t *testing.T) {
	t.SkipNow()
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.TestCallPanicViewEPFromFull(ctx)
		f.Func.Post()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), sbtestsc.MsgViewPanic)

		verifyErrorInReceipts(t, ctx, sbtestsc.MsgViewPanic)
	})
}

func TestCallPanicViewFromView(t *testing.T) {
	t.SkipNow()
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.TestCallPanicViewEPFromView(ctx)
		f.Func.Call()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), sbtestsc.MsgViewPanic)

		verifyErrorInReceipts(t, ctx, "")
	})
}
