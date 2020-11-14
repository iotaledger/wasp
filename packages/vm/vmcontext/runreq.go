package vmcontext

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
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
	vmctx.setRequestContext(reqRef, timestamp)

	defer func() {
		vmctx.virtualState.ApplyStateUpdate(vmctx.stateUpdate)

		vmctx.log.Debugw("runTheRequest OUT",
			"reqId", vmctx.reqRef.RequestID().Short(),
			"entry point", vmctx.reqRef.RequestSection().EntryPointCode().String(),
			"state update", vmctx.stateUpdate.String(),
		)
	}()

	if !vmctx.handleNodeRewards() {
		return vmctx.stateUpdate
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				vmctx.log.Errorf("Recovered from panic in VM: %v", r)
				debug.PrintStack()
				if _, ok := r.(buffered.DBError); ok {
					// There was an error accessing the DB
					// TODO invalidate the whole block?
				}
				vmctx.Rollback()
			}
		}()
		vmctx.callFromRequest()
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
}

func (vmctx *VMContext) Rollback() {
	vmctx.txBuilder = vmctx.saveTxBuilder
	vmctx.stateUpdate.Clear()
}

// handleNodeRewards return true if to continue with request processing
// rewards are "rewards for node", so smart contract sending request to itself might need
// to pay rewards too
func (vmctx *VMContext) handleNodeRewards() bool {
	if !vmctx.nodeRewardsEnabled {
		// nothing to do
		return true
	}
	var err error

	reqTxId := vmctx.reqRef.Tx.ID()
	// determining how many iotas have been left in the request transaction
	availableIotas := vmctx.txBuilder.GetInputBalanceFromTransaction(balance.ColorIOTA, reqTxId)

	var proceed bool
	// taking into account 1 request token which will be recolored back to iota
	// and will be send to the node reward address (if enabled)
	var sendToRewardAddress int64
	if availableIotas+1 >= vmctx.minimumReward {
		sendToRewardAddress = vmctx.minimumReward - 1
		proceed = true
	} else {
		sendToRewardAddress = availableIotas
		// if reward is not enough, the state update will be empty, i.e. NOP (the fee will be taken)
		proceed = false
	}
	err = vmctx.txBuilder.MoveToAddressFromTransaction(vmctx.rewardAddress, balance.ColorIOTA, sendToRewardAddress, reqTxId)

	if err != nil {
		vmctx.log.Error("can't move reward tokens: %v", err)
		proceed = false
	}
	return proceed
}

func (vmctx *VMContext) FinalizeTransaction(blockIndex uint32, stateHash *hashing.HashValue, timestamp int64) (*sctransaction.Transaction, error) {
	// add state block
	err := vmctx.txBuilder.SetStateParams(blockIndex, stateHash, timestamp)
	if err != nil {
		return nil, err
	}
	// create result transaction
	tx, err := vmctx.txBuilder.Build(false)
	if err != nil {
		return nil, err
	}
	// check semantic just in case
	if _, err := tx.Properties(); err != nil {
		return nil, err
	}
	return tx, nil
}
