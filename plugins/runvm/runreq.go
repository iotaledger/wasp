package runvm

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/builtin"
)

// runTheRequest:
// - handles request token
// - processes reward logic
// - checks authorisations for protected requests
// - redirects reserved request codes (is supported) to hardcoded processing
// - redirects not reserved codes (is supported) to SC VM
// - in case of something not correct the whole operation is NOP, however
//   all the sent fees and other funds remains in the SC address (this may change).
func runTheRequest(ctx *vm.VMContext) {
	mustHandleRequestToken(ctx)

	if !handleRewards(ctx) {
		return
	}

	reqBlock := ctx.RequestRef.RequestBlock()
	if reqBlock.RequestCode().IsProtected() && !ctx.RequestRef.IsAuthorised(&ctx.OwnerAddress) {
		// if protected call is not authorised by the containing transaction, do nothing
		// the result will be taking all iotas and no effect
		// Maybe it is nice to return back all iotas exceeding minimum reward ??? TODO
		return
	}

	if reqBlock.RequestCode().IsReserved() {
		// processing of built-in requests
		entryPoint, ok := builtin.Processor.GetEntryPoint(reqBlock.RequestCode())
		if !ok {
			return
		}
		entryPoint.Run(vm.NewSandbox(ctx))
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
	entryPoint.Run(vm.NewSandbox(ctx))
}

func mustHandleRequestToken(ctx *vm.VMContext) {
	// destroy token corresponding to request
	// NOTE: it is assumed here that balances contain all necessary request token balances
	// it is checked in the dispatcher.dispatchAddressUpdate
	err := ctx.TxBuilder.EraseColor(ctx.Address, (balance.Color)(ctx.RequestRef.Tx.ID()), 1)
	if err != nil {
		// not enough balance for requests tokens
		// major inconsistency, it must had been checked before
		ctx.Log.Panicf("something wrong with request token for reqid = %s. Not all requests were processed: %v",
			ctx.RequestRef.RequestId().String(), err)
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

	totalIotaOutput := sctransaction.BalanceOfOutputToColor(ctx.RequestRef.Tx, ctx.Address, balance.ColorIOTA)
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
