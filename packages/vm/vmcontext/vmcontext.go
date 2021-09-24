package vmcontext

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
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
	MaxBlockOutputCount = uint8(ledgerstate.MaxOutputCount - 1) // -1 for the chain transition output
	MaxBlockInputCount  = uint8(ledgerstate.MaxInputCount - 2)  // // 125 (126 limit -1 for the previous state utxo)
)

// VMContext represents state of the chain during one run of the VM while processing
// a batch of requests. VMContext object mutates with each request in the bathc.
// The VMContext is created from immutable vm.VMTask object and UTXO state of the
// chain address contained in the statetxbuilder.Builder
type VMContext struct {
	// same for the block
	chainID              iscp.ChainID
	chainOwnerID         iscp.AgentID
	chainInput           *ledgerstate.AliasOutput
	processors           *processors.Cache
	txBuilder            *utxoutil.Builder
	virtualState         state.VirtualState
	solidStateBaseline   coreutil.StateBaseline
	remainingAfterFees   colored.Balances
	blockContext         map[iscp.Hname]*blockContext
	blockContextCloseSeq []iscp.Hname
	log                  *logger.Logger
	blockOutputCount     uint8
	// fee related
	validatorFeeTarget iscp.AgentID // provided by validator
	feeColor           colored.Color
	ownerFee           uint64
	validatorFee       uint64
	// events related
	maxEventSize    uint16
	maxEventsPerReq uint16
	// request context
	req                      iscp.Request
	requestIndex             uint16
	requestEventIndex        uint16
	requestOutputCount       uint8
	currentStateUpdate       state.StateUpdate
	entropy                  hashing.HashValue // mutates with each request
	contractRecord           *root.ContractRecord
	lastError                error     // mutated
	lastResult               dict.Dict // mutated. Used only by 'solo'
	lastTotalAssets          colored.Balances
	callStack                []*callContext
	exceededBlockOutputLimit bool
}

type callContext struct {
	isRequestContext bool             // is called from the request (true) or from another SC (false)
	caller           iscp.AgentID     // calling agent
	contract         iscp.Hname       // called contract
	params           dict.Dict        // params passed
	transfer         colored.Balances // transfer passed
}

type blockContext struct {
	obj     interface{}
	onClose func(interface{})
}

// CreateVMContext a constructor
func CreateVMContext(task *vm.VMTask, txb *utxoutil.Builder) (*VMContext, error) {
	chainID, err := iscp.ChainIDFromAddress(task.ChainInput.Address())
	if err != nil {
		task.Log.Panicf("CreateVMContext: %v", err)
	}

	// assert consistency
	stateHash, err := hashing.HashValueFromBytes(task.ChainInput.GetStateData())
	if err != nil {
		return nil, xerrors.Errorf("CreateVMContext: can't parse state hash from chain input %w", err)
	}
	if stateHash != task.VirtualState.StateCommitment() {
		return nil, xerrors.New("CreateVMContext: state hash mismatch")
	}
	if task.VirtualState.BlockIndex() != task.ChainInput.GetStateIndex() {
		return nil, xerrors.Errorf("CreateVMContext: state index is inconsistent: block: #%d != chain output: #%d",
			task.VirtualState.BlockIndex(), task.ChainInput.GetStateIndex())
	}
	{
		openingStateUpdate := state.NewStateUpdateWithBlocklogValues(task.VirtualState.BlockIndex()+1, task.Timestamp, stateHash)
		task.VirtualState.ApplyStateUpdates(openingStateUpdate)
	}
	ret := &VMContext{
		chainID:              *chainID,
		chainInput:           task.ChainInput,
		txBuilder:            txb,
		virtualState:         task.VirtualState,
		solidStateBaseline:   task.SolidStateBaseline,
		validatorFeeTarget:   task.ValidatorFeeTarget,
		processors:           task.Processors,
		blockContext:         make(map[iscp.Hname]*blockContext),
		blockContextCloseSeq: make([]iscp.Hname, 0),
		log:                  task.Log,
		entropy:              task.Entropy,
		callStack:            make([]*callContext, 0),
	}
	// consume chain input
	err = txb.ConsumeAliasInput(task.ChainInput.Address())
	if err != nil {
		return nil, xerrors.Errorf("CreateVMContext: consume chain input %w", err)
	}
	return ret, nil
}

//nolint:revive
func (vmctx *VMContext) GetResult() (dict.Dict, colored.Balances, error, bool) {
	return vmctx.lastResult, vmctx.lastTotalAssets, vmctx.lastError, vmctx.exceededBlockOutputLimit
}

func (vmctx *VMContext) BuildTransactionEssence(stateHash hashing.HashValue, timestamp time.Time) (*ledgerstate.TransactionEssence, error) {
	if err := vmctx.txBuilder.AddAliasOutputAsRemainder(vmctx.chainID.AsAddress(), stateHash[:]); err != nil {
		return nil, xerrors.Errorf("mustFinalizeRequestCall: %v", err)
	}
	tx, _, err := vmctx.txBuilder.WithTimestamp(timestamp).BuildEssence()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CloseVMContext does the closing actions on the block
// return nil for normal block and rotation address for rotation block
func (vmctx *VMContext) CloseVMContext(numRequests, numSuccess, numOffLedger uint16) ledgerstate.Address {
	rotationAddr := vmctx.mustSaveBlockInfo(numRequests, numSuccess, numOffLedger)
	vmctx.closeBlockContexts()
	return rotationAddr
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
		vmctx.log.Panicf("CloseVMContext: inconsistent block index")
	}
	if vmctx.virtualState.PreviousStateHash() != blockInfo.PreviousStateHash {
		vmctx.log.Panicf("CloseVMContext: inconsistent previous state hash")
	}
	blocklog.SaveControlAddressesIfNecessary(
		vmctx.State(),
		vmctx.StateAddress(),
		vmctx.GoverningAddress(),
		vmctx.chainInput.GetStateIndex(),
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
