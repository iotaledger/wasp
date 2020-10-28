package runvm

import (
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/processor"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
)

// runTheRequest:
// - handles request token
// - processes reward logic
func runTheRequest(vmctx *vm.VMContext) {
	if !handleNodeRewards(vmctx) {
		return
	}

	reqBlock := vmctx.RequestRef.RequestBlock()
	// request requires user-defined program on VM
	proc, err := processor.Acquire(vmctx.ProgramHash.String())
	if err != nil {
		vmctx.Log.Warn(err)
		return
	}
	defer processor.Release(vmctx.ProgramHash.String())

	entryPoint, ok := proc.GetEntryPoint(reqBlock.EntryPointCode())
	if !ok {
		vmctx.Log.Warnf("can't find entry point for request code %s in the user-defined processor prog hash: %s",
			reqBlock.EntryPointCode(), vmctx.ProgramHash.String())
		return
	}
	vmctx.Log.Debugf("processing entry point %s for processor prog hash: %s",
		reqBlock.EntryPointCode().String(), vmctx.ProgramHash.String())

	sandbox := sandbox.New(vmctx)
	func() {
		defer func() {
			if r := recover(); r != nil {
				vmctx.Log.Errorf("Recovered from panic in SC: %v", r)
				if _, ok := r.(buffered.DBError); ok {
					// There was an error accessing the DB
					// TODO invalidate the whole batch?
				}
				sandbox.Rollback()
			}
		}()
		_, err := entryPoint.Call(sandbox, reqBlock.Args())
		if err != nil {
			vmctx.Log.Warnw("call to entry point",
				"err", err,
				"reqId", vmctx.RequestRef.RequestID().Short(),
			)
		}
	}()

	defer vmctx.Log.Debugw("runTheRequest OUT USER DEFINED",
		"reqId", vmctx.RequestRef.RequestID().Short(),
		"programHash", vmctx.ProgramHash.String(),
		"entry point", vmctx.RequestRef.RequestBlock().EntryPointCode().String(),
		"state update", vmctx.StateUpdate.String(),
	)
}
