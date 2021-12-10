package vmcontext

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"golang.org/x/xerrors"
)

// TODO temporary place for the constant. In the future must be shared with pruning module
//  the number is approximate assumed maximum number of requests in the batch
//  The node must guarantee at least this number of last requests processed recorded in the state
//  for each address
const OffLedgerNonceStrictOrderTolerance = 10000

// RunTheRequest processes any request based on the Extended output, even if it
// doesn't parse correctly as a SC request
func (vmctx *VMContext) RunTheRequest(req iscp.Request, requestIndex uint16) {
	defer vmctx.mustFinalizeRequestCall()

	// snapshot tx builder to be able to rollbck and not include this request in the current block
	snapshotTxBuilderWithoutInput := vmctx.txBuilder.Clone()
	vmctx.mustSetUpRequestContext(req, requestIndex)

	// guard against replaying off-ledger requests here to prevent replaying fee deduction
	// also verifies that account for off-ledger request exists
	if !vmctx.validateRequest() {
		vmctx.log.Debugw("vmctx.RunTheRequest: request failed validation", "id", vmctx.req.ID())
		return
	}

	if vmctx.isInitChainRequest() {
		vmctx.chainOwnerID = vmctx.req.SenderAccount().Clone()
	} else {
		vmctx.getChainConfigFromState()
		enoughFees := vmctx.mustHandleFees()
		if !enoughFees {
			return
		}
	}

	// snapshot state baseline for rollback in case of panic
	snapshotTxBuilder := vmctx.txBuilder.Clone()
	vmctx.virtualState.ApplyStateUpdates(vmctx.currentStateUpdate)
	// request run updates will be collected to the new state update
	vmctx.currentStateUpdate = state.NewStateUpdate()

	vmctx.lastError = nil
	func() {
		// panic catcher for the whole call from request to the VM
		defer func() {
			r := recover()
			if r == nil {
				return
			}

			switch err := r.(type) {
			case *kv.DBError:
				panic(err)
			case error:
				if errors.Is(err, coreutil.ErrorStateInvalidated) {
					panic(err)
				}
			}
			vmctx.lastResult = nil
			vmctx.lastError = xerrors.Errorf("panic in VM: %v", r)
			vmctx.Debugf("%v", vmctx.lastError)
			vmctx.Debugf(string(debug.Stack()))
		}()
		vmctx.mustCallFromRequest()
	}()

	if vmctx.blockOutputCount > MaxBlockOutputCount {
		vmctx.exceededBlockOutputLimit = true
		vmctx.blockOutputCount -= vmctx.requestOutputCount
		vmctx.Debugf("outputs produced by this request do not fit inside the current block, reqID: %s", vmctx.req.ID().Base58())
		// rollback request processing, don't consume output or send funds back as this request should be processed in a following batch
		vmctx.txBuilder = snapshotTxBuilderWithoutInput
		vmctx.currentStateUpdate = state.NewStateUpdate()
	}

	if vmctx.lastError != nil {
		// treating panic and error returned from request the same way
		// restore the txbuilder and dispose mutations in the last state update
		vmctx.txBuilder = snapshotTxBuilder
		vmctx.currentStateUpdate = state.NewStateUpdate()

		vmctx.mustSendBack(vmctx.remainingAfterFees)
	}
}

// mustSetUpRequestContext sets up VMContext for request
func (vmctx *VMContext) mustSetUpRequestContext(req iscp.Request, requestIndex uint16) {
	if _, ok := req.Params(); !ok {
		vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: request args should had been solidified")
	}
	vmctx.req = req
	vmctx.requestIndex = requestIndex
	vmctx.requestEventIndex = 0
	vmctx.requestOutputCount = 0
	vmctx.exceededBlockOutputLimit = false

	if !req.IsOffLedger() {
		vmctx.txBuilder.AddConsumable(vmctx.req.(*request.OnLedger).Output())
		if err := vmctx.txBuilder.ConsumeInputByOutputID(req.(*request.OnLedger).Output().ID()); err != nil {
			vmctx.log.Panicf("mustSetUpRequestContext.inconsistency : %v", err)
		}
	}
	ts := vmctx.virtualState.Timestamp().Add(1 * time.Nanosecond)
	vmctx.currentStateUpdate = state.NewStateUpdate(ts)

	vmctx.entropy = hashing.HashData(vmctx.entropy[:])
	vmctx.callStack = vmctx.callStack[:0]

	if isRequestTimeLockedNow(req, ts) {
		vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: input is time locked. Nowis: %v\nInput: %s\n", ts, req.ID().String())
	}
	if !req.IsOffLedger() {
		// on-ledger request
		reqt := req.(*request.OnLedger)
		if input, ok := reqt.Output().(*ledgerstate.ExtendedLockedOutput); ok {
			// it is an on-ledger request
			if !input.UnlockAddressNow(ts).Equals(vmctx.chainID.AsAddress()) {
				vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: input cannot be unlocked at %v.\nInput: %s\n chainID: %s",
					ts, input.String(), vmctx.chainID.String())
			}
		} else {
			vmctx.log.Panicf("mustSetUpRequestContext.inconsistency: unexpected UTXO type")
		}
		vmctx.remainingAfterFees = colored.BalancesFromL1Balances(reqt.Output().Balances())
	} else {
		// off-ledger request
		vmctx.remainingAfterFees = vmctx.adjustOffLedgerTransfer()
	}

	targetContract, _ := req.Target()
	var ok bool
	vmctx.contractRecord, ok = vmctx.findContractByHname(targetContract)
	if !ok {
		vmctx.log.Warnf("contract not found: %s", targetContract)
	}
	if vmctx.contractRecord.Hname() == 0 {
		vmctx.log.Warn("default contract will be called")
	}
}

func (vmctx *VMContext) adjustOffLedgerTransfer() colored.Balances {
	req, ok := vmctx.req.(*request.OffLedger)
	if !ok {
		vmctx.log.Panicf("adjustOffLedgerTransfer.inconsistency: unexpected request type")
	}
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	// take sender-provided token transfer info and adjust it to
	// reflect what is actually available in the local sender account
	sender := req.SenderAccount()
	transfers := colored.NewBalances()
	// deterministic order of iteration is not necessary
	req.Tokens().ForEachRandomly(func(col colored.Color, bal uint64) bool {
		available := accounts.GetBalance(vmctx.State(), sender, col)
		if bal > available {
			vmctx.log.Warn(
				"adjusting transfer from ", bal,
				" to available ", available,
				" for ", sender.String(),
				" req ", vmctx.Request().ID().String(),
			)
			bal = available
		}
		if bal > 0 {
			transfers.Set(col, bal)
		}
		return true
	})
	return transfers
}

func (vmctx *VMContext) validateRequest() bool {
	req, ok := vmctx.req.(*request.OffLedger)
	if !ok {
		// on-ledger request is always valid
		return true
	}

	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	// off-ledger account must exist, i.e. it should have non zero balance on the chain
	if _, exists := accounts.GetAccountBalances(vmctx.State(), req.SenderAccount()); !exists {
		vmctx.lastError = fmt.Errorf("validateRequest: unverified account %s for %s", req.SenderAccount(), req.ID().String())
		return false
	}

	// this is a replay protection measure for off-ledger requests assuming in the batch order of requests is random.
	// See rfc [replay-off-ledger.md]

	maxAssumed := accounts.GetMaxAssumedNonce(vmctx.State(), req.SenderAddress())

	vmctx.log.Debugf("vmctx.validateRequest - nonce check - maxAssumed: %d, tolerance: %d, request nonce: %d ",
		maxAssumed, OffLedgerNonceStrictOrderTolerance, req.Nonce())

	if maxAssumed < OffLedgerNonceStrictOrderTolerance {
		return true
	}
	return req.Nonce() > maxAssumed-OffLedgerNonceStrictOrderTolerance
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
		vmctx.grabFee(vmctx.validatorFeeTarget, vmctx.validatorFee) {
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
func (vmctx *VMContext) grabFee(account *iscp.AgentID, amount uint64) bool {
	if amount == 0 {
		return true
	}

	// determine how much fees we can actually take
	available := vmctx.remainingAfterFees.Get(vmctx.feeColor)
	if available == 0 {
		return false
	}
	enoughFees := available >= amount
	if !enoughFees {
		// just take whatever is there
		amount = available
	}
	vmctx.remainingAfterFees.SubNoOverflow(vmctx.feeColor, amount)

	// get ready to transfer the fees
	transfer := colored.NewBalancesForColor(vmctx.feeColor, amount)

	if !vmctx.req.IsFeePrepaid() {
		vmctx.creditToAccount(account, transfer)
		return enoughFees
	}

	// fees should have been deposited in sender account on chain
	sender := vmctx.req.SenderAccount()
	return vmctx.moveBetweenAccounts(sender, account, transfer) && enoughFees
}

func (vmctx *VMContext) mustSendBack(tokens colored.Balances) {
	if len(tokens) == 0 || vmctx.req.IsOffLedger() {
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
	metadata := request.NewMetadata().WithTarget(backToContract)
	err := vmctx.txBuilder.AddExtendedOutputSpend(backToAddress, metadata.Bytes(), colored.ToL1Map(tokens), nil)
	if err != nil {
		vmctx.log.Errorf("mustSendBack: %v", err)
	}
}

// mustCallFromRequest is the call itself. Assumes sc exists
func (vmctx *VMContext) mustCallFromRequest() {
	vmctx.log.Debugf("mustCallFromRequest: %s", vmctx.req.ID().String())

	vmctx.mustUpdateOffledgerRequestMaxAssumedNonce()

	// calling only non view entry points. Calling the view will trigger error and fallback
	_, entryPoint := vmctx.req.Target()
	targetContract := vmctx.contractRecord.Hname()
	params, _ := vmctx.req.Params()
	vmctx.lastResult, vmctx.lastError = vmctx.callNonViewByProgramHash(
		targetContract, entryPoint, params, vmctx.remainingAfterFees, vmctx.contractRecord.ProgramHash)
}

func (vmctx *VMContext) mustUpdateOffledgerRequestMaxAssumedNonce() {
	if offl, ok := vmctx.req.(*request.OffLedger); ok {
		vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
		defer vmctx.popCallContext()

		accounts.RecordMaxAssumedNonce(vmctx.State(), offl.SenderAddress(), offl.Nonce())
	}
}

func (vmctx *VMContext) mustFinalizeRequestCall() {
	if vmctx.exceededBlockOutputLimit {
		return
	}
	vmctx.mustLogRequestToBlockLog(vmctx.lastError) // panic not caught
	vmctx.lastTotalAssets = vmctx.totalAssets()

	vmctx.virtualState.ApplyStateUpdates(vmctx.currentStateUpdate)
	vmctx.currentStateUpdate = nil

	_, ep := vmctx.req.Target()
	vmctx.log.Debug("runTheRequest OUT. ",
		"reqId: ", vmctx.req.ID().Short(),
		" entry point: ", ep.String(),
	)
}

// getChainConfigFromState only makes sense if chain is already deployed
func (vmctx *VMContext) getChainConfigFromState() {
	cfg := vmctx.getChainInfo()
	if !cfg.ChainID.Equals(vmctx.chainID) {
		vmctx.log.Panicf("getChainConfigFromState: major inconsistency of chainID")
	}
	vmctx.chainOwnerID = cfg.ChainOwnerID
	vmctx.maxEventSize = cfg.MaxEventSize
	vmctx.maxEventsPerReq = cfg.MaxEventsPerReq
	vmctx.feeColor, vmctx.ownerFee, vmctx.validatorFee = vmctx.getFeeInfo()
}

func (vmctx *VMContext) isInitChainRequest() bool {
	targetContract, entryPoint := vmctx.req.Target()
	return targetContract == root.Contract.Hname() && entryPoint == iscp.EntryPointInit
}

func isRequestTimeLockedNow(req iscp.Request, nowis time.Time) bool {
	if req.IsOffLedger() {
		return false
	}
	if req.(*request.OnLedger).TimeLock().IsZero() {
		return false
	}
	return req.(*request.OnLedger).TimeLock().After(nowis)
}
