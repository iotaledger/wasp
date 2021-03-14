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
)

// VMContext represents state of the chain during one run of the VM while processing
// a batch of requests. VMContext object mutates with each request in the bathc.
// The VMContext is created from immutable vm.VMTask object and UTXO state of the
// chain address contained in the statetxbuilder.Builder
type VMContext struct {
	// same for the block
	chainID      coretypes.ChainID
	chainOwnerID coretypes.AgentID
	processors   *processors.ProcessorCache
	txBuilder    *utxoutil.Builder  // mutated
	virtualState state.VirtualState // mutated
	log          *logger.Logger
	// fee related
	validatorFeeTarget coretypes.AgentID // provided by validator
	feeColor           ledgerstate.Color
	ownerFee           int64
	validatorFee       int64
	// request context
	entropy        hashing.HashValue // mutates with each request
	req            *sctransaction.Request
	reqHname       coretypes.Hname
	contractRecord *root.ContractRecord
	timestamp      int64
	stateUpdate    state.StateUpdate
	lastError      error     // mutated
	lastResult     dict.Dict // mutated. Used only by 'solo'
	callStack      []*callContext
}

type callContext struct {
	isRequestContext bool                      // is called from the request (true) or from another SC (false)
	caller           coretypes.AgentID         // calling agent
	contract         coretypes.Hname           // called contract
	params           dict.Dict                 // params passed
	transfer         coretypes.ColoredBalances // transfer passed
}

// MustNewVMContext a constructor
func MustNewVMContext(task *vm.VMTask, txb *utxoutil.Builder) *VMContext {
	chainID, err := coretypes.NewChainIDFromAddress(task.ChainInput.Address())
	if err != nil {
		task.Log.Panicf("MustNewVMContext: %v", err)
	}
	return &VMContext{
		chainID:      chainID,
		txBuilder:    txb,
		virtualState: task.VirtualState.Clone(),
		processors:   task.Processors,
		log:          task.Log,
		entropy:      task.Entropy,
		callStack:    make([]*callContext, 0),
	}
}

func (vmctx *VMContext) GetResult() (state.StateUpdate, dict.Dict, error) {
	return vmctx.stateUpdate, vmctx.lastResult, vmctx.lastError
}
