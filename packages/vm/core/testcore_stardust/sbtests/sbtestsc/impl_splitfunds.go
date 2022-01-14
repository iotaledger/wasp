package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// testEventLogGenericData is called several times in log_test.go
func testSplitFunds(ctx iscp.Sandbox) (dict.Dict, error) {
	iotas := ctx.AllowanceAvailable().Iotas
	for i := uint64(0); i < iotas; i += 200 {
		ctx.TransferAllowedFunds(ctx.AccountID(), iscp.NewAssets(200, nil)) // claim 200 iotas from allowance at a time
		ctx.Send(
			iscp.RequestParameters{
				TargetAddress: ctx.Caller().Address(),
				Assets:        iscp.NewAssets(200, nil),
			},
		)
	}
	return nil, nil
}
