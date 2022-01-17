package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// testSplitFunds calls Send in a loop by sending 200 iotas back to the caller
func testSplitFunds(ctx iscp.Sandbox) (dict.Dict, error) {
	for !ctx.AllowanceAvailable().IsEmpty() && ctx.AllowanceAvailable().Iotas >= 200 {
		// claim 200 iotas from allowance at a time
		// send back to caller's address
		// depending on the amount of iotas, it will exceed number of outputs or not
		ctx.TransferAllowedFunds(ctx.AccountID(), iscp.NewAssetsIotas(200))
		ctx.Send(
			iscp.RequestParameters{
				TargetAddress: ctx.Caller().Address(),
				Assets:        iscp.NewAssetsIotas(200),
			},
		)
	}
	return nil, nil
}

func pingAllowanceBack(ctx iscp.Sandbox) (dict.Dict, error) {
	// assert caller is L1 address, not a SC
	ctx.Requiref(!ctx.Caller().Address().Equal(ctx.ChainID().AsAddress()) && ctx.Caller().Hname() == 0,
		"pingAllowanceBack: caller expected to be a L1 address")
	// save allowance budget because after transfer it will be modified
	toSend := ctx.AllowanceAvailable()
	if toSend.IsEmpty() {
		// nothing to send back, NOP
		return nil, nil
	}
	// claim all transfer to the current account
	left := ctx.TransferAllowedFunds(ctx.AccountID())
	// assert what has left is empty. Only for testing
	ctx.Requiref(left.IsEmpty(), "pingAllowanceBack: inconsistency")

	// send the funds to the caller L1 address on-ledger
	ctx.Send(
		iscp.RequestParameters{
			TargetAddress: ctx.Caller().Address(),
			Assets:        toSend,
		},
	)
	return nil, nil
}
