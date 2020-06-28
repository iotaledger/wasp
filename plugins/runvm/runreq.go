package runvm

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/builtin"
	"github.com/iotaledger/wasp/packages/vm/processor"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
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
	ctx.Log.Debugw("runTheRequest IN",
		"reqId", ctx.RequestRef.RequestId().Short(),
		"programHash", ctx.ProgramHash.String(),
		"code", ctx.RequestRef.RequestBlock().RequestCode().String(),
	)
	defer ctx.Log.Debugw("runTheRequest OUT",
		"reqId", ctx.RequestRef.RequestId().Short(),
		"programHash", ctx.ProgramHash.String(),
		"code", ctx.RequestRef.RequestBlock().RequestCode().String(),
		"state update", ctx.StateUpdate.String(),
	)

	mustHandleRequestToken(ctx)

	if !handleRewards(ctx) {
		return
	}

	reqBlock := ctx.RequestRef.RequestBlock()
	if reqBlock.RequestCode().IsProtected() {
		// check authorisation
		if !ctx.RequestRef.IsAuthorised(&ctx.OwnerAddress) {
			// if protected call is not authorised by the containing transaction, do nothing
			// the result will be taking all iotas and no effect on state
			// Maybe it is nice to return back all iotas exceeding minimum reward ??? TODO

			ctx.Log.Warnf("protected request %s (code %s) is not authorised by %s",
				ctx.RequestRef.RequestId().String(), reqBlock.RequestCode(), ctx.OwnerAddress.String(),
			)
			ctx.Log.Debugw("protected request is not authorised",
				"req", ctx.RequestRef.RequestId().String(),
				"code", reqBlock.RequestCode(),
				"owner", ctx.OwnerAddress.String(),
				"inputs", util.InputsToStringByAddress(ctx.RequestRef.Tx.Inputs()),
			)
			return
		}
		if ctx.VirtualState.StateIndex() > 0 && !ctx.VirtualState.InitiatedBy(&ctx.OwnerAddress) {
			// for states after #0 it is required to have record about initiator's address in the solid state
			// to prevent attack when owner (initiator) address is overwritten in the quorum of bootup records
			// TODO protection must also be set at the lowest level of the solid state. i.e. some metadata that variable
			// is protected by some address and authorisation with that address is needed to modify the value

			ctx.Log.Errorf("inconsistent state: variable '%s' != owner record from bootup record '%s'",
				vmconst.VarNameOwnerAddress, ctx.OwnerAddress.String())

			return
		}
	}
	// authorisation check passed

	if reqBlock.RequestCode().IsReserved() {
		// finding and running builtin entry point
		entryPoint, ok := builtin.Processor.GetEntryPoint(reqBlock.RequestCode())
		if !ok {
			ctx.Log.Warnf("can't find entry point for request code %s in the builtin processor", reqBlock.RequestCode())
			return
		}
		entryPoint.Run(sandbox.NewSandbox(ctx))
		return
	}

	// request requires user-defined program on VM
	proc, err := processor.Acquire(ctx.ProgramHash.String())
	if err != nil {
		ctx.Log.Warn(err)
		return
	}
	defer processor.Release(ctx.ProgramHash.String())

	entryPoint, ok := proc.GetEntryPoint(reqBlock.RequestCode())
	if !ok {
		ctx.Log.Warnf("can't find entry point for request code %s in the user-defined processor prog hash: %s",
			reqBlock.RequestCode(), ctx.ProgramHash.String())
		return
	}
	entryPoint.Run(sandbox.NewSandbox(ctx))
}

func mustHandleRequestToken(ctx *vm.VMContext) {
	// destroy token corresponding to request
	// NOTE: it is assumed here that balances contain all necessary request token balances
	// it is checked in the dispatcher.dispatchAddressUpdate
	err := ctx.TxBuilder.EraseColor(ctx.Address, (balance.Color)(ctx.RequestRef.Tx.ID()), 1)
	if err != nil {
		//
		ctx.Log.Errorf("dump balances:\n%s\n", ctx.TxBuilder.Dump())
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
	if ctx.RequestRef.IsAuthorised(&ctx.Address) {
		// no need for rewards from itself
		return true
	}

	totalIotaOutput := sctransaction.OutputValueOfColor(ctx.RequestRef.Tx, ctx.Address, balance.ColorIOTA)
	var err error

	var proceed bool
	if totalIotaOutput >= ctx.MinimumReward {
		err = ctx.TxBuilder.MoveToAddress(ctx.RewardAddress, balance.ColorIOTA, ctx.MinimumReward)
		proceed = true
	} else {
		// if reward is not enough, the state update will be empty, i.e. NOP (but the fee will be taken)
		err = ctx.TxBuilder.MoveToAddress(ctx.RewardAddress, balance.ColorIOTA, totalIotaOutput)
		proceed = false
	}
	if err != nil {
		ctx.Log.Panicf("can't move reward tokens: %v", err)
	}
	return proceed
}
