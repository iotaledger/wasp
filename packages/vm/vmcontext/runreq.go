package vmcontext

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"runtime/debug"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"golang.org/x/xerrors"
)

// RunTheRequest processes any request based on the Extended output, even if it
// doesn't parse correctly as a SC request
func (vmctx *VMContext) RunTheRequest(req coretypes.Request, inputIndex int) {
	defer vmctx.finalizeRequestCall()

	vmctx.mustSetUpRequestContext(req)

	// guard against replaying off-ledger requests here to prevent replaying fee deduction
	// also verifies that account for off-ledger request exists
	if !vmctx.validRequest() {
		return
	}

	if !vmctx.isInitChainRequest() {
		vmctx.mustGetBaseValues()
		enoughFees := vmctx.mustHandleFees()
		if !enoughFees {
			return
		}
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
				vmctx.lastError = xerrors.Errorf("%s: recovered from panic in VM: %v", req.ID(), r)
				vmctx.Debugf(string(debug.Stack()))
				if dberr, ok := r.(*kv.DBError); ok {
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
	if _, ok := req.Params(); !ok {
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

	vmctx.entropy = hashing.HashData(vmctx.entropy[:])
	vmctx.stateUpdate = state.NewStateUpdate(req.ID()).WithTimestamp(vmctx.timestamp)
	vmctx.callStack = vmctx.callStack[:0]

	if isRequestTimeLockedNow(req, t) {
		vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: input is time locked. Nowis: %v\nInput: %s\n", t, req.ID().String())
	}
	if req.Output() != nil {
		// on-ledger request
		if input, ok := req.Output().(*ledgerstate.ExtendedLockedOutput); ok {
			// it is an on-ledger request
			if !input.UnlockAddressNow(t).Equals(vmctx.chainID.AsAddress()) {
				vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: input cannot be unlocked at %v.\nInput: %s\n chainID: %s",
					t, input.String(), vmctx.chainID.String())
			}
		} else {
			vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: unexpected UTXO type")
		}
		vmctx.remainingAfterFees = req.Output().Balances().Clone()
	} else {
		// off-ledger request
		vmctx.remainingAfterFees = vmctx.adjustOffLedgerTransfer()
	}

	targetContract, _ := req.Target()
	var ok bool
	if vmctx.contractRecord, ok = vmctx.findContractByHname(targetContract); !ok {
		vmctx.log.Panicf("inconsistency: findContractByHname")
	}
	if vmctx.contractRecord.Hname() == 0 {
		vmctx.log.Warn("default contract will be called")
	}
}

func (vmctx *VMContext) adjustOffLedgerTransfer() *ledgerstate.ColoredBalances {
	req, ok := vmctx.req.(*request.RequestOffLedger)
	if !ok {
		vmctx.log.Panicf("adjustOffLedgerTransfer.inconsistency: unexpected request type")
	}
	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	// take sender-provided token transfer info and adjust it to
	// reflect what is actually available in the local sender account
	sender := req.SenderAccount()
	transfers := make(map[ledgerstate.Color]uint64)
	if tokens := req.Tokens(); tokens != nil {
		tokens.ForEach(func(color ledgerstate.Color, balance uint64) bool {
			available := accounts.GetBalance(vmctx.State(), sender, color)
			if balance > available {
				vmctx.log.Warn("adjusting transfer from ", balance, " to ", available)
				balance = available
			}
			if balance > 0 {
				transfers[color] = balance
			}
			return true
		})
	}
	return ledgerstate.NewColoredBalances(transfers)
}

func (vmctx *VMContext) validRequest() bool {
	req, ok := vmctx.req.(*request.RequestOffLedger)
	if !ok {
		// on-ledger request is always valid
		return true
	}

	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	// off-ledger account must exist
	if _, exists := accounts.GetAccountBalances(vmctx.State(), req.SenderAccount()); !exists {
		vmctx.lastError = fmt.Errorf("validRequest: unverified account for %s", req.ID().String())
		return false
	}

	// order of requests must always increase
	if vmctx.req.Order() <= accounts.GetOrder(vmctx.State(), req.SenderAddress()) {
		vmctx.lastError = fmt.Errorf("validRequest: invalid order for %s", req.ID().String())
		return false
	}

	return true
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

	// process fees for owner and validator
	if vmctx.grabFee(vmctx.commonAccount(), vmctx.ownerFee) &&
		vmctx.grabFee(&vmctx.validatorFeeTarget, vmctx.validatorFee) {
		// there were enough fees for both
		return true
	}

	// not enough fees available
	vmctx.mustSendBack(vmctx.remainingAfterFees)
	vmctx.remainingAfterFees = nil
	vmctx.lastError = fmt.Errorf("mustHandleFees: not enough fees for request %s. Remaining tokens were sent back to %s",
		vmctx.req.ID(), vmctx.req.SenderAddress().Base58())
	return false
}

// Return false if not enough fees
func (vmctx *VMContext) grabFee(account *coretypes.AgentID, amount uint64) bool {
	if amount == 0 {
		return true
	}

	// determine how much fees we can actually take
	available, _ := vmctx.remainingAfterFees.Get(vmctx.feeColor)
	if available == 0 {
		return false
	}
	enoughFees := available >= amount
	if !enoughFees {
		// just take whatever is there
		amount = available
	}
	available -= amount

	// take fee from remainingAfterFees
	remaining := vmctx.remainingAfterFees.Map()
	if available == 0 {
		delete(remaining, vmctx.feeColor)
	} else {
		remaining[vmctx.feeColor] = available
	}
	vmctx.remainingAfterFees = ledgerstate.NewColoredBalances(remaining)

	// get ready to transfer the fees
	transfer := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		vmctx.feeColor: amount,
	})

	if !vmctx.req.IsFeePrepaid() {
		vmctx.creditToAccount(account, transfer)
		return enoughFees
	}

	// fees should have been deposited in sender account on chain
	sender := vmctx.req.SenderAccount()
	return vmctx.moveBetweenAccounts(sender, account, transfer) && enoughFees
}

func (vmctx *VMContext) mustSendBack(tokens *ledgerstate.ColoredBalances) {
	if tokens == nil || tokens.Size() == 0 || vmctx.req.Output() == nil {
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
	backToContract := sender.Hname()
	metadata := request.NewRequestMetadata().WithTarget(backToContract)
	err := vmctx.txBuilder.AddExtendedOutputSpend(backToAddress, metadata.Bytes(), tokens.Map())
	if err != nil {
		vmctx.log.Errorf("mustSendBack: %v", err)
	}
}

// mustCallFromRequest is the call itself. Assumes sc exists
func (vmctx *VMContext) mustCallFromRequest() {
	vmctx.log.Debugf("mustCallFromRequest: %s", vmctx.req.ID().String())

	vmctx.mustSaveRequestOrder()

	// calling only non view entry points. Calling the view will trigger error and fallback
	targetContract, entryPoint := vmctx.req.Target()
	params, _ := vmctx.req.Params()
	vmctx.lastResult, vmctx.lastError = vmctx.callNonViewByProgramHash(
		targetContract, entryPoint, params, vmctx.remainingAfterFees, vmctx.contractRecord.ProgramHash)
}

func (vmctx *VMContext) mustSaveRequestOrder() {
	if _, ok := vmctx.req.(*request.RequestOffLedger); ok {
		vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil)
		defer vmctx.popCallContext()

		address := vmctx.req.SenderAddress()
		order := vmctx.req.Order()
		accounts.SetOrder(vmctx.State(), address, order)
	}
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
	reqStr := coretypes.OID(vmctx.req.ID().OutputID())
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

func isRequestTimeLockedNow(req coretypes.Request, nowis time.Time) bool {
	if req.TimeLock().IsZero() {
		return false
	}
	return req.TimeLock().After(nowis)
}

func (vmctx *VMContext) BuildTransactionEssence(stateHash hashing.HashValue, timestamp time.Time) (*ledgerstate.TransactionEssence, error) {
	if err := vmctx.txBuilder.AddAliasOutputAsRemainder(vmctx.chainID.AsAddress(), stateHash[:]); err != nil {
		return nil, xerrors.Errorf("finalizeRequestCall: %v", err)
	}
	tx, _, err := vmctx.txBuilder.WithTimestamp(timestamp).BuildEssence()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (vmctx *VMContext) StoreBlockInfo(blockIndex uint32, blockInfo *blocklog.BlockInfo) error {
	vmctx.pushCallContext(blocklog.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	idx := blocklog.SaveNextBlockInfo(vmctx.State(), blockInfo)
	if idx != blockIndex {
		return xerrors.New("StoreBlockInfo: inconsistent block index")
	}
	return nil
}
