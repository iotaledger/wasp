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
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
	"golang.org/x/xerrors"
)

// RunTheRequest processes each iscp.RequestData in the batch
func (vmctx *VMContext) RunTheRequest(req iscp.RequestData, requestIndex uint16) error {
	// prepare context for the request
	vmctx.req = req
	vmctx.requestIndex = requestIndex
	vmctx.requestEventIndex = 0
	vmctx.entropy = hashing.HashData(vmctx.entropy[:])
	vmctx.callStack = vmctx.callStack[:0]

	if err := vmctx.earlyCheckReasonToSkip(); err != nil {
		return err
	}
	vmctx.loadChainConfig()
	vmctx.locateTargetContract()

	// at this point state update is empty
	// so far there were no panics except optimistic reader
	// No prepare state update (buffer) for mutations and panics

	txsnapshot := vmctx.createTxBuilderSnapshot()
	vmctx.currentStateUpdate = state.NewStateUpdate(vmctx.virtualState.Timestamp().Add(1 * time.Nanosecond))

	// catches error which is not the request or contract fault
	// If it occurs, the request is just skipped
	err := util.CatchPanicReturnError(func() {
		// transfer all attached assets to the sender's account
		vmctx.creditAssetsToChain()
		// load gas and fee policy, calculate and set gas budget
		vmctx.prepareGasBudget()
		// run the contract program
		vmctx.callTheContract()
	}, vmtxbuilder.ErrInputLimitExceeded, vmtxbuilder.ErrOutputLimitExceeded)

	if err != nil {
		// transaction limits exceeded. Skipping the request. Rollback
		vmctx.restoreTxBuilderSnapshot(txsnapshot)
		vmctx.currentStateUpdate = nil
		return err
	}
	vmctx.virtualState.ApplyStateUpdates(vmctx.currentStateUpdate)
	vmctx.currentStateUpdate = nil
	return nil
}

// creditAssetsToChain credits L1 accounts with attached assets and accrues all of them to the sender's account on-chain
func (vmctx *VMContext) creditAssetsToChain() {
	if vmctx.req.Type() == iscp.TypeOffLedger {
		// off ledger requests does not bring any deposit
		return
	}
	// update transaction builder
	vmctx.txbuilder.AddDeltaIotas(vmctx.req.Request().Assets().Iotas)
	for _, nt := range vmctx.req.Request().Assets().Tokens {
		vmctx.txbuilder.AddDeltaNativeToken(nt.ID, nt.Amount)
	}
	// sender account will be CommonAccount if sender address is not available
	vmctx.creditToAccount(vmctx.req.Request().SenderAccount(), vmctx.req.Request().Assets())
}

func (vmctx *VMContext) prepareGasBudget() {
	vmctx.loadGasPolicy()
	vmctx.calculateAffordableGasBudget()
	vmctx.gasSetBudget(vmctx.gasBudgetAffordable)
}

// callTheContract runs the contract. It catches and processes all panics except the one which cancel the whole block
func (vmctx *VMContext) callTheContract() {
	txsnapshot := vmctx.createTxBuilderSnapshot()
	snapMutations := vmctx.currentStateUpdate.Clone()

	if vmctx.req.Type() == iscp.TypeOffLedger {
		vmctx.updateOffLedgerRequestMaxAssumedNonce()
	}

	vmctx.lastError = nil
	func() {
		defer func() {
			vmctx.lastError = checkVMPluginPanic()
			if vmctx.lastError == nil {
				return
			}
			vmctx.lastResult = nil
			vmctx.Debugf("%v", vmctx.lastError)
			vmctx.Debugf(string(debug.Stack()))
		}()
		vmctx.callFromRequest()
	}()
	if vmctx.lastError != nil {
		// panic happened during VM plugin call
		// restore the state
		vmctx.restoreTxBuilderSnapshot(txsnapshot)
		vmctx.currentStateUpdate = snapMutations
	}
	vmctx.chargeGasFee()
}

// callFromRequest is the call itself. Assumes sc exists
func (vmctx *VMContext) callFromRequest() {
	vmctx.Debugf("callFromRequest: %s", vmctx.req.Request().ID().String())

	// calling only non view entry points. Calling the view will trigger error and fallback
	entryPoint := vmctx.req.Request().Target().EntryPoint
	targetContract := vmctx.contractRecord.Hname()
	vmctx.lastResult, vmctx.lastError = vmctx.callNonViewByProgramHash(
		targetContract,
		entryPoint,
		vmctx.req.Request().Params(),
		vmctx.req.Request().Transfer(),
		vmctx.contractRecord.ProgramHash,
	)
}

// chargeGasFee takes burned tokens from the sender's account
// It should always be enough because gas budget is set affordable
func (vmctx *VMContext) chargeGasFee() {
	if vmctx.req.Request().SenderAddress() == nil {
		panic("inconsistency: vmctx.req.Request().SenderAddress() == nil")
	}
	tokensToMove := vmctx.GasBurned() / vmctx.gasPolicyGasPerGasToken
	transferToValidator := &iscp.Assets{}
	if vmctx.gasFeeTokenNotIota {
		transferToValidator.Tokens = iotago.NativeTokens{{vmctx.gasFeeTokenID, new(big.Int).SetUint64(tokensToMove)}}
	} else {
		transferToValidator.Iotas = tokensToMove
	}
	sender := vmctx.req.Request().SenderAccount()
	vmctx.mustMoveBetweenAccounts(sender, vmctx.task.ValidatorFeeTarget, transferToValidator)
}

func checkVMPluginPanic() error {
	r := recover()
	if r == nil {
		return nil
	}
	// re-panic-ing if error it not user nor VM plugin fault.
	// Otherwise, the panic is wrapped into the returned error, including gas-related panic
	switch err := r.(type) {
	case *kv.DBError:
		panic(err)
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
	}
	return xerrors.Errorf("exception: %w", r)
}

// calculateAffordableGasBudget checks the account of the sender and calculates affordable gas budget
func (vmctx *VMContext) calculateAffordableGasBudget() {
	if vmctx.req.Request().SenderAddress() == nil {
		panic("inconsistency: sender must be defined when running request")
	}
	tokensAvailable := uint64(0)
	if vmctx.gasFeeTokenNotIota {
		tokensAvailableBig := vmctx.GetTokenBalance(vmctx.req.Request().SenderAccount(), &vmctx.gasFeeTokenID)
		if tokensAvailableBig != nil {
			if tokensAvailableBig.IsUint64() {
				tokensAvailable = tokensAvailableBig.Uint64()
			}
		} else {
			tokensAvailable = math.MaxUint64
		}
	} else {
		tokensAvailable = vmctx.GetIotaBalance(vmctx.req.Request().SenderAccount())
	}
	// safe arithmetics
	if tokensAvailable < math.MaxUint64/vmctx.gasPolicyGasPerGasToken {
		vmctx.gasBudgetAffordable = tokensAvailable * vmctx.gasPolicyGasPerGasToken
	} else {
		vmctx.gasBudgetAffordable = math.MaxUint64
	}

	// TODO introduce minimum balance on account
	vmctx.gasBudgetFromRequest = vmctx.req.Request().GasBudget()
	vmctx.gasBudget = vmctx.gasBudgetFromRequest
	if vmctx.gasBudget > vmctx.gasBudgetAffordable {
		vmctx.gasBudget = vmctx.gasBudgetAffordable
	}
}

func (vmctx *VMContext) loadGasPolicy() {
	// TODO load from governance contract
	vmctx.gasFeeTokenNotIota = false
	vmctx.gasFeeTokenID = iotago.NativeTokenID{}
	vmctx.gasPolicyFixedBudget = false
	vmctx.gasPolicyGasPerGasToken = 100
}

func (vmctx *VMContext) locateTargetContract() {
	// find target contract
	targetContract := vmctx.req.Request().Target().Contract
	var ok bool
	vmctx.contractRecord, ok = vmctx.findContractByHname(targetContract)
	if !ok {
		vmctx.Warnf("contract not found: %s", targetContract)
	}
	if vmctx.contractRecord.Hname() == 0 {
		vmctx.Warnf("default contract will be called")
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
