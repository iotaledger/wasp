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
// The AO chain maintained by the LocalView is somewhat orthogonal to the journal LogIndexes.
// While the new entries are added by publishing AOs, the chain matches the LogIndexes, but
// if a the local view is reset based on a rejection or externally made AO transition, then
// the direct mapping with the log indexes is lost. New AO will be considered on the next LogIndex.

// TODO: Keep some history of published alias outputs just to handle out-of-order delivery
// of messages on AO confirmation/rejection.

package journal

import (
	"bytes"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

type LocalView interface {
	//
	// Returns alias output to produce next transaction on, or nil if we should wait.
	// In the case of nil, we either wait for the first AO to receive, or we are
	// still recovering from a TX rejection.
	GetBaseAliasOutputID() *iotago.OutputID
	//
	// Corresponds to the `ao_received` event in the specification.
	AliasOutputConfirmed(confirmed *isc.AliasOutputWithID)
	//
	// Corresponds to the `tx_rejected` event in the specification.
	AliasOutputRejected(rejected *isc.AliasOutputWithID)
	//
	// Corresponds to the `tx_posted` event in the specification.
	AliasOutputPublished(consumed iotago.OutputID, published *isc.AliasOutputWithID)
	//
	// For serialization.
	AsBytes() ([]byte, error)
}

type localViewEntry struct {
	outputID   iotago.OutputID
	stateIndex uint32
	rejected   bool
}

func newLocalViewEntryFromBytes(data []byte) (*localViewEntry, error) {
	r := bytes.NewBuffer(data)
	var outputID iotago.OutputID
	if _, err := r.Read(outputID[:]); err != nil {
		return nil, err
	}
	var stateIndex uint32
	if err := util.ReadUint32(r, &stateIndex); err != nil {
		return nil, err
	}
	var rejected bool
	if err := util.ReadBoolByte(r, &rejected); err != nil {
		return nil, err
	}
	return &localViewEntry{outputID, stateIndex, rejected}, nil
}

func (lve *localViewEntry) AsBytes() ([]byte, error) {
	w := bytes.NewBuffer([]byte{})
	if _, err := w.Write(lve.outputID[:]); err != nil {
		return nil, err
	}
	if err := util.WriteUint32(w, lve.stateIndex); err != nil {
		return nil, err
	}
	if err := util.WriteBoolByte(w, lve.rejected); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

type localViewImpl struct {
	entries []*localViewEntry
}

func NewLocalView() LocalView {
	return &localViewImpl{
		entries: []*localViewEntry{},
	}
}

func NewLocalViewFromBytes(data []byte) (LocalView, error) {
	r := bytes.NewBuffer(data)
	var entriesLen uint16
	if err := util.ReadUint16(r, &entriesLen); err != nil {
		return nil, err
	}
	entries := make([]*localViewEntry, entriesLen)
	for i := range entries {
		entryBytes, err := util.ReadBytes16(r)
		if err != nil {
			return nil, err
		}
		entries[i], err = newLocalViewEntryFromBytes(entryBytes)
		if err != nil {
			return nil, err
		}
	}
	return &localViewImpl{entries}, nil
}

func (lvi *localViewImpl) AsBytes() ([]byte, error) {
	w := bytes.NewBuffer([]byte{})
	if err := util.WriteUint16(w, uint16(len(lvi.entries))); err != nil {
		return nil, err
	}
	for _, e := range lvi.entries {
		entryBytes, err := e.AsBytes()
		if err != nil {
			return nil, err
		}
		if err := util.WriteBytes16(w, entryBytes); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// Return latest AO to be used as an input for the following TX.
// nil means we have to wait: either we have no AO, or we have some rejections and waiting until a re-sync.
func (lvi *localViewImpl) GetBaseAliasOutputID() *iotago.OutputID {
	if len(lvi.entries) == 0 {
		return nil
	}
	for _, e := range lvi.entries {
		if e.rejected {
			return nil
		}
	}
	return &lvi.entries[len(lvi.entries)-1].outputID
}

// A confirmed AO is received from L1. Base on that, we either truncate our local
// history until the received AO (if we know it was posted before), or we replace
// the entire history with an unseen AO (probably produced not by this chain×cmt).
func (lvi *localViewImpl) AliasOutputConfirmed(confirmed *isc.AliasOutputWithID) {
	foundIdx := -1
	for i := range lvi.entries {
		if lvi.entries[i].outputID == confirmed.OutputID() {
			foundIdx = i
			break
		}
	}
	if foundIdx == -1 {
		lvi.entries = []*localViewEntry{
			{
				outputID:   confirmed.OutputID(),
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
		if lvi.entries[i].outputID == rejected.OutputID() {
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
	if lvi.entries[len(lvi.entries)-1].outputID != consumed {
		// Some other output was published in parallel?
		// Just ignore this call, it is outdated.
		return
	}
	lvi.entries = append(lvi.entries, &localViewEntry{
		outputID:   published.OutputID(),
		stateIndex: published.GetStateIndex(),
		rejected:   false,
	})
}
