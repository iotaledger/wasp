// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"

	"github.com/iotaledger/wasp/packages/chain/statemanager"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

// Tracks a single chain of state transitions. We will have 2 instances of it:
//   - one for tracking the active state. It is needed for mempool to clear the requests.
//   - one for the committed state to await for committed request receipts.
type StateTracker interface {
	//
	// The main functions provided by this component.
	TrackAliasOutput(ao *isc.AliasOutputWithID)
	AwaitRequestReceipt(query *awaitReceiptReq)
	//
	// The following 2 functions are only to move the channel receive loop to the main ChainNode thread.
	ChainNodeAwaitStateMgrCh() <-chan *smInputs.ChainFetchStateDiffResults
	ChainNodeStateMgrResponse(*smInputs.ChainFetchStateDiffResults)
}

type StateTrackerStepCB = func(st state.State, from, till *isc.AliasOutputWithID, added, removed []state.Block)

type stateTrackerImpl struct {
	ctx          context.Context
	stateMgr     statemanager.StateMgr
	haveLatestCB StateTrackerStepCB
	haveAOState  state.State
	haveAO       *isc.AliasOutputWithID // We have a state ready for this AO.
	nextAO       *isc.AliasOutputWithID // For this state a query was made, but the response not received yet.
	nextAOCancel context.CancelFunc     // Cancel for a context used to query for the nextAO state.
	nextAOWaitCh <-chan *smInputs.ChainFetchStateDiffResults
	awaitReceipt AwaitReceipt
}

var _ StateTracker = &stateTrackerImpl{}

func NewStateTracker(
	ctx context.Context,
	stateMgr statemanager.StateMgr,
	haveLatestCB StateTrackerStepCB,
) StateTracker {
	return &stateTrackerImpl{
		ctx:          ctx,
		stateMgr:     stateMgr,
		haveLatestCB: haveLatestCB,
		haveAOState:  nil,
		haveAO:       nil,
		nextAO:       nil,
		nextAOCancel: nil,
		nextAOWaitCh: nil,
		awaitReceipt: NewAwaitReceipt(awaitReceiptCleanupEvery),
	}
}

func (sti *stateTrackerImpl) TrackAliasOutput(ao *isc.AliasOutputWithID) {
	if ao.Equals(sti.nextAO) {
		return
	}
	if ao.Equals(sti.haveAO) {
		sti.nextAO = sti.haveAO // All done, state is already received.
		sti.cancelQuery()       // Cancel the request, if pending.
		return
	}
	nextAOCtx, nextAOCancel := context.WithCancel(sti.ctx)
	sti.nextAO = ao
	sti.nextAOCancel = nextAOCancel
	sti.nextAOWaitCh = sti.stateMgr.ChainFetchStateDiff(nextAOCtx, sti.haveAO, sti.nextAO)
}

func (sti *stateTrackerImpl) AwaitRequestReceipt(query *awaitReceiptReq) {
	sti.awaitReceipt.Await(query)
}

// To be used in the select loop at the chain node.
func (sti *stateTrackerImpl) ChainNodeAwaitStateMgrCh() <-chan *smInputs.ChainFetchStateDiffResults {
	return sti.nextAOWaitCh
}

// This is assumed to be called right after the `ChainNodeAwaitStateMgrCh()`,
// thus no additional checks are present here.
func (sti *stateTrackerImpl) ChainNodeStateMgrResponse(results *smInputs.ChainFetchStateDiffResults) {
	sti.cancelQuery()
	sti.haveLatestCB(results.GetNewState(), sti.haveAO, sti.nextAO, results.GetAdded(), results.GetRemoved())
	sti.haveAO = sti.nextAO
	sti.haveAOState = results.GetNewState()
	sti.awaitReceipt.ConsiderState(results.GetNewState(), results.GetAdded())
}

func (sti *stateTrackerImpl) cancelQuery() {
	if sti.nextAOCancel == nil {
		return
	}
	sti.nextAOCancel()
	sti.nextAOCancel = nil
	sti.nextAOWaitCh = nil
}
