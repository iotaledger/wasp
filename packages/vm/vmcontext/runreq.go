package vmcontext

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// runTheRequest:
// - handles request token
// - processes reward logic
func (vmctx *VMContext) RunTheRequest(req *sctransaction.Request, timestamp int64) {
	vmctx.initRequestContext(req, timestamp)
	vmctx.mustHandleRequestToken()

	if !vmctx.isInitChainRequest() {
		vmctx.mustGetBaseValues()
		vmctx.mustHandleFees()
	}
	vmctx.mustHandleFreeTokens()
	defer vmctx.finalizeRequestCall()

	if vmctx.contractRecord == nil {
		// sc does not exist, stop here
		vmctx.lastResult = nil
		vmctx.lastError = fmt.Errorf("smart contract '%s' does not exist", vmctx.reqHname)
		return
	}
	// snapshot state baseline for rollback in case of panic
	snapshotTxBuilder := vmctx.txBuilder.Clone()
	snapshotStateUpdate := vmctx.stateUpdate.Clone()

	vmctx.lastError = nil
	func() {
		// panic catcher for the whole call from request to the VM
		defer func() {
			if r := recover(); r != nil {
				vmctx.lastResult = nil
				vmctx.lastError = fmt.Errorf("recovered from panic in VM: %v", r)
				if dberr, ok := r.(buffered.DBError); ok {
					// There was an error accessing the DB
					// The world stops
					vmctx.Panicf("DB error: %v", dberr)
				}
			}
		}()
		vmctx.mustCallFromRequest()
	}()

	if vmctx.lastError != nil {
		// treating panic and error returned from request the same way
		vmctx.txBuilder = snapshotTxBuilder
		vmctx.stateUpdate = snapshotStateUpdate

		vmctx.mustHandleFallback()
	}
}

// mustHandleRequestToken handles the request token
// it will panic on inconsistency because consistency of the request token must be checked well before
func (vmctx *VMContext) mustHandleRequestToken() {
	reqColor := balance.Color(vmctx.req.Tx.ID())
	if vmctx.txBuilder.Balance(reqColor) == 0 {
		// must be checked before, while validating transaction
		vmctx.log.Panicf("mustHandleRequestToken: request token not found: %s", reqColor.String())
	}
	if !vmctx.txBuilder.Erase1TokenToChain(reqColor) {
		vmctx.log.Panicf("mustHandleRequestToken: can't erase request token: %s", reqColor.String())
	}
	// always accrue 1 uncolored iota to the sender on-chain. This makes completely fee-less requests possible
	vmctx.creditToAccount(vmctx.req.SenderAgentID(), coretypes.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 1,
	}))
	vmctx.remainingAfterFees = vmctx.req.RequestSection().Transfer()
	vmctx.log.Debugf("mustHandleFees: 1 request token accrued to the sender: %s\n", vmctx.req.SenderAgentID())
}

// mustHandleFees:
// - handles request token
// - handles node fee, including fallback if not enough
func (vmctx *VMContext) mustHandleFees() {
	transfer := vmctx.req.RequestSection().Transfer()
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
		sender := vmctx.req.SenderAgentID()
		vmctx.creditToAccount(sender, transfer)
		vmctx.lastError = fmt.Errorf("mustHandleFees: not enough fees for request %s. Transfer accrued to %s",
			vmctx.req.RequestID().Short(), sender.String())
		vmctx.remainingAfterFees = coretypes.NewFromMap(nil)
		return
	}
	// enough fees. Split between owner and validator
	if vmctx.ownerFee > 0 {
		vmctx.creditToAccount(vmctx.ChainOwnerID(), coretypes.NewFromMap(map[balance.Color]int64{
			vmctx.feeColor: vmctx.ownerFee,
		}))
	}
	if vmctx.validatorFee > 0 {
		vmctx.creditToAccount(vmctx.validatorFeeTarget, coretypes.NewFromMap(map[balance.Color]int64{
			vmctx.feeColor: vmctx.validatorFee,
		}))
	}
	// subtract fees from the transfer
	remaining := map[balance.Color]int64{
		vmctx.feeColor: -totalFee,
	}
	transfer.AddToMap(remaining)
	vmctx.remainingAfterFees = coretypes.NewFromMap(remaining)
}

// mustHandleFreeTokens free tokens accrued to the chain owner
func (vmctx *VMContext) mustHandleFreeTokens() {
	if vmctx.req.FreeTokens == nil || vmctx.req.FreeTokens.Len() == 0 {
		return
	}
	vmctx.creditToAccount(vmctx.ChainOwnerID(), vmctx.req.FreeTokens)
}

// mustHandleFallback all remaining tokens are:
// -- if sender is address, sent to that address
// -- otherwise accrue to the sender on-chain
func (vmctx *VMContext) mustHandleFallback() {
	sender := vmctx.req.SenderAgentID()
	if sender.IsAddress() {
		err := vmctx.txBuilder.TransferToAddress(sender.MustAddress(), vmctx.remainingAfterFees)
		if err != nil {
			vmctx.log.Panicf("mustHandleFallback: transferring tokens to address %s", sender.MustAddress().String())
		}
	} else {
		vmctx.creditToAccount(sender, vmctx.remainingAfterFees)
	}
}

// mustCallFromRequest is the call itself. Assumes sc exists
func (vmctx *VMContext) mustCallFromRequest() {
	req := vmctx.req.RequestSection()
	vmctx.log.Debugf("mustCallFromRequest: %s -- %s\n", vmctx.req.RequestID().String(), req.String())

	// calling only non vew entry points. Calling the view will trigger error and fallback
	vmctx.lastResult, vmctx.lastError = vmctx.callNonViewByProgramHash(
		vmctx.reqHname, req.EntryPointCode(), req.SolidArgs(), vmctx.remainingAfterFees, vmctx.contractRecord.ProgramHash)
}

func (vmctx *VMContext) finalizeRequestCall() {
	vmctx.mustRequestToEventLog(vmctx.lastError)
	vmctx.virtualState.ApplyStateUpdate(vmctx.stateUpdate)

	vmctx.log.Debugw("runTheRequest OUT",
		"reqId", vmctx.req.RequestID().Short(),
		"entry point", vmctx.req.RequestSection().EntryPointCode().String(),
	)
}

func (vmctx *VMContext) mustRequestToEventLog(err error) {
	if err != nil {
		vmctx.log.Error(err)
	}
	e := "Ok"
	if err != nil {
		e = err.Error()
	}
	msg := fmt.Sprintf("[req] %s: %s", vmctx.req.RequestID().String(), e)
	vmctx.log.Infof("eventlog -> '%s'", msg)
	vmctx.StoreToEventLog(vmctx.reqHname, []byte(msg))
}

// mustGetBaseValues only makes sense if chain is already deployed
func (vmctx *VMContext) mustGetBaseValues() {
	info := vmctx.mustGetChainInfo()
	if info.ChainID != vmctx.chainID {
		vmctx.log.Panicf("initRequestContext: major inconsistency of chainID")
	}
	vmctx.chainOwnerID = info.ChainOwnerID
	vmctx.feeColor, vmctx.ownerFee, vmctx.validatorFee = vmctx.getFeeInfo()
}

// initRequestContext initializes VMContext for request and returns  if contract exists
func (vmctx *VMContext) initRequestContext(req *sctransaction.Request, timestamp int64) {
	vmctx.req = req
	vmctx.reqHname = req.GetMetadata().TargetContractHname

	vmctx.timestamp = timestamp
	vmctx.stateUpdate = state.NewStateUpdate(req.Output().ID()).WithTimestamp(timestamp)
	vmctx.callStack = vmctx.callStack[:0]
	vmctx.entropy = hashing.HashData(vmctx.entropy[:])

	vmctx.contractRecord, _ = vmctx.findContractByHname(vmctx.reqHname)
}

func (vmctx *VMContext) isInitChainRequest() bool {
	return vmctx.req.GetMetadata().TargetContractHname == root.Interface.Hname() &&
		vmctx.req.GetMetadata().EntryPoint == coretypes.EntryPointInit
}

func (vmctx *VMContext) FinalizeTransactionEssence(blockIndex uint32, stateHash hashing.HashValue, timestamp int64) (*sctransaction_old.TransactionEssence, error) {
	// add state block
	err := vmctx.txBuilder.SetStateParams(blockIndex, stateHash, timestamp)
	if err != nil {
		return nil, err
	}
	tx, err := vmctx.txBuilder.Build()
	if err != nil {
		return nil, err
	}
	return tx, nil
}
