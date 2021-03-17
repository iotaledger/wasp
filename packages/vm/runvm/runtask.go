package runvm

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
)

// MustRunComputationsAsync runs computations for the batch of requests in the background
// This is the main entry point to the VM
// TODO timeout for VM. Gas limit
func MustRunComputationsAsync(ctx *vm.VMTask) {
	if len(ctx.Requests) == 0 {
		ctx.Log.Panicf("MustRunComputationsAsync: must be at least 1 request")
	}
	outputs := sctransaction.OutputsFromRequests(ctx.Requests...)
	txb := utxoutil.NewBuilder(append(outputs, ctx.ChainInput)...)

	go runTask(ctx, txb)
}

// runTask runs batch of requests on VM
func runTask(task *vm.VMTask, txb *utxoutil.Builder) {
	task.Log.Debugw("runTask IN",
		"chainID", task.ChainInput.Address().String(),
		"timestamp", task.Timestamp,
		"block index", task.VirtualState.BlockIndex(),
		"num req", len(task.Requests),
	)
	vmctx, err := vmcontext.MustNewVMContext(task, txb)
	if err != nil {
		task.Log.Panicf("runTask: %v", err)
	}

	stateUpdates := make([]state.StateUpdate, 0, len(task.Requests))
	var lastResult dict.Dict
	var lastErr error
	var lastStateUpdate state.StateUpdate

	// loop over the batch of requests and run each request on the VM.
	// the result accumulates in the VMContext and in the list of stateUpdates
	for i, req := range task.Requests {
		vmctx.RunTheRequest(req, i)
		lastStateUpdate, lastResult, lastErr = vmctx.GetResult()

		stateUpdates = append(stateUpdates, lastStateUpdate)
	}

	// create block from state updates.
	task.ResultBlock, err = state.NewBlock(stateUpdates)
	if err != nil {
		task.OnFinish(nil, nil, fmt.Errorf("RunVM.NewBlock: %v", err))
		return
	}
	task.ResultBlock.WithBlockIndex(task.VirtualState.BlockIndex() + 1)

	// calculate resulting state hash
	vsClone := task.VirtualState.Clone()
	if err = vsClone.ApplyBlock(task.ResultBlock); err != nil {
		task.OnFinish(nil, nil, fmt.Errorf("RunVM.ApplyBlock: %v", err))
		return
	}

	task.ResultTransaction, err = vmctx.BuildTransactionEssence(vsClone.BlockIndex(), vsClone.Hash())
	if err != nil {
		task.OnFinish(nil, nil, fmt.Errorf("RunVM.BuildTransactionEssence: %v", err))
		return
	}
	// Note: can't take tx ID!!
	task.Log.Debugw("runTask OUT",
		"batch size", task.ResultBlock.Size(),
		"block index", task.ResultBlock.StateIndex(),
		"variable state hash", vsClone.Hash().String(),
		"tx essence hash", hashing.HashData(task.ResultTransaction.Bytes()).String(),
		"tx finalTimestamp", task.ResultTransaction.Timestamp(),
	)
	task.OnFinish(lastResult, lastErr, nil)
}
