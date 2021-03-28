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
func (vmctx *VMContext) RunTheRequest(req coretypes.Request, inputIndex int) {
	defer vmctx.finalizeRequestCall()

	vmctx.mustSetUpRequestContext(req)

	enoughFees := true
	if !vmctx.isInitChainRequest() {
		vmctx.mustGetBaseValues()
		enoughFees = vmctx.mustHandleFees()
	}

	if !enoughFees {
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
		// restore the txbuilder and state back to the moment before calling VM plugin
		vmctx.txBuilder = snapshotTxBuilder
		vmctx.stateUpdate = snapshotStateUpdate

		vmctx.mustSendBack(vmctx.remainingAfterFees)
	}
}

// mustSetUpRequestContext sets up VMContext for request
func (vmctx *VMContext) mustSetUpRequestContext(req coretypes.Request) {
	if req.Params() == nil {
		vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: request args should had been solidified")
	}
	vmctx.req = req
	if req.Output() != nil {
		if err := vmctx.txBuilder.ConsumeInputByOutputID(req.Output().ID()); err != nil {
			vmctx.log.Panicf("mustSetUpRequestContext.inconsistency : %v", err)
		}

	}
	vmctx.timestamp += 1
	t := time.Unix(0, vmctx.timestamp)
	if input, ok := req.Output().(*ledgerstate.ExtendedLockedOutput); ok {
		// it is an on-ledger request
		if input.TimeLockedNow(t) {
			vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: input is time locked. Nowis: %v\nInput: %s\n", t, input.String())
		}
		if !input.UnlockAddressNow(t).Equals(vmctx.chainID.AsAddress()) {
			vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: input cannot be unlocked at %v.\nInput: %s\n chainID: %s",
				t, input.String(), vmctx.chainID.String())
		}
	}

	vmctx.remainingAfterFees = req.Output().Balances().Clone()
	vmctx.entropy = hashing.HashData(vmctx.entropy[:])
	vmctx.stateUpdate = state.NewStateUpdate(req.Output().ID()).WithTimestamp(vmctx.timestamp)
	vmctx.callStack = vmctx.callStack[:0]

	targetContract, _ := req.Target()
	var ok bool
	if vmctx.contractRecord, ok = vmctx.findContractByHname(targetContract); !ok {
		vmctx.log.Panicf("inconsistency: findContractByHname")
	}
	if vmctx.contractRecord.Hname() == 0 {
		vmctx.log.Warn("default contract will be called")
	}
}

// mustHandleFees handles node fees. If not enough, takes as much as it can, the rest sends back
// Return false if not enough fees
func (vmctx *VMContext) mustHandleFees() bool {
	totalFee := vmctx.ownerFee + vmctx.validatorFee
	if totalFee == 0 || vmctx.requesterIsLocal() {
		// no fees enabled or the caller is the chain owner
		vmctx.log.Debugf("mustHandleFees: no fees charged")
		return true
	}
	// handle fees
	availableForFees, _ := vmctx.remainingAfterFees.Get(vmctx.feeColor)
	if availableForFees < totalFee {
		// take as much as available, the rest send back
		rem := vmctx.remainingAfterFees.Map()
		delete(rem, vmctx.feeColor)
		if availableForFees > 0 {
			accrue := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
				vmctx.feeColor: availableForFees,
			})
			vmctx.creditToAccount(vmctx.commonAccount(), accrue)
		}
		vmctx.mustSendBack(ledgerstate.NewColoredBalances(rem))
		vmctx.lastError = fmt.Errorf("mustHandleFees: not enough fees for request %s. Remaining tokens were sent back to %s",
			vmctx.req.ID(), vmctx.req.SenderAddress().Base58())
		vmctx.remainingAfterFees = nil
		return false
	}
	// enough fees. Split between owner and validator
	if vmctx.ownerFee > 0 {
		t := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
			vmctx.feeColor: vmctx.ownerFee,
		})
		// send to common account
		vmctx.creditToAccount(vmctx.commonAccount(), t)
	}
	if vmctx.validatorFee > 0 {
		t := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
			vmctx.feeColor: vmctx.validatorFee,
		})
		vmctx.creditToAccount(&vmctx.validatorFeeTarget, t)
	}
	// subtract fees from the transfer
	remaining := vmctx.remainingAfterFees.Map()
	s, _ := remaining[vmctx.feeColor]
	if s > totalFee {
		remaining[vmctx.feeColor] = s - totalFee
	} else {
		delete(remaining, vmctx.feeColor)
	}
	vmctx.remainingAfterFees = ledgerstate.NewColoredBalances(remaining)
	return true
}

func (vmctx *VMContext) mustSendBack(tokens *ledgerstate.ColoredBalances) {
	if tokens == nil || tokens.Size() == 0 {
		return
	}
	sender := vmctx.req.SenderAccount()
	if sender.Address().Equals(vmctx.chainID.AsAddress()) {
		// if sender is on the same chain, just accrue tokens back to it
		vmctx.creditToAccount(vmctx.adjustAccount(sender), tokens)
		return
	}
	// send tokens back
	// the logic is to send to original aliasAddress and to original contract is any
	// otherwise will be sent to _default contract. In case if sender
	// is ordinary wallet the tokens (less fees) will be returned back
	backToAddress := sender.Address()
	backToContract := vmctx.req.SenderAccount().Hname()
	metadata := sctransaction.NewRequestMetadata().WithTarget(backToContract)
	err := vmctx.txBuilder.AddExtendedOutputSpend(backToAddress, metadata.Bytes(), tokens.Map())
	if err != nil {
		vmctx.log.Errorf("mustSendBack: %v", err)
	}
}

// mustCallFromRequest is the call itself. Assumes sc exists
func (vmctx *VMContext) mustCallFromRequest() {
	vmctx.log.Debugf("mustCallFromRequest: %s", vmctx.req.ID().String())

	// calling only non view entry points. Calling the view will trigger error and fallback
	targetContract, entryPoint := vmctx.req.Target()
	vmctx.lastResult, vmctx.lastError = vmctx.callNonViewByProgramHash(
		targetContract, entryPoint, vmctx.req.Params(), vmctx.remainingAfterFees, vmctx.contractRecord.ProgramHash)
}

func (vmctx *VMContext) finalizeRequestCall() {
	vmctx.mustRequestToEventLog(vmctx.lastError)
	vmctx.lastTotalAssets = vmctx.totalAssets()
	vmctx.virtualState.ApplyStateUpdate(vmctx.stateUpdate)

	_, ep := vmctx.req.Target()
	vmctx.log.Debugw("runTheRequest OUT",
		"reqId", vmctx.req.ID().Short(),
		"entry point", ep.String(),
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
	reqStr := coretypes.RequestID(vmctx.req.Output().ID()).String()
	msg := fmt.Sprintf("[req] %s: %s", reqStr, e)
	vmctx.log.Infof("eventlog -> '%s'", msg)
	targetContract, _ := vmctx.req.Target()
	vmctx.StoreToEventLog(targetContract, []byte(msg))
}

// mustGetBaseValues only makes sense if chain is already deployed
func (vmctx *VMContext) mustGetBaseValues() {
	info := vmctx.mustGetChainInfo()
	if !info.ChainID.Equals(&vmctx.chainID) {
		vmctx.log.Panicf("mustSetUpRequestContext: major inconsistency of chainID")
	}
	vmctx.chainOwnerID = info.ChainOwnerID
	vmctx.feeColor, vmctx.ownerFee, vmctx.validatorFee = vmctx.getFeeInfo()
}

func (vmctx *VMContext) isInitChainRequest() bool {
	targetContract, entryPoint := vmctx.req.Target()
	return targetContract == root.Interface.Hname() && entryPoint == coretypes.EntryPointInit
}

func (vmctx *VMContext) BuildTransactionEssence(stateHash hashing.HashValue) (*ledgerstate.TransactionEssence, error) {
	if err := vmctx.txBuilder.AddAliasOutputAsReminder(vmctx.chainID.AsAddress(), stateHash[:]); err != nil {
		return nil, xerrors.Errorf("finalizeRequestCall: %v", err)
	}
	tx, _, err := vmctx.txBuilder.BuildEssence()
	if err != nil {
		return nil, err
	}
	return tx, nil
}
