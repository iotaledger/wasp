// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bp

import (
	"time"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

type BatchProposal struct {
	nodeIndex               uint16                 `bcs:""` // Just for a double-check.
	baseAliasOutput         *iscmove.AnchorWithRef `bcs:""` // Proposed Base AliasOutput to use.
	dssIndexProposal        util.BitVector         `bcs:""` // DSS Index proposal.
	timeData                time.Time              `bcs:""` // Our view of time.
	validatorFeeDestination isc.AgentID            `bcs:""` // Proposed destination for fees.
	requestRefs             []*isc.RequestRef      `bcs:""` // Requests we propose to include into the execution.
	//
	// TODO: Add these fields? How to aggregate them?
	//
	// - gasPayments []*sui.ObjectRef, // optional
	// - gasPrice uint64,
	// - gasBudget uint64,
}

func NewBatchProposal(
	nodeIndex uint16,
	baseAliasOutput *iscmove.AnchorWithRef,
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
	return bcs.MustMarshal(b)
}
