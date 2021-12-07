package vmcontext

import (
	"math/big"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
	"golang.org/x/xerrors"
)

// VMContext represents state of the chain during one run of the VM while processing
// a batch of requests. VMContext object mutates with each request in the bathc.
// The VMContext is created from immutable vm.VMTask object and UTXO state of the
// chain address contained in the statetxbuilder.Builder
type VMContext struct {
	task *vm.VMTask
	// same for the block
	chainOwnerID         *iscp.AgentID
	virtualState         state.VirtualStateAccess
	finalStateTimestamp  time.Time
	blockContext         map[iscp.Hname]*blockContext
	blockContextCloseSeq []iscp.Hname
	blockOutputCount     uint8
	txbuilder            *vmtxbuilder.AnchorTransactionBuilder
	// fee related
	feeAssetID   []byte
	ownerFee     uint64
	validatorFee uint64
	// events related
	maxEventSize    uint16
	maxEventsPerReq uint16
	// ---- request context
	req                iscp.RequestData
	requestIndex       uint16
	requestEventIndex  uint16
	currentStateUpdate state.StateUpdate
	entropy            hashing.HashValue
	contractRecord     *root.ContractRecord
	lastError          error
	lastResult         dict.Dict
	callStack          []*callContext
	// --- gas related
	// gas from the request
	gasBudgetFromRequest uint64
	// max gas budget capped by the number of tokens in the sender's account
	gasBudgetAffordable uint64
	// final gas budget set for the run
	gasBudget uint64
	// gas already burned
	gasBurned uint64
	// gas policy
	gasPolicy *governance.GasFeePolicy
}

type callContext struct {
	caller   *iscp.AgentID // calling agent
	contract iscp.Hname    // called contract
	params   dict.Dict     // params passed
	transfer *iscp.Assets  // transfer passed
}

type blockContext struct {
	obj     interface{}
	onClose func(interface{})
}

// CreateVMContext creates a context for the whole batch run
func CreateVMContext(task *vm.VMTask) *VMContext {
	// assert consistency. It is a bit redundant double check
	if len(task.Requests) == 0 {
		// should never happen
		panic(xerrors.Errorf("CreateVMContext.invalid params: must be at least 1 request"))
	}
	stateData, err := iscp.StateDataFromBytes(task.AnchorOutput.StateMetadata)
	if err != nil {
		// should never happen
		panic(xerrors.Errorf("CreateVMContext: can't parse state data from chain input %w", err))
	}
	// we create optimistic state access wrapper to be used inside the VM call.
	// It will panic any time the state is accessed.
	// The panic will be caught above and VM call will be abandoned peacefully
	optimisticStateAccess := state.WrapMustOptimisticVirtualStateAccess(task.VirtualStateAccess, task.SolidStateBaseline)

	// assert consistency
	stateHashFromState := optimisticStateAccess.StateCommitment()
	blockIndex := optimisticStateAccess.BlockIndex()
	if stateData.Commitment != stateHashFromState || blockIndex != task.AnchorOutput.StateIndex {
		// leaving earlier, state is not consistent and optimistic reader sync didn't catch it
		panic(coreutil.ErrorStateInvalidated)
	}
	openingStateUpdate := state.NewStateUpdateWithBlocklogValues(blockIndex+1, task.TimeAssumption.Time, stateData.Commitment)
	optimisticStateAccess.ApplyStateUpdates(openingStateUpdate)
	finalStateTimestamp := task.TimeAssumption.Time.Add(time.Duration(len(task.Requests)+1) * time.Nanosecond)

	ret := &VMContext{
		task:                 task,
		virtualState:         optimisticStateAccess,
		finalStateTimestamp:  finalStateTimestamp,
		blockContext:         make(map[iscp.Hname]*blockContext),
		blockContextCloseSeq: make([]iscp.Hname, 0),
		entropy:              task.Entropy,
		callStack:            make([]*callContext, 0),
	}
	nativeTokenBalanceLoader := func(id iotago.NativeTokenID) (*big.Int, *iotago.UTXOInput) {
		return ret.loadNativeTokensOnChain(id)
	}
	ret.txbuilder = vmtxbuilder.NewAnchorTransactionBuilder(
		task.AnchorOutput,
		&task.AnchorOutputID,
		task.AnchorOutput.Amount,
		nativeTokenBalanceLoader,
	)
	return ret
}

//nolint:revive
func (vmctx *VMContext) GetResult() (dict.Dict, error) {
	return vmctx.lastResult, vmctx.lastError
}

// CloseVMContext does the closing actions on the block
// return nil for normal block and rotation address for rotation block
func (vmctx *VMContext) CloseVMContext(numRequests, numSuccess, numOffLedger uint16) (uint32, hashing.HashValue, time.Time, iotago.Address) {
	rotationAddr := vmctx.mustSaveBlockInfo(numRequests, numSuccess, numOffLedger)
	vmctx.closeBlockContexts()
	vmctx.saveTokenIDInternalIndices()

	blockIndex := vmctx.virtualState.BlockIndex()
	stateCommitment := vmctx.virtualState.StateCommitment()
	timestamp := vmctx.virtualState.Timestamp()

	return blockIndex, stateCommitment, timestamp, rotationAddr
}

func (vmctx *VMContext) checkRotationAddress() iotago.Address {
	vmctx.pushCallContext(governance.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return governance.GetRotationAddress(vmctx.State())
}

// mustSaveBlockInfo is in the blocklog partition context
// returns rotation address if this block is a rotation block
func (vmctx *VMContext) mustSaveBlockInfo(numRequests, numSuccess, numOffLedger uint16) iotago.Address {
	vmctx.currentStateUpdate = state.NewStateUpdate() // need this before to make state valid

	if rotationAddress := vmctx.checkRotationAddress(); rotationAddress != nil {
		// block was marked fake by the governance contract because it is a committee rotation.
		// There was only on request in the block
		// We skip saving block information in order to avoid inconsistencies
		return rotationAddress
	}
	// block info will be stored into the separate state update
	vmctx.pushCallContext(blocklog.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	prevStateData, err := iscp.StateDataFromBytes(vmctx.task.AnchorOutput.StateMetadata)
	if err != nil {
		panic(err)
	}
	blockInfo := &blocklog.BlockInfo{
		BlockIndex:            vmctx.virtualState.BlockIndex(),
		Timestamp:             vmctx.virtualState.Timestamp(),
		TotalRequests:         numRequests,
		NumSuccessfulRequests: numSuccess,
		NumOffLedgerRequests:  numOffLedger,
		PreviousStateHash:     prevStateData.Commitment,
	}

	blocklog.SetAnchorTransactionIDOfLatestBlock(vmctx.State(), vmctx.task.AnchorOutputID.TransactionID)

	idx := blocklog.SaveNextBlockInfo(vmctx.State(), blockInfo)
	if idx != blockInfo.BlockIndex {
		panic("CloseVMContext: inconsistent block index")
	}
	if vmctx.virtualState.PreviousStateHash() != blockInfo.PreviousStateHash {
		panic("CloseVMContext: inconsistent previous state hash")
	}
	blocklog.SaveControlAddressesIfNecessary(
		vmctx.State(),
		vmctx.task.AnchorOutput.StateController,
		vmctx.task.AnchorOutput.GovernanceController,
		vmctx.task.AnchorOutput.StateIndex,
	)
	vmctx.virtualState.ApplyStateUpdates(vmctx.currentStateUpdate)
	vmctx.currentStateUpdate = nil // invalidate

	return nil
}

// closeBlockContexts closing block contexts in deterministic FIFO sequence
func (vmctx *VMContext) closeBlockContexts() {
	vmctx.currentStateUpdate = state.NewStateUpdate()
	for _, hname := range vmctx.blockContextCloseSeq {
		b := vmctx.blockContext[hname]
		b.onClose(b.obj)
	}
	vmctx.virtualState.ApplyStateUpdates(vmctx.currentStateUpdate)
	vmctx.currentStateUpdate = nil
}

func (vmctx *VMContext) saveTokenIDInternalIndices() {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	tokenUtxoIndices := vmctx.txbuilder.SortedListOfTokenIDsForOutputs()
	stateIndex := vmctx.task.AnchorOutput.StateIndex + 1 // UTXOs for the native tokens will be produced together with the next block (state update)
	accounts.SetAssetsUtxoIndices(vmctx.State(), stateIndex, tokenUtxoIndices)
}
