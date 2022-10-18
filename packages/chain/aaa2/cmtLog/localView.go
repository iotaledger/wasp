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
// The AO chain maintained by the LocalView is somewhat orthogonal to the LogIndexes.
// While the new entries are added by publishing AOs, the chain matches the LogIndexes, but
// if a the local view is reset based on a rejection or externally made AO transition, then
// the direct mapping with the log indexes is lost. New AO will be considered on the next LogIndex.

// TODO: Keep some history of published alias outputs just to handle out-of-order delivery
// of messages on AO confirmation/rejection.

package cmtLog

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

type LocalView interface {
	//
	// Returns alias output to produce next transaction on, or nil if we should wait.
	// In the case of nil, we either wait for the first AO to receive, or we are
	// still recovering from a TX rejection.
	GetBaseAliasOutput() *isc.AliasOutputWithID
	//
	// Corresponds to the `ao_received` event in the specification.
	AliasOutputConfirmed(confirmed *isc.AliasOutputWithID)
	//
	// Corresponds to the `tx_rejected` event in the specification.
	AliasOutputRejected(rejected *isc.AliasOutputWithID)
	//
	// Corresponds to the `tx_posted` event in the specification.
	AliasOutputPublished(consumed iotago.OutputID, published *isc.AliasOutputWithID)
}

type localViewEntry struct {
	output     *isc.AliasOutputWithID
	stateIndex uint32
	rejected   bool
}

type localViewImpl struct {
	entries []*localViewEntry
}

func NewLocalView() LocalView {
	return &localViewImpl{
		entries: []*localViewEntry{},
	}
}

// Return latest AO to be used as an input for the following TX.
// nil means we have to wait: either we have no AO, or we have some rejections and waiting until a re-sync.
func (lvi *localViewImpl) GetBaseAliasOutput() *isc.AliasOutputWithID {
	if len(lvi.entries) == 0 {
		return nil
	}
	for _, e := range lvi.entries {
		if e.rejected {
			return nil
		}
	}
	return lvi.entries[len(lvi.entries)-1].output
}

// A confirmed AO is received from L1. Base on that, we either truncate our local
// history until the received AO (if we know it was posted before), or we replace
// the entire history with an unseen AO (probably produced not by this chain×cmt).
func (lvi *localViewImpl) AliasOutputConfirmed(confirmed *isc.AliasOutputWithID) {
	foundIdx := -1
	for i := range lvi.entries {
		if lvi.entries[i].output.Equals(confirmed) {
			foundIdx = i
			break
		}
	}
	if foundIdx == -1 {
		lvi.entries = []*localViewEntry{
			{
				output:     confirmed,
				stateIndex: confirmed.GetStateIndex(),
				rejected:   false,
			},
		}
		return
	}
	lvi.entries = lvi.entries[foundIdx:]
}

// Mark the specified AO as rejected.
// Trim the suffix of rejected AOs.
func (lvi *localViewImpl) AliasOutputRejected(rejected *isc.AliasOutputWithID) {
	rejectedIdx := -1
	remainingRejected := true
	for i := range lvi.entries {
		if lvi.entries[i].output.Equals(rejected) {
			lvi.entries[i].rejected = true
			rejectedIdx = i
		}
		if rejectedIdx != -1 && i > rejectedIdx {
			remainingRejected = remainingRejected && lvi.entries[i].rejected
		}
	}
	if rejectedIdx == -1 {
		// Not found, maybe outdated info.
		return
	}
	if remainingRejected {
		lvi.entries = lvi.entries[0:rejectedIdx]
	}
}

func (lvi *localViewImpl) AliasOutputPublished(consumed iotago.OutputID, published *isc.AliasOutputWithID) {
	if len(lvi.entries) == 0 {
		// Have we done reset recently?
		// Just ignore this call, it is outdated.
		return
	}
	if !lvi.entries[len(lvi.entries)-1].output.OutputID().UTXOInput().Equals(consumed.UTXOInput()) {
		// Some other output was published in parallel?
		// Just ignore this call, it is outdated.
		return
	}
	lvi.entries = append(lvi.entries, &localViewEntry{
		output:     published,
		stateIndex: published.GetStateIndex(),
		rejected:   false,
	})
}
