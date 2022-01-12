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
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
	"golang.org/x/xerrors"
)

// RunTheRequest processes each iscp.Request in the batch
func (vmctx *VMContext) RunTheRequest(req iscp.Request, requestIndex uint16) error {
	// prepare context for the request
	vmctx.req = req
	vmctx.requestIndex = requestIndex
	vmctx.requestEventIndex = 0
	vmctx.entropy = hashing.HashData(vmctx.entropy[:])
	vmctx.callStack = vmctx.callStack[:0]
	vmctx.gasBudget = 0
	vmctx.gasBurned = 0
	vmctx.gasFeeCharged = 0
	vmctx.gasBurnEnable(false)

	vmctx.currentStateUpdate = state.NewStateUpdate(vmctx.virtualState.Timestamp().Add(1 * time.Nanosecond))
	defer func() { vmctx.currentStateUpdate = nil }()

	if err := vmctx.earlyCheckReasonToSkip(); err != nil {
		return err
	}
	vmctx.loadChainConfig()

	// at this point state update is empty
	// so far there were no panics except optimistic reader
	txsnapshot := vmctx.createTxBuilderSnapshot()

	// catches error which is not the request or contract fault
	// If it occurs, the request is just skipped
	err := util.CatchPanicReturnError(
		func() {
			// transfer all attached assets to the sender's account
			vmctx.creditAssetsToChain()
			// load gas and fee policy, calculate and set gas budget
			vmctx.prepareGasBudget()
			// run the contract program
			vmctx.callTheContract()
		},
		vmtxbuilder.ErrInputLimitExceeded,
		vmtxbuilder.ErrOutputLimitExceeded,
		vmtxbuilder.ErrNotEnoughFundsForInternalDustDeposit,
		vmtxbuilder.ErrNumberOfNativeTokensLimitExceeded,
	)
	if err != nil {
		// transaction limits exceeded or not enough funds for internal dust deposit. Skipping the request. Rollback
		vmctx.restoreTxBuilderSnapshot(txsnapshot)
		return err
	}
	vmctx.virtualState.ApplyStateUpdates(vmctx.currentStateUpdate)
	vmctx.assertConsistentL2WithL1TxBuilder("end RunTheRequest")
	return nil
}

// creditAssetsToChain credits L1 accounts with attached assets and accrues all of them to the sender's account on-chain
func (vmctx *VMContext) creditAssetsToChain() {
	vmctx.assertConsistentL2WithL1TxBuilder("begin creditAssetsToChain")

	if vmctx.req.IsOffLedger() {
		// off ledger request does not bring any deposit
		return
	}
	// Consume the output. Adjustment in L2 is needed because of the dust in the internal UTXOs
	dustAdjustmentOfTheCommonAccount := vmctx.txbuilder.Consume(vmctx.req)
	// update the state, the account ledger
	// NOTE: sender account will be CommonAccount if sender address is not available
	// It means any random sends to the chain end up in the common account
	vmctx.creditToAccount(vmctx.req.SenderAccount(), vmctx.req.Assets())

	// adjust the common account with the dust consumed or returned by internal UTXOs
	// If common account does not contain enough funds for internal dust, it panics with
	// vmtxbuilder.ErrNotEnoughFundsForInternalDustDeposit and the request will be skipped
	vmctx.adjustL2IotasIfNeeded(dustAdjustmentOfTheCommonAccount)
	// here transaction builder must be consistent itself and be consistent with the state (the accounts)
	vmctx.assertConsistentL2WithL1TxBuilder("end creditAssetsToChain")
}

func (vmctx *VMContext) prepareGasBudget() {
	if vmctx.isInitChainRequest() {
		return
	}
	vmctx.calculateAffordableGasBudget()
	vmctx.gasSetBudget(vmctx.gasBudget)
	vmctx.gasBurnEnable(true)
}

// callTheContract runs the contract. It catches and processes all panics except the one which cancel the whole block
func (vmctx *VMContext) callTheContract() {
	txsnapshot := vmctx.createTxBuilderSnapshot()
	snapMutations := vmctx.currentStateUpdate.Clone()

	if vmctx.req.IsOffLedger() {
		vmctx.updateOffLedgerRequestMaxAssumedNonce()
	}
	vmctx.lastError = nil
	func() {
		defer func() {
			panicErr := checkVMPluginPanic(recover())
			if panicErr == nil {
				return
			}
			vmctx.lastError = panicErr
			vmctx.lastResult = nil
			vmctx.Debugf("%v", vmctx.lastError)
			vmctx.Debugf(string(debug.Stack()))
		}()
		vmctx.lastResult, vmctx.lastError = vmctx.callFromRequest()
	}()
	if vmctx.lastError != nil {
		// panic happened during VM plugin call. Restore the state
		vmctx.restoreTxBuilderSnapshot(txsnapshot)
		vmctx.currentStateUpdate = snapMutations
	}
	// charge gas fee no matter what
	vmctx.chargeGasFee()
	// write receipt no matter what
	vmctx.writeReceiptToBlockLog(vmctx.lastError)
}

func checkVMPluginPanic(r interface{}) error {
	if r == nil {
		return nil
	}
	// re-panic-ing if error it not user nor VM plugin fault.
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
		if errors.Is(err, vmtxbuilder.ErrOutputLimitExceeded) {
			panic(err)
		}
		if errors.Is(err, vmtxbuilder.ErrInputLimitExceeded) {
			panic(err)
		}
		if errors.Is(err, vmtxbuilder.ErrNumberOfNativeTokensLimitExceeded) {
			panic(err)
		}
	}
	return xerrors.Errorf("%v", r)
}

// callFromRequest is the call itself. Assumes sc exists
func (vmctx *VMContext) callFromRequest() (dict.Dict, error) {
	vmctx.Debugf("callFromRequest: %s", vmctx.req.ID().String())

	// calling only non view entry points. Calling the view will trigger error and fallback
	entryPoint := vmctx.req.CallTarget().EntryPoint
	targetContract := vmctx.targetContract()
	if targetContract == nil {
		vmctx.GasBurn(gas.NotFoundTarget)
		panic(xerrors.Errorf("%v: target = %s", ErrTargetContractNotFound, vmctx.req.CallTarget().Contract))
	}
	return vmctx.callByProgramHash(
		targetContract.Hname(),
		entryPoint,
		vmctx.req.Params(),
		vmctx.req.Allowance(),
		targetContract.ProgramHash,
	)
}

// calculateAffordableGasBudget checks the account of the sender and calculates affordable gas budget
// Affordable gas budget is calculated from gas budget provided in the request by the user and taking into account
// how many tokens the sender has in its account and how many are allowed for the target.
// Safe arithmetics is used
func (vmctx *VMContext) calculateAffordableGasBudget() {
	if vmctx.req.SenderAddress() == nil {
		panic("inconsistency: vmctx.req.SenderAddress() == nil")
	}
	// calculate how many tokens for gas fee can be guaranteed after taking into account the allowance
	tokensGuaranteed := uint64(0)
	if vmctx.chainInfo.GasFeePolicy.GasFeeTokenID != nil {
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
	} else {
		// Iotas are used to pay the gas fee
		tokensGuaranteed = vmctx.GetIotaBalance(vmctx.req.SenderAccount())
		// safely subtract the transfer from the sender to the target
		if transfer := vmctx.req.Allowance(); transfer != nil {
			if tokensGuaranteed < transfer.Iotas {
				tokensGuaranteed = 0
			} else {
				tokensGuaranteed -= transfer.Iotas
			}
		}
	}
	vmctx.gasMaxTokensAvailableForGasFee = tokensGuaranteed
	if tokensGuaranteed < vmctx.chainInfo.GasFeePolicy.GasPricePerNominalUnit {
		// it will not proceed but will charge at least the minimum
		panic(ErrNotEnoughTokensFor1GasNominalUnit)
	}
	var gasBudgetAffordable uint64
	if vmctx.chainInfo.GasFeePolicy.GasPricePerNominalUnit == 0 {
		gasBudgetAffordable = math.MaxUint64
	} else {
		nominalUnitsOfGas := tokensGuaranteed / vmctx.chainInfo.GasFeePolicy.GasPricePerNominalUnit
		if nominalUnitsOfGas > math.MaxUint64/vmctx.chainInfo.GasFeePolicy.GasNominalUnit {
			gasBudgetAffordable = math.MaxUint64
		} else {
			gasBudgetAffordable = nominalUnitsOfGas * vmctx.chainInfo.GasFeePolicy.GasNominalUnit
		}
	}
	vmctx.gasBudget = util.MinUint64(vmctx.req.GasBudget(), gasBudgetAffordable)
}

// chargeGasFee takes burned tokens from the sender's account
// It should always be enough because gas budget is set affordable
func (vmctx *VMContext) chargeGasFee() {
	vmctx.gasBurnEnable(false)
	if vmctx.req.SenderAddress() == nil {
		panic("inconsistency: vmctx.req.Request().SenderAddress() == nil")
	}
	if vmctx.isInitChainRequest() {
		// do not charge gas fees if init request
		return
	}
	// total fees to charge
	sendToOwner, sendToValidator := vmctx.chainInfo.GasFeePolicy.FeeFromGas(vmctx.GasBurned(), vmctx.gasMaxTokensAvailableForGasFee)
	vmctx.gasFeeCharged = sendToOwner + sendToValidator

	// calc totals
	vmctx.gasBurnedTotal += vmctx.gasBurned
	vmctx.gasFeeChargedTotal += vmctx.gasFeeCharged

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

func (vmctx *VMContext) targetContract() *root.ContractRecord {
	// find target contract
	targetContract := vmctx.req.CallTarget().Contract
	ret := vmctx.findContractByHname(targetContract)
	if ret == nil {
		vmctx.Warnf("contract not found: %s", targetContract)
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
	target := vmctx.req.CallTarget()
	return target.Contract == root.Contract.Hname() && target.EntryPoint == iscp.EntryPointInit
}
