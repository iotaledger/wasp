package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/stretchr/testify/require"
)

func TestBasicBlockContext1(t *testing.T) {
	ctx := deployTestCore(t, false)

	f := testcore.ScFuncs.TestBlockContext1(ctx)
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestBasicBlockContext2(t *testing.T) {
	ctx := deployTestCore(t, false)

	f := testcore.ScFuncs.TestBlockContext2(ctx)
	f.Func.Post()
	require.NoError(t, ctx.Err)

	v := testcore.ScFuncs.GetStringValue(ctx)
	v.Params.VarName().SetValue("atTheEndKey")
	v.Func.Call()
	require.NoError(t, ctx.Err)
	value := v.Results.Vars().GetString("atTheEndKey")
	require.True(t, value.Exists())
	require.EqualValues(t, "atTheEndValue", value.Value())
}
