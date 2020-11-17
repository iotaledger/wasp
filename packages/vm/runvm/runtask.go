package runvm

import (
	"fmt"
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
		return fmt.Errorf("must be at least 1 request")
	}

	txb, err := statetxbuilder.New(ctx.Color)
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
		"state index", task.VirtualState.StateIndex(),
		"num req", len(task.Requests),
		"leader", task.LeaderPeerIndex,
	)

	vmctx, err := vmcontext.NewVMContext(task, txb)
	if err != nil {
		task.OnFinish(fmt.Errorf("runTask.createVMContext: %v", err))
		return
	}
	stateUpdates := make([]state.StateUpdate, 0, len(task.Requests))
	timestamp := task.Timestamp
	for _, reqRef := range task.Requests {
		stateUpdate := vmctx.RunTheRequest(reqRef, timestamp)

		stateUpdates = append(stateUpdates, stateUpdate)
		// update state
		if timestamp != 0 {
			// increasing (nonempty) timestamp for 1 nanosecond for each request in the batch
			// the reason is to provide a different timestamp for each VM call and remain deterministic
			timestamp += 1
		}
	}
	if len(stateUpdates) == 0 {
		// should not happen
		task.OnFinish(fmt.Errorf("RunVM: no state updates were produced"))
		return
	}

	// create block from state updates.
	task.ResultBlock, err = state.NewBlock(stateUpdates)
	if err != nil {
		task.OnFinish(fmt.Errorf("RunVM.NewBlock: %v", err))
		return
	}
	task.ResultBlock.WithBlockIndex(task.VirtualState.StateIndex() + 1)

	// calculate resulting state hash
	vsClone := task.VirtualState.Clone()
	if err = vsClone.ApplyBatch(task.ResultBlock); err != nil {
		task.OnFinish(fmt.Errorf("RunVM.ApplyBatch: %v", err))
		return
	}
	stateHash := vsClone.Hash()
	task.ResultTransaction, err = vmctx.FinalizeTransactionEssence(
		task.VirtualState.StateIndex()+1,
		stateHash,
		vsClone.Timestamp(),
	)
	if err != nil {
		task.OnFinish(fmt.Errorf("RunVM.FinalizeTransactionEssence: %v", err))
		return
	}
	// Note: can't take tx ID!!
	task.Log.Debugw("runTask OUT",
		"result batch size", task.ResultBlock.Size(),
		"result batch state index", task.ResultBlock.StateIndex(),
		"result variable state hash", stateHash.String(),
		"result essence hash", hashing.HashData(task.ResultTransaction.EssenceBytes()).String(),
		"result tx finalTimestamp", time.Unix(0, task.ResultTransaction.MustState().Timestamp()),
	)
	task.OnFinish(nil)
}
