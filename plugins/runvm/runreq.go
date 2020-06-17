package runvm

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/builtin"
)

// runTheRequest is a wrapper which
// - check and validates the context
// - checks authorisations for protected requests
// - checks if requests are protected
// - handles request token
// - processes reward logic
// - redirects reserved request codes (is supported) to hardcoded processing
// - redirects not reserved codes (is supported) to SC VM
// - in case of something not correct:
//  --- sends all funds transferred by the request transaction to the SC address less reward back
//  --- the whole operation is NOP wrt the SC state.
func runTheRequest(ctx *vm.VMContext) {
	mustHandleRequestToken(ctx)

	if !handleRewards(ctx) {
		return
	}

	reqBlock := ctx.Request.RequestBlock()
	if reqBlock.RequestCode().IsProtected() && !ctx.Request.IsAuthorised(&ctx.OwnerAddress) {
		// if protected call is not authorised by the containing transaction, do nothing
		// the result will be taking all iotas and no effect
		// Maybe it is nice to return back all iotas exceeding reward and minimum ??? TODO
		return
	}

	if reqBlock.RequestCode().IsReserved() {
		entryPoint, ok := builtin.Processor.GetEntryPoint(reqBlock.RequestCode())
		if !ok {
			return
		}
		entryPoint.Run(ctx)
		return
	}

	// here reqBlock.RequestCode().IsUserDefined()

	processor, err := getProcessor(ctx.ProgramHash.String())
	if err != nil {
		// it should not come to this point if processor is not ready
		ctx.Log.Panicf("major inconsistency: %v", err)
	}
	entryPoint, ok := processor.GetEntryPoint(reqBlock.RequestCode())
	if !ok {
		return
	}
	entryPoint.Run(ctx)
}

func mustHandleRequestToken(ctx *vm.VMContext) {
	// destroy token corresponding to request
	// NOTE: it is assumed here that balances contain all necessary request token balances
	// it is checked in the dispatcher.dispatchAddressUpdate
	err := ctx.TxBuilder.EraseColor(ctx.Address, (balance.Color)(ctx.Request.Tx.ID()), 1)
	if err != nil {
		// not enough balance for requests tokens
		// major inconsistency, it must had been checked before
		ctx.Log.Panicf("something wrong with request token for reqid = %s. Not all requests were processed: %v",
			ctx.Request.RequestId().String(), err)
	}
}

// handleRewards return true if to continue with request processing
func handleRewards(ctx *vm.VMContext) bool {
	if ctx.RewardAddress[0] == 0 {
		// first byte is never 0 for the correct address
		return true
	}
	if ctx.MinimumReward <= 0 {
		return true
	}

	totalIotaOutput := sctransaction.BalanceOfOutputToColor(ctx.Request.Tx, ctx.Address, balance.ColorIOTA)
	var err error

	var proceed bool
	if totalIotaOutput >= ctx.MinimumReward {
		err = ctx.TxBuilder.MoveTokens(ctx.RewardAddress, balance.ColorIOTA, ctx.MinimumReward)
		proceed = true
	} else {
		// if reward is not enough, the state update will be empty, i.e. NOP (but the fee will be taken)
		err = ctx.TxBuilder.MoveTokens(ctx.RewardAddress, balance.ColorIOTA, totalIotaOutput)
		proceed = false
	}
	if err != nil {
		ctx.Log.Panicf("can't move reward tokens: %v", err)
	}
	return proceed
}
