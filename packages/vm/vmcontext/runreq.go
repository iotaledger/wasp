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
	"golang.org/x/xerrors"
	"time"
)

// RunTheRequest processes any request based on the Extended output, even if it
// doesn't parse correctly as a SC request
func (vmctx *VMContext) RunTheRequest(req *sctransaction.Request, inputIndex int) {
	vmctx.mustSetUpRequestContext(req, inputIndex)
	vmctx.mustPreProcessRequest()

	defer vmctx.finalizeRequestCall()

	if vmctx.contractRecord == nil {
		// sc does not exist, stop here
		vmctx.lastResult = nil
		vmctx.lastError = fmt.Errorf("smart contract '%s' does not exist", vmctx.req.GetMetadata().TargetContract())
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
					// There was an error accessing the DB. The world stops
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

// mustSetUpRequestContext sets up VMContext for request
func (vmctx *VMContext) mustSetUpRequestContext(req *sctransaction.Request, inputIndex int) {
	if req.SolidArgs() == nil {
		vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: request args should had been solidified")
	}
	vmctx.req = req
	inp, err := vmctx.txBuilder.InputByIndex(inputIndex)
	if err != nil {
		vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: %v", err)
	}
	input, ok := inp.(*ledgerstate.ExtendedLockedOutput)
	if !ok {
		vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: unexpected input type")
	}
	if err := vmctx.txBuilder.ConsumeInputByIndex(inputIndex); err != nil {
		vmctx.log.Panicf("mustSetUpRequestContext.inconsistency : %v", err)
	}
	vmctx.timestamp += 1
	t := time.Unix(0, vmctx.timestamp)
	if input.TimeLockedNow(t) {
		vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: input is time locked. Nowis: %v\nInput: %s\n", t, input.String())
	}
	if !input.UnlockAddressNow(t).Equals(vmctx.chainID.AsAddress()) {
		vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: input cannot be unlocked at %v.\nInput: %s\n chainID: %s",
			t, input.String(), vmctx.chainID.String())
	}

	vmctx.remainingAfterFees = coretypes.NewColoredBalances(*req.Output().Balances())
	vmctx.entropy = hashing.HashData(vmctx.entropy[:])
	vmctx.stateUpdate = state.NewStateUpdate(req.Output().ID()).WithTimestamp(vmctx.timestamp)
	vmctx.callStack = vmctx.callStack[:0]

	vmctx.contractRecord = nil
	if req.GetMetadata().ParsedOk() {
		vmctx.contractRecord, _ = vmctx.findContractByHname(req.GetMetadata().TargetContract())
	}
}

func (vmctx *VMContext) mustPreProcessRequest() {
	if vmctx.isInitChainRequest() {
		return
	}
	vmctx.mustGetBaseValues()
	vmctx.mustHandleFees()
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
	if vmctx.remainingAfterFees.Balance(vmctx.feeColor) < totalFee {
		// TODO more sophisticated policy, for example taking fees to chain owner, the rest returned to sender
		// fallback: not enough fees. Accrue everything to the sender
		sender := vmctx.req.SenderAgentID()
		vmctx.creditToAccount(sender, &vmctx.remainingAfterFees)
		vmctx.lastError = fmt.Errorf("mustHandleFees: not enough fees for request %s. Transfer accrued to %s",
			vmctx.req.Output().ID().Base58(), sender.String())
		vmctx.remainingAfterFees = coretypes.NewColoredBalancesFromMap(nil)
		return
	}
	// enough fees. Split between owner and validator
	if vmctx.ownerFee > 0 {
		t := coretypes.NewColoredBalancesFromMap(map[ledgerstate.Color]uint64{
			vmctx.feeColor: vmctx.ownerFee,
		})
		vmctx.creditToAccount(vmctx.ChainOwnerID(), &t)
	}
	if vmctx.validatorFee > 0 {
		t := coretypes.NewColoredBalancesFromMap(map[ledgerstate.Color]uint64{
			vmctx.feeColor: vmctx.validatorFee,
		})
		vmctx.creditToAccount(&vmctx.validatorFeeTarget, &t)
	}
	// subtract fees from the transfer
	remaining := vmctx.remainingAfterFees.Map()
	s, _ := remaining[vmctx.feeColor]
	remaining[vmctx.feeColor] = s - totalFee
	vmctx.remainingAfterFees = coretypes.NewColoredBalancesFromMap(remaining)
}

// mustHandleFallback all remaining tokens are:
// -- if sender is address, sent to that address
// -- otherwise accrue to the sender on-chain
func (vmctx *VMContext) mustHandleFallback() {
	sender := vmctx.req.SenderAgentID()
	if !sender.IsContract() {
		err := vmctx.txBuilder.AddExtendedOutputSimple(
			sender.AsAddress(),
			[]byte("returned due to error"),
			vmctx.remainingAfterFees.Map(),
		)
		if err != nil {
			vmctx.log.Panicf("mustHandleFallback: transferring tokens to address %s", sender.AsAddress().String())
		}
	} else {
		vmctx.creditToAccount(sender, &vmctx.remainingAfterFees)
	}
}

// mustCallFromRequest is the call itself. Assumes sc exists
func (vmctx *VMContext) mustCallFromRequest() {
	vmctx.log.Debugf("mustCallFromRequest: %s\n", vmctx.req.Output().ID().String())

	// calling only non vew entry points. Calling the view will trigger error and fallback
	md := vmctx.req.GetMetadata()
	vmctx.lastResult, vmctx.lastError = vmctx.callNonViewByProgramHash(
		md.TargetContract(), md.EntryPoint(), vmctx.req.SolidArgs(), &vmctx.remainingAfterFees, vmctx.contractRecord.ProgramHash)
}

func (vmctx *VMContext) finalizeRequestCall() {
	vmctx.mustRequestToEventLog(vmctx.lastError)
	vmctx.virtualState.ApplyStateUpdate(vmctx.stateUpdate)

	vmctx.log.Debugw("runTheRequest OUT",
		"reqId", vmctx.req.Output().ID().String(),
		"entry point", vmctx.req.GetMetadata().EntryPoint().String(),
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
	msg := fmt.Sprintf("[req] %s: %s", vmctx.req.Output().ID().Base58(), e)
	vmctx.log.Infof("eventlog -> '%s'", msg)
	vmctx.StoreToEventLog(vmctx.req.GetMetadata().TargetContract(), []byte(msg))
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

func (vmctx *VMContext) isInitChainRequest() bool {
	return vmctx.req.GetMetadata().TargetContract() == root.Interface.Hname() &&
		vmctx.req.GetMetadata().EntryPoint() == coretypes.EntryPointInit
}

func (vmctx *VMContext) BuildTransactionEssence(blockIndex uint32, stateHash hashing.HashValue) (*ledgerstate.TransactionEssence, error) {
	stateMetadata := sctransaction.NewStateMetadata(blockIndex, stateHash)
	if err := vmctx.txBuilder.AddChainOutputAsReminder(vmctx.chainID.AsAddress(), stateMetadata.Bytes()); err != nil {
		return nil, xerrors.Errorf("finalizeRequestCall: %v", err)
	}
	tx, _, err := vmctx.txBuilder.BuildEssence()
	if err != nil {
		return nil, err
	}
	return tx, nil
}
