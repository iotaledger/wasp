package sbtestsc

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
)

// testSplitFunds calls Send in a loop by sending 200 base tokens back to the caller
func testSplitFunds(ctx isc.Sandbox) {
	addr, ok := isc.AddressFromAgentID(ctx.Caller())
	ctx.Requiref(ok, "caller must have L1 address")
	// claim 1Mi base tokens from allowance at a time
	var baseTokensToTransfer coin.Value = 1 * isc.Million
	for !ctx.AllowanceAvailable().IsEmpty() && ctx.AllowanceAvailable().BaseTokens() >= baseTokensToTransfer {
		// send back to caller's address
		// depending on the amount of base tokens, it will exceed number of outputs or not
		ctx.TransferAllowedFunds(ctx.AccountID(), isc.NewAssets(baseTokensToTransfer))
		ctx.Send(
			isc.RequestParameters{
				TargetAddress: addr,
				Assets:        isc.NewAssets(baseTokensToTransfer),
			},
		)
	}
}

// testSplitFundsNativeTokens calls Send for each Native token
func testSplitFundsNativeTokens(ctx isc.Sandbox) {
	addr, ok := isc.AddressFromAgentID(ctx.Caller())
	ctx.Requiref(ok, "caller must have L1 address")
	// claims all base tokens from allowance
	accountID := ctx.AccountID()
	ctx.TransferAllowedFunds(accountID, isc.NewAssets(ctx.AllowanceAvailable().BaseTokens()))
	for coinType, coinValue := range ctx.AllowanceAvailable().Coins.Iterate() {
		for coinValue > 0 {
			// claim 1 token from allowance at a time
			// send back to caller's address
			// depending on the amount of tokens, it will exceed number of outputs or not
			assets := isc.NewEmptyAssets().AddCoin(coinType, 1)
			rem := ctx.TransferAllowedFunds(accountID, assets)
			fmt.Printf("%s\n", rem)
			ctx.Send(
				isc.RequestParameters{
					TargetAddress: addr,
					Assets:        assets,
				},
			)
		}
	}
}

func pingAllowanceBack(ctx isc.Sandbox) {
	caller := ctx.Caller()
	addr, ok := isc.AddressFromAgentID(caller)
	// assert caller is L1 address, not a SC
	ctx.Requiref(ok,
		"pingAllowanceBack: caller expected to be a L1 address")
	// save allowance budget because after transfer it will be modified
	toSend := ctx.AllowanceAvailable()
	if toSend.IsEmpty() {
		// nothing to send back, NOP
		return
	}
	// claim all transfer to the current account
	left := ctx.TransferAllowedFunds(ctx.AccountID())
	// assert what has left is empty. Only for testing
	ctx.Requiref(left.IsEmpty(), "pingAllowanceBack: inconsistency")

	// send the funds to the caller L1 address on-ledger
	ctx.Send(
		isc.RequestParameters{
			TargetAddress: addr,
			Assets:        toSend,
		},
	)
}

// tries to sendback whatever objectss are specified in allowance
func sendObjectsBack(ctx isc.Sandbox) {
	addr, ok := isc.AddressFromAgentID(ctx.Caller())
	ctx.Requiref(ok, "caller must have L1 address")

	allowance := ctx.AllowanceAvailable()
	ctx.TransferAllowedFunds(ctx.AccountID())
	for obj := range allowance.Objects.Iterate() {
		ctx.Send(isc.RequestParameters{
			TargetAddress: addr,
			Assets:        isc.NewEmptyAssets().AddObject(obj),
		})
	}
}

// just claims everything from allowance and does nothing with it
func claimAllowance(ctx isc.Sandbox) {
	initialObjects := ctx.OwnedObjects()
	allowance := ctx.AllowanceAvailable()
	ctx.TransferAllowedFunds(ctx.AccountID())
	ctx.Requiref(len(ctx.OwnedObjects())-len(initialObjects) == allowance.Objects.Size(), "must get all objects from allowance")
}
