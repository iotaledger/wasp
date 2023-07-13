// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bp

import (
	"io"
	"time"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type BatchProposal struct {
	nodeIndex               uint16                 // Just for a double-check.
	baseAliasOutput         *isc.AliasOutputWithID // Proposed Base AliasOutput to use.
	dssIndexProposal        util.BitVector         // DSS Index proposal.
	timeData                time.Time              // Our view of time.
	validatorFeeDestination isc.AgentID            // Proposed destination for fees.
	requestRefs             []*isc.RequestRef      // Requests we propose to include into the execution.
}

func NewBatchProposal(
	nodeIndex uint16,
	baseAliasOutput *isc.AliasOutputWithID,
	dssIndexProposal util.BitVector,
	timeData time.Time,
	validatorFeeDestination isc.AgentID,
	requestRefs []*isc.RequestRef,
) *BatchProposal {
	return &BatchProposal{
		nodeIndex:               nodeIndex,
		baseAliasOutput:         baseAliasOutput,
		dssIndexProposal:        dssIndexProposal,
		timeData:                timeData,
		validatorFeeDestination: validatorFeeDestination,
		requestRefs:             requestRefs,
	}
}

func (b *BatchProposal) Bytes() []byte {
	return rwutil.WriteToBytes(b)
}

func (b *BatchProposal) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	b.nodeIndex = rr.ReadUint16()
	b.baseAliasOutput = new(isc.AliasOutputWithID)
	rr.Read(b.baseAliasOutput)
	b.dssIndexProposal = util.NewFixedSizeBitVector(0)
	rr.Read(b.dssIndexProposal)
	b.timeData = time.Unix(0, rr.ReadInt64())
	b.validatorFeeDestination = isc.AgentIDFromReader(rr)
	size := rr.ReadSize16()
	b.requestRefs = make([]*isc.RequestRef, size)
	for i := range b.requestRefs {
		b.requestRefs[i] = new(isc.RequestRef)
		rr.ReadN(b.requestRefs[i].ID[:])
		rr.ReadN(b.requestRefs[i].Hash[:])
	}
	return rr.Err
}

func (b *BatchProposal) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteUint16(b.nodeIndex)
	ww.Write(b.baseAliasOutput)
	ww.Write(b.dssIndexProposal)
	ww.WriteInt64(b.timeData.UnixNano())
	ww.Write(b.validatorFeeDestination)
	ww.WriteSize16(len(b.requestRefs))
	for i := range b.requestRefs {
		ww.WriteN(b.requestRefs[i].ID[:])
		ww.WriteN(b.requestRefs[i].Hash[:])
	}
	return ww.Err
}
