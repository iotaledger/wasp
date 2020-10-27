package runvm

import (
	"github.com/iotaledger/wasp/packages/kv"
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

	sndbox := sandbox.New(vmctx)
	func() {
		defer func() {
			if r := recover(); r != nil {
				vmctx.Log.Errorf("Recovered from panic in SC: %v", r)
				if _, ok := r.(kv.DBError); ok {
					// There was an error accessing the DB
					// TODO invalidate the whole batch?
				}
				sndbox.Rollback()
			}
		}()
		entryPoint.Call(sndbox, nil)
	}()

	defer vmctx.Log.Debugw("runTheRequest OUT USER DEFINED",
		"reqId", vmctx.RequestRef.RequestID().Short(),
		"programHash", vmctx.ProgramHash.String(),
		"code", vmctx.RequestRef.RequestBlock().EntryPointCode().String(),
		"state update", vmctx.StateUpdate.String(),
	)
}
