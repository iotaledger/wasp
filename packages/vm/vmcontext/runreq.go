package vmcontext

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
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
func (vmctx *VMContext) RunTheRequest(req *sctransaction.Request, inputIndex int) {
	processable := vmctx.mustSetUpRequestContext(req, inputIndex)
	if !processable {
		// wrong metadata or contract does not exist, just handle tokens
		vmctx.mustProcessBadRequest()
		return
	}
	if !vmctx.isInitChainRequest() {
		vmctx.mustGetBaseValues()
		vmctx.mustHandleFees()
	}
	defer vmctx.finalizeRequestCall()

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

func (vmctx *VMContext) mustProcessBadRequest() {
	// TODO accrue all tokens to chain owner
	if vmctx.contractRecord == nil {
		// sc does not exist, stop here
		vmctx.lastResult = nil
		vmctx.lastError = fmt.Errorf("smart contract '%s' does not exist", vmctx.targetContractHname)
		return
	}
}

// mustHandleFees:
// - handles node fee, including fallback if not enough
func (vmctx *VMContext) mustHandleFees() {
	totalFee := vmctx.ownerFee + vmctx.validatorFee
	if totalFee == 0 || vmctx.requesterIsChainOwner() {
		// no fees enabled or the caller is the chain owner
		vmctx.log.Debugf("mustHandleFees: no fees charged\n")
		return
	}
	// handle fees
	if vmctx.incoming.Balance(vmctx.feeColor) < totalFee {
		// TODO more sophisticated policy, for example taking fees to chain owner, the rest returned to sender
		// fallback: not enough fees. Accrue everything to the sender
		sender := vmctx.req.SenderAgentID()
		vmctx.creditToAccount(sender, vmctx.incoming)
		vmctx.lastError = fmt.Errorf("mustHandleFees: not enough fees for request %s. Transfer accrued to %s",
			vmctx.req.Output().ID().Base58(), sender.String())
		vmctx.incoming = coretypes.NewFromMap(nil)
		return
	}
	// enough fees. Split between owner and validator
	if vmctx.ownerFee > 0 {
		vmctx.creditToAccount(vmctx.ChainOwnerID(), coretypes.NewFromMap(map[ledgerstate.Color]uint64{
			vmctx.feeColor: vmctx.ownerFee,
		}))
	}
	if vmctx.validatorFee > 0 {
		vmctx.creditToAccount(vmctx.validatorFeeTarget, coretypes.NewFromMap(map[ledgerstate.Color]uint64{
			vmctx.feeColor: vmctx.validatorFee,
		}))
	}
	// subtract fees from the transfer
	remaining := vmctx.incoming.Map()
	s, _ := remaining[vmctx.feeColor]
	remaining[vmctx.feeColor] = s - totalFee
	vmctx.incoming = coretypes.NewFromMap(remaining)
}

// mustHandleFallback all remaining tokens are:
// -- if sender is address, sent to that address
// -- otherwise accrue to the sender on-chain
func (vmctx *VMContext) mustHandleFallback() {
	sender := vmctx.req.SenderAgentID()
	if sender.IsNonAliasAddress() {
		err := vmctx.txBuilder.AddExtendedOutputSimple(
			sender.MustAddress(),
			[]byte("returned due to error"),
			vmctx.incoming.Map(),
		)
		if err != nil {
			vmctx.log.Panicf("mustHandleFallback: transferring tokens to address %s", sender.MustAddress().String())
		}
	} else {
		vmctx.creditToAccount(sender, vmctx.incoming)
	}
}

// mustCallFromRequest is the call itself. Assumes sc exists
func (vmctx *VMContext) mustCallFromRequest() {
	req := vmctx.req.RequestSection()
	vmctx.log.Debugf("mustCallFromRequest: %s -- %s\n", vmctx.req.RequestID().String(), req.String())

	// calling only non vew entry points. Calling the view will trigger error and fallback
	vmctx.lastResult, vmctx.lastError = vmctx.callNonViewByProgramHash(
		vmctx.targetContractHname, req.EntryPointCode(), req.SolidArgs(), vmctx.remainingAfterFees, vmctx.contractRecord.ProgramHash)
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
	vmctx.StoreToEventLog(vmctx.targetContractHname, []byte(msg))
}

// mustGetBaseValues only makes sense if chain is already deployed
func (vmctx *VMContext) mustGetBaseValues() {
	info := vmctx.mustGetChainInfo()
	if info.ChainID != vmctx.chainID {
		vmctx.log.Panicf("mustSetUpRequestContext: major inconsistency of chainID")
	}
	vmctx.chainOwnerID = info.ChainOwnerID
	vmctx.feeColor, vmctx.ownerFee, vmctx.validatorFee = vmctx.getFeeInfo()
}

// mustSetUpRequestContext sets up VMContext for request
func (vmctx *VMContext) mustSetUpRequestContext(req *sctransaction.Request, inputIndex int) bool {
	if req.SolidArgs() == nil {
		vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: request args should had been solidified")
	}
	vmctx.req = req
	vmctx.targetContractHname = req.GetMetadata().TargetContractHname

	if err := vmctx.txBuilder.ConsumeByIndex(inputIndex); err != nil {
		vmctx.log.Panicf("mustSetUpRequestContext.inconsistency : %v", err)
	}
	vmctx.incoming = coretypes.NewColoredBalances(*req.Output().Balances())
	vmctx.timestamp += 1
	vmctx.entropy = hashing.HashData(vmctx.entropy[:])
	vmctx.stateUpdate = state.NewStateUpdate(req.Output().ID()).WithTimestamp(vmctx.timestamp)
	vmctx.callStack = vmctx.callStack[:0]

	if !req.ParsedOk() {
		return false
	}
	vmctx.contractRecord, _ = vmctx.findContractByHname(vmctx.targetContractHname)
	if vmctx.contractRecord == nil {
		return false
	}
	return true
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
