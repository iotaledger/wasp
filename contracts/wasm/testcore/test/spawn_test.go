package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/stretchr/testify/require"
)

func TestSpawn(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.Spawn(ctx)
		f.Params.ProgHash().SetValue(ctx.Convertor.ScHash(ctx.Hprog))
		f.Func.TransferIotas(1).Post()
		require.NoError(t, ctx.Err)

		ctxSpawn := ctx.SoloContextForCore(t, testcore.ScName+"_spawned", testcore.OnLoad)
		require.NoError(t, ctxSpawn.Err)
		v := testcore.ScFuncs.GetCounter(ctxSpawn)
		v.Func.Call()
		require.EqualValues(t, 5, v.Results.Counter().Value())

		_, _, recs := ctx.Chain.GetInfo()
		require.EqualValues(t, len(core.AllCoreContractsByHash)+2, len(recs))
	})
}
