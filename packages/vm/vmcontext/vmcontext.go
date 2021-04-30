package vmcontext

import (
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
	chainID            coretypes.ChainID
	chainOwnerID       coretypes.AgentID
	processors         *processors.ProcessorCache
	txBuilder          *utxoutil.Builder
	stateUpdates       []state.StateUpdate
	virtualState       state.VirtualState
	remainingAfterFees *ledgerstate.ColoredBalances
	log                *logger.Logger
	// fee related
	validatorFeeTarget coretypes.AgentID // provided by validator
	feeColor           ledgerstate.Color
	ownerFee           uint64
	validatorFee       uint64
	// request context
	req             coretypes.Request
	requestIndex    uint16
	entropy         hashing.HashValue // mutates with each request
	contractRecord  *root.ContractRecord
	stateUpdate     state.StateUpdate
	lastError       error     // mutated
	lastResult      dict.Dict // mutated. Used only by 'solo'
	lastTotalAssets *ledgerstate.ColoredBalances
	callStack       []*callContext
}

type callContext struct {
	isRequestContext bool                         // is called from the request (true) or from another SC (false)
	caller           coretypes.AgentID            // calling agent
	contract         coretypes.Hname              // called contract
	params           dict.Dict                    // params passed
	transfer         *ledgerstate.ColoredBalances // transfer passed
}

// MustNewVMContext a constructor
func MustNewVMContext(task *vm.VMTask, txb *utxoutil.Builder) (*VMContext, error) {
	chainID, err := coretypes.ChainIDFromAddress(task.ChainInput.Address())
	if err != nil {
		return nil, xerrors.Errorf("MustNewVMContext: %v", err)
	}

	{
		// assert consistency
		stateHash, err := hashing.HashValueFromBytes(task.ChainInput.GetStateData())
		if err != nil {
			// chain input must always be present
			return nil, xerrors.Errorf("MustNewVMContext: can't parse state hash %v", err)
		}
		if stateHash != task.VirtualState.Hash() {
			return nil, xerrors.New("MustNewVMContext: state hash mismatch")
		}
		if task.VirtualState.BlockIndex() != task.ChainInput.GetStateIndex() {
			return nil, xerrors.New("MustNewVMContext: state index is inconsistent")
		}
	}

	vs := task.VirtualState.Clone()
	blockUpdate := state.NewStateUpdateWithBlockIndexMutation(task.VirtualState.BlockIndex()+1, task.Timestamp)
	vs.ApplyStateUpdate(blockUpdate)
	ret := &VMContext{
		chainID:      *chainID,
		txBuilder:    txb,
		stateUpdates: make([]state.StateUpdate, 0, len(task.Requests)+1),
		virtualState: vs,
		processors:   task.Processors,
		log:          task.Log,
		entropy:      task.Entropy,
		callStack:    make([]*callContext, 0),
	}
	ret.stateUpdates = append(ret.stateUpdates, blockUpdate)

	err = txb.ConsumeAliasInput(task.ChainInput.Address())
	if err != nil {
		// chain input must always be present
		return nil, xerrors.Errorf("MustNewVMContext: can't find chain input %w", err)
	}
	return ret, nil
}

func (vmctx *VMContext) GetStateUpdates() []state.StateUpdate {
	return vmctx.stateUpdates
}

func (vmctx *VMContext) GetResult() (dict.Dict, *ledgerstate.ColoredBalances, error) {
	return vmctx.lastResult, vmctx.lastTotalAssets, vmctx.lastError
}
