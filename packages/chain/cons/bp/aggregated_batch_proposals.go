// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bp

import (
	"bytes"
	"sort"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// Here we store just an aggregated info.
type AggregatedBatchProposals struct {
	shouldBeSkipped        bool
	batchProposalSet       batchProposalSet
	decidedIndexProposals  map[gpa.NodeID][]int
	decidedBaseAliasOutput *isc.AliasOutputWithID
	decidedRequestRefs     []*isc.RequestRef
	aggregatedTime         time.Time
}

func AggregateBatchProposals(inputs map[gpa.NodeID][]byte, nodeIDs []gpa.NodeID, f int, log *logger.Logger) *AggregatedBatchProposals {
	bps := batchProposalSet{}
	//
	// Parse and validate the batch proposals. Skip the invalid ones.
	for nid := range inputs {
		var batchProposal *BatchProposal
		batchProposal, err := rwutil.ReadFromBytes(inputs[nid], new(BatchProposal))
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
		log.Debugf("Cant' aggregate batch proposal: have 0 batch proposals.")
		return &AggregatedBatchProposals{shouldBeSkipped: true}
	}
	aggregatedTime := bps.aggregatedTime(f)
	decidedBaseAliasOutput := bps.decidedBaseAliasOutput(f)
	abp := &AggregatedBatchProposals{
		batchProposalSet:       bps,
		decidedIndexProposals:  bps.decidedDSSIndexProposals(),
		decidedBaseAliasOutput: decidedBaseAliasOutput,
		decidedRequestRefs:     bps.decidedRequestRefs(f, decidedBaseAliasOutput),
		aggregatedTime:         aggregatedTime,
	}
	if abp.decidedBaseAliasOutput == nil || len(abp.decidedRequestRefs) == 0 || abp.aggregatedTime.IsZero() {
		log.Debugf(
			"Cant' aggregate batch proposal: decidedBaseAliasOutput=%v, |decidedRequestRefs|=%v, aggregatedTime=%v",
			abp.decidedBaseAliasOutput, len(abp.decidedRequestRefs), abp.aggregatedTime,
		)
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

func (abp *AggregatedBatchProposals) DecidedBaseAliasOutput() *isc.AliasOutputWithID {
	if abp.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return abp.decidedBaseAliasOutput
}

func (abp *AggregatedBatchProposals) AggregatedTime() time.Time {
	if abp.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return abp.aggregatedTime
}

func (abp *AggregatedBatchProposals) ValidatorFeeTarget(randomness hashing.HashValue) isc.AgentID {
	if abp.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return abp.batchProposalSet.selectedFeeDestination(abp.aggregatedTime, randomness)
}

func (abp *AggregatedBatchProposals) DecidedRequestRefs() []*isc.RequestRef {
	if abp.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return abp.decidedRequestRefs
}

// TODO should this be moved to the VM?
func (abp *AggregatedBatchProposals) OrderedRequests(requests []isc.Request, randomness hashing.HashValue) []isc.Request {
	type sortStruct struct {
		key hashing.HashValue
		ref *isc.RequestRef
		req isc.Request
	}

	sortBuf := make([]*sortStruct, len(abp.decidedRequestRefs))
	for i := range abp.decidedRequestRefs {
		ref := abp.decidedRequestRefs[i]
		var found isc.Request
		for j := range requests {
			if ref.IsFor(requests[j]) {
				found = requests[j]
				break
			}
		}
		if found == nil {
			panic("request was not provided by mempool")
		}
		sortBuf[i] = &sortStruct{
			key: hashing.HashDataBlake2b(ref.ID.Bytes(), ref.Hash[:], randomness[:]),
			ref: ref,
			req: found,
		}
	}
	sort.Slice(sortBuf, func(i, j int) bool {
		return bytes.Compare(sortBuf[i].key[:], sortBuf[j].key[:]) < 0
	})

	// Make sure the requests are sorted such way, that the nonces per account are increasing.
	// This is needed to handle several requests per batch for the VMs that expect the in-order nonces.
	// We make a second pass here to tain the overall ordering of requests (module account) without
	// making requests from a single account grouped together while sorting.
	for i := range sortBuf {
		oi, ok := sortBuf[i].req.(isc.OffLedgerRequest)
		if !ok {
			continue
		}
		for j := i + 1; j < len(sortBuf); j++ {
			oj, ok := sortBuf[j].req.(isc.OffLedgerRequest)
			if !ok {
				continue
			}
			if oi.SenderAccount().Equals(oj.SenderAccount()) && oi.Nonce() > oj.Nonce() {
				sortBuf[i], sortBuf[j] = sortBuf[j], sortBuf[i]
				oi = oj
			}
		}
	}

	sorted := make([]isc.Request, len(abp.decidedRequestRefs))
	for i := range sortBuf {
		sorted[i] = sortBuf[i].req
	}
	return sorted
}
