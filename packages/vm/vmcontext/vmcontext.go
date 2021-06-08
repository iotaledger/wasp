package vmcontext

import (
	"time"

	"github.com/iotaledger/wasp/packages/coretypes/coreutil"

	"github.com/iotaledger/wasp/packages/vm/core/blocklog"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"golang.org/x/xerrors"
)

// VMContext represents state of the chain during one run of the VM while processing
// a batch of requests. VMContext object mutates with each request in the bathc.
// The VMContext is created from immutable vm.VMTask object and UTXO state of the
// chain address contained in the statetxbuilder.Builder
type VMContext struct {
	// same for the block
	chainID              coretypes.ChainID
	chainOwnerID         coretypes.AgentID
	processors           *processors.ProcessorCache
	txBuilder            *utxoutil.Builder
	virtualState         state.VirtualState
	solidStateBaseline   coreutil.StateBaseline
	remainingAfterFees   *ledgerstate.ColoredBalances
	blockContext         map[coretypes.Hname]*blockContext
	blockContextCloseSeq []coretypes.Hname
	log                  *logger.Logger
	// fee related
	validatorFeeTarget coretypes.AgentID // provided by validator
	feeColor           ledgerstate.Color
	ownerFee           uint64
	validatorFee       uint64
	// request context
	req                coretypes.Request
	requestIndex       uint16
	currentStateUpdate state.StateUpdate
	entropy            hashing.HashValue // mutates with each request
	contractRecord     *root.ContractRecord
	lastError          error     // mutated
	lastResult         dict.Dict // mutated. Used only by 'solo'
	lastTotalAssets    *ledgerstate.ColoredBalances
	callStack          []*callContext
}

type callContext struct {
	isRequestContext bool                         // is called from the request (true) or from another SC (false)
	caller           coretypes.AgentID            // calling agent
	contract         coretypes.Hname              // called contract
	params           dict.Dict                    // params passed
	transfer         *ledgerstate.ColoredBalances // transfer passed
}

type blockContext struct {
	obj     interface{}
	onClose func(interface{})
}

// CreateVMContext a constructor
func CreateVMContext(task *vm.VMTask, txb *utxoutil.Builder) (*VMContext, error) {
	chainID, err := coretypes.ChainIDFromAddress(task.ChainInput.Address())
	if err != nil {
		task.Log.Panicf("CreateVMContext: %v", err)
	}
	{
		// assert consistency
		stateHash, err := hashing.HashValueFromBytes(task.ChainInput.GetStateData())
		if err != nil {
			return nil, xerrors.Errorf("CreateVMContext: can't parse state hash from chain input %w", err)
		}
		if stateHash != task.VirtualState.Hash() {
			return nil, xerrors.New("CreateVMContext: state hash mismatch")
		}
		if task.VirtualState.BlockIndex() != task.ChainInput.GetStateIndex() {
			return nil, xerrors.New("CreateVMContext: state index is inconsistent")
		}
	}
	{
		openingStateUpdate := state.NewStateUpdateWithBlockIndexMutation(task.VirtualState.BlockIndex()+1, task.Timestamp)
		task.VirtualState.ApplyStateUpdates(openingStateUpdate)
	}
	ret := &VMContext{
		chainID:              *chainID,
		txBuilder:            txb,
		virtualState:         task.VirtualState,
		solidStateBaseline:   task.SolidStateBaseline,
		processors:           task.Processors,
		blockContext:         make(map[coretypes.Hname]*blockContext),
		blockContextCloseSeq: make([]coretypes.Hname, 0),
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

func (vmctx *VMContext) GetResult() (dict.Dict, *ledgerstate.ColoredBalances, error) {
	return vmctx.lastResult, vmctx.lastTotalAssets, vmctx.lastError
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
func (vmctx *VMContext) CloseVMContext(numRequests, numSuccess, numOffLedger uint16) {
	vmctx.mustSaveBlockInfo(numRequests, numSuccess, numOffLedger)
	vmctx.closeBlockContexts()
}

func (vmctx *VMContext) mustSaveBlockInfo(numRequests, numSuccess, numOffLedger uint16) {
	// block info will be stored into the separate state update
	vmctx.currentStateUpdate = state.NewStateUpdate()
	vmctx.pushCallContext(blocklog.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	blockInfo := &blocklog.BlockInfo{
		BlockIndex:            vmctx.virtualState.BlockIndex(),
		Timestamp:             vmctx.virtualState.Timestamp(),
		TotalRequests:         numRequests,
		NumSuccessfulRequests: numSuccess,
		NumOffLedgerRequests:  numOffLedger,
	}

	idx := blocklog.SaveNextBlockInfo(vmctx.State(), blockInfo)
	if idx != blockInfo.BlockIndex {
		vmctx.log.Panicf("CloseVMContext: inconsistent block index")
	}
	vmctx.virtualState.ApplyStateUpdates(vmctx.currentStateUpdate)
	vmctx.currentStateUpdate = nil // invalidate
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
