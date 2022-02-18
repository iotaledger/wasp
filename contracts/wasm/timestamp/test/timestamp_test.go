package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/timestamp/go/timestamp"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, timestamp.ScName, timestamp.OnLoad)
	require.NoError(t, ctx.ContractExists(timestamp.ScName))
}

func TestStamp(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, timestamp.ScName, timestamp.OnLoad)

	v := timestamp.ScFuncs.GetTimestamp(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	t1 := v.Results.Timestamp().Value()

	v = timestamp.ScFuncs.GetTimestamp(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, t1, v.Results.Timestamp().Value())

	f := timestamp.ScFuncs.Now(ctx)
	f.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	v = timestamp.ScFuncs.GetTimestamp(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	t2 := v.Results.Timestamp().Value()
	require.Greater(t, t2, t1)

	v = timestamp.ScFuncs.GetTimestamp(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, t2, v.Results.Timestamp().Value())

	f = timestamp.ScFuncs.Now(ctx)
	f.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	v = timestamp.ScFuncs.GetTimestamp(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	t3 := v.Results.Timestamp().Value()
	require.Greater(t, t3, t2)

	v = timestamp.ScFuncs.GetTimestamp(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, t3, v.Results.Timestamp().Value())
}
