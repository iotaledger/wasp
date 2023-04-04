// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Here we implement the local view of a chain, maintained by a committee to decide which
// alias output to propose to the ACS. The alias output decided by the ACS will be used
// as an input for TX we build.
//
// The LocalView maintains a list of Alias Outputs (AOs). The are chained based on consumed/produced
// AOs in a transaction we publish. The goal here is to tract the unconfirmed alias outputs, update
// the list based on confirmations/rejections from the L1.
//
// In overall, the LocalView acts as a filter between the L1 and LogIndex assignment in varLogIndex.
// It has to distinguish between AOs that are confirming a prefix of the posted transaction (pipelining),
// from other changes in L1 (rotations, rollbacks, rejections, etc.).
//
// We have several inputs:
//
//   - **Alias Output Confirmed**.
//     It can be AO posted by this committee,
//     as well as by other committee (e.g. chain was rotated to other committee and then back)
//     or a user (e.g. external rotation TX).
//
//   - **Alias Output Rejected**.
//     These events are always for TXes posted by this committee.
//     We assume for each TX we will get either Confirmation or Rejection.
//
//   - **Consensus Done**.
//     Consensus produced a TX, and will post it to the L1.
//
//   - **Consensus Skip**.
//     Consensus completed without producing a TX and a block. So the previous AO is left unspent.
//
//   - **Consensus Recover**.
//     Consensus is still running, but it takes long time, so maybe something is wrong
//     and we should consider spawning another consensus for the same base AO.
//
// On the pipelining:
//
//   - During the normal operation, if consensus produces a TX, it can use the produced AO
//     to build next TX on it. That's pipelining. It allows to produce multiple blocks per
//     L1 milestone. This component tracks the AOs build in this way and not confirmed yet.
//
//   - If AO produced by the consensus is rejected, then all the AOs build on top of it will
//     be rejected eventually as well, because they use the rejected AO as an input. On the
//     other hand, it is unclear if unconfirmed AOs before the rejected one will be confirmed
//     or rejected, we will wait until L1 decides on all of them.
//
//   - If we get a confirmed AO, that is not one of the AOs we have posted (and still waiting
//     for a decision), then someone from the outside of a committee transitioned the chain.
//     In this case all our produced/pending transactions are not meaningful anymore and we
//     have to start building all the chain from the newly received AO.
//
//   - Recovery notice is received from a consensus (with SI/LI...) a new consensus will
//     be started after agreeing on the next LI. The new consensus will take the same AO as
//     an input and therefore will race with the existing one (maybe it has stuck for some
//     reason, that's a fallback). In this case we stop building an unconfirmed chain for
//     the future state indexes and will wait for some AO to be confirmed or all the concurrent
//     consensus TX'es to be rejected.
//
// Note on the AO as an input for a consensus. The provided AO is just a proposal. After ACS
// is completed, the participants will select the actual AO, which can differ from the one
// proposed by this node.
package cmtLog

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

type VarLocalView interface {
	//
	// Returns alias output to produce next transaction on, or nil if we should wait.
	// In the case of nil, we either wait for the first AO to receive, or we are
	// still recovering from a TX rejection.
	Value() *isc.AliasOutputWithID
	//
	// Corresponds to the `tx_posted` event in the specification.
	// Returns true, if the proposed BaseAliasOutput has changed.
	ConsensusOutputDone(logIndex LogIndex, consumed iotago.OutputID, published *isc.AliasOutputWithID) (*isc.AliasOutputWithID, bool) // TODO: Recheck, if consumed AO is the decided one.
	//
	// Corresponds to the `ao_received` event in the specification.
	// Returns true, if the proposed BaseAliasOutput has changed.
	AliasOutputConfirmed(confirmed *isc.AliasOutputWithID) (*isc.AliasOutputWithID, bool)
	//
	// Corresponds to the `tx_rejected` event in the specification.
	// Returns true, if the proposed BaseAliasOutput has changed.
	AliasOutputRejected(rejected *isc.AliasOutputWithID) (*isc.AliasOutputWithID, bool)
	//
	// Support functions.
	StatusString() string
}

type varLocalViewEntry struct {
	output   *isc.AliasOutputWithID // The AO published.
	consumed iotago.OutputID        // The AO used as an input for the TX.
	rejected bool                   // True, if the AO as rejected. We keep them to detect the other rejected AOs.
}

type varLocalViewImpl struct {
	// The latest confirmed AO, as received from L1.
	// All the pending entries are built on top ot this one.
	// It can be nil, if the latest AO is unclear (either not received yet, or some rejections happened).
	confirmed *isc.AliasOutputWithID
	// AOs produced by this committee, but not confirmed yet.
	// It is possible to have several AOs for a StateIndex in the case of
	// Recovery/Timeout notices. Then the next consensus is started o build a TX.
	// Both of them can still produce a TX, but only one of them will be confirmed.
	pending *shrinkingmap.ShrinkingMap[uint32, []*varLocalViewEntry]
	// Just a logger.
	log *logger.Logger
}

func NewVarLocalView(log *logger.Logger) VarLocalView {
	return &varLocalViewImpl{
		confirmed: nil,
		pending:   shrinkingmap.New[uint32, []*varLocalViewEntry](),
		log:       log,
	}
}

// Return latest AO to be used as an input for the following TX.
// nil means we have to wait: either we have no AO, or we have some rejections and waiting until a re-sync.
func (lvi *varLocalViewImpl) Value() *isc.AliasOutputWithID {
	return lvi.findLatestPending()
}

func (lvi *varLocalViewImpl) ConsensusOutputDone(logIndex LogIndex, consumed iotago.OutputID, published *isc.AliasOutputWithID) (*isc.AliasOutputWithID, bool) {
	lvi.log.Debugf("ConsensusOutputDone: logIndex=%v, consumed.ID=%v, published=%v", logIndex, consumed.ToHex(), published)
	stateIndex := published.GetStateIndex()
	prevLatest := lvi.findLatestPending()
	//
	// Check, if not outdated.
	if lvi.confirmed == nil {
		lvi.log.Debugf("⊳ Ignoring it, have no confirmed AO.")
		return prevLatest, false
	}
	confirmedStateIndex := lvi.confirmed.GetStateIndex()
	if stateIndex <= confirmedStateIndex {
		lvi.log.Debugf("⊳ Ignoring it, outdated, current confirmed=%v", lvi.confirmed)
		return prevLatest, false
	}
	//
	// Add it to the pending list.
	var entries []*varLocalViewEntry
	entries, ok := lvi.pending.Get(stateIndex)
	if !ok {
		entries = []*varLocalViewEntry{}
	}
	if lo.ContainsBy(entries, func(e *varLocalViewEntry) bool { return e.output.Equals(published) }) {
		lvi.log.Debugf("⊳ Ignoring it, duplicate.")
		return prevLatest, false
	}
	entries = append(entries, &varLocalViewEntry{
		output:   published,
		consumed: consumed,
		rejected: false,
	})
	lvi.pending.Set(stateIndex, entries)
	//
	// Check, if the added AO is a new tip for the chain.
	if published.Equals(lvi.findLatestPending()) {
		lvi.log.Debugf("⊳ Will consider consensusOutput=%v as a tip, the current confirmed=%v.", published, lvi.confirmed)
		return published, true
	}
	lvi.log.Debugf("⊳ That's not a tip.")
	return lvi.Value(), false
}

// A confirmed AO is received from L1. Base on that, we either truncate our local
// history until the received AO (if we know it was posted before), or we replace
// the entire history with an unseen AO (probably produced not by this chain×cmt).
func (lvi *varLocalViewImpl) AliasOutputConfirmed(confirmed *isc.AliasOutputWithID) (*isc.AliasOutputWithID, bool) {
	lvi.log.Debugf("AliasOutputConfirmed: confirmed=%v", confirmed)
	stateIndex := confirmed.GetStateIndex()
	oldTip := lvi.findLatestPending()
	lvi.confirmed = confirmed
	if lvi.isAliasOutputPending(confirmed) {
		lvi.pending.ForEach(func(si uint32, es []*varLocalViewEntry) bool {
			if si <= stateIndex {
				for _, e := range es {
					lvi.log.Debugf("⊳ Removing[%v≤%v] %v", si, stateIndex, e.output)
				}
				lvi.pending.Delete(si)
			}
			return true
		})
		lvi.clearPendingIfAllRejected()
	} else {
		lvi.pending.ForEach(func(si uint32, es []*varLocalViewEntry) bool {
			for _, e := range es {
				lvi.log.Debugf("⊳ Removing[all] %v", si, stateIndex, e.output)
			}
			lvi.pending.Delete(si)
			return true
		})
	}
	return lvi.outputIfChanged(oldTip, lvi.findLatestPending())
}

// Mark the specified AO as rejected.
// Trim the suffix of rejected AOs.
func (lvi *varLocalViewImpl) AliasOutputRejected(rejected *isc.AliasOutputWithID) (*isc.AliasOutputWithID, bool) {
	lvi.log.Debugf("AliasOutputRejected: rejected=%v", rejected)
	stateIndex := rejected.GetStateIndex()
	oldTip := lvi.findLatestPending()
	//
	// Mark the output as rejected, as well as all the outputs depending on it.
	if entries, ok := lvi.pending.Get(stateIndex); ok {
		for _, entry := range entries {
			if entry.output.Equals(rejected) {
				lvi.log.Debugf("⊳ Entry marked as rejected.")
				entry.rejected = true
				lvi.markDependentAsRejected(rejected)
			}
		}
	}
	//
	// If all remaining are rejected, remove them, and proceed from the confirmed one.
	lvi.clearPendingIfAllRejected()
	return lvi.outputIfChanged(oldTip, lvi.findLatestPending())
}

func (lvi *varLocalViewImpl) markDependentAsRejected(ao *isc.AliasOutputWithID) {
	accRejected := map[iotago.OutputID]struct{}{ao.OutputID(): {}}
	for si := ao.GetStateIndex() + 1; ; si++ {
		es, esFound := lvi.pending.Get(si)
		if !esFound {
			break
		}
		for _, e := range es {
			if _, ok := accRejected[e.consumed]; ok && !e.rejected {
				lvi.log.Debugf("⊳ Also marking %v as rejected.", e.output)
				e.rejected = true
				accRejected[e.output.OutputID()] = struct{}{}
			}
		}
	}
}

func (lvi *varLocalViewImpl) clearPendingIfAllRejected() {
	if !lvi.allRejected() || lvi.pending.IsEmpty() {
		return
	}
	lvi.log.Debugf("⊳ All entries are rejected, clearing them.")
	lvi.pending.ForEach(func(si uint32, es []*varLocalViewEntry) bool {
		for _, e := range es {
			lvi.log.Debugf("⊳ Clearing %v", e.output)
		}
		lvi.pending.Delete(si)
		return true
	})
}

func (lvi *varLocalViewImpl) outputIfChanged(oldTip, newTip *isc.AliasOutputWithID) (*isc.AliasOutputWithID, bool) {
	if oldTip == nil && newTip == nil {
		lvi.log.Debugf("⊳ Tip remains nil.")
		return nil, false
	}
	if oldTip == nil || newTip == nil {
		lvi.log.Debugf("⊳ New tip=%v, was %v", newTip, oldTip)
		return newTip, true
	}
	if oldTip.Equals(newTip) {
		lvi.log.Debugf("⊳ Tip remains %v.", newTip)
		return newTip, false
	}
	lvi.log.Debugf("⊳ New tip=%v, was %v", newTip, oldTip)
	return newTip, true
}

func (lvi *varLocalViewImpl) StatusString() string {
	return fmt.Sprintf("{varLocalView: confirmed=%v, tip=%v, |pendingSIs|=%v}", lvi.confirmed, lvi.findLatestPending(), lvi.pending.Size())
}

// Latest pending AO is only considered existing, if the current pending
// set of AOs is a chain, with no gaps, or alternatives, and all the AOs
// are not rejected.
func (lvi *varLocalViewImpl) findLatestPending() *isc.AliasOutputWithID {
	if lvi.confirmed == nil {
		return nil
	}
	latest := lvi.confirmed
	confirmedSI := lvi.confirmed.GetStateIndex()
	pendingSICount := uint32(lvi.pending.Size())
	for i := uint32(0); i < pendingSICount; i++ {
		entries, ok := lvi.pending.Get(confirmedSI + i + 1)
		if !ok {
			return nil // That's a gap.
		}
		if len(entries) != 1 {
			return nil // Alternatives exist.
		}
		if entries[0].rejected {
			return nil // Some are rejected.
		}
		if latest.OutputID() != entries[0].consumed {
			return nil // Don't form a chain.
		}
		latest = entries[0].output
	}
	return latest
}

func (lvi *varLocalViewImpl) isAliasOutputPending(ao *isc.AliasOutputWithID) bool {
	found := false
	lvi.pending.ForEach(func(si uint32, es []*varLocalViewEntry) bool {
		found = lo.ContainsBy(es, func(e *varLocalViewEntry) bool {
			return e.output.Equals(ao)
		})
		return !found
	})
	return found
}

func (lvi *varLocalViewImpl) allRejected() bool {
	allRejected := true
	lvi.pending.ForEach(func(si uint32, es []*varLocalViewEntry) bool {
		containsPending := lo.ContainsBy(es, func(e *varLocalViewEntry) bool {
			return !e.rejected
		})
		allRejected = !containsPending
		return !containsPending
	})
	return allRejected
}
