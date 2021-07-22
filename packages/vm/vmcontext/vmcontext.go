package vmcontext

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
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
	remainingAfterFees   *ledgerstate.ColoredBalances
	blockContext         map[iscp.Hname]*blockContext
	blockContextCloseSeq []iscp.Hname
	log                  *logger.Logger
	// fee related
	validatorFeeTarget iscp.AgentID // provided by validator
	feeColor           ledgerstate.Color
	ownerFee           uint64
	validatorFee       uint64
	// request context
	req                iscp.Request
	requestIndex       uint16
	requestEventIndex  uint16
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
	caller           iscp.AgentID                 // calling agent
	contract         iscp.Hname                   // called contract
	params           dict.Dict                    // params passed
	transfer         *ledgerstate.ColoredBalances // transfer passed
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
		chainInput:           task.ChainInput,
		txBuilder:            txb,
		virtualState:         task.VirtualState,
		solidStateBaseline:   task.SolidStateBaseline,
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
	vmctx.currentStateUpdate = state.NewStateUpdate() // need ths before to make state valid

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
	}

	idx := blocklog.SaveNextBlockInfo(vmctx.State(), blockInfo)
	if idx != blockInfo.BlockIndex {
		vmctx.log.Panicf("CloseVMContext: inconsistent block index")
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
