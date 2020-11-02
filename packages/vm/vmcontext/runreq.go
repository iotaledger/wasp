package vmcontext

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
)

// runTheRequest:
// - handles request token
// - processes reward logic
func (vmctx *VMContext) RunTheRequest(reqRef sctransaction.RequestRef, timestamp int64) state.StateUpdate {
	vmctx.setRequestContext(reqRef, timestamp)

	defer func() {
		vmctx.virtualState.ApplyStateUpdate(vmctx.stateUpdate)

		vmctx.log.Debugw("runTheRequest OUT USER DEFINED",
			"reqId", vmctx.reqRef.RequestID().Short(),
			"entry point", vmctx.reqRef.RequestSection().EntryPointCode().String(),
			"state update", vmctx.stateUpdate.String(),
		)
	}()

	if !vmctx.handleNodeRewards() {
		return vmctx.stateUpdate
	}
	entryPoint, ok := vmctx.getEntryPoint()
	if !ok {
		return vmctx.stateUpdate
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				vmctx.log.Errorf("Recovered from panic in SC: %v", r)
				if _, ok := r.(buffered.DBError); ok {
					// There was an error accessing the DB
					// TODO invalidate the whole block?
				}
				vmctx.Rollback()
			}
		}()
		// TODO remove param passing
		_, err := entryPoint.Call(NewSandbox(vmctx), nil)
		if err != nil {
			vmctx.log.Warnw("call to entry point",
				"err", err,
				"reqId", vmctx.reqRef.RequestID().Short(),
			)
		}
	}()
	return vmctx.stateUpdate
}

func (vmctx *VMContext) setRequestContext(reqRef sctransaction.RequestRef, timestamp int64) {
	vmctx.saveTxBuilder = vmctx.txBuilder.Clone()

	vmctx.reqRef = reqRef
	vmctx.timestamp = timestamp
	vmctx.stateUpdate = state.NewStateUpdate(reqRef.RequestID()).WithTimestamp(timestamp)
	vmctx.callStack = vmctx.callStack[:0]
	vmctx.entropy = *hashing.HashData(vmctx.entropy[:])
	// TODO budget
	vmctx.PushCallContext(vmctx.reqRef.RequestSection().Target().Index(), vmctx.reqRef.RequestSection().Args(), nil)
}

func (vmctx *VMContext) Rollback() {
	vmctx.txBuilder = vmctx.saveTxBuilder
	vmctx.stateUpdate.Clear()
}
