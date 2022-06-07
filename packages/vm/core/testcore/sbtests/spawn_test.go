package sbtests

import (
	"testing"

	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"
	"github.com/stretchr/testify/require"
)

func TestSpawn(t *testing.T) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, false)

	ch.MustDepositIotasToL2(10_000, nil)

	req := solo.NewCallParams(ScName, sbtestsc.FuncSpawn.Name,
		sbtestsc.ParamProgHash, sbtestsc.Contract.ProgramHash).
		WithGasBudget(100_000)
	_, err := ch.PostRequestSync(req, nil)
	require.NoError(t, err)

	ret, err := ch.CallView(ScName+"_spawned", sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)
	res := kvdecoder.New(ret, ch.Log())
	counter := res.MustGetUint64(sbtestsc.VarCounter)
	require.EqualValues(t, 5, counter)

	_, _, recs := ch.GetInfo()
	require.EqualValues(t, len(corecontracts.All)+2, len(recs))
}
