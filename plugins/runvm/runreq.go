package runvm

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/builtin"
	"github.com/iotaledger/wasp/packages/vm/processor"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
)

// runTheRequest:
// - handles request token
// - processes reward logic
// - checks authorisations for protected requests
// - redirects reserved request codes (is supported) to hardcoded processing
// - redirects not reserved codes (is supported) to SC VM
// - in case of something not correct the whole operation is NOP, however
//   all the sent fees and other funds remains in the SC address (this may change).
func runTheRequest(vmctx *vm.VMContext) {
	vmctx.Log.Debugf("runTheRequest IN:\n%s\n", vmctx.RequestRef.RequestBlock().String(vmctx.RequestRef.RequestId()))

	if !handleNodeRewards(vmctx) {
		return
	}

	reqBlock := vmctx.RequestRef.RequestBlock()
	if reqBlock.RequestCode().IsProtected() {
		// check authorisation
		if !vmctx.RequestRef.IsAuthorised(&vmctx.OwnerAddress) {
			// if protected call is not authorised by the containing transaction, do nothing
			// the result will be taking all iotas and no effect on state
			// Maybe it is nice to return back all iotas exceeding minimum reward ??? TODO

			vmctx.Log.Warnf("protected request %s (code %s) is not authorised by the SC owner %s. Sender: %s",
				vmctx.RequestRef.RequestId().String(), reqBlock.RequestCode(),
				vmctx.OwnerAddress.String(), vmctx.RequestRef.Sender().String(),
			)
			vmctx.Log.Debugw("protected request is not authorised",
				"req", vmctx.RequestRef.RequestId().String(),
				"code", reqBlock.RequestCode(),
				"owner", vmctx.OwnerAddress.String(),
				"inputs", util.InputsToStringByAddress(vmctx.RequestRef.Tx.Inputs()),
			)
			return
		}
		if vmctx.VirtualState.StateIndex() > 0 && !vmctx.VirtualState.InitiatedBy(&vmctx.OwnerAddress) {
			// for states after #0 it is required to have record about initiator's address in the solid state
			// to prevent attack when owner (initiator) address is overwritten in the quorum of bootup records
			// TODO protection may also be set at the lowest level of the solid state. i.e. some metadata that variable
			// is protected by some address and authorisation with that address is needed to modify the value

			vmctx.Log.Errorf("inconsistent state: variable '%s' != owner record from bootup record '%s'",
				vmconst.VarNameOwnerAddress, vmctx.OwnerAddress.String())

			return
		}
	}
	// authorisation check passed
	if reqBlock.RequestCode().IsReserved() {
		// finding and running builtin entry point
		entryPoint, ok := builtin.Processor.GetEntryPoint(reqBlock.RequestCode())
		if !ok {
			vmctx.Log.Warnf("can't find entry point for request code %s in the builtin processor", reqBlock.RequestCode())
			return
		}
		entryPoint.Run(sandbox.NewSandbox(vmctx))

		defer vmctx.Log.Debugw("runTheRequest OUT BUILTIN",
			"reqId", vmctx.RequestRef.RequestId().Short(),
			"programHash", vmctx.ProgramHash.String(),
			"code", vmctx.RequestRef.RequestBlock().RequestCode().String(),
			"state update", vmctx.StateUpdate.String(),
		)
		return
	}

	// request requires user-defined program on VM
	proc, err := processor.Acquire(vmctx.ProgramHash.String())
	if err != nil {
		vmctx.Log.Warn(err)
		return
	}
	defer processor.Release(vmctx.ProgramHash.String())

	entryPoint, ok := proc.GetEntryPoint(reqBlock.RequestCode())
	if !ok {
		vmctx.Log.Warnf("can't find entry point for request code %s in the user-defined processor prog hash: %s",
			reqBlock.RequestCode(), vmctx.ProgramHash.String())
		return
	}

	sndbox := sandbox.NewSandbox(vmctx)
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
		entryPoint.Run(sndbox)
	}()

	defer vmctx.Log.Debugw("runTheRequest OUT USER DEFINED",
		"reqId", vmctx.RequestRef.RequestId().Short(),
		"programHash", vmctx.ProgramHash.String(),
		"code", vmctx.RequestRef.RequestBlock().RequestCode().String(),
		"state update", vmctx.StateUpdate.String(),
	)
}
