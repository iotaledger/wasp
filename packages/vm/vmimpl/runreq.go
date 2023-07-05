package vmimpl

import (
	"errors"
	"fmt"
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
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
)

// runRequest processes a single isc.Request in the batch
func (vmctx *vmContext) runRequest(req isc.Request, requestIndex uint16) (
	res *vm.RequestResult,
	unprocessableToRetry []isc.OnLedgerRequest,
	err error,
) {
	if len(vmctx.callStack) != 0 {
		panic("expected empty callstack")
	}

	vmctx.reqCtx = &requestContext{
		req:          req,
		requestIndex: requestIndex,
		entropy:      hashing.HashData(append(codec.EncodeUint16(requestIndex), vmctx.task.Entropy[:]...)),
	}
	defer func() { vmctx.reqCtx = nil }()

	if vmctx.task.EnableGasBurnLogging {
		vmctx.reqCtx.gas.burnLog = gas.NewGasBurnLog()
	}

	vmctx.GasBurnEnable(false)

	initialGasBurnedTotal := vmctx.blockGas.burned
	initialGasFeeChargedTotal := vmctx.blockGas.feeCharged

	if vmctx.currentStateUpdate != nil {
		panic("expected currentStateUpdate == nil")
	}
	vmctx.currentStateUpdate = buffered.NewMutations()
	defer func() { vmctx.currentStateUpdate = nil }()

	vmctx.chainState().Set(kv.Key(coreutil.StatePrefixTimestamp), codec.EncodeTime(vmctx.taskResult.StateDraft.Timestamp().Add(1*time.Nanosecond)))

	if err = vmctx.earlyCheckReasonToSkip(); err != nil {
		return nil, nil, err
	}
	vmctx.loadChainConfig()

	// at this point state update is empty
	// so far there were no panics except optimistic reader
	txsnapshot := vmctx.createTxBuilderSnapshot()

	var result *vm.RequestResult
	err = vmctx.catchRequestPanic(
		func() {
			// transfer all attached assets to the sender's account
			vmctx.creditAssetsToChain()
			// load gas and fee policy, calculate and set gas budget
			vmctx.prepareGasBudget()
			// run the contract program
			receipt, callRet := vmctx.callTheContract()
			vmctx.mustCheckTransactionSize()
			result = &vm.RequestResult{
				Request: req,
				Receipt: receipt,
				Return:  callRet,
			}
		},
	)
	if err != nil {
		// protocol exception triggered. Skipping the request. Rollback
		vmctx.restoreTxBuilderSnapshot(txsnapshot)
		vmctx.blockGas.burned = initialGasBurnedTotal
		vmctx.blockGas.feeCharged = initialGasFeeChargedTotal

		if errors.Is(vmexceptions.ErrNotEnoughFundsForSD, err) {
			vmctx.unprocessable = append(vmctx.unprocessable, req.(isc.OnLedgerRequest))
		}

		return nil, nil, err
	}

	vmctx.chainState().Apply()
	return result, vmctx.reqCtx.unprocessableToRetry, nil
}

func (vmctx *vmContext) payoutAgentID() isc.AgentID {
	var payoutAgentID isc.AgentID
	vmctx.callCore(governance.Contract, func(s kv.KVStore) {
		payoutAgentID = governance.MustGetPayoutAgentID(s)
	})
	return payoutAgentID
}

// creditAssetsToChain credits L1 accounts with attached assets and accrues all of them to the sender's account on-chain
func (vmctx *vmContext) creditAssetsToChain() {
	req := vmctx.reqCtx.req
	if req.IsOffLedger() {
		// off ledger request does not bring any deposit
		return
	}
	// Consume the output. Adjustment in L2 is needed because of storage deposit in the internal UTXOs
	storageDepositNeeded := vmctx.txbuilder.Consume(req.(isc.OnLedgerRequest))

	// if sender is specified, all assets goes to sender's sender
	// Otherwise it all goes to the common sender and panics is logged in the SC call
	sender := req.SenderAccount()
	if sender == nil {
		if req.IsOffLedger() {
			panic("nil sender on offledger requests should never happen")
		}
		// onleger request with no sender, send all assets to the payoutAddress
		payoutAgentID := vmctx.payoutAgentID()
		vmctx.creditNFTToAccount(payoutAgentID, req.NFT())
		vmctx.creditToAccount(payoutAgentID, req.Assets())

		// debit any SD required for accounting UTXOs
		if storageDepositNeeded > 0 {
			vmctx.debitFromAccount(payoutAgentID, isc.NewAssetsBaseTokens(storageDepositNeeded))
		}
		return
	}

	senderBaseTokens := req.Assets().BaseTokens + vmctx.GetBaseTokensBalance(sender)

	if senderBaseTokens < storageDepositNeeded {
		// user doesn't have enough funds to pay for the SD needs of this request
		panic(vmexceptions.ErrNotEnoughFundsForSD)
	}

	vmctx.creditToAccount(sender, req.Assets())
	vmctx.creditNFTToAccount(sender, req.NFT())
	if storageDepositNeeded > 0 {
		vmctx.reqCtx.sdCharged = storageDepositNeeded
		vmctx.debitFromAccount(sender, isc.NewAssetsBaseTokens(storageDepositNeeded))
	}
}

func (vmctx *vmContext) catchRequestPanic(f func()) error {
	err := panicutil.CatchPanic(f)
	if err == nil {
		return nil
	}
	// catches protocol exception error which is not the request or contract fault
	// If it occurs, the request is just skipped
	for _, targetError := range vmexceptions.AllProtocolLimits {
		if errors.Is(err, targetError) {
			return err
		}
	}
	// panic again with more information about the error
	panic(fmt.Errorf(
		"panic when running request #%d ID:%s, requestbytes:%s err:%w",
		vmctx.reqCtx.requestIndex,
		vmctx.reqCtx.req.ID(),
		iotago.EncodeHex(vmctx.reqCtx.req.Bytes()),
		err,
	))
}

// checkAllowance ensure there are enough funds to cover the specified allowance
// panics if not enough funds
func (vmctx *vmContext) checkAllowance() {
	if !vmctx.HasEnoughForAllowance(vmctx.reqCtx.req.SenderAccount(), vmctx.reqCtx.req.Allowance()) {
		panic(vm.ErrNotEnoughFundsForAllowance)
	}
}

func (vmctx *vmContext) shouldChargeGasFee() bool {
	if vmctx.reqCtx.req.SenderAccount() == nil {
		return false
	}
	if vmctx.reqCtx.req.SenderAccount().Equals(vmctx.chainOwnerID) && vmctx.reqCtx.req.CallTarget().Contract == governance.Contract.Hname() {
		return false
	}
	return true
}

func (vmctx *vmContext) prepareGasBudget() {
	if !vmctx.shouldChargeGasFee() {
		return
	}
	vmctx.gasSetBudget(vmctx.calculateAffordableGasBudget())
	vmctx.GasBurnEnable(true)
}

// callTheContract runs the contract. It catches and processes all panics except the one which cancel the whole block
func (vmctx *vmContext) callTheContract() (receipt *blocklog.RequestReceipt, callRet dict.Dict) {
	txsnapshot := vmctx.createTxBuilderSnapshot()
	snapMutations := vmctx.currentStateUpdate.Clone()

	var callErr *isc.VMError
	func() {
		defer func() {
			panicErr := vmctx.checkVMPluginPanic(recover())
			if panicErr == nil {
				return
			}
			callErr = panicErr
			vmctx.Debugf("recovered panic from contract call: %v", panicErr)
			if vmctx.task.WillProduceBlock() {
				vmctx.Debugf(string(debug.Stack()))
			}
		}()
		// ensure there are enough funds to cover the specified allowance
		vmctx.checkAllowance()

		callRet = vmctx.callFromRequest()
		// ensure at least the minimum amount of gas is charged
		vmctx.GasBurn(gas.BurnCodeMinimumGasPerRequest1P, vmctx.GasBurned())
	}()
	if callErr != nil {
		// panic happened during VM plugin call. Restore the state
		vmctx.restoreTxBuilderSnapshot(txsnapshot)
		vmctx.currentStateUpdate = snapMutations
	}
	// charge gas fee no matter what
	vmctx.chargeGasFee()

	// write receipt no matter what
	receipt = vmctx.writeReceiptToBlockLog(callErr)

	if vmctx.reqCtx.req.IsOffLedger() {
		vmctx.updateOffLedgerRequestNonce()
	}

	return receipt, callRet
}

func (vmctx *vmContext) checkVMPluginPanic(r interface{}) *isc.VMError {
	if r == nil {
		return nil
	}
	// re-panic-ing if error it not user nor VM plugin fault.
	if vmexceptions.IsSkipRequestException(r) {
		panic(r)
	}
	// Otherwise, the panic is wrapped into the returned error, including gas-related panic
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
func (vmctx *vmContext) callFromRequest() dict.Dict {
	req := vmctx.reqCtx.req
	vmctx.Debugf("callFromRequest: %s", req.ID().String())

	if req.SenderAccount() == nil {
		// if sender unknown, follow panic path
		panic(vm.ErrSenderUnknown)
	}

	contract := req.CallTarget().Contract
	entryPoint := req.CallTarget().EntryPoint

	return vmctx.callProgram(
		contract,
		entryPoint,
		req.Params(),
		req.Allowance(),
	)
}

func (vmctx *vmContext) getGasBudget() uint64 {
	gasBudget, isEVM := vmctx.reqCtx.req.GasBudget()
	if !isEVM || gasBudget == 0 {
		return gasBudget
	}

	var gasRatio util.Ratio32
	vmctx.callCore(governance.Contract, func(s kv.KVStore) {
		gasRatio = governance.MustGetGasFeePolicy(s).EVMGasRatio
	})
	return gas.EVMGasToISC(gasBudget, &gasRatio)
}

// calculateAffordableGasBudget checks the account of the sender and calculates affordable gas budget
// Affordable gas budget is calculated from gas budget provided in the request by the user and taking into account
// how many tokens the sender has in its account and how many are allowed for the target.
// Safe arithmetics is used
func (vmctx *vmContext) calculateAffordableGasBudget() (budget, maxTokensToSpendForGasFee uint64) {
	gasBudget := vmctx.getGasBudget()

	if vmctx.task.EstimateGasMode && gasBudget == 0 {
		// gas budget 0 means its a view call, so we give it max gas and tokens
		return vmctx.chainInfo.GasLimits.MaxGasExternalViewCall, math.MaxUint64
	}

	// make sure the gasBuget is at least >= than the allowed minimum
	if gasBudget < vmctx.chainInfo.GasLimits.MinGasPerRequest {
		gasBudget = vmctx.chainInfo.GasLimits.MinGasPerRequest
	}

	// calculate how many tokens for gas fee can be guaranteed after taking into account the allowance
	guaranteedFeeTokens := vmctx.calcGuaranteedFeeTokens()
	// calculate how many tokens maximum will be charged taking into account the budget
	f1, f2 := vmctx.chainInfo.GasFeePolicy.FeeFromGasBurned(gasBudget, guaranteedFeeTokens)
	maxTokensToSpendForGasFee = f1 + f2
	// calculate affordableGas gas budget
	affordableGas := vmctx.chainInfo.GasFeePolicy.GasBudgetFromTokens(guaranteedFeeTokens)
	// adjust gas budget to what is affordable
	affordableGas = util.MinUint64(gasBudget, affordableGas)
	// cap gas to the maximum allowed per tx
	return util.MinUint64(affordableGas, vmctx.chainInfo.GasLimits.MaxGasPerRequest), maxTokensToSpendForGasFee
}

// calcGuaranteedFeeTokens return the maximum tokens (base tokens or native) can be guaranteed for the fee,
// taking into account allowance (which must be 'reserved')
func (vmctx *vmContext) calcGuaranteedFeeTokens() uint64 {
	tokensGuaranteed := vmctx.GetBaseTokensBalance(vmctx.reqCtx.req.SenderAccount())
	// safely subtract the allowed from the sender to the target
	if allowed := vmctx.reqCtx.req.Allowance(); allowed != nil {
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
func (vmctx *vmContext) chargeGasFee() {
	defer func() {
		// add current request gas burn to the total of the block
		vmctx.blockGas.burned += vmctx.reqCtx.gas.burned
	}()

	// ensure at least the minimum amount of gas is charged
	minGas := gas.BurnCodeMinimumGasPerRequest1P.Cost(0)
	if vmctx.reqCtx.gas.burned < minGas {
		vmctx.reqCtx.gas.burned = minGas
	}

	vmctx.GasBurnEnable(false)

	if !vmctx.shouldChargeGasFee() {
		return
	}

	availableToPayFee := vmctx.reqCtx.gas.maxTokensToSpendForGasFee
	if !vmctx.task.EstimateGasMode && !vmctx.chainInfo.GasFeePolicy.IsEnoughForMinimumFee(availableToPayFee) {
		// user didn't specify enough base tokens to cover the minimum request fee, charge whatever is present in the user's account
		availableToPayFee = vmctx.GetSenderTokenBalanceForFees()
	}

	// total fees to charge
	sendToPayout, sendToValidator := vmctx.chainInfo.GasFeePolicy.FeeFromGasBurned(vmctx.GasBurned(), availableToPayFee)
	vmctx.reqCtx.gas.feeCharged = sendToPayout + sendToValidator

	// calc gas totals
	vmctx.blockGas.feeCharged += vmctx.reqCtx.gas.feeCharged

	if vmctx.task.EstimateGasMode {
		// If estimating gas, compute the gas fee but do not attempt to charge
		return
	}

	sender := vmctx.reqCtx.req.SenderAccount()
	if sendToValidator != 0 {
		transferToValidator := &isc.Assets{}
		transferToValidator.BaseTokens = sendToValidator
		vmctx.mustMoveBetweenAccounts(sender, vmctx.task.ValidatorFeeTarget, transferToValidator)
	}

	// ensure common account has at least minBalanceInCommonAccount, and transfer the rest of gas fee to payout AgentID
	// if the payout AgentID is not set in governance contract, then chain owner will be used
	var minBalanceInCommonAccount uint64
	vmctx.callCore(governance.Contract, func(s kv.KVStore) {
		minBalanceInCommonAccount = governance.MustGetMinCommonAccountBalance(s)
	})
	commonAccountBal := vmctx.GetBaseTokensBalance(accounts.CommonAccount())
	if commonAccountBal < minBalanceInCommonAccount {
		// pay to common account since the balance of common account is less than minSD
		transferToCommonAcc := sendToPayout
		sendToPayout = 0
		if commonAccountBal+transferToCommonAcc > minBalanceInCommonAccount {
			excess := (commonAccountBal + transferToCommonAcc) - minBalanceInCommonAccount
			transferToCommonAcc -= excess
			sendToPayout = excess
		}
		vmctx.mustMoveBetweenAccounts(sender, accounts.CommonAccount(), isc.NewAssetsBaseTokens(transferToCommonAcc))
	}
	if sendToPayout > 0 {
		payoutAgentID := vmctx.payoutAgentID()
		vmctx.mustMoveBetweenAccounts(sender, payoutAgentID, isc.NewAssetsBaseTokens(sendToPayout))
	}
}

func (vmctx *vmContext) GetContractRecord(contractHname isc.Hname) (ret *root.ContractRecord) {
	ret = vmctx.findContractByHname(contractHname)
	if ret == nil {
		vmctx.GasBurn(gas.BurnCodeCallTargetNotFound)
		panic(vm.ErrContractNotFound.Create(contractHname))
	}
	return ret
}

func (vmctx *vmContext) getOrCreateContractRecord(contractHname isc.Hname) (ret *root.ContractRecord) {
	return vmctx.GetContractRecord(contractHname)
}

// loadChainConfig only makes sense if chain is already deployed
func (vmctx *vmContext) loadChainConfig() {
	vmctx.chainInfo = vmctx.getChainInfo()
	vmctx.chainOwnerID = vmctx.chainInfo.ChainOwnerID
}

// mustCheckTransactionSize panics with ErrMaxTransactionSizeExceeded if the estimated transaction size exceeds the limit
func (vmctx *vmContext) mustCheckTransactionSize() {
	essence, _ := vmctx.BuildTransactionEssence(state.L1CommitmentNil, false)
	tx := transaction.MakeAnchorTransaction(essence, &iotago.Ed25519Signature{})
	if tx.Size() > parameters.L1().MaxPayloadSize {
		panic(vmexceptions.ErrMaxTransactionSizeExceeded)
	}
}
