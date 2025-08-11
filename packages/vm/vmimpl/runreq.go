package vmimpl

import (
	"math"
	"os"
	"runtime/debug"
	"time"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/buffered"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/util/panicutil"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
	"github.com/iotaledger/wasp/v2/packages/vm/processors"
	"github.com/iotaledger/wasp/v2/packages/vm/vmexceptions"
)

// runRequest processes a single isc.Request in the batch, returning an error means the request will be skipped
func (vmctx *vmContext) runRequest(req isc.Request, requestIndex uint16, maintenanceMode bool) (
	res *vm.RequestResult,
	err error,
) {
	reqctx := vmctx.newRequestContext(req, requestIndex)

	if vmctx.task.EnableGasBurnLogging {
		reqctx.gas.burnLog = gas.NewGasBurnLog()
	}

	initialGasBurnedTotal := vmctx.blockGas.burned
	initialGasFeeChargedTotal := vmctx.blockGas.feeCharged

	reqctx.uncommittedState.Set(
		kv.Key(coreutil.StatePrefixTimestamp),
		codec.Encode(vmctx.stateDraft.Timestamp().Add(1*time.Nanosecond)),
	)

	if err = reqctx.earlyCheckReasonToSkip(maintenanceMode); err != nil {
		return nil, err
	}
	vmctx.loadChainConfig()

	// at this point state update is empty
	// so far there were no panics except optimistic reader
	txsnapshot := vmctx.createTxBuilderSnapshot()

	result, err := reqctx.callTheContract()
	if err == nil {
		err = vmctx.txbuilder.CheckTransactionSize()
	}
	if err != nil {
		// skip the request / rollback tx builder (no need to rollback the state, because the mutations will never be applied)
		vmctx.restoreTxBuilderSnapshot(txsnapshot)
		vmctx.blockGas.burned = initialGasBurnedTotal
		vmctx.blockGas.feeCharged = initialGasFeeChargedTotal
		return nil, err
	}

	reqctx.uncommittedState.Mutations().ApplyTo(vmctx.stateDraft)
	return result, nil
}

func (vmctx *vmContext) newRequestContext(req isc.Request, requestIndex uint16) *requestContext {
	reqctx := &requestContext{
		vm:               vmctx,
		req:              req,
		requestIndex:     requestIndex,
		entropy:          hashing.HashData(append(codec.Encode(requestIndex), vmctx.task.Entropy[:]...)),
		uncommittedState: buffered.NewBufferedKVStore(vmctx.stateDraft),
	}
	if vmctx.task.EnforceGasBurned != nil {
		reqctx.gas.enforceGasBurned = &vmctx.task.EnforceGasBurned[requestIndex]
	}
	return reqctx
}

func (vmctx *vmContext) payoutAgentID() isc.AgentID {
	return governance.NewStateReaderFromChainState(vmctx.stateDraft).GetPayoutAgentID()
}

// consumeRequest Consumes incoming request and updating the sender's L2 Balance according to the current request.
// In the Iota ver impl, For L1's perspective ISC anchor credits all the assets attached on the
// requests, when calling txbuilder.BuildTransactionEssence
func (reqctx *requestContext) consumeRequest() {
	req, ok := reqctx.req.(isc.OnLedgerRequest)
	if !ok {
		// off ledger request does not bring any deposit
		return
	}
	reqctx.vm.task.Log.LogDebugf("consumeRequest: %s with %s", req.ID(), req.Assets())
	reqctx.vm.txbuilder.ConsumeRequest(req)

	// if sender is specified, all assets goes to sender's sender
	// Otherwise it all goes to the common sender and panic is logged in the SC call
	sender := req.SenderAccount()
	if sender == nil {
		// should not happen, but just in case...
		// onledger request with no sender, send all assets to the payoutAddress
		payoutAgentID := reqctx.vm.payoutAgentID()
		reqctx.creditObjectsToAccount(payoutAgentID, req.Assets().Objects.Sorted())
		reqctx.creditToAccount(payoutAgentID, req.Assets().Coins)
		return
	}

	senderBaseTokens := req.Assets().BaseTokens() + reqctx.GetBaseTokensBalanceDiscardRemainder(sender)

	// check if the sender has enough balance to cover the minimum gas fee
	if reqctx.shouldChargeGasFee() {
		minReqCost := reqctx.ChainInfo().GasFeePolicy.MinFee(isc.RequestGasPrice(reqctx.req), parameters.BaseTokenDecimals)
		if senderBaseTokens < minReqCost {
			// TODO: this should probably not skip the request, and also the check
			// should be done in L1 so the request is rejected before it reaches the mempool
			panic(vmexceptions.ErrNotEnoughFundsForMinFee)
		}
	}

	reqctx.creditObjectsToAccount(sender, req.Assets().Objects.Sorted())
	reqctx.creditToAccount(sender, req.Assets().Coins)
}

// checkAllowance panics if the allowance is invalid or there are not enough
// funds to cover it
func (reqctx *requestContext) checkAllowance() {
	allowance, err := reqctx.req.Allowance()
	if err != nil {
		panic(vm.ErrInvalidAllowance)
	}
	if !reqctx.HasEnoughForAllowance(reqctx.req.SenderAccount(), allowance) {
		panic(vm.ErrNotEnoughFundsForAllowance)
	}
}

func (reqctx *requestContext) shouldChargeGasFee() bool {
	// freeGasPerToken checks whether we charge token per gas
	// If it is free, then we will still burn the gas, but it doesn't charge tokens
	// NOT FOR PUBLIC NETWORK
	var freeGasPerToken bool
	reqctx.callCore(governance.Contract, func(s kv.KVStore) {
		gasPerToken := governance.NewStateReader(s).GetGasFeePolicy().GasPerToken
		freeGasPerToken = gasPerToken.A == 0 && gasPerToken.B == 0
	})
	if freeGasPerToken {
		return false
	}
	if reqctx.req.SenderAccount() == nil {
		return false
	}
	if reqctx.req.SenderAccount().Equals(reqctx.vm.ChainAdmin()) && reqctx.req.Message().Target.Contract == governance.Contract.Hname() {
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
	// pre execution ---------------------------------------------------------------
	err := panicutil.CatchPanic(func() {
		// transfer all attached assets to the sender's account
		reqctx.consumeRequest()
		// load gas and fee policy, calculate and set gas budget
		reqctx.prepareGasBudget()
		// run the contract program
	})
	if err != nil {
		// this should never happen. something is wrong here, SKIP the request
		reqctx.vm.task.Log.LogErrorf("panic before request execution (reqid: %s): %v", reqctx.req.ID(), err)
		return nil, err
	}

	// execution ---------------------------------------------------------------

	result := &vm.RequestResult{Request: reqctx.req}

	txSnapshot := reqctx.vm.createTxBuilderSnapshot() // take the txbuilder snapshot **after** the request has been consumed (in `consumeRequest`)
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
			if os.Getenv("DEBUG") != "" || reqctx.vm.task.WillProduceBlock() {
				reqctx.Debugf(string(debug.Stack()))
			}
		}()
		// ensure there are enough funds to cover the specified allowance
		reqctx.checkAllowance()

		reqctx.GasBurnEnable(true)
		result.Return = reqctx.callFromRequest()

		if reqctx.gas.enforceGasBurned != nil {
			// this is a stardust request being traced; we must make sure that we charge
			// the exact same amount of gas units
			if reqctx.gas.enforceGasBurned.Error != nil {
				panic(reqctx.gas.enforceGasBurned.Error)
			}
			reqctx.gas.burned = reqctx.gas.enforceGasBurned.GasBurned
		}

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
		reqctx.vm.task.Log.LogErrorf("panic after request execution (reqid: %s, executionErr: %s): %v", reqctx.req.ID(), callErrStr, err)
		reqctx.vm.task.Log.LogDebug(string(debug.Stack()))
		return nil, err
	}

	return result, nil
}

func recoverFromExecutionError(r any) *isc.VMError {
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
func (reqctx *requestContext) callFromRequest() isc.CallArguments {
	req := reqctx.req
	reqctx.Debugf("callFromRequest: %s", req.ID().String())

	if req.SenderAccount() == nil {
		// if sender unknown, follow panic path
		panic(vm.ErrSenderUnknown)
	}

	allowance, err := req.Allowance()
	if err != nil {
		panic(vm.ErrInvalidAllowance)
	}

	return reqctx.callProgram(req.Message(), allowance, req.SenderAccount())
}

func (reqctx *requestContext) getGasBudget() uint64 {
	gasBudget, isEVM := reqctx.req.GasBudget()
	if !isEVM || gasBudget == 0 {
		return gasBudget
	}

	var gasRatio util.Ratio32
	reqctx.callCore(governance.Contract, func(s kv.KVStore) {
		gasRatio = governance.NewStateReader(s).GetGasFeePolicy().EVMGasRatio
	})
	return gas.EVMGasToISC(gasBudget, &gasRatio)
}

// calculateAffordableGasBudget checks the account of the sender and calculates affordable gas budget
// Affordable gas budget is calculated from gas budget provided in the request by the user and taking into account
// how many tokens the sender has in its account and how many are allowed for the target.
// Safe arithmetics is used
func (reqctx *requestContext) calculateAffordableGasBudget() (budget uint64, maxTokensToSpendForGasFee coin.Value) {
	gasBudget := reqctx.getGasBudget()

	if reqctx.vm.task.EstimateGasMode && gasBudget == 0 {
		// gas budget 0 means its a view call, so we give it max gas and tokens
		return reqctx.vm.chainInfo.GasLimits.MaxGasExternalViewCall, math.MaxUint64
	}

	// make sure the gasBudget is at least >= than the allowed minimum
	if gasBudget < reqctx.vm.chainInfo.GasLimits.MinGasPerRequest {
		gasBudget = reqctx.vm.chainInfo.GasLimits.MinGasPerRequest
	}

	// make sure the gasBudget is less than the allowed maximum
	if gasBudget > reqctx.vm.chainInfo.GasLimits.MaxGasPerRequest {
		gasBudget = reqctx.vm.chainInfo.GasLimits.MaxGasPerRequest
	}

	if reqctx.vm.task.EstimateGasMode {
		return gasBudget, math.MaxUint64
	}

	// calculate how many tokens for gas fee can be guaranteed after taking into account the allowance
	guaranteedFeeTokens := reqctx.calcGuaranteedFeeTokens()
	// calculate how many tokens maximum will be charged taking into account the budget
	gasPrice := isc.RequestGasPrice(reqctx.req)
	f1, f2 := reqctx.vm.chainInfo.GasFeePolicy.FeeFromGasBurned(
		gasBudget,
		guaranteedFeeTokens,
		gasPrice,
		parameters.BaseTokenDecimals,
	)
	maxTokensToSpendForGasFee = f1 + f2
	// calculate affordableGas gas budget
	var affordableGas uint64
	if gasPrice != nil {
		affordableGas = reqctx.vm.chainInfo.GasFeePolicy.GasBudgetFromTokensWithGasPrice(
			guaranteedFeeTokens,
			gasPrice,
			parameters.BaseTokenDecimals,
		)
	} else {
		affordableGas = reqctx.vm.chainInfo.GasFeePolicy.GasBudgetFromTokens(guaranteedFeeTokens)
	}
	// adjust gas budget to what is affordable
	affordableGas = min(gasBudget, affordableGas)
	// cap gas to the maximum allowed per tx
	return affordableGas, maxTokensToSpendForGasFee
}

// calcGuaranteedFeeTokens return the maximum tokens (base tokens or native) can be guaranteed for the fee,
// taking into account allowance (which must be 'reserved')
func (reqctx *requestContext) calcGuaranteedFeeTokens() coin.Value {
	tokensGuaranteed := reqctx.GetBaseTokensBalanceDiscardRemainder(reqctx.req.SenderAccount())

	// safely subtract the allowed from the sender to the target
	allowance, err := reqctx.req.Allowance()
	if err != nil {
		reqctx.vm.task.Log.LogDebugf("error decoding allowance: %s", err.Error())
		// ignore the error, since we cannot panic here. The request will fail
		// anyway when checkAllowance is called.
		allowance = isc.NewEmptyAssets()
	}

	allowed := allowance.BaseTokens()
	if tokensGuaranteed < allowed {
		tokensGuaranteed = 0
	} else {
		tokensGuaranteed -= allowed
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
	gasPrice := isc.RequestGasPrice(reqctx.req)
	if !reqctx.vm.task.EstimateGasMode && !reqctx.vm.chainInfo.GasFeePolicy.IsEnoughForMinimumFee(
		availableToPayFee,
		gasPrice,
		parameters.BaseTokenDecimals,
	) {
		// user didn't specify enough base tokens to cover the minimum request fee, charge whatever is present in the user's account
		availableToPayFee = reqctx.GetSenderTokenBalanceForFees()
	}

	// total fees to charge
	sendToPayout, sendToValidator := reqctx.vm.chainInfo.GasFeePolicy.FeeFromGasBurned(
		reqctx.GasBurned(),
		availableToPayFee,
		gasPrice,
		parameters.BaseTokenDecimals,
	)
	reqctx.gas.feeCharged = sendToPayout + sendToValidator

	// calc gas totals
	reqctx.vm.blockGas.feeCharged += reqctx.gas.feeCharged

	if reqctx.vm.task.EstimateGasMode {
		// If estimating gas, compute the gas fee but do not attempt to charge
		return
	}

	sender := reqctx.req.SenderAccount()
	if sendToValidator != 0 {
		reqctx.mustMoveBetweenAccounts(
			sender,
			reqctx.vm.task.ValidatorFeeTarget,
			isc.NewAssets(sendToValidator),
			false,
		)
	}

	// ensure common account has at least GasCoinTargetValue, and transfer the rest of gas fee to payout AgentID
	// if the payout AgentID is not set in governance contract, then chain admin will be used
	targetCommonAccountBalance := governance.NewStateReaderFromChainState(reqctx.uncommittedState).GetGasCoinTargetValue()
	commonAccountBal := reqctx.GetBaseTokensBalanceDiscardRemainder(accounts.CommonAccount())
	reqctx.vm.task.Log.LogDebugf("common account balance: %d, targetCommonAccountBalance: %d", commonAccountBal, targetCommonAccountBalance)

	if commonAccountBal < targetCommonAccountBalance {
		// pay to common account since the balance of common account is less than min
		transferToCommonAcc := sendToPayout
		sendToPayout = 0
		if commonAccountBal+transferToCommonAcc > targetCommonAccountBalance {
			excess := (commonAccountBal + transferToCommonAcc) - targetCommonAccountBalance
			transferToCommonAcc -= excess
			sendToPayout = excess
		}
		reqctx.vm.task.Log.LogDebugf("transferring %d to common account", transferToCommonAcc)
		reqctx.mustMoveBetweenAccounts(
			sender,
			accounts.CommonAccount(),
			isc.NewAssets(transferToCommonAcc),
			false,
		)
	}
	if sendToPayout > 0 {
		payoutAgentID := reqctx.vm.payoutAgentID()
		reqctx.mustMoveBetweenAccounts(
			sender,
			payoutAgentID,
			isc.NewAssets(sendToPayout),
			false,
		)
	}
}

func (reqctx *requestContext) Processors() *processors.Config {
	return reqctx.vm.task.Processors
}

func (reqctx *requestContext) GetContractRecord(contractHname isc.Hname) (ret *root.ContractRecord) {
	ret = findContractByHname(reqctx.chainStateWithGasBurn(), contractHname)
	if ret == nil {
		reqctx.GasBurn(gas.BurnCodeCallTargetNotFound)
		panic(vm.ErrContractNotFound.Create(uint32(contractHname)))
	}
	return ret
}

func (vmctx *vmContext) loadChainConfig() {
	vmctx.chainInfo = governance.NewStateReaderFromChainState(vmctx.stateDraft).GetChainInfo()
}
