// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bp

import (
	"time"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

// Here we store just an aggregated info.
type AggregatedBatchProposals struct {
	shouldBeSkipped          bool
	decidedIndexProposals    map[gpa.NodeID][]int
	decidedBaseAliasOutputID *iotago.OutputID
	decidedRequestRefs       []*isc.RequestRef
	aggregatedTime           time.Time
	validatorFeeTarget       isc.AgentID
}

func AggregateBatchProposals(inputs map[gpa.NodeID][]byte, nodeIDs []gpa.NodeID, f int, log *logger.Logger) *AggregatedBatchProposals {
	bps := batchProposalSet{}
	//
	// Parse and validate the batch proposals. Skip the invalid ones.
	for nid := range inputs {
		batchProposal, err := batchProposalFromBytes(inputs[nid])
		if err != nil {
			log.Warnf("cannot decode BatchProposal from %v: %v", nid, err)
			continue
		}
		if int(batchProposal.nodeIndex) >= len(nodeIDs) || nodeIDs[batchProposal.nodeIndex] != nid {
			log.Warnf("invalid nodeIndex=%v in batchProposal from %v", batchProposal.nodeIndex, nid)
			continue
		}
		bps[nid] = batchProposal
	}
	//
	// Store the aggregated values.
	if len(bps) == 0 {
		return &AggregatedBatchProposals{shouldBeSkipped: true}
	}
	aggregatedTime := bps.aggregatedTime(f)
	abp := &AggregatedBatchProposals{
		decidedIndexProposals:    bps.decidedDSSIndexProposals(),
		decidedBaseAliasOutputID: bps.decidedBaseAliasOutputID(f),
		decidedRequestRefs:       bps.decidedRequestRefs(f),
		aggregatedTime:           aggregatedTime,
		validatorFeeTarget:       bps.selectedFeeDestination(aggregatedTime),
	}
	if abp.decidedBaseAliasOutputID == nil || len(abp.decidedRequestRefs) == 0 || abp.aggregatedTime.IsZero() {
		abp.shouldBeSkipped = true
	}
	return abp
}

func (abp *AggregatedBatchProposals) ShouldBeSkipped() bool {
	return abp.shouldBeSkipped
}

func (abp *AggregatedBatchProposals) DecidedDSSIndexProposals() map[gpa.NodeID][]int {
	if abp.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return abp.decidedIndexProposals
}

func (abp *AggregatedBatchProposals) DecidedBaseAliasOutputID() *iotago.OutputID {
	if abp.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return abp.decidedBaseAliasOutputID
}

func (abp *AggregatedBatchProposals) AggregatedTime() time.Time {
	if abp.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return abp.aggregatedTime
}

func (abp *AggregatedBatchProposals) ValidatorFeeTarget() isc.AgentID {
	if abp.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return abp.validatorFeeTarget
}

func (abp *AggregatedBatchProposals) DecidedRequestRefs() []*isc.RequestRef {
	if abp.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return abp.decidedRequestRefs
}
