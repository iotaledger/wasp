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
		f.Func.Post()
		require.NoError(t, ctx.Err)

		spawnedName := testcore.ScName + "_spawned"
		ctxSpawn := ctx.SoloContextForCore(t, spawnedName, testcore.OnLoad)
		require.NoError(t, ctxSpawn.Err)
		v := testcore.ScFuncs.GetCounter(ctxSpawn)
		v.Func.Call()
		require.EqualValues(t, 5, v.Results.Counter().Value())

		_, _, recs := ctx.Chain.GetInfo()
		require.EqualValues(t, len(corecontracts.All)+2, len(recs))
	})
}
