package vmcontext

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
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
	virtualState       state.VirtualState
	remainingAfterFees *ledgerstate.ColoredBalances
	log                *logger.Logger
	// fee related
	validatorFeeTarget coretypes.AgentID // provided by validator
	feeColor           ledgerstate.Color
	ownerFee           uint64
	validatorFee       uint64
	// request context
	req             *sctransaction.Request
	entropy         hashing.HashValue // mutates with each request
	contractRecord  *root.ContractRecord
	timestamp       int64
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
	chainID, err := coretypes.NewChainIDFromAddress(task.ChainInput.Address())
	if err != nil {
		return nil, xerrors.Errorf("MustNewVMContext: %v", err)
	}
	ret := &VMContext{
		chainID:      *chainID,
		txBuilder:    txb,
		virtualState: task.VirtualState.Clone(),
		processors:   task.Processors,
		log:          task.Log,
		entropy:      task.Entropy,
		timestamp:    task.Timestamp.UnixNano(),
		callStack:    make([]*callContext, 0),
	}
	stateHash, err := hashing.HashValueFromBytes(task.ChainInput.GetStateData())
	if err != nil {
		// chain input must always be present
		return nil, xerrors.Errorf("MustNewVMContext: can't parse state hash %v", err)
	}
	if stateHash != ret.virtualState.Hash() {
		return nil, xerrors.New("MustNewVMContext: state hash mismatch")
	}
	if ret.virtualState.BlockIndex() != task.ChainInput.GetStateIndex() {
		return nil, xerrors.New("MustNewVMContext: state index is inconsistent")
	}
	err = txb.ConsumeAliasInput(task.ChainInput.Address())
	if err != nil {
		// chain input must always be present
		return nil, xerrors.Errorf("MustNewVMContext: can't find chain input %v", err)
	}
	return ret, nil
}

func (vmctx *VMContext) GetResult() (state.StateUpdate, dict.Dict, *ledgerstate.ColoredBalances, error) {
	return vmctx.stateUpdate, vmctx.lastResult, vmctx.lastTotalAssets, vmctx.lastError
}
