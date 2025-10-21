// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/chain/statemanager"
	"github.com/iotaledger/wasp/v2/packages/chain/statemanager/gpa/inputs"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/state"
)

type StateTrackerStepCB = func(st state.State, from, till *isc.StateAnchor, added, removed []state.Block)

// StateTracker tracks a single chain of state transitions. We will have 2 instances of it:
//   - one for tracking the active state. It is needed for mempool to clear the requests.
//   - one for the committed state to await for committed request receipts.
type StateTracker struct {
	ctx                    context.Context
	stateMgr               statemanager.StateMgr
	haveLatestCB           StateTrackerStepCB
	haveAnchorState        state.State
	haveAnchor             *isc.StateAnchor   // We have a state ready for this Anchor.
	nextAnchor             *isc.StateAnchor   // For this state a query was made, but the response not received yet.
	nextAnchorCancel       context.CancelFunc // Cancel for a context used to query for the nextAnchor state.
	nextAnchorWaitCh       <-chan *inputs.ChainFetchStateDiffResults
	awaitReceipt           AwaitReceipt
	metricWantStateIndexCB func(uint32)
	metricHaveStateIndexCB func(uint32)
	log                    log.Logger
}

func NewStateTracker(
	ctx context.Context,
	stateMgr statemanager.StateMgr,
	haveLatestCB StateTrackerStepCB,
	metricWantStateIndexCB func(uint32),
	metricHaveStateIndexCB func(uint32),
	log log.Logger,
) *StateTracker {
	return &StateTracker{
		ctx:                    ctx,
		stateMgr:               stateMgr,
		haveLatestCB:           haveLatestCB,
		haveAnchorState:        nil,
		haveAnchor:             nil,
		nextAnchor:             nil,
		nextAnchorCancel:       nil,
		nextAnchorWaitCh:       nil,
		awaitReceipt:           NewAwaitReceipt(AwaitReceiptCleanupEvery, log),
		metricWantStateIndexCB: metricWantStateIndexCB,
		metricHaveStateIndexCB: metricHaveStateIndexCB,
		log:                    log,
	}
}

func (sti *StateTracker) TrackAnchor(ao *isc.StateAnchor, strict bool) {
	if ao == nil {
		// We don't have the latest Anchor while we are still synching.
		return
	}
	sti.log.LogDebugf("TrackAnchor[strict=%v], ao=%v, haveAnchor=%v, nextAnchor=%v", strict, ao, sti.haveAnchor, sti.nextAnchor)
	if !strict && sti.haveAnchor != nil && sti.haveAnchor.GetStateIndex() >= ao.GetStateIndex() {
		return
	}
	if ao.Equals(sti.nextAnchor) {
		return
	}
	sti.metricWantStateIndexCB(ao.GetStateIndex())
	if ao.Equals(sti.haveAnchor) {
		sti.nextAnchor = sti.haveAnchor // All done, state is already received.
		sti.cancelQuery()               // Cancel the request, if pending.
		return
	}
	nextAnchorCtx, nextAnchorCancel := context.WithCancel(sti.ctx)
	sti.nextAnchor = ao
	sti.nextAnchorCancel = nextAnchorCancel
	sti.nextAnchorWaitCh = sti.stateMgr.ChainFetchStateDiff(nextAnchorCtx, sti.haveAnchor, sti.nextAnchor)
}

func (sti *StateTracker) AwaitRequestReceipt(query *awaitReceiptReq) {
	sti.log.LogDebugf("AwaitRequestReceipt, query.requestID=%v", query.requestID)
	sti.awaitReceipt.Await(query)
}

// ChainNodeAwaitStateMgrCh is to be used in the select loop at the chain node.
func (sti *StateTracker) ChainNodeAwaitStateMgrCh() <-chan *inputs.ChainFetchStateDiffResults {
	return sti.nextAnchorWaitCh
}

// ChainNodeStateMgrResponse is assumed to be called right after the `ChainNodeAwaitStateMgrCh()`,
// thus no additional checks are present here.
func (sti *StateTracker) ChainNodeStateMgrResponse(results *inputs.ChainFetchStateDiffResults) {
	sti.cancelQuery()
	newState := results.GetNewState()
	sti.log.LogDebugf(
		"Have latest state for %v, state.BlockIndex=%v, state.trieRoot=%v, previous=%v, |blocksAdded|=%v, |blockRemoved|=%v",
		sti.nextAnchor, newState.BlockIndex(), newState.TrieRoot(), sti.haveAnchor, len(results.GetAdded()), len(results.GetRemoved()),
	)
	sti.haveLatestCB(newState, sti.haveAnchor, sti.nextAnchor, results.GetAdded(), results.GetRemoved())
	sti.haveAnchor = sti.nextAnchor
	sti.haveAnchorState = newState
	sti.metricHaveStateIndexCB(newState.BlockIndex())
	sti.awaitReceipt.ConsiderState(newState, results.GetAdded())
}

func (sti *StateTracker) cancelQuery() {
	if sti.nextAnchorCancel == nil {
		return
	}
	sti.nextAnchorCancel()
	sti.nextAnchorCancel = nil
	sti.nextAnchorWaitCh = nil
}
