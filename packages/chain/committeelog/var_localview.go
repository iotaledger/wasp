// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package committeelog implements the local view of a chain, maintained by a committee to decide which
// achor object to propose to the ACS. The achor object decided by the ACS will be used
// as an input for TX we build.
//
// The LocalView maintains a list of achor objects (Anchors). The are chained based on consumed/produced
// Anchors in a transaction we publish. The goal here is to tract the unconfirmed achor objects, update
// the list based on confirmations/rejections from the L1.
//
// In overall, the LocalView acts as a filter between the L1 and LogIndex assignment in logIndexVariable.
// It has to distinguish between Anchors that are confirming a prefix of the posted transaction (pipelining),
// from other changes in L1 (rotations, rollbacks, rejections, etc.).
//
// We have several inputs:
//
//   - **Achor Object Confirmed**.
//     It can be Anchor posted by this committee,
//     as well as by other committee (e.g. chain was rotated to other committee and then back)
//     or a user (e.g. external rotation TX).
//
//   - **Achor Object Rejected**.
//     These events are always for TXes posted by this committee.
//     We assume for each TX we will get either Confirmation or Rejection.
//
//   - **Consensus Done**.
//     Consensus produced a TX, and will post it to the L1.
//
//   - **Consensus Skip**.
//     Consensus completed without producing a TX and a block. So the previous Anchor is left actual.
//
//   - **Consensus Recover**.
//     Consensus is still running, but it takes long time, so maybe something is wrong
//     and we should consider spawning another consensus for the same base Anchor.
//
// On the pipelining -- the current L1 model don't allow us to do any kind of pipelining,
// apart from creating L2 blocks without committing them to the L1. But that's not the
// responsibility of the local view component.
//
// Note on the Anchor as an input for a consensus. The provided Anchor is just a proposal. After ACS
// is completed, the participants will select the actual Anchor, which can differ from the one
// proposed by this node.
package committeelog

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type LocalViewVariable interface {
	//
	// Called in the case of new Anchor from L1.
	AnchorConfirmed(confirmedAnchor *isc.StateAnchor) gpa.OutMessages
	//
	// Called by the consensus to determine if a produced TX can be posted to the L1.
	TransactionProduced(logIndex LogIndex, consumedAnchor *isc.StateAnchor, tx *iotasigner.SignedTransaction) gpa.OutMessages // TODO: Call it.
	//
	// Called if a TX is rejected.
	// This will always be called after TransactionProduced.
	TransactionRejected(logIndex LogIndex) gpa.OutMessages // TODO: Call it.
	//
	// Support functions.
	StatusString() string
}

type localViewVariableEntry struct {
	logIndex       LogIndex
	consumedAnchor *isc.StateAnchor
	transaction    *iotasigner.SignedTransaction
}

type localViewVariableImpl struct {
	latestTip *isc.StateAnchor
	// The latest confirmed Anchor, as received from L1.
	// It can be nil, if the latest Anchor is unclear (either not received yet).
	confirmedAnchor *isc.StateAnchor
	// Transactions that are ready to be posted.
	pendingTXes *shrinkingmap.ShrinkingMap[uint32, []*localViewVariableEntry]
	// Callback for the TIP changes.
	tipUpdatedCB func(ao *isc.StateAnchor) gpa.OutMessages
	// Just a logger.
	log log.Logger
}

func NewLocalViewVariable(pipeliningLimit int, tipUpdatedCB func(ao *isc.StateAnchor) gpa.OutMessages, log log.Logger) LocalViewVariable {
	log.LogDebugf("NewLocalViewVariable, pipeliningLimit=%v", pipeliningLimit)
	return &localViewVariableImpl{
		latestTip:       nil,
		confirmedAnchor: nil,
		pendingTXes:     shrinkingmap.New[uint32, []*localViewVariableEntry](),
		tipUpdatedCB:    tipUpdatedCB,
		log:             log,
	}
}

func (lvi *localViewVariableImpl) AnchorConfirmed(confirmedAnchor *isc.StateAnchor) gpa.OutMessages {
	lvi.confirmedAnchor = confirmedAnchor
	return lvi.processIt()
}

func (lvi *localViewVariableImpl) TransactionProduced(logIndex LogIndex, consumedAnchor *isc.StateAnchor, tx *iotasigner.SignedTransaction) gpa.OutMessages {
	stateIndex := consumedAnchor.GetStateIndex()
	stateIndexEntries, _ := lvi.pendingTXes.GetOrCreate(stateIndex, func() []*localViewVariableEntry { return []*localViewVariableEntry{} })
	contains := lo.ContainsBy(stateIndexEntries, func(entry *localViewVariableEntry) bool {
		return lo.Must(tx.Digest()).Equals(*lo.Must(entry.transaction.Digest()))
	})
	if !contains {
		stateIndexEntries = append(stateIndexEntries, &localViewVariableEntry{
			logIndex:       logIndex,
			consumedAnchor: consumedAnchor,
			transaction:    tx,
		})
		lvi.pendingTXes.Set(stateIndex, stateIndexEntries)
	}
	return lvi.processIt()
}

func (lvi *localViewVariableImpl) TransactionRejected(logIndex LogIndex) gpa.OutMessages {
	lvi.pendingTXes.ForEach(func(stateIndex uint32, entries []*localViewVariableEntry) bool {
		entries = lo.Filter(entries, func(entry *localViewVariableEntry, index int) bool {
			return entry.logIndex != logIndex
		})
		if len(entries) == 0 {
			lvi.pendingTXes.Delete(stateIndex)
		} else {
			lvi.pendingTXes.Set(stateIndex, entries)
		}
		return true
	})
	return lvi.processIt()
}

func (lvi *localViewVariableImpl) StatusString() string {
	return fmt.Sprintf("{localViewVariable: confirmedAnchor=%v, |pendingTxIndexes|=%v}", lvi.confirmedAnchor, lvi.pendingTXes.Size())
}

func (lvi *localViewVariableImpl) processIt() gpa.OutMessages {
	if lvi.confirmedAnchor == nil {
		lvi.updateVal(nil)
		return nil
	}
	confirmedStateIndex := lvi.confirmedAnchor.GetStateIndex()

	//
	// Cleanup outdated.
	lvi.pendingTXes.ForEachKey(func(pendingStateIndex uint32) bool {
		if pendingStateIndex < confirmedStateIndex {
			lvi.pendingTXes.Delete(pendingStateIndex)
		}
		return true
	})

	entries, found := lvi.pendingTXes.Get(confirmedStateIndex)
	if found && len(entries) > 0 {
		return lvi.updateVal(nil)
	}

	return lvi.updateVal(lvi.confirmedAnchor)
}

func (lvi *localViewVariableImpl) updateVal(tip *isc.StateAnchor) gpa.OutMessages {
	if tip == nil && lvi.latestTip == nil {
		return nil
	}
	if tip != nil && lvi.latestTip != nil && tip.GetObjectRef().Equals(lvi.latestTip.GetObjectRef()) {
		return nil
	}
	lvi.latestTip = tip
	return lvi.tipUpdatedCB(tip)
}
