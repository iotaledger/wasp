package sbtests

import (
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"
	"github.com/stretchr/testify/require"
)

func TestBasicBlockContext1(t *testing.T) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, false)

	ch.MustDepositIotasToL2(10_000, nil)

	req := solo.NewCallParams(ScName, sbtestsc.FuncTestBlockContext1.Name).
		WithGasBudget(100_000)
	_, err := ch.PostRequestSync(req, nil)
	require.NoError(t, err)
}

func TestBasicBlockContext2(t *testing.T) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, false)

	ch.MustDepositIotasToL2(10_000, nil)

	req := solo.NewCallParams(ScName, sbtestsc.FuncTestBlockContext2.Name).
		WithGasBudget(100_000)
	_, err := ch.PostRequestSync(req, nil)
	require.NoError(t, err)

	res, err := ch.CallView(ScName, sbtestsc.FuncGetStringValue.Name,
		sbtestsc.ParamVarName, "atTheEndKey")
	require.NoError(t, err)
	b, err := res.Get("atTheEndKey")
	require.NoError(t, err)
	require.EqualValues(t, "atTheEndValue", string(b))
}
