package vmcontext

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

// runTheRequest:
// - handles request token
// - processes reward logic
func (vmctx *VMContext) RunTheRequest(reqRef sctransaction.RequestRef, timestamp int64) {
	vmctx.prepareForRequest(reqRef, timestamp)

	if vmctx.isInitChainRequest() {
		vmctx.mustPreprocessChainInitRequest()
	} else {
		if !vmctx.mustPreprocessRequest() {
			// may not be enough fees or contract not found
			return
		}
	}
	// snapshot state baseline for rollback in case of panic
	snapshotTxBuilder := vmctx.txBuilder.Clone()
	snapshotStateUpdate := vmctx.stateUpdate.Clone()

	vmctx.lastError = nil
	func() {
		// panic catcher for the whole call from request to the VM
		defer func() {
			if r := recover(); r != nil {
				vmctx.lastResult = nil
				vmctx.lastError = fmt.Errorf("recovered from panic in VM: %v", r)
				//debug.PrintStack()
				if dberr, ok := r.(buffered.DBError); ok {
					// There was an error accessing the DB
					// The world stops
					vmctx.Panicf("DB error: %v", dberr)
				}
			}
		}()
		vmctx.mustCallFromRequest()
	}()

	if vmctx.lastError != nil {
		// treating panic and error, returned from request the same way: rollback to the checkpoint after fees
		vmctx.txBuilder = snapshotTxBuilder
		vmctx.stateUpdate = snapshotStateUpdate

		// TODO fallback policy for after-fee transfer

		//switch vmctx.lastError {
		//case ErrContractNotFound, ErrEntryPointNotFound, ErrProcessorNotFound:
		//	// TODO more sophisticated policy
		//	// if sent to the wrong contract or entry point, accrue the transfer to the sender' account on the chain
		//	// the sender can withdraw it at any time
		//	vmctx.creditToAccount(vmctx.reqRef.SenderAgentID(), vmctx.remainingAfterFees)
		//default:
		//	// TODO what to do with the transfer in case of error
		//}
	}
	vmctx.mustRequestToEventLog(vmctx.lastError)
	vmctx.virtualState.ApplyStateUpdate(vmctx.stateUpdate)

	vmctx.log.Debugw("runTheRequest OUT",
		"reqId", vmctx.reqRef.RequestID().Short(),
		"entry point", vmctx.reqRef.RequestSection().EntryPointCode().String(),
	)
}

func (vmctx *VMContext) mustRequestToEventLog(err error) {
	if err != nil {
		vmctx.log.Error(err)
	}
	e := "Ok"
	if err != nil {
		e = err.Error()
	}
	msg := fmt.Sprintf("[req] %s: %s", vmctx.reqRef.RequestID().String(), e)
	vmctx.log.Infof("chainlog -> '%s'", msg)
	vmctx.StoreToChainLog(vmctx.reqHname, []byte(msg))
}

// mustPreprocessChainInitRequest prepares VMContext for the initialization of the chain
func (vmctx *VMContext) mustPreprocessChainInitRequest() {
	vmctx.mustHandleRequestToken()
}

// mustGetBaseValues only makes sense if chain is already deployed
func (vmctx *VMContext) mustGetBaseValues() {
	// ordinary request, only makes sense when chain is already deployed
	info, err := vmctx.getChainInfo()
	if err != nil {
		vmctx.log.Panicf("mustPrepareForRequest: %s", err)
	}
	if info.ChainID != vmctx.chainID {
		vmctx.log.Panicf("mustPrepareForRequest: major inconsistency of chainID")
	}
	vmctx.chainOwnerID = info.ChainOwnerID
}

func (vmctx *VMContext) prepareFeeInfo() bool {
	var ok bool
	vmctx.feeColor, vmctx.ownerFee, vmctx.validatorFee, ok = vmctx.getFeeInfo(vmctx.reqHname)
	if !ok {
		vmctx.log.Errorf("prepareFeeInfo: not found contract '%s'",
			vmctx.reqRef.RequestSection().Target().Hname().String())
	}
	return ok
}

// mustPreprocessRequest prepares VMContext for the precessing if the ordinary request
// returns false when contract not found or not enough fees
func (vmctx *VMContext) mustPreprocessRequest() bool {
	vmctx.mustHandleRequestToken()
	vmctx.mustGetBaseValues()
	if !vmctx.prepareFeeInfo() {
		return false
	}
	vmctx.mustHandleFees()
	return true
}

func (vmctx *VMContext) prepareForRequest(reqRef sctransaction.RequestRef, timestamp int64) {
	reqHname := reqRef.RequestSection().Target().Hname()
	vmctx.reqRef = reqRef
	vmctx.reqHname = reqHname

	vmctx.timestamp = timestamp
	vmctx.stateUpdate = state.NewStateUpdate(reqRef.RequestID()).WithTimestamp(timestamp)
	vmctx.callStack = vmctx.callStack[:0]
	vmctx.entropy = *hashing.HashData(vmctx.entropy[:])
	vmctx.remainingAfterFees = cbalances.NewFromMap(nil)
}

func (vmctx *VMContext) isInitChainRequest() bool {
	s := vmctx.reqRef.RequestSection()
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
