package sbtests

import (
	"errors"
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
	"github.com/stretchr/testify/require"
)

func TestTooManyOutputsInASingleCall(t *testing.T) {
	env := solo.New(t).WithNativeContract(sbtestsc.Processor)
	ch := env.NewChain(nil, "chain1")
	err := ch.DeployContract(nil, sbtestsc.Contract.Name, sbtestsc.Contract.ProgramHash)
	require.NoError(t, err)

	// send 1 tx will 1_000_000 iotas which should result in too mant outputs, so the request must fail
	wallet, address := env.NewKeyPairWithFunds(env.NewSeedFromIndex(1))
	env.GetFundsFromFaucet(address, 10_000_000)

	req := solo.NewCallParams(sbtestsc.Contract.Name, sbtestsc.FuncSplitFunds.Name).
		AddAssetsIotas(10_000_000).
		AddAllowance(iscp.NewAssets(40_000, nil)). // 40k iotas = 200 ouputs
		WithGasBudget(10_000_000)
	_, err = ch.PostRequestSync(req, wallet)
	require.Error(t, err)
	require.True(t, errors.Is(err, vmtxbuilder.ErrOutputLimitInSingleCallExceeded))
	require.NotContains(t, err.Error(), "skipped")
}
