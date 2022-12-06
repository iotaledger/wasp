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

package cmtLog

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

type VarLocalView interface {
	//
	// Returns alias output to produce next transaction on, or nil if we should wait.
	// In the case of nil, we either wait for the first AO to receive, or we are
	// still recovering from a TX rejection.
	GetBaseAliasOutput() *isc.AliasOutputWithID
	//
	// Corresponds to the `ao_received` event in the specification.
	// Returns true, if the proposed BaseAliasOutput has changed.
	AliasOutputConfirmed(confirmed *isc.AliasOutputWithID) bool
	//
	// Corresponds to the `tx_rejected` event in the specification.
	// Returns true, if the proposed BaseAliasOutput has changed.
	AliasOutputRejected(rejected *isc.AliasOutputWithID) bool
	//
	// Corresponds to the `tx_posted` event in the specification.
	// Returns true, if the proposed BaseAliasOutput has changed.
	ConsensusOutputDone(consumed iotago.OutputID, published *isc.AliasOutputWithID) bool
	//
	// Support functions.
	StatusString() string
}

type varLocalViewEntry struct {
	output   *isc.AliasOutputWithID
	rejected bool
}

type varLocalViewImpl struct {
	// The first element in the entries list is the latest confirmed output.
	// All the remaining are either pending or rejected.
	entries []*varLocalViewEntry
	// Resync is true, if we got a rejection and now we are waiting for all the
	// pending outputs to be either confirmed or rejected.
	resync bool
}

func NewVarLocalView() VarLocalView {
	return &varLocalViewImpl{
		entries: []*varLocalViewEntry{},
		resync:  true,
	}
}

// Return latest AO to be used as an input for the following TX.
// nil means we have to wait: either we have no AO, or we have some rejections and waiting until a re-sync.
func (lvi *varLocalViewImpl) GetBaseAliasOutput() *isc.AliasOutputWithID {
	if len(lvi.entries) == 0 {
		return nil
	}
	if lvi.resync {
		return nil
	}
	return lvi.entries[len(lvi.entries)-1].output
}

// A confirmed AO is received from L1. Base on that, we either truncate our local
// history until the received AO (if we know it was posted before), or we replace
// the entire history with an unseen AO (probably produced not by this chain×cmt).
func (lvi *varLocalViewImpl) AliasOutputConfirmed(confirmed *isc.AliasOutputWithID) bool {
	foundIdx := -1
	for i := range lvi.entries {
		if lvi.entries[i].output.OutputID() == confirmed.OutputID() {
			foundIdx = i
			break
		}
	}
	if foundIdx == -1 {
		lvi.entries = []*varLocalViewEntry{
			{
				output:   confirmed,
				rejected: false,
			},
		}
		lvi.resync = false
		return true
	}
	lvi.entries = lvi.entries[foundIdx:]
	if lvi.resync && len(lvi.entries) == 1 {
		lvi.resync = false
		return true
	}
	return false
}

// Mark the specified AO as rejected.
// Trim the suffix of rejected AOs.
func (lvi *varLocalViewImpl) AliasOutputRejected(rejected *isc.AliasOutputWithID) bool {
	for i := range lvi.entries {
		if lvi.entries[i].output.OutputID() == rejected.OutputID() {
			lvi.entries = lvi.entries[0:i]
			lvi.resync = i > 1
			return !lvi.resync
		}
	}
	// That was an outdated rejection.
	return false
}

func (lvi *varLocalViewImpl) ConsensusOutputDone(consumed iotago.OutputID, published *isc.AliasOutputWithID) bool {
	if len(lvi.entries) == 0 || lvi.resync {
		// Have we done reset recently?
		// Just ignore this call, it is outdated.
		return false
	}
	if lvi.entries[len(lvi.entries)-1].output.OutputID() != consumed {
		// Some other output was published in parallel?
		// Just ignore this call, it is outdated.
		return false
	}
	lvi.entries = append(lvi.entries, &varLocalViewEntry{
		output:   published,
		rejected: false,
	})
	return true
}

func (lvi *varLocalViewImpl) StatusString() string {
	str := fmt.Sprintf("varLocalView[resync=%v, entries=%v]:", lvi.resync, len(lvi.entries))
	for _, e := range lvi.entries {
		oid := e.output.OutputID()
		str += fmt.Sprintf(" %v/%v", oid[0:4], e.rejected)
	}
	return str
}
