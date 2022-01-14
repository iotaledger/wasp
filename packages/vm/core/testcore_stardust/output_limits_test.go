package testcore

import (
	"errors"
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
	"github.com/stretchr/testify/require"
)

const paramShouldEmitEvent = "e"

var (
	manyOutputsContract = coreutil.NewContract("Many outputs", "")

	funcSplitFunds = coreutil.Func("splitFunds")

	manyOutputsProcessor = manyOutputsContract.Processor(nil,
		// splits incoming iotas into many outputs (chunks of 200)
		funcSplitFunds.WithHandler(func(ctx iscp.Sandbox) (dict.Dict, error) {
			par := kvdecoder.New(ctx.Params())
			shouldEmitEvent := par.MustGetInt16(paramShouldEmitEvent, 0) == 1
			iotas := ctx.AllowanceAvailable().Iotas
			for i := uint64(0); i < iotas; i += 200 {
				ctx.TransferAllowedFunds(ctx.AccountID(), iscp.NewAssets(200, nil)) // claim 200 iotas from allowance at a time
				ctx.Send(
					iscp.RequestParameters{
						TargetAddress: ctx.Caller().Address(),
						Assets:        iscp.NewAssets(200, nil),
					},
				)
				if shouldEmitEvent {
					ctx.Event(fmt.Sprintf("reqid: %s, sent funds", ctx.Request().ID()))
				}
			}
			return nil, nil
		}),
	)
)

func TestTooManyOutputsInASingleCall(t *testing.T) {
	env := solo.New(t).WithNativeContract(manyOutputsProcessor)
	ch := env.NewChain(nil, "chain1")
	err := ch.DeployContract(nil, manyOutputsContract.Name, manyOutputsContract.ProgramHash)
	require.NoError(t, err)

	// send 1 tx will 1_000_000 iotas which should result in too mant outputs, so the request must fail
	wallet, _ := env.NewKeyPairWithFunds(env.NewSeedFromIndex(1))

	req := solo.NewCallParams(manyOutputsContract.Name, funcSplitFunds.Name).
		AddAssetsIotas(1_000_000).
		AddAllowance(iscp.NewAssets(40_000, nil)). // 40k iotas = 200 ouputs
		WithGasBudget(1_000_000)
	_, err = ch.PostRequestSync(req, wallet)
	require.Error(t, err)
	require.True(t, errors.Is(err, vmtxbuilder.ErrOutputLimitInSingleCallExceeded))
	require.NotContains(t, err.Error(), "skipped")
}
