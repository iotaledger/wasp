package vmcontext

import (
	"math/big"
	"time"

	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"golang.org/x/xerrors"
)

var (
	MaxBlockOutputCount = uint8(iotago.MaxOutputsCount - 1) // -1 for the chain transition output
	MaxBlockInputCount  = uint8(iotago.MaxInputsCount - 2)  // // 125 (126 limit -1 for the previous state utxo)
)

// VMContext represents state of the chain during one run of the VM while processing
// a batch of requests. VMContext object mutates with each request in the bathc.
// The VMContext is created from immutable vm.VMTask object and UTXO state of the
// chain address contained in the statetxbuilder.Builder
type VMContext struct {
	// same for the block
	chainID              iscp.ChainID
	chainOwnerID         *iscp.AgentID
	chainInput           *iotago.AliasOutput
	chainInputID         iotago.UTXOInput
	processors           *processors.Cache
	virtualState         state.VirtualStateAccess
	solidStateBaseline   coreutil.StateBaseline
	remainingAfterFees   colored.Balances
	blockContext         map[iscp.Hname]*blockContext
	blockContextCloseSeq []iscp.Hname
	log                  *logger.Logger
	blockOutputCount     uint8
	// txbuilder
	txbuilder *vmtxbuilder.AnchorTransactionBuilder
	// fee related
	validatorFeeTarget *iscp.AgentID // provided by validator
	feeColor           colored.Color
	ownerFee           uint64
	validatorFee       uint64
	// events related
	maxEventSize    uint16
	maxEventsPerReq uint16
	// ---- request context
	req                      iscp.Request
	requestIndex             uint16
	requestEventIndex        uint16
	requestOutputCount       uint8
	currentStateUpdate       state.StateUpdate
	entropy                  hashing.HashValue
	contractRecord           *root.ContractRecord
	lastError                error
	lastResult               dict.Dict
	lastTotalAssets          colored.Balances
	callStack                []*callContext
	exceededBlockOutputLimit bool
	// gas related
	gas              int64
	gasBudgetEnabled bool
}

type callContext struct {
	isRequestContext bool             // is called from the request (true) or from another SC (false)
	caller           *iscp.AgentID    // calling agent
	contract         iscp.Hname       // called contract
	params           dict.Dict        // params passed
	transfer         colored.Balances // transfer passed
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
	// we create optimistic state access wrapper to be used inside the VM call.
	// It will panic any time the state is accessed.
	// The panic will be caught above and VM call will be abandoned peacefully
	optimisticStateAccess := state.WrapMustOptimisticVirtualStateAccess(task.VirtualStateAccess, task.SolidStateBaseline)

	// assert consistency
	stateHash, err := hashing.HashValueFromBytes(task.AnchorOutput.StateMetadata)
	if err != nil {
		// should never happen
		panic(xerrors.Errorf("CreateVMContext: can't parse state hash from chain input %w", err))
	}
	stateHashFromState := optimisticStateAccess.StateCommitment()
	blockIndex := optimisticStateAccess.BlockIndex()
	if stateHash != stateHashFromState || blockIndex != task.AnchorOutput.StateIndex {
		// leaving earlier, state is not consistent and optimistic reader sync didn't catch it
		panic(coreutil.ErrorStateInvalidated)
	}
	openingStateUpdate := state.NewStateUpdateWithBlocklogValues(blockIndex+1, task.Timestamp, stateHash)
	optimisticStateAccess.ApplyStateUpdates(openingStateUpdate)

	ret := &VMContext{
		chainID:              iscp.NewChainID(task.AnchorOutput.AliasID),
		chainInput:           task.AnchorOutput,
		chainInputID:         task.AnchorOutputID,
		virtualState:         optimisticStateAccess,
		solidStateBaseline:   task.SolidStateBaseline,
		validatorFeeTarget:   task.ValidatorFeeTarget,
		processors:           task.Processors,
		blockContext:         make(map[iscp.Hname]*blockContext),
		blockContextCloseSeq: make([]iscp.Hname, 0),
		log:                  task.Log,
		entropy:              task.Entropy,
		callStack:            make([]*callContext, 0),
	}
	ret.txbuilder = vmtxbuilder.NewAnchorTransactionBuilder(task.AnchorOutput, task.AnchorOutputID, task.AnchorOutput.Amount, func(id iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
		return ret.loadNativeTokensOnChain(id)
	})
	return ret
}

//nolint:revive
func (vmctx *VMContext) GetResult() (dict.Dict, colored.Balances, error, bool) {
	return vmctx.lastResult, vmctx.lastTotalAssets, vmctx.lastError, vmctx.exceededBlockOutputLimit
}

// CloseVMContext does the closing actions on the block
// return nil for normal block and rotation address for rotation block
func (vmctx *VMContext) CloseVMContext(numRequests, numSuccess, numOffLedger uint16) (uint32, hashing.HashValue, time.Time, ledgerstate.Address) {
	rotationAddr := vmctx.mustSaveBlockInfo(numRequests, numSuccess, numOffLedger)
	vmctx.closeBlockContexts()

	blockIndex := vmctx.virtualState.BlockIndex()
	stateCommitment := vmctx.virtualState.StateCommitment()
	timestamp := vmctx.virtualState.Timestamp()

	return blockIndex, stateCommitment, timestamp, rotationAddr
}

func (vmctx *VMContext) checkRotationAddress() ledgerstate.Address {
	vmctx.pushCallContext(governance.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return governance.GetRotationAddress(vmctx.State())
}

// mustSaveBlockInfo is in the blocklog partition context
// returns rotation address if this block is a rotation block
func (vmctx *VMContext) mustSaveBlockInfo(numRequests, numSuccess, numOffLedger uint16) ledgerstate.Address {
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

	blockInfo := &blocklog.BlockInfo{
		BlockIndex:            vmctx.virtualState.BlockIndex(),
		Timestamp:             vmctx.virtualState.Timestamp(),
		TotalRequests:         numRequests,
		NumSuccessfulRequests: numSuccess,
		NumOffLedgerRequests:  numOffLedger,
		PreviousStateHash:     vmctx.StateHash(),
	}

	idx := blocklog.SaveNextBlockInfo(vmctx.State(), blockInfo)
	if idx != blockInfo.BlockIndex {
		panic("CloseVMContext: inconsistent block index")
	}
	if vmctx.virtualState.PreviousStateHash() != blockInfo.PreviousStateHash {
		panic("CloseVMContext: inconsistent previous state hash")
	}
	blocklog.SaveControlAddressesIfNecessary(
		vmctx.State(),
		vmctx.StateAddress(),
		vmctx.GoverningAddress(),
		vmctx.chainInput.StateIndex,
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
