package vmcontext

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/iotaledger/wasp/packages/util"

	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"

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

// RunTheRequestOld processes each iscp.RequestData in the batch
func (vmctx *VMContext) RunTheRequest(req iscp.RequestData, requestIndex uint16) {
	// prepare context for the request
	vmctx.req = req
	vmctx.requestIndex = requestIndex
	vmctx.requestEventIndex = 0
	vmctx.entropy = hashing.HashData(vmctx.entropy[:])
	vmctx.callStack = vmctx.callStack[:0]

	if err := vmctx.checkReasonToSkip(); err != nil {
		vmctx.Log().Warnf("request skipped (ignored) by the VM: %s, reason: %v", req.Request().ID().String(), err)
		return
	}
	vmctx.loadChainConfig()
	vmctx.locateTargetContract()

	// at this point state update is empty
	// so far there ware no panics except optimistic reader
	// No prepare state update (buffer) for mutations and panics
	// TODO start handling panics

	txsnapshot := vmctx.createTxBuilderSnapshot()
	vmctx.currentStateUpdate = state.NewStateUpdate(vmctx.virtualState.Timestamp().Add(1 * time.Nanosecond))

	if err := vmctx.creditDepositToChain(); err != nil {
		// have to skip the request. Rollback
		vmctx.restoreTxBuilderSnapshot(txsnapshot)
		vmctx.currentStateUpdate = nil
		vmctx.Log().Warnf("request skipped due to not being able to accrue deposit: %s, reason: %v",
			req.Request().ID().String(), err)
		return
	}
	// create new transaction snapshot
	txsnapshot = vmctx.createTxBuilderSnapshot()
	// apply state update to the state because at this point it is consistent
	// If further on panic will occur, the chain will be left with all assets on the sender's account
	vmctx.virtualState.ApplyStateUpdates(vmctx.currentStateUpdate)
	vmctx.currentStateUpdate = state.NewStateUpdate()

	// TODO
	vmctx.callTheContract()

	// TODO call SC, handle VM panics, finalize the request
}

func (vmctx *VMContext) callTheContract() error {
	// handle panics, call the contract
}

// creditDepositToChain credits L1 accounts with attached assets and accrues all of them to the sender's account on-chain
func (vmctx *VMContext) creditDepositToChain() error {
	if vmctx.req.Type() == iscp.TypeOffLedger {
		// off ledger requests does not bring any deposit
		return nil
	}
	// catches panics in txbuilder
	err := util.CatchPanicReturnError(func() {
		// update transaction builder
		vmctx.txbuilder.AddDeltaIotas(vmctx.req.Request().Assets().Iotas)
		for _, nt := range vmctx.req.Request().Assets().Tokens {
			vmctx.txbuilder.AddDeltaNativeToken(nt.ID, nt.Amount)
		}
		// sender account will be CommonAccount if sender address is not available
		vmctx.creditToAccount(vmctx.req.Request().SenderAccount(), vmctx.req.Request().Assets())
	}, vmtxbuilder.ErrInputLimitExceeded, vmtxbuilder.ErrOutputLimitExceeded)

	return err
}

func (vmctx *VMContext) locateTargetContract() {
	// find target contract
	targetContract := vmctx.req.Request().Target().Contract
	var ok bool
	vmctx.contractRecord, ok = vmctx.findContractByHname(targetContract)
	if !ok {
		vmctx.Log().Warnf("contract not found: %s", targetContract)
	}
	if vmctx.contractRecord.Hname() == 0 {
		vmctx.Log().Warn("default contract will be called")
	}
}

// loadChainConfig only makes sense if chain is already deployed
func (vmctx *VMContext) loadChainConfig() {
	if vmctx.isInitChainRequest() {
		vmctx.chainOwnerID = vmctx.req.Request().SenderAccount()
		return
	}
	cfg := vmctx.getChainInfo()
	vmctx.chainOwnerID = cfg.ChainOwnerID
	vmctx.maxEventSize = cfg.MaxEventSize
	vmctx.maxEventsPerReq = cfg.MaxEventsPerReq
	//vmctx.feeColor, vmctx.ownerFee, vmctx.validatorFee = vmctx.getFeeInfo()  // TODO fee policy
}

func (vmctx *VMContext) isInitChainRequest() bool {
	target := vmctx.req.Request().Target()
	return target.Contract == root.Contract.Hname() && target.EntryPoint == iscp.EntryPointInit
}

// new code from here up
//======================================================================================
// deprecated code from here down

// RunTheRequestOld processes each iscp.RequestData in the batch
func (vmctx *VMContext) RunTheRequestOld(req iscp.RequestData, requestIndex uint16) {

	if req.Unwrap().UTXO() != nil && vmctx.txbuilder.InputsAreFull() {
		// ignore the UTXO request. Exceeded limit of input in the anchorOutput transaction
		return
	}

	defer vmctx.mustFinalizeRequestCall()

	snap := vmctx.createTxBuilderSnapshot()
	if vmctx.preprocessRequestData(req, requestIndex) {
		// request does not require invocation of smart contract
		return
	}

	// snapshot tx builder to be able to rollback and not include this request in the current block

	// guard against replaying off-ledger requests here to prevent replaying fee deduction
	// also verifies that account for off-ledger request exists
	if !vmctx.validateRequest() {
		vmctx.log.Debugw("vmctx.RunTheRequestOld: request failed validation", "id", vmctx.req.ID())
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
	vmctx.createTxBuilderSnapshot(txBuilderSnapshotBeforeCall)
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
		vmctx.restoreTxBuilderSnapshot(txBuilderSnapshotWithoutInput)
		vmctx.currentStateUpdate = state.NewStateUpdate()
	}

	if vmctx.lastError != nil {
		// treating panic and error returned from request the same way
		// restore the txbuilder and dispose mutations in the last state update
		vmctx.restoreTxBuilderSnapshot(txBuilderSnapshotBeforeCall)
		vmctx.currentStateUpdate = state.NewStateUpdate()

		vmctx.mustSendBack(vmctx.remainingAfterFees)
	}
}

// mustSetUpRequestContextOld sets up VMContext for request
// Deprecated:
func (vmctx *VMContext) mustSetUpRequestContextOld(req iscp.RequestData, requestIndex uint16) {
	vmctx.req = req
	vmctx.requestIndex = requestIndex
	vmctx.requestEventIndex = 0
	vmctx.requestOutputCount = 0
	vmctx.exceededBlockOutputLimit = false

	if !req.IsOffLedger() {
		vmctx.txBuilder.AddConsumable(vmctx.req.(*request.OnLedger).Output())
		if err := vmctx.txBuilder.ConsumeInputByOutputID(req.(*request.OnLedger).Output().ID()); err != nil {
			vmctx.log.Panicf("mustSetUpRequestContextOld.inconsistency : %v", err)
		}
	}
	ts := vmctx.virtualState.Timestamp().Add(1 * time.Nanosecond)
	vmctx.currentStateUpdate = state.NewStateUpdate(ts)

	vmctx.entropy = hashing.HashData(vmctx.entropy[:])
	vmctx.callStack = vmctx.callStack[:0]

	if isRequestTimeLockedNow(req, ts) {
		vmctx.log.Panicf("mustSetUpRequestContextOld.inconsistency: input is time locked. Nowis: %v\nInput: %s\n", ts, req.ID().String())
	}
	if !req.IsOffLedger() {
		// on-ledger request
		reqt := req.(*request.OnLedger)
		if input, ok := reqt.Output().(*ledgerstate.ExtendedLockedOutput); ok {
			// it is an on-ledger request
			if !input.UnlockAddressNow(ts).Equals(vmctx.chainID.AsAddress()) {
				vmctx.log.Panicf("mustSetUpRequestContextOld.inconsistency: input cannot be unlocked at %v.\nInput: %s\n chainID: %s",
					ts, input.String(), vmctx.chainID.String())
			}
		} else {
			vmctx.log.Panicf("mustSetUpRequestContextOld.inconsistency: unexpected UTXO type")
		}
		vmctx.remainingAfterFees = colored.BalancesFromL1Balances(reqt.Output().Balances())
	} else {
		// off-ledger request
		vmctx.remainingAfterFees = vmctx.adjustOffLedgerTransfer()
	}

	targetContract := req.Target().Contract
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
	entryPoint := vmctx.req.Target().EntryPoint
	targetContract := vmctx.contractRecord.Hname()
	vmctx.lastResult, vmctx.lastError = vmctx.callNonViewByProgramHash(
		targetContract, entryPoint, vmctx.req.Args(), vmctx.remainingAfterFees, vmctx.contractRecord.ProgramHash)
}

func (vmctx *VMContext) mustUpdateOffledgerRequestMaxAssumedNonce() {
	if offl, ok := vmctx.req.(*request.OffLedger); ok {
		vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
		defer vmctx.popCallContext()

		accounts.RecordMaxAssumedNonce(vmctx.State(), offl.SenderAddress(), offl.Nonce())
	}
}

func (vmctx *VMContext) mustFinalizeRequestCall() {
	vmctx.clearTxBuilderSnapshots()

	if vmctx.exceededBlockOutputLimit {
		return
	}
	vmctx.mustLogRequestToBlockLog(vmctx.lastError) // panic not caught
	vmctx.lastTotalAssets = vmctx.totalAssets()

	vmctx.virtualState.ApplyStateUpdates(vmctx.currentStateUpdate)
	vmctx.currentStateUpdate = nil

	vmctx.log.Debug("runTheRequest OUT. ",
		"reqId: ", vmctx.req.ID().Short(),
		" entry point: ", vmctx.req.Target().EntryPoint.String(),
	)
}
