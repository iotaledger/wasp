package vmimpl

import (
	"math"
	"runtime/debug"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
)

// runRequest processes a single isc.Request in the batch, returning an error means the request will be skipped
func (vmctx *vmContext) runRequest(req isc.Request, requestIndex uint16, maintenanceMode bool) (
	res *vm.RequestResult,
	unprocessableToRetry []isc.OnLedgerRequest,
	err error,
) {
	reqctx := &requestContext{
		vm:               vmctx,
		req:              req,
		requestIndex:     requestIndex,
		entropy:          hashing.HashData(append(codec.EncodeUint16(requestIndex), vmctx.task.Entropy[:]...)),
		uncommittedState: buffered.NewBufferedKVStore(vmctx.stateDraft),
	}

	if vmctx.task.EnableGasBurnLogging {
		reqctx.gas.burnLog = gas.NewGasBurnLog()
	}

	initialGasBurnedTotal := vmctx.blockGas.burned
	initialGasFeeChargedTotal := vmctx.blockGas.feeCharged

	reqctx.uncommittedState.Set(
		kv.Key(coreutil.StatePrefixTimestamp),
		codec.EncodeTime(vmctx.stateDraft.Timestamp().Add(1*time.Nanosecond)),
	)

	if err = reqctx.earlyCheckReasonToSkip(maintenanceMode); err != nil {
		return nil, nil, err
	}
	vmctx.loadChainConfig()

	// at this point state update is empty
	// so far there were no panics except optimistic reader
	txsnapshot := vmctx.createTxBuilderSnapshot()

	result, err := reqctx.callTheContract()
	if err == nil {
		err = vmctx.checkTransactionSize()
	}
	if err != nil {
		// skip the request / rollback tx builder (no need to rollback the state, because the mutations will never be applied)
		vmctx.restoreTxBuilderSnapshot(txsnapshot)
		vmctx.blockGas.burned = initialGasBurnedTotal
		vmctx.blockGas.feeCharged = initialGasFeeChargedTotal
		return nil, nil, err
	}

	reqctx.uncommittedState.Mutations().ApplyTo(vmctx.stateDraft)
	return result, reqctx.unprocessableToRetry, nil
}

func (vmctx *vmContext) payoutAgentID() isc.AgentID {
	var payoutAgentID isc.AgentID
	withContractState(vmctx.stateDraft, governance.Contract, func(s kv.KVStore) {
		payoutAgentID = governance.MustGetPayoutAgentID(s)
	})
	return payoutAgentID
}

// creditAssetsToChain credits L1 accounts with attached assets and accrues all of them to the sender's account on-chain
func (reqctx *requestContext) creditAssetsToChain() {
	req, ok := reqctx.req.(isc.OnLedgerRequest)
	if !ok {
		// off ledger request does not bring any deposit
		return
	}
	// Consume the output. Adjustment in L2 is needed because of storage deposit in the internal UTXOs
	storageDepositNeeded := reqctx.vm.txbuilder.Consume(req)

	// if sender is specified, all assets goes to sender's sender
	// Otherwise it all goes to the common sender and panic is logged in the SC call
	sender := req.SenderAccount()
	if sender == nil {
		if storageDepositNeeded > req.Assets().BaseTokens {
			panic(vmexceptions.ErrNotEnoughFundsForSD) // if sender is not specified, and extra tokens are needed to pay for SD, the request cannot be processed.
		}
		// onleger request with no sender, send all assets to the payoutAddress
		payoutAgentID := reqctx.vm.payoutAgentID()
		creditNFTToAccount(reqctx.uncommittedState, payoutAgentID, req, reqctx.ChainID())
		creditToAccount(reqctx.uncommittedState, payoutAgentID, req.Assets(), reqctx.ChainID())
		if storageDepositNeeded > 0 {
			debitFromAccount(reqctx.uncommittedState, payoutAgentID, isc.NewAssetsBaseTokens(storageDepositNeeded), reqctx.ChainID())
		}
		return
	}

	senderBaseTokens := req.Assets().BaseTokens + reqctx.GetBaseTokensBalance(sender)

	minReqCost := reqctx.ChainInfo().GasFeePolicy.MinFee()
	if senderBaseTokens < storageDepositNeeded+minReqCost {
		// user doesn't have enough funds to pay for the SD needs of this request
		panic(vmexceptions.ErrNotEnoughFundsForSD)
	}

	creditToAccount(reqctx.uncommittedState, sender, req.Assets(), reqctx.ChainID())
	creditNFTToAccount(reqctx.uncommittedState, sender, req, reqctx.ChainID())
	if storageDepositNeeded > 0 {
		reqctx.sdCharged = storageDepositNeeded
		debitFromAccount(reqctx.uncommittedState, sender, isc.NewAssetsBaseTokens(storageDepositNeeded), reqctx.ChainID())
	}
}

// checkAllowance ensure there are enough funds to cover the specified allowance
// panics if not enough funds
func (reqctx *requestContext) checkAllowance() {
	if !reqctx.HasEnoughForAllowance(reqctx.req.SenderAccount(), reqctx.req.Allowance()) {
		panic(vm.ErrNotEnoughFundsForAllowance)
	}
}

func (reqctx *requestContext) shouldChargeGasFee() bool {
	// freeGasPerToken checks whether we charge token per gas
	// If it is free, then we will still burn the gas, but it doesn't charge tokens
	// NOT FOR PUBLIC NETWORK
	var freeGasPerToken bool
	reqctx.callCore(governance.Contract, func(s kv.KVStore) {
		gasPerToken := governance.MustGetGasFeePolicy(s).GasPerToken
		freeGasPerToken = gasPerToken.A == 0 && gasPerToken.B == 0
	})
	if freeGasPerToken {
		return false
	}
	if reqctx.req.SenderAccount() == nil {
		return false
	}
	if reqctx.req.SenderAccount().Equals(reqctx.vm.ChainOwnerID()) && reqctx.req.CallTarget().Contract == governance.Contract.Hname() {
		return false
	}
	return true
}

func (reqctx *requestContext) prepareGasBudget() {
	if !reqctx.shouldChargeGasFee() {
		return
	}
	reqctx.gasSetBudget(reqctx.calculateAffordableGasBudget())
}

// callTheContract runs the contract. if an error is returned, the request will be skipped
func (reqctx *requestContext) callTheContract() (*vm.RequestResult, error) {
	// TODO: do not mutate vmContext's txbuilder

	// pre execution ---------------------------------------------------------------
	err := panicutil.CatchPanic(func() {
		// transfer all attached assets to the sender's account
		reqctx.creditAssetsToChain()
		// load gas and fee policy, calculate and set gas budget
		reqctx.prepareGasBudget()
		// run the contract program
	})
	if err != nil {
		// this should never happen. something is wrong here, SKIP the request
		reqctx.vm.task.Log.Errorf("panic before request execution (reqid: %s): %v", reqctx.req.ID(), err)
		return nil, err
	}

	// execution ---------------------------------------------------------------

	result := &vm.RequestResult{Request: reqctx.req}

	txSnapshot := reqctx.vm.createTxBuilderSnapshot() // take the txbuilder snapshot **after** the request has been consumed (in `creditAssetsToChain`)
	stateSnapshot := reqctx.uncommittedState.Clone()

	rollback := func() {
		reqctx.vm.restoreTxBuilderSnapshot(txSnapshot)
		reqctx.uncommittedState = stateSnapshot
	}

	var executionErr *isc.VMError
	var skipRequestErr error
	func() {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			skipRequestErr = vmexceptions.IsSkipRequestException(r)
			executionErr = recoverFromExecutionError(r)
			reqctx.Debugf("recovered panic from contract call: %v", executionErr)
			if reqctx.vm.task.WillProduceBlock() {
				reqctx.Debugf(string(debug.Stack()))
			}
		}()
		// ensure there are enough funds to cover the specified allowance
		reqctx.checkAllowance()

		reqctx.GasBurnEnable(true)
		result.Return = reqctx.callFromRequest()
		// ensure at least the minimum amount of gas is charged
		reqctx.GasBurn(gas.BurnCodeMinimumGasPerRequest1P, reqctx.GasBurned())
	}()
	reqctx.GasBurnEnable(false)
	if skipRequestErr != nil {
		return nil, skipRequestErr
	}

	// post execution ---------------------------------------------------------------

	// execution over, save receipt, update nonces, etc
	// if anything goes wrong here, state must be rolled back and the request must be skipped
	err = panicutil.CatchPanic(func() {
		if executionErr != nil {
			// panic happened during VM plugin call. Restore the state
			rollback()
		}
		// charge gas fee no matter what
		reqctx.chargeGasFee()

		// write receipt no matter what
		result.Receipt = reqctx.writeReceiptToBlockLog(executionErr)

		if reqctx.req.IsOffLedger() {
			reqctx.updateOffLedgerRequestNonce()
		}
	})

	if err != nil {
		rollback()
		callErrStr := ""
		if executionErr != nil {
			callErrStr = executionErr.Error()
		}
		reqctx.vm.task.Log.Errorf("panic after request execution (reqid: %s, executionErr: %s): %v", reqctx.req.ID(), callErrStr, err)
		reqctx.vm.task.Log.Debug(string(debug.Stack()))
		return nil, err
	}

	return result, nil
}

func recoverFromExecutionError(r interface{}) *isc.VMError {
	switch err := r.(type) {
	case *isc.VMError:
		return r.(*isc.VMError)
	case isc.VMError:
		e := r.(isc.VMError)
		return &e
	case *kv.DBError:
		panic(err)
	case string:
		return coreerrors.ErrUntypedError.Create(err)
	case error:
		return coreerrors.ErrUntypedError.Create(err.Error())
	}
	return nil
}

// callFromRequest is the call itself. Assumes sc exists
func (reqctx *requestContext) callFromRequest() dict.Dict {
	req := reqctx.req
	reqctx.Debugf("callFromRequest: %s", req.ID().String())

	if req.SenderAccount() == nil {
		// if sender unknown, follow panic path
		panic(vm.ErrSenderUnknown)
	}

	contract := req.CallTarget().Contract
	entryPoint := req.CallTarget().EntryPoint

	return reqctx.callProgram(
		contract,
		entryPoint,
		req.Params(),
		req.Allowance(),
		req.SenderAccount(),
	)
}

func (reqctx *requestContext) getGasBudget() uint64 {
	gasBudget, isEVM := reqctx.req.GasBudget()
	if !isEVM || gasBudget == 0 {
		return gasBudget
	}

	var gasRatio util.Ratio32
	reqctx.callCore(governance.Contract, func(s kv.KVStore) {
		gasRatio = governance.MustGetGasFeePolicy(s).EVMGasRatio
	})
	return gas.EVMGasToISC(gasBudget, &gasRatio)
}

// calculateAffordableGasBudget checks the account of the sender and calculates affordable gas budget
// Affordable gas budget is calculated from gas budget provided in the request by the user and taking into account
// how many tokens the sender has in its account and how many are allowed for the target.
// Safe arithmetics is used
func (reqctx *requestContext) calculateAffordableGasBudget() (budget, maxTokensToSpendForGasFee uint64) {
	gasBudget := reqctx.getGasBudget()

	if reqctx.vm.task.EstimateGasMode && gasBudget == 0 {
		// gas budget 0 means its a view call, so we give it max gas and tokens
		return reqctx.vm.chainInfo.GasLimits.MaxGasExternalViewCall, math.MaxUint64
	}

	// make sure the gasBudget is at least >= than the allowed minimum
	if gasBudget < reqctx.vm.chainInfo.GasLimits.MinGasPerRequest {
		gasBudget = reqctx.vm.chainInfo.GasLimits.MinGasPerRequest
	}

	if reqctx.vm.task.EstimateGasMode {
		return gasBudget, math.MaxUint64
	}

	// calculate how many tokens for gas fee can be guaranteed after taking into account the allowance
	guaranteedFeeTokens := reqctx.calcGuaranteedFeeTokens()
	// calculate how many tokens maximum will be charged taking into account the budget
	f1, f2 := reqctx.vm.chainInfo.GasFeePolicy.FeeFromGasBurned(gasBudget, guaranteedFeeTokens)
	maxTokensToSpendForGasFee = f1 + f2
	// calculate affordableGas gas budget
	affordableGas := reqctx.vm.chainInfo.GasFeePolicy.GasBudgetFromTokens(guaranteedFeeTokens)
	// adjust gas budget to what is affordable
	affordableGas = min(gasBudget, affordableGas)
	// cap gas to the maximum allowed per tx
	return min(affordableGas, reqctx.vm.chainInfo.GasLimits.MaxGasPerRequest), maxTokensToSpendForGasFee
}

// calcGuaranteedFeeTokens return the maximum tokens (base tokens or native) can be guaranteed for the fee,
// taking into account allowance (which must be 'reserved')
func (reqctx *requestContext) calcGuaranteedFeeTokens() uint64 {
	tokensGuaranteed := reqctx.GetBaseTokensBalance(reqctx.req.SenderAccount())
	// safely subtract the allowed from the sender to the target
	if allowed := reqctx.req.Allowance(); allowed != nil {
		if tokensGuaranteed < allowed.BaseTokens {
			tokensGuaranteed = 0
		} else {
			tokensGuaranteed -= allowed.BaseTokens
		}
	}
	return tokensGuaranteed
}

// chargeGasFee takes burned tokens from the sender's account
// It should always be enough because gas budget is set affordable
func (reqctx *requestContext) chargeGasFee() {
	defer func() {
		// add current request gas burn to the total of the block
		reqctx.vm.blockGas.burned += reqctx.gas.burned
	}()

	// ensure at least the minimum amount of gas is charged
	minGas := gas.BurnCodeMinimumGasPerRequest1P.Cost(0)
	if reqctx.gas.burned < minGas {
		reqctx.gas.burned = minGas
	}

	if !reqctx.shouldChargeGasFee() {
		return
	}

	availableToPayFee := reqctx.gas.maxTokensToSpendForGasFee
	if !reqctx.vm.task.EstimateGasMode && !reqctx.vm.chainInfo.GasFeePolicy.IsEnoughForMinimumFee(availableToPayFee) {
		// user didn't specify enough base tokens to cover the minimum request fee, charge whatever is present in the user's account
		availableToPayFee = reqctx.GetSenderTokenBalanceForFees()
	}

	// total fees to charge
	sendToPayout, sendToValidator := reqctx.vm.chainInfo.GasFeePolicy.FeeFromGasBurned(reqctx.GasBurned(), availableToPayFee)
	reqctx.gas.feeCharged = sendToPayout + sendToValidator

	// calc gas totals
	reqctx.vm.blockGas.feeCharged += reqctx.gas.feeCharged

	if reqctx.vm.task.EstimateGasMode {
		// If estimating gas, compute the gas fee but do not attempt to charge
		return
	}

	sender := reqctx.req.SenderAccount()
	if sendToValidator != 0 {
		transferToValidator := &isc.Assets{}
		transferToValidator.BaseTokens = sendToValidator
		mustMoveBetweenAccounts(
			reqctx.uncommittedState,
			sender,
			reqctx.vm.task.ValidatorFeeTarget,
			transferToValidator,
			reqctx.ChainID(),
		)
	}

	// ensure common account has at least minBalanceInCommonAccount, and transfer the rest of gas fee to payout AgentID
	// if the payout AgentID is not set in governance contract, then chain owner will be used
	var minBalanceInCommonAccount uint64
	withContractState(reqctx.uncommittedState, governance.Contract, func(s kv.KVStore) {
		minBalanceInCommonAccount = governance.MustGetMinCommonAccountBalance(s)
	})
	commonAccountBal := reqctx.GetBaseTokensBalance(accounts.CommonAccount())
	if commonAccountBal < minBalanceInCommonAccount {
		// pay to common account since the balance of common account is less than minSD
		transferToCommonAcc := sendToPayout
		sendToPayout = 0
		if commonAccountBal+transferToCommonAcc > minBalanceInCommonAccount {
			excess := (commonAccountBal + transferToCommonAcc) - minBalanceInCommonAccount
			transferToCommonAcc -= excess
			sendToPayout = excess
		}
		mustMoveBetweenAccounts(reqctx.uncommittedState,
			sender,
			accounts.CommonAccount(),
			isc.NewAssetsBaseTokens(transferToCommonAcc),
			reqctx.ChainID(),
		)
	}
	if sendToPayout > 0 {
		payoutAgentID := reqctx.vm.payoutAgentID()
		mustMoveBetweenAccounts(
			reqctx.uncommittedState,
			sender,
			payoutAgentID,
			isc.NewAssetsBaseTokens(sendToPayout),
			reqctx.ChainID(),
		)
	}
}

func (reqctx *requestContext) LocateProgram(programHash hashing.HashValue) (vmtype string, binary []byte, err error) {
	return reqctx.vm.locateProgram(reqctx.chainStateWithGasBurn(), programHash)
}

func (reqctx *requestContext) Processors() *processors.Cache {
	return reqctx.vm.task.Processors
}

func (reqctx *requestContext) GetContractRecord(contractHname isc.Hname) (ret *root.ContractRecord) {
	ret = findContractByHname(reqctx.chainStateWithGasBurn(), contractHname)
	if ret == nil {
		reqctx.GasBurn(gas.BurnCodeCallTargetNotFound)
		panic(vm.ErrContractNotFound.Create(contractHname))
	}
	return ret
}

func (vmctx *vmContext) loadChainConfig() {
	vmctx.chainInfo = governance.NewStateAccess(vmctx.stateDraft).ChainInfo(vmctx.ChainID())
}

// checkTransactionSize panics with ErrMaxTransactionSizeExceeded if the estimated transaction size exceeds the limit
func (vmctx *vmContext) checkTransactionSize() error {
	essence, _ := vmctx.BuildTransactionEssence(state.L1CommitmentNil, false)
	tx := transaction.MakeAnchorTransaction(essence, &iotago.Ed25519Signature{})
	if tx.Size() > parameters.L1().MaxPayloadSize {
		return vmexceptions.ErrMaxTransactionSizeExceeded
	}
	return nil
}
