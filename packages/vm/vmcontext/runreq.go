package vmcontext

import (
	"errors"
	"math"
	"math/big"
	"runtime/debug"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmexceptions"
	"golang.org/x/xerrors"
)

// RunTheRequest processes each iscp.Request in the batch
func (vmctx *VMContext) RunTheRequest(req iscp.Request, requestIndex uint16) (result *vm.RequestResult, err error) {
	// prepare context for the request
	vmctx.req = req
	vmctx.numPostedOutputs = 0
	vmctx.requestIndex = requestIndex
	vmctx.requestEventIndex = 0
	vmctx.entropy = hashing.HashData(vmctx.entropy[:])
	vmctx.callStack = vmctx.callStack[:0]
	vmctx.gasBudgetAdjusted = 0
	vmctx.gasBurned = 0
	vmctx.gasFeeCharged = 0
	vmctx.gasBurnEnable(false)

	vmctx.currentStateUpdate = state.NewStateUpdate(vmctx.virtualState.Timestamp().Add(1 * time.Nanosecond))
	defer func() { vmctx.currentStateUpdate = nil }()

	if err := vmctx.earlyCheckReasonToSkip(); err != nil {
		return nil, err
	}
	vmctx.loadChainConfig()

	// at this point state update is empty
	// so far there were no panics except optimistic reader
	txsnapshot := vmctx.createTxBuilderSnapshot()

	// catches protocol exception error which is not the request or contract fault
	// If it occurs, the request is just skipped
	err = util.CatchPanicReturnError(
		func() {
			// transfer all attached assets to the sender's account
			vmctx.creditAssetsToChain()
			// load gas and fee policy, calculate and set gas budget
			vmctx.prepareGasBudget()
			// run the contract program
			receipt, callRet, callErr := vmctx.callTheContract()
			result = &vm.RequestResult{
				Request: req,
				Receipt: receipt,
				Return:  callRet,
				Error:   callErr,
			}
		}, vmexceptions.AllProtocolLimits...,
	)
	if err != nil {
		// transaction limits exceeded or not enough funds for internal dust deposit. Skipping the request. Rollback
		vmctx.restoreTxBuilderSnapshot(txsnapshot)
		return nil, err
	}
	vmctx.virtualState.ApplyStateUpdates(vmctx.currentStateUpdate)
	vmctx.assertConsistentL2WithL1TxBuilder("end RunTheRequest")
	return result, nil
}

// creditAssetsToChain credits L1 accounts with attached assets and accrues all of them to the sender's account on-chain
func (vmctx *VMContext) creditAssetsToChain() {
	vmctx.assertConsistentL2WithL1TxBuilder("begin creditAssetsToChain")

	if vmctx.req.IsOffLedger() {
		// off ledger request does not bring any deposit
		return
	}
	// Consume the output. Adjustment in L2 is needed because of the dust in the internal UTXOs
	dustAdjustment := vmctx.txbuilder.Consume(vmctx.req)
	if dustAdjustment > 0 {
		panic("`dustAdjustment > 0`: assertion failed, expected always non-positive dust adjustment")
	}

	// if sender is specified, all assets goes to sender's account
	// Otherwise it all goes to the common account and panics is logged in the SC call
	account := vmctx.req.SenderAccount()
	if account == nil {
		account = commonaccount.Get(vmctx.ChainID())
	}
	vmctx.creditToAccount(account, vmctx.req.Assets())

	// adjust the sender's account with the dust consumed or returned by internal UTXOs
	// if iotas in the sender's account is not enough for the dust deposit of newly created TNT outputs
	// it will panic with exceptions.ErrNotEnoughFundsForInternalDustDeposit
	// TNT outputs will use dust deposit from the caller
	// TODO remove attack vector when iotas for dust deposit is not enough and the request keeps being skipped
	vmctx.adjustL2IotasIfNeeded(dustAdjustment, account)

	// here transaction builder must be consistent itself and be consistent with the state (the accounts)
	vmctx.assertConsistentL2WithL1TxBuilder("end creditAssetsToChain")
}

func (vmctx *VMContext) prepareGasBudget() {
	if vmctx.req.SenderAccount() == nil {
		return
	}
	if vmctx.isInitChainRequest() {
		return
	}
	vmctx.calculateAffordableGasBudget()
	vmctx.gasSetBudget(vmctx.gasBudgetAdjusted)
	vmctx.gasBurnEnable(true)
}

// callTheContract runs the contract. It catches and processes all panics except the one which cancel the whole block
func (vmctx *VMContext) callTheContract() (receipt *blocklog.RequestReceipt, callRet dict.Dict, callErr error) {
	vmctx.txsnapshot = vmctx.createTxBuilderSnapshot()
	snapMutations := vmctx.currentStateUpdate.Clone()

	if vmctx.req.IsOffLedger() {
		vmctx.updateOffLedgerRequestMaxAssumedNonce()
	}
	func() {
		defer func() {
			panicErr := vmctx.checkVMPluginPanic(recover())
			if panicErr == nil {
				return
			}
			callErr = panicErr
			vmctx.Debugf("recovered panic from contract call: %v", panicErr)
			vmctx.Debugf(string(debug.Stack()))
		}()
		callRet = vmctx.callFromRequest()
		// ensure at least the minimum amount of gas is charged
		if vmctx.GasBurned() < gas.BurnCodeMinimumGasPerRequest1P.Cost() {
			vmctx.GasBurn(gas.BurnCodeMinimumGasPerRequest1P, vmctx.GasBurned())
		}
	}()
	if callErr != nil {
		// panic happened during VM plugin call. Restore the state
		vmctx.restoreTxBuilderSnapshot(vmctx.txsnapshot)
		vmctx.currentStateUpdate = snapMutations
	}
	// charge gas fee no matter what
	vmctx.chargeGasFee()
	// write receipt no matter what
	receipt = vmctx.writeReceiptToBlockLog(callErr)
	return receipt, callRet, callErr
}

func (vmctx *VMContext) checkVMPluginPanic(r interface{}) error {
	if r == nil {
		return nil
	}
	// re-panic-ing if error it not user nor VM plugin fault.
	if vmexceptions.IsSkipRequestException(r) {
		panic(r)
	}
	// Otherwise, the panic is wrapped into the returned error, including gas-related panic
	switch err := r.(type) {
	case *kv.DBError:
		panic(err)
	case string:
		r = errors.New(err)
	case error:
		if errors.Is(err, coreutil.ErrorStateInvalidated) {
			panic(err)
		}
	}
	return xerrors.Errorf("%v", r)
}

// callFromRequest is the call itself. Assumes sc exists
func (vmctx *VMContext) callFromRequest() dict.Dict {
	vmctx.Debugf("callFromRequest: %s", vmctx.req.ID().String())

	if vmctx.req.SenderAccount() == nil {
		// if sender unknown, follow panic path
		panic(ErrSenderUnknown)
	}
	// TODO check if the comment below holds true
	// calling only non view entry points. Calling the view will trigger error and fallback
	contract := vmctx.req.CallTarget().Contract
	entryPoint := vmctx.req.CallTarget().EntryPoint

	return vmctx.callProgram(
		contract,
		entryPoint,
		vmctx.req.Params(),
		vmctx.req.Allowance(),
	)
}

// calculateAffordableGasBudget checks the account of the sender and calculates affordable gas budget
// Affordable gas budget is calculated from gas budget provided in the request by the user and taking into account
// how many tokens the sender has in its account and how many are allowed for the target.
// Safe arithmetics is used
func (vmctx *VMContext) calculateAffordableGasBudget() {
	// when estimating gas, if maxUint64 is provided, use the maximum gas budget possible
	if vmctx.task.EstimateGasMode && vmctx.req.GasBudget() == math.MaxUint64 {
		vmctx.gasBudgetAdjusted = gas.MaxGasPerCall
		vmctx.gasMaxTokensToSpendForGasFee = math.MaxUint64
		return
	}

	if vmctx.req.SenderAddress() == nil {
		panic("inconsistency: vmctx.req.SenderAddress() == nil")
	}
	// calculate how many tokens for gas fee can be guaranteed after taking into account the allowance
	guaranteedFeeTokens := vmctx.calcGuaranteedFeeTokens()
	// calculate how many tokens maximum will be charged taking into account the budget
	f1, f2 := vmctx.chainInfo.GasFeePolicy.FeeFromGas(vmctx.req.GasBudget(), guaranteedFeeTokens)
	vmctx.gasMaxTokensToSpendForGasFee = f1 + f2
	// calculate affordable gas budget
	affordable := vmctx.chainInfo.GasFeePolicy.AffordableGasBudgetFromAvailableTokens(guaranteedFeeTokens)
	// adjust gas budget to what is affordable
	affordable = util.MinUint64(vmctx.req.GasBudget(), affordable)
	// cap gas to the maximum allowed per tx
	vmctx.gasBudgetAdjusted = util.MinUint64(affordable, gas.MaxGasPerCall)
}

// calcGuaranteedFeeTokens return hiw maximum tokens (iotas or native) can be guaranteed for the fee,
// taking into account allowance (which must be 'reserved')
func (vmctx *VMContext) calcGuaranteedFeeTokens() uint64 {
	var tokensGuaranteed uint64

	if vmctx.chainInfo.GasFeePolicy.GasFeeTokenID == nil {
		// iotas are used as gas tokens
		tokensGuaranteed = vmctx.GetIotaBalance(vmctx.req.SenderAccount())
		// safely subtract the allowed from the sender to the target
		if allowed := vmctx.req.Allowance(); allowed != nil {
			if tokensGuaranteed < allowed.Iotas {
				tokensGuaranteed = 0
			} else {
				tokensGuaranteed -= allowed.Iotas
			}
		}
		return tokensGuaranteed
	}
	// native tokens are used for gas fee
	tokenID := vmctx.chainInfo.GasFeePolicy.GasFeeTokenID
	// to pay for gas chain is configured to use some native token, not IOTA
	tokensAvailableBig := vmctx.GetNativeTokenBalance(vmctx.req.SenderAccount(), tokenID)
	if tokensAvailableBig != nil {
		// safely subtract the transfer from the sender to the target
		if transfer := vmctx.req.Allowance(); transfer != nil {
			if transferTokens := iscp.FindNativeTokenBalance(transfer.Tokens, tokenID); transferTokens != nil {
				if tokensAvailableBig.Cmp(transferTokens) < 0 {
					tokensAvailableBig.SetUint64(0)
				} else {
					tokensAvailableBig.Sub(tokensAvailableBig, transferTokens)
				}
			}
		}
		if tokensAvailableBig.IsUint64() {
			tokensGuaranteed = tokensAvailableBig.Uint64()
		} else {
			tokensGuaranteed = math.MaxUint64
		}
	}
	return tokensGuaranteed
}

// chargeGasFee takes burned tokens from the sender's account
// It should always be enough because gas budget is set affordable
func (vmctx *VMContext) chargeGasFee() {
	// disable gas burn
	vmctx.gasBurnEnable(false)
	if vmctx.req.SenderAccount() == nil {
		// no charging if sender is unknown
		return
	}
	if vmctx.isInitChainRequest() {
		// do not charge gas fees if init request
		return
	}

	availableToPayFee := vmctx.gasMaxTokensToSpendForGasFee
	if !vmctx.task.EstimateGasMode && !vmctx.chainInfo.GasFeePolicy.IsEnoughForMinimumFee(availableToPayFee) {
		// user didn't specify enough iotas to cover the minimum request fee, charge whatever is present in the user's account
		availableToPayFee = vmctx.GetSenderTokenBalanceForFees()
	}

	// total fees to charge
	sendToOwner, sendToValidator := vmctx.chainInfo.GasFeePolicy.FeeFromGas(vmctx.GasBurned(), availableToPayFee)
	vmctx.gasFeeCharged = sendToOwner + sendToValidator

	// calc gas totals
	vmctx.gasFeeChargedTotal += vmctx.gasFeeCharged

	if vmctx.task.EstimateGasMode {
		// If estimating gas, compute the gas fee but do not attempt to charge
		return
	}

	transferToValidator := &iscp.Assets{}
	transferToOwner := &iscp.Assets{}
	if vmctx.chainInfo.GasFeePolicy.GasFeeTokenID != nil {
		transferToValidator.Tokens = iotago.NativeTokens{
			&iotago.NativeToken{ID: *vmctx.chainInfo.GasFeePolicy.GasFeeTokenID, Amount: big.NewInt(int64(sendToValidator))},
		}
		transferToOwner.Tokens = iotago.NativeTokens{
			&iotago.NativeToken{ID: *vmctx.chainInfo.GasFeePolicy.GasFeeTokenID, Amount: big.NewInt(int64(sendToOwner))},
		}
	} else {
		transferToValidator.Iotas = sendToValidator
		transferToOwner.Iotas = sendToOwner
	}
	sender := vmctx.req.SenderAccount()

	vmctx.mustMoveBetweenAccounts(sender, vmctx.task.ValidatorFeeTarget, transferToValidator)
	vmctx.mustMoveBetweenAccounts(sender, commonaccount.Get(vmctx.ChainID()), transferToOwner)
}

func (vmctx *VMContext) getContractRecord(contractHname iscp.Hname) *root.ContractRecord {
	ret := vmctx.findContractByHname(contractHname)
	if ret == nil {
		vmctx.GasBurn(gas.BurnCodeCallTargetNotFound)
		panic(xerrors.Errorf("%v: contract = %s", ErrTargetContractNotFound, contractHname))
	}
	return ret
}

// loadChainConfig only makes sense if chain is already deployed
func (vmctx *VMContext) loadChainConfig() {
	if vmctx.isInitChainRequest() {
		vmctx.chainOwnerID = vmctx.req.SenderAccount()
		vmctx.chainInfo = nil
		return
	}
	vmctx.chainInfo = vmctx.getChainInfo()
	vmctx.chainOwnerID = vmctx.chainInfo.ChainOwnerID
}

func (vmctx *VMContext) isInitChainRequest() bool {
	if vmctx.req == nil {
		return false
	}
	target := vmctx.req.CallTarget()
	return target.Contract == root.Contract.Hname() && target.EntryPoint == iscp.EntryPointInit
}
