package vmcontext

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"runtime/debug"
)

// runTheRequest:
// - handles request token
// - processes reward logic
func (vmctx *VMContext) RunTheRequest(reqRef sctransaction.RequestRef, timestamp int64) {
	if !vmctx.prepareRequestContext(reqRef, timestamp) {
		return
	}
	var err error
	func() {
		// panic catcher for the whole call from request to the VM
		defer func() {
			if r := recover(); r != nil {
				vmctx.lastResult = nil
				vmctx.lastError = fmt.Errorf("recovered from panic in VM: %v", r)
				vmctx.log.Error(vmctx.lastError)
				debug.PrintStack()
				if dberr, ok := r.(buffered.DBError); ok {
					// There was an error accessing the DB
					// The world stops
					vmctx.Panicf("DB error: %v", dberr)
				}
				vmctx.txBuilder = vmctx.baselineTxBuilder
				vmctx.txBuilder.MustValidate()
				vmctx.stateUpdate.Clear()
			}
			err = vmctx.lastError
		}()
		vmctx.mustCallFromRequest()
	}()

	if err != nil {
		// TODO fallback processing on any error returned from the call
	}
	vmctx.chainlogRequest(err)
	vmctx.virtualState.ApplyStateUpdate(vmctx.stateUpdate)

	vmctx.log.Debugw("runTheRequest OUT",
		"reqId", vmctx.reqRef.RequestID().Short(),
		"entry point", vmctx.reqRef.RequestSection().EntryPointCode().String(),
		//"state update", vmctx.stateUpdate.String(),
	)
}

func (vmctx *VMContext) chainlogRequest(err error) {
	e := "Ok"
	if err != nil {
		e = err.Error()
	}
	msg := fmt.Sprintf("[req] %s: %s", vmctx.reqRef.RequestID().String(), e)
	vmctx.log.Infof("chainlog -> '%s'", msg)
	vmctx.StoreToChainLog(vmctx.reqHname, []byte(msg))
}

func (vmctx *VMContext) prepareRequestContext(reqRef sctransaction.RequestRef, timestamp int64) bool {
	reqHname := reqRef.RequestSection().Target().Hname()
	vmctx.reqRef = reqRef
	vmctx.reqHname = reqHname

	vmctx.timestamp = timestamp
	vmctx.stateUpdate = state.NewStateUpdate(reqRef.RequestID()).WithTimestamp(timestamp)
	vmctx.callStack = vmctx.callStack[:0]
	vmctx.entropy = *hashing.HashData(vmctx.entropy[:])

	// TODO handle tokens and save the state update baseline
	vmctx.baselineTxBuilder = vmctx.txBuilder.Clone()

	if isInitChainRequest(reqRef) {
		return true
	}
	// ordinary request, only makes sense when chain is already deployed
	info, err := vmctx.getChainInfo()
	if err != nil {
		vmctx.log.Errorf("prepareRequestContext: %s", err)
		return false
	}
	if info.ChainID != vmctx.chainID {
		vmctx.log.Panicf("prepareRequestContext: major inconsistency of chainID")
		return false
	}
	vmctx.chainOwnerID = info.ChainOwnerID
	feeColor, ownerFee, validatorFee, ok := vmctx.getFeeInfo(reqHname)
	if !ok {
		vmctx.log.Errorf("not found contract '%s', request %s",
			reqRef.RequestSection().Target().Hname().String(), vmctx.reqRef.RequestID().Short())
		return false
	}
	vmctx.feeColor = feeColor
	vmctx.ownerFee = ownerFee
	vmctx.validatorFee = validatorFee

	return true
}

func isInitChainRequest(reqRef sctransaction.RequestRef) bool {
	s := reqRef.RequestSection()
	return s.Target().Hname() == root.Interface.Hname() && s.EntryPointCode() == coretypes.EntryPointInit
}

func (vmctx *VMContext) FinalizeTransactionEssence(blockIndex uint32, stateHash *hashing.HashValue, timestamp int64) (*sctransaction.Transaction, error) {
	// add state block
	err := vmctx.txBuilder.SetStateParams(blockIndex, stateHash, timestamp)
	if err != nil {
		return nil, err
	}
	tx, err := vmctx.txBuilder.Build()
	if err != nil {
		return nil, err
	}
	return tx, nil
}
