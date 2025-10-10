// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package batchproposal

import (
	"bytes"
	"sort"
	"time"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/parameters"
)

// AggregatedBatchProposals stores just an aggregated info.
type AggregatedBatchProposals struct {
	shouldBeSkipped       bool
	batchProposalSet      batchProposalSet
	decidedIndexProposals map[gpa.NodeID][]int
	decidedBaseAnchor     *isc.StateAnchor
	decidedRequestRefs    []*isc.RequestRef
	decidedRotateTo       *iotago.Address
	aggregatedTime        time.Time
	aggregatedGasCoins    []*coin.CoinWithRef
	aggregatedL1Params    *parameters.L1Params
}

func AggregateBatchProposals(inputs map[gpa.NodeID][]byte, nodeIDs []gpa.NodeID, f int, log log.Logger) *AggregatedBatchProposals {
	batchProposals := batchProposalSet{}
	//
	// Parse and validate the batch proposals. Skip the invalid ones.
	nilCount := 0
	for nid := range inputs {
		if len(inputs[nid]) == 0 {
			log.LogWarnf("cannot decode empty BatchProposal from %v", nid)
			continue
		}
		batchProposal, err := bcs.Unmarshal[*BatchProposal](inputs[nid])
		if err != nil {
			log.LogWarnf("cannot decode BatchProposal from %v: %v", nid, err)
			continue
		}
		if batchProposal.baseAnchor == nil {
			nilCount++
		}
		if int(batchProposal.nodeIndex) >= len(nodeIDs) || nodeIDs[batchProposal.nodeIndex] != nid {
			log.LogWarnf("invalid nodeIndex=%v in batchProposal from %v", batchProposal.nodeIndex, nid)
			continue
		}
		batchProposals[nid] = batchProposal
	}
	//
	// Store the aggregated values.
	if nilCount > f {
		log.LogDebugf("Can't aggregate batch proposal: have >= f+1 nil proposals.")
		return &AggregatedBatchProposals{shouldBeSkipped: true}
	}
	if len(batchProposals) == 0 {
		log.LogDebugf("Can't aggregate batch proposal: have 0 batch proposals.")
		return &AggregatedBatchProposals{shouldBeSkipped: true}
	}
	aggregatedTime := batchProposals.aggregatedTime(f)
	decidedBaseAnchor := batchProposals.decidedBaseAnchor(f)
	aggregatedGasCoins := batchProposals.aggregatedGasCoins(f)
	aggregatedL1Params := batchProposals.aggregatedL1Params(f)
	aggregatedBatchProposals := &AggregatedBatchProposals{
		batchProposalSet:      batchProposals,
		decidedIndexProposals: batchProposals.decidedDistributedSignatureIndexProposals(),
		decidedBaseAnchor:     decidedBaseAnchor,
		decidedRequestRefs:    batchProposals.decidedRequestRefs(f, decidedBaseAnchor),
		decidedRotateTo:       batchProposals.decidedRotateTo(f),
		aggregatedTime:        aggregatedTime,
		aggregatedGasCoins:    aggregatedGasCoins,
		aggregatedL1Params:    aggregatedL1Params,
	}
	if aggregatedBatchProposals.decidedBaseAnchor == nil ||
		len(aggregatedBatchProposals.decidedRequestRefs) == 0 ||
		// No need to check the rotateTo field here.
		aggregatedBatchProposals.aggregatedTime.IsZero() ||
		len(aggregatedBatchProposals.aggregatedGasCoins) == 0 ||
		aggregatedBatchProposals.aggregatedL1Params == nil {
		log.LogDebugf(
			"Can't aggregate batch proposal: decidedBaseAnchor=%v, |decidedRequestRefs|=%v, |aggregatedGasCoins|=%v, |aggregatedL1Params|=%v , aggregatedTime=%v",
			aggregatedBatchProposals.decidedBaseAnchor, len(aggregatedBatchProposals.decidedRequestRefs), len(aggregatedBatchProposals.aggregatedGasCoins), aggregatedBatchProposals.aggregatedL1Params, aggregatedBatchProposals.aggregatedTime,
		)
		aggregatedBatchProposals.shouldBeSkipped = true
	}
	return aggregatedBatchProposals
}

func (p *AggregatedBatchProposals) ShouldBeSkipped() bool {
	return p.shouldBeSkipped
}

func (p *AggregatedBatchProposals) DecidedDistributedSignatureIndexProposals() map[gpa.NodeID][]int {
	if p.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return p.decidedIndexProposals
}

func (p *AggregatedBatchProposals) DecidedBaseAnchor() *isc.StateAnchor {
	if p.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return p.decidedBaseAnchor
}

func (p *AggregatedBatchProposals) DecidedRotateTo() *iotago.Address {
	if p.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return p.decidedRotateTo
}

func (p *AggregatedBatchProposals) AggregatedTime() time.Time {
	if p.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return p.aggregatedTime
}

func (p *AggregatedBatchProposals) ValidatorFeeTarget(randomness hashing.HashValue) isc.AgentID {
	if p.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return p.batchProposalSet.selectedFeeDestination(p.aggregatedTime, randomness)
}

func (p *AggregatedBatchProposals) DecidedRequestRefs() []*isc.RequestRef {
	if p.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return p.decidedRequestRefs
}

// OrderedRequests returns ordered requests.
// TODO: should this be moved to the VM?
func (p *AggregatedBatchProposals) OrderedRequests(requests []isc.Request, randomness hashing.HashValue) []isc.Request {
	type sortStruct struct {
		key hashing.HashValue
		ref *isc.RequestRef
		req isc.Request
	}

	sortBuf := make([]*sortStruct, len(p.decidedRequestRefs))
	for i := range p.decidedRequestRefs {
		ref := p.decidedRequestRefs[i]
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

	sorted := make([]isc.Request, len(p.decidedRequestRefs))
	for i := range sortBuf {
		sorted[i] = sortBuf[i].req
	}
	return sorted
}

func (p *AggregatedBatchProposals) AggregatedGasCoins() []*coin.CoinWithRef {
	if p.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return p.aggregatedGasCoins
}

func (p *AggregatedBatchProposals) AggregatedL1Params() *parameters.L1Params {
	if p.shouldBeSkipped {
		panic("trying to use aggregated proposal marked to be skipped")
	}
	return p.aggregatedL1Params
}
