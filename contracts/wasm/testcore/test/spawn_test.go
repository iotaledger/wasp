package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcoreimpl"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreroot"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
)

func TestSpawn(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		// first turn off default required deploy permission
		ctxr := ctx.SoloContextForCore(t, coreroot.ScName, coreroot.OnDispatch)
		require.NoError(t, ctxr.Err)
		f := coreroot.ScFuncs.RequireDeployPermissions(ctxr)
		f.Params.DeployPermissionsEnabled().SetValue(false)
		f.Func.Post()
		require.NoError(t, ctxr.Err)

		s := testcore.ScFuncs.Spawn(ctx)
		s.Params.ProgHash().SetValue(ctx.Cvt.ScHash(ctx.Hprog))
		s.Func.Post()
		require.NoError(t, ctx.Err)

		spawnedName := testcore.ScName + "_spawned"
		ctxSpawn := ctx.SoloContextForCore(t, spawnedName, testcoreimpl.OnDispatch)
		require.NoError(t, ctxSpawn.Err)
		v := testcore.ScFuncs.GetCounter(ctxSpawn)
		v.Func.Call()
		require.EqualValues(t, 5, v.Results.Counter().Value())

		_, _, recs := ctx.Chain.GetInfo()
		require.EqualValues(t, len(corecontracts.All)+2, len(recs))
	})
}
