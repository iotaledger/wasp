// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bp

import (
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

type BatchProposal struct {
	nodeIndex        uint16                 // Just for a double-check.
	baseAliasOutput  *isc.AliasOutputWithID // Proposed Base AliasOutput to use.
	dssIndexProposal util.BitVector         // DSS Index proposal.
	timeData         time.Time              // Our view of time.
	feeDestination   isc.AgentID            // Proposed destination for fees.
	requestRefs      []*isc.RequestRef      // Requests we propose to include into the execution.
}

func NewBatchProposal(
	nodeIndex uint16,
	baseAliasOutput *isc.AliasOutputWithID,
	dssIndexProposal util.BitVector,
	timeData time.Time,
	feeDestination isc.AgentID,
	requestRefs []*isc.RequestRef,
) *BatchProposal {
	return &BatchProposal{
		nodeIndex:        nodeIndex,
		baseAliasOutput:  baseAliasOutput,
		dssIndexProposal: dssIndexProposal,
		timeData:         timeData,
		feeDestination:   feeDestination,
		requestRefs:      requestRefs,
	}
}

func batchProposalFromBytes(data []byte) (*BatchProposal, error) {
	return batchProposalFromMarshalUtil(marshalutil.New(data))
}

func batchProposalFromMarshalUtil(mu *marshalutil.MarshalUtil) (*BatchProposal, error) {
	errFmt := "batchProposalFromMarshalUtil: %w"
	ret := &BatchProposal{}
	var err error
	ret.nodeIndex, err = mu.ReadUint16()
	if err != nil {
		return nil, fmt.Errorf(errFmt, err)
	}
	if ret.baseAliasOutput, err = isc.NewAliasOutputWithIDFromMarshalUtil(mu); err != nil {
		return nil, fmt.Errorf(errFmt, err)
	}
	if ret.dssIndexProposal, err = util.NewFixedSizeBitVectorFromMarshalUtil(mu); err != nil {
		return nil, fmt.Errorf(errFmt, err)
	}
	ret.timeData, err = mu.ReadTime()
	if err != nil {
		return nil, fmt.Errorf(errFmt, err)
	}
	ret.feeDestination, err = isc.AgentIDFromMarshalUtil(mu)
	if err != nil {
		return nil, fmt.Errorf(errFmt, err)
	}
	requestCount, err := mu.ReadUint16()
	if err != nil {
		return nil, fmt.Errorf(errFmt, err)
	}
	ret.requestRefs = make([]*isc.RequestRef, requestCount)
	for i := range ret.requestRefs {
		ret.requestRefs[i] = &isc.RequestRef{}
		ret.requestRefs[i].ID, err = isc.RequestIDFromMarshalUtil(mu)
		if err != nil {
			return nil, fmt.Errorf(errFmt, err)
		}
		hashBytes, err := mu.ReadBytes(32)
		copy(ret.requestRefs[i].Hash[:], hashBytes)
		if err != nil {
			return nil, fmt.Errorf(errFmt, err)
		}
	}
	return ret, nil
}

func (b *BatchProposal) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint16(b.nodeIndex).
		Write(b.baseAliasOutput).
		Write(b.dssIndexProposal).
		WriteTime(b.timeData).
		Write(b.feeDestination).
		WriteUint16(uint16(len(b.requestRefs)))
	for i := range b.requestRefs {
		mu.Write(b.requestRefs[i].ID)
		mu.WriteBytes(b.requestRefs[i].Hash[:])
	}
	return mu.Bytes()
}
