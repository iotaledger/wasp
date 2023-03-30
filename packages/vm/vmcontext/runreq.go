package vmcontext

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
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmexceptions"
)

// RunTheRequest processes each isc.Request in the batch
func (vmctx *VMContext) RunTheRequest(req isc.Request, requestIndex uint16) (*vm.RequestResult, error) {
	// prepare context for the request
	vmctx.req = req
	defer func() { vmctx.req = nil }() // in case `getToBeCaller()` is called afterwards

	vmctx.NumPostedOutputs = 0
	vmctx.requestIndex = requestIndex
	vmctx.requestEventIndex = 0
	vmctx.entropy = hashing.HashData(append(codec.EncodeUint16(requestIndex), vmctx.task.Entropy[:]...))
	vmctx.callStack = vmctx.callStack[:0]
	vmctx.gasBudgetAdjusted = 0
	vmctx.gasBurned = 0
	vmctx.gasFeeCharged = 0
	vmctx.GasBurnEnable(false)

	vmctx.currentStateUpdate = NewStateUpdate()
	vmctx.chainState().Set(kv.Key(coreutil.StatePrefixTimestamp), codec.EncodeTime(vmctx.task.StateDraft.Timestamp().Add(1*time.Nanosecond)))
	defer func() { vmctx.currentStateUpdate = nil }()

	if err2 := vmctx.earlyCheckReasonToSkip(); err2 != nil {
		return nil, err2
	}
	vmctx.loadChainConfig()

	// at this point state update is empty
	// so far there were no panics except optimistic reader
	txsnapshot := vmctx.createTxBuilderSnapshot()

	var result *vm.RequestResult
	err := vmctx.catchRequestPanic(
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
		return nil, err
	}
	vmctx.chainState().Apply()
	return result, nil
}

// creditAssetsToChain credits L1 accounts with attached assets and accrues all of them to the sender's account on-chain
func (vmctx *VMContext) creditAssetsToChain() {
	if vmctx.req.IsOffLedger() {
		// off ledger request does not bring any deposit
		return
	}
	// Consume the output. Adjustment in L2 is needed because of storage deposit in the internal UTXOs
	storageDepositNeeded := vmctx.txbuilder.Consume(vmctx.req.(isc.OnLedgerRequest))

	// if sender is specified, all assets goes to sender's sender
	// Otherwise it all goes to the common sender and panics is logged in the SC call
	sender := vmctx.req.SenderAccount()
	if sender == nil {
		// TODO this should never happen... can we just panic here?
		// this is probably an artifact from the "originTx"
		sender = accounts.CommonAccount()
	}

	senderBaseTokens := vmctx.req.Assets().BaseTokens + vmctx.GetBaseTokensBalance(sender)

	if senderBaseTokens < storageDepositNeeded {
		panic("TODO, not enough funds to pay for the SD NEEDED, THIS REQUEST MUST BE IGNORED OR SAVED FOR LATER SOMEHOW")
		// ...if not enough to pay for all the SD, this request needs to be flagged as "TO PROCESS LATER"... (do we consume it or not?)
	}

	vmctx.creditToAccount(sender, vmctx.req.Assets())
	vmctx.creditNFTToAccount(sender, vmctx.req.NFT())
	if storageDepositNeeded > 0 {
		// TODO the charged SD should be included in the receipt
		vmctx.debitFromAccount(sender, isc.NewAssetsBaseTokens(storageDepositNeeded))
	}
}

func (vmctx *VMContext) catchRequestPanic(f func()) error {
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
		vmctx.requestIndex,
		vmctx.req.ID(),
		iotago.EncodeHex(vmctx.req.Bytes()),
		err,
	))
}

// checkAllowance ensure there are enough funds to cover the specified allowance
// panics if not enough funds
func (vmctx *VMContext) checkAllowance() {
	if !vmctx.HasEnoughForAllowance(vmctx.req.SenderAccount(), vmctx.req.Allowance()) {
		panic(vm.ErrNotEnoughFundsForAllowance)
	}
}

func (vmctx *VMContext) shouldChargeGasFee() bool {
	if vmctx.req.SenderAccount() == nil {
		return false
	}
	if vmctx.req.SenderAccount().Equals(vmctx.chainOwnerID) && vmctx.req.CallTarget().Contract == governance.Contract.Hname() {
		return false
	}
	return true
}

func (vmctx *VMContext) prepareGasBudget() {
	if !vmctx.shouldChargeGasFee() {
		return
	}
	vmctx.gasSetBudget(vmctx.calculateAffordableGasBudget())
	vmctx.GasBurnEnable(true)
}

// callTheContract runs the contract. It catches and processes all panics except the one which cancel the whole block
func (vmctx *VMContext) callTheContract() (receipt *blocklog.RequestReceipt, callRet dict.Dict) {
	vmctx.txsnapshot = vmctx.createTxBuilderSnapshot()
	snapMutations := vmctx.currentStateUpdate.Clone()

	if vmctx.req.IsOffLedger() {
		vmctx.updateOffLedgerRequestMaxAssumedNonce()
	}
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
		if vmctx.GasBurned() < gas.BurnCodeMinimumGasPerRequest1P.Cost() {
			vmctx.gasBurnedTotal -= vmctx.gasBurned
			vmctx.gasBurned = 0
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
	return receipt, callRet
}

func (vmctx *VMContext) checkVMPluginPanic(r interface{}) *isc.VMError {
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
func (vmctx *VMContext) callFromRequest() dict.Dict {
	vmctx.Debugf("callFromRequest: %s", vmctx.req.ID().String())

	if vmctx.req.SenderAccount() == nil {
		// if sender unknown, follow panic path
		panic(vm.ErrSenderUnknown)
	}

	contract := vmctx.req.CallTarget().Contract
	entryPoint := vmctx.req.CallTarget().EntryPoint

	return vmctx.callProgram(
		contract,
		entryPoint,
		vmctx.req.Params(),
		vmctx.req.Allowance(),
	)
}

func (vmctx *VMContext) getGasBudget() uint64 {
	gasBudget, isEVM := vmctx.req.GasBudget()
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
func (vmctx *VMContext) calculateAffordableGasBudget() uint64 {
	gasBudget := vmctx.getGasBudget()

	if vmctx.task.EstimateGasMode && gasBudget == 0 {
		// gas budget 0 means its a view call, so we give it max gas and tokens
		vmctx.gasMaxTokensToSpendForGasFee = math.MaxUint64
		return vmctx.chainInfo.GasLimits.MaxGasExternalViewCall
	}

	// make sure the gasBuget is at least >= than the allowed minimum
	if gasBudget < vmctx.chainInfo.GasLimits.MinGasPerRequest {
		gasBudget = vmctx.chainInfo.GasLimits.MinGasPerRequest
	}

	// calculate how many tokens for gas fee can be guaranteed after taking into account the allowance
	guaranteedFeeTokens := vmctx.calcGuaranteedFeeTokens()
	// calculate how many tokens maximum will be charged taking into account the budget
	f1, f2 := vmctx.chainInfo.GasFeePolicy.FeeFromGasBurned(gasBudget, guaranteedFeeTokens)
	vmctx.gasMaxTokensToSpendForGasFee = f1 + f2
	// calculate affordableGas gas budget
	affordableGas := vmctx.chainInfo.GasFeePolicy.GasBudgetFromTokens(guaranteedFeeTokens)
	// adjust gas budget to what is affordable
	affordableGas = util.MinUint64(gasBudget, affordableGas)
	// cap gas to the maximum allowed per tx
	return util.MinUint64(affordableGas, vmctx.chainInfo.GasLimits.MaxGasPerRequest)
}

// calcGuaranteedFeeTokens return the maximum tokens (base tokens or native) can be guaranteed for the fee,
// taking into account allowance (which must be 'reserved')
func (vmctx *VMContext) calcGuaranteedFeeTokens() uint64 {
	tokensGuaranteed := vmctx.GetBaseTokensBalance(vmctx.req.SenderAccount())
	// safely subtract the allowed from the sender to the target
	if allowed := vmctx.req.Allowance(); allowed != nil {
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
func (vmctx *VMContext) chargeGasFee() {
	// ensure at least the minimum amount of gas is charged
	minGas := gas.BurnCodeMinimumGasPerRequest1P.Cost()
	if vmctx.GasBurned() < minGas {
		currentGas := vmctx.gasBurned
		vmctx.gasBurned = minGas
		vmctx.gasBurnedTotal += minGas - currentGas
	}

	vmctx.GasBurnEnable(false)

	if !vmctx.shouldChargeGasFee() {
		return
	}

	availableToPayFee := vmctx.gasMaxTokensToSpendForGasFee
	if !vmctx.task.EstimateGasMode && !vmctx.chainInfo.GasFeePolicy.IsEnoughForMinimumFee(availableToPayFee) {
		// user didn't specify enough base tokens to cover the minimum request fee, charge whatever is present in the user's account
		availableToPayFee = vmctx.GetSenderTokenBalanceForFees()
	}

	// total fees to charge
	sendToOwner, sendToValidator := vmctx.chainInfo.GasFeePolicy.FeeFromGasBurned(vmctx.GasBurned(), availableToPayFee)
	vmctx.gasFeeCharged = sendToOwner + sendToValidator

	// calc gas totals
	vmctx.gasFeeChargedTotal += vmctx.gasFeeCharged

	if vmctx.task.EstimateGasMode {
		// If estimating gas, compute the gas fee but do not attempt to charge
		return
	}

	transferToValidator := &isc.Assets{}
	transferToOwner := &isc.Assets{}
	transferToValidator.BaseTokens = sendToValidator
	transferToOwner.BaseTokens = sendToOwner
	sender := vmctx.req.SenderAccount()

	vmctx.mustMoveBetweenAccounts(sender, vmctx.task.ValidatorFeeTarget, transferToValidator)
	vmctx.mustMoveBetweenAccounts(sender, accounts.CommonAccount(), transferToOwner)
}

func (vmctx *VMContext) GetContractRecord(contractHname isc.Hname) (ret *root.ContractRecord) {
	ret = vmctx.findContractByHname(contractHname)
	if ret == nil {
		vmctx.GasBurn(gas.BurnCodeCallTargetNotFound)
		panic(vm.ErrContractNotFound.Create(contractHname))
	}
	return ret
}

func (vmctx *VMContext) getOrCreateContractRecord(contractHname isc.Hname) (ret *root.ContractRecord) {
	return vmctx.GetContractRecord(contractHname)
}

// loadChainConfig only makes sense if chain is already deployed
func (vmctx *VMContext) loadChainConfig() {
	vmctx.chainInfo = vmctx.getChainInfo()
	vmctx.chainOwnerID = vmctx.chainInfo.ChainOwnerID
}

// mustCheckTransactionSize panics with ErrMaxTransactionSizeExceeded if the estimated transaction size exceeds the limit
func (vmctx *VMContext) mustCheckTransactionSize() {
	essence, _ := vmctx.BuildTransactionEssence(state.L1CommitmentNil, false)
	tx := transaction.MakeAnchorTransaction(essence, &iotago.Ed25519Signature{})
	if tx.Size() > parameters.L1().MaxPayloadSize {
		panic(vmexceptions.ErrMaxTransactionSizeExceeded)
	}
}
