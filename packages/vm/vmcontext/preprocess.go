package vmcontext

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
)

// mustHandleRequestToken handles the request token
// it will panic on inconsistency because consistency of the request token must be checked well before
func (vmctx *VMContext) mustHandleRequestToken() {
	reqColor := balance.Color(vmctx.reqRef.Tx.ID())
	if vmctx.txBuilder.Balance(reqColor) == 0 {
		// must be checked before, while validating transaction
		vmctx.log.Panicf("mustHandleRequestToken: request token not found: %s", reqColor.String())
	}
	if !vmctx.txBuilder.Erase1TokenToChain(reqColor) {
		vmctx.log.Panicf("mustHandleRequestToken: can't erase request token: %s", reqColor.String())
	}
	// always accrue 1 uncolored iota to the sender on-chain. This makes completely fee-less requests possible
	vmctx.creditToAccount(vmctx.reqRef.SenderAgentID(), cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 1,
	}))
	vmctx.remainingAfterFees = vmctx.reqRef.RequestSection().Transfer()
	vmctx.log.Debugf("mustHandleFees: 1 request token accrued to the sender: %s\n", vmctx.reqRef.SenderAgentID())
}

// mustHandleFees:
// - handles request token
// - handles node fee, including fallback if not enough
func (vmctx *VMContext) mustHandleFees() {
	transfer := vmctx.reqRef.RequestSection().Transfer()
	totalFee := vmctx.ownerFee + vmctx.validatorFee
	if totalFee == 0 || vmctx.requesterIsChainOwner() {
		// no fees enabled or the caller is the chain owner
		vmctx.log.Debugf("mustHandleFees: no fees charged\n")
		vmctx.remainingAfterFees = transfer
		return
	}
	// handle fees
	if transfer.Balance(vmctx.feeColor) < totalFee {
		// TODO more sophisticated policy, for example taking fees to chain owner, the rest returned to sender
		// fallback: not enough fees. Accrue everything to the sender
		sender := vmctx.reqRef.SenderAgentID()
		vmctx.creditToAccount(sender, transfer)
		vmctx.lastError = fmt.Errorf("mustHandleFees: not enough fees for request %s. Transfer accrued to %s",
			vmctx.reqRef.RequestID().Short(), sender.String())
		vmctx.remainingAfterFees = cbalances.NewFromMap(nil)
		return
	}
	// enough fees. Split between owner and validator
	if vmctx.ownerFee > 0 {
		vmctx.creditToAccount(vmctx.ChainOwnerID(), cbalances.NewFromMap(map[balance.Color]int64{
			vmctx.feeColor: vmctx.ownerFee,
		}))
	}
	if vmctx.validatorFee > 0 {
		vmctx.creditToAccount(vmctx.validatorFeeTarget, cbalances.NewFromMap(map[balance.Color]int64{
			vmctx.feeColor: vmctx.validatorFee,
		}))
	}
	// subtract fees from the transfer
	remaining := map[balance.Color]int64{
		vmctx.feeColor: -totalFee,
	}
	transfer.AddToMap(remaining)
	vmctx.remainingAfterFees = cbalances.NewFromMap(remaining)
}
