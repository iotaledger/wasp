package cmt_log

// TODO: Remove it?
// The problem with this approach is that the StateMgr.ChainFetchStateDiff calculates a diff between the L1Commitments, not AOs.
// And several AOs can have the same L1Commitment. At least in the governance TX cases.

import (
	"context"
	"sync"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_inputs"
	"github.com/iotaledger/wasp/packages/isc"
)

type diffFunc func(ctx context.Context, prevAO, nextAO *isc.AliasOutputWithID) <-chan *sm_inputs.ChainFetchStateDiffResults // Stands for StateMgr.ChainFetchStateDiff

type TipTracker struct {
	ctx      context.Context
	diffFunc diffFunc
	l1AO     *isc.AliasOutputWithID       // Latest confirmed AO from L1.
	coAOs    map[iotago.OutputID]*tipInfo // A list of AOs received from Consensus Output.
	lock     *sync.RWMutex                // TODO: Make it lock/thread free.
}

type tipInfo struct {
	ao        *isc.AliasOutputWithID
	ctxCancel context.CancelFunc
}

func NewTipTracker(ctx context.Context, diffFunc diffFunc) *TipTracker {
	return &TipTracker{
		ctx:      ctx,
		diffFunc: diffFunc,
		l1AO:     nil,
		coAOs:    map[iotago.OutputID]*tipInfo{},
		lock:     &sync.RWMutex{},
	}
}

func (tt *TipTracker) CurrentTip() *isc.AliasOutputWithID {
	tt.lock.RLock()
	defer tt.lock.RUnlock()
	if tt.l1AO == nil {
		return nil
	}
	if len(tt.coAOs) == 0 {
		return tt.l1AO
	}
	if len(tt.coAOs) > 1 {
		return nil // Forked, have to wait for confirms/rejects.
	}
	for _, ti := range tt.coAOs { // Loop over a single element actually.
		if ti.ctxCancel != nil {
			return nil // Still waiting for a diff result.
		}
		return ti.ao // This is a tip.
	}
	panic("above loop should iterate exactly once, thus return")
}

func (tt *TipTracker) L1AliasOutputConfirmed(ao *isc.AliasOutputWithID) {
	if ao.Equals(tt.l1AO) {
		return
	}
	tt.l1AO = ao
	tt.lock.Lock()
	defer tt.lock.Unlock()
	//
	// Cancel all the ongoing diff queries.
	for _, ti := range tt.coAOs {
		if ti.ctxCancel != nil {
			ti.ctxCancel()
			ti.ctxCancel = nil
		}
	}
	//
	// Issue new diff queries.
	for _, ti := range tt.coAOs {
		tiCtx, tiCancel := context.WithCancel(tt.ctx)
		ti.ctxCancel = tiCancel
		diffCh := tt.diffFunc(tiCtx, tt.l1AO, ti.ao)
		tiRef := ti
		go tt.awaitDiffResult(tiCtx, tiRef, diffCh)
	}
}

func (tt *TipTracker) awaitDiffResult(tiCtx context.Context, tiRef *tipInfo, diffCh <-chan *sm_inputs.ChainFetchStateDiffResults) {
	select {
	case <-tiCtx.Done():
		tt.lock.Lock()
		defer tt.lock.Unlock()
		tiRef.ctxCancel = nil
	case diff := <-diffCh:
		tt.lock.Lock()
		defer tt.lock.Unlock()
		if tiRef.ctxCancel == nil {
			return // This is outdated.
		}
		tiRef.ctxCancel = nil
		if len(diff.GetRemoved()) > 0 {
			// TODO: Drop this coAO, it is outdated.
			return
		}
		if len(diff.GetAdded()) > 0 {
			// TODO: This coAO is ahead of this l1AO.
			return
		}
		// TODO: This coAO got confirmed.
	}
}

func (tt *TipTracker) ConsensusOutputProduced(ao *isc.AliasOutputWithID) {
	// TODO: Implement.
}

func (tt *TipTracker) ConsensusOutputRejected(ao *isc.AliasOutputWithID) {
	// TODO: Not clear in this case, how rejections should be tracked,
	// the diff don't give the intermediate AOs.
	// It only gives the L1Commitments/Blocks.
}
