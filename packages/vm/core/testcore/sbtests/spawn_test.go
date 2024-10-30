package sbtests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestSpawn(t *testing.T) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil)

	ch.MustDepositBaseTokensToL2(10_000, nil)

	err := sbtestsc.FuncSpawn.Call(sbtestsc.Contract.ProgramHash, func(msg isc.Message) (isc.CallArguments, error) {
		req := solo.NewCallParams(msg, ScName).
			WithGasBudget(100_000)
		return ch.PostRequestSync(req, nil)
	})

	require.NoError(t, err)

	ret, err := ch.CallViewEx(ScName+"_spawned", sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)
	counterResult, err := sbtestsc.FuncGetCounter.DecodeOutput(ret)
	require.NoError(t, err)

	require.EqualValues(t, 5, counterResult)

	_, _, recs := ch.GetInfo()
	require.EqualValues(t, len(corecontracts.All)+2, len(recs))
}
