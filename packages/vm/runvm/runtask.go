package runvm

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/statetxbuilder"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
	"time"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
)

// RunComputationsAsync runs computations for the batch of requests in the background
func RunComputationsAsync(ctx *vm.VMTask) error {
	if len(ctx.Requests) == 0 {
		return fmt.Errorf("RunComputationsAsync: must be at least 1 request")
	}

	txb, err := statetxbuilder.New(address.Address(ctx.ChainID), ctx.Color, ctx.Balances)
	if err != nil {
		ctx.Log.Debugf("statetxbuilder.New: %v", err)
		return err
	}

	// TODO 1 graceful shutdown of the running VM task (with daemon)
	// TODO 2 timeout for VM. Gas limit

	go runTask(ctx, txb)

	return nil
}

// runTask runs batch of requests on VM
func runTask(task *vm.VMTask, txb *statetxbuilder.Builder) {
	task.Log.Debugw("runTask IN",
		"chainID", task.ChainID.String(),
		"timestamp", task.Timestamp,
		"state index", task.VirtualState.BlockIndex(),
		"num req", len(task.Requests),
	)
	vmctx, err := vmcontext.NewVMContext(task, txb)
	if err != nil {
		task.OnFinish(nil, nil, fmt.Errorf("runTask.createVMContext: %v", err))
		return
	}

	stateUpdates := make([]state.StateUpdate, 0, len(task.Requests))
	var lastResult dict.Dict
	var lastErr error
	var lastStateUpdate state.StateUpdate

	// loop over the batch of requests and run each request on the VM.
	// the result accumulates in the VMContext and in the list of stateUpdates
	timestamp := task.Timestamp
	for _, reqRef := range task.Requests {
		vmctx.RunTheRequest(reqRef, timestamp)
		lastStateUpdate, lastResult, lastErr = vmctx.GetResult()

		stateUpdates = append(stateUpdates, lastStateUpdate)
		if timestamp != 0 {
			// increasing (nonempty) timestamp for 1 nanosecond for each request in the batch
			// the reason is to provide a different timestamp for each VM call and remain deterministic
			timestamp += 1
		}
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
	stateHash := vsClone.Hash()
	task.ResultTransaction, err = vmctx.FinalizeTransactionEssence(
		task.VirtualState.BlockIndex()+1,
		stateHash,
		vsClone.Timestamp(),
	)
	if err != nil {
		task.OnFinish(nil, nil, fmt.Errorf("RunVM.FinalizeTransactionEssence: %v", err))
		return
	}
	// Note: can't take tx ID!!
	task.Log.Debugw("runTask OUT",
		"batch size", task.ResultBlock.Size(),
		"block index", task.ResultBlock.StateIndex(),
		"variable state hash", stateHash.String(),
		"tx essence hash", hashing.HashData(task.ResultTransaction.EssenceBytes()).String(),
		"tx finalTimestamp", time.Unix(0, task.ResultTransaction.MustState().Timestamp()),
	)
	task.OnFinish(lastResult, lastErr, nil)
}
