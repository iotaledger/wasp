package vmcontext

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"runtime/debug"
)

// runTheRequest:
// - handles request token
// - processes reward logic
func (vmctx *VMContext) RunTheRequest(reqRef sctransaction.RequestRef, timestamp int64) state.StateUpdate {
	if !vmctx.setRequestContext(reqRef, timestamp) {
		vmctx.log.Errorf("not found contract '%s' in %s",
			reqRef.RequestSection().Target().Hname().String(), vmctx.reqRef.RequestID().Short())
		return state.NewStateUpdate(reqRef.RequestID()).WithTimestamp(timestamp)
	}

	defer func() {
		vmctx.virtualState.ApplyStateUpdate(vmctx.stateUpdate)

		vmctx.log.Debugw("runTheRequest OUT",
			"reqId", vmctx.reqRef.RequestID().Short(),
			"entry point", vmctx.reqRef.RequestSection().EntryPointCode().String(),
			//"state update", vmctx.stateUpdate.String(),
		)
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				vmctx.lastResult = nil
				vmctx.lastError = fmt.Errorf("recovered from panic in VM: %v", r)
				vmctx.log.Error(vmctx.lastError)
				debug.PrintStack()
				if _, ok := r.(buffered.DBError); ok {
					// There was an error accessing the DB
					// TODO invalidate the whole block?
				}
				vmctx.Rollback()
			}
		}()
		vmctx.mustCallFromRequest()
	}()
	return vmctx.stateUpdate
}

func (vmctx *VMContext) setRequestContext(reqRef sctransaction.RequestRef, timestamp int64) bool {
	reqHname := reqRef.RequestSection().Target().Hname()
	vmctx.saveTxBuilder = vmctx.txBuilder.Clone()
	vmctx.reqRef = reqRef
	vmctx.reqHname = reqHname

	vmctx.timestamp = timestamp
	vmctx.stateUpdate = state.NewStateUpdate(reqRef.RequestID()).WithTimestamp(timestamp)
	vmctx.callStack = vmctx.callStack[:0]
	vmctx.entropy = *hashing.HashData(vmctx.entropy[:])
	feeColor, ownerFee, validatorFee, ok := vmctx.getFeeInfo(reqHname)
	if !ok {
		return false
	}
	vmctx.feeColor = feeColor
	vmctx.ownerFee = ownerFee
	vmctx.validatorFee = validatorFee
	return true
}

func (vmctx *VMContext) Rollback() {
	vmctx.txBuilder = vmctx.saveTxBuilder
	vmctx.stateUpdate.Clear()
}

func (vmctx *VMContext) FinalizeTransactionEssence(blockIndex uint32, stateHash *hashing.HashValue, timestamp int64) (*sctransaction.Transaction, error) {
	// add state block
	err := vmctx.txBuilder.SetStateParams(blockIndex, stateHash, timestamp)
	if err != nil {
		return nil, err
	}
	// create result transaction
	//vmctx.log.Debugf("-- before building the tx:\n%s\n", vmctx.txBuilder.Dump(true))
	tx, err := vmctx.txBuilder.Build()
	if err != nil {
		return nil, err
	}
	return tx, nil
}
