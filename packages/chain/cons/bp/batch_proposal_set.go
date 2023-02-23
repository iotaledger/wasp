// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bp

import (
	"sort"
	"time"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

type batchProposalSet map[gpa.NodeID]*BatchProposal

func (bps batchProposalSet) decidedDSSIndexProposals() map[gpa.NodeID][]int {
	ips := map[gpa.NodeID][]int{}
	for nid, bp := range bps {
		ips[nid] = bp.dssIndexProposal.AsInts()
	}
	return ips
}

// Decided Base Alias Output is the one, that was proposed by F+1 nodes or more.
// If there is more that 1 such ID, we refuse to use all of them.
func (bps batchProposalSet) decidedBaseAliasOutput(f int) *isc.AliasOutputWithID {
	counts := map[hashing.HashValue]int{}
	values := map[hashing.HashValue]*isc.AliasOutputWithID{}
	for _, bp := range bps {
		h := bp.baseAliasOutput.Hash()
		counts[h]++
		if _, ok := values[h]; !ok {
			values[h] = bp.baseAliasOutput
		}
	}

	var found *isc.AliasOutputWithID
	for h, count := range counts {
		if count > f {
			if found != nil {
				// Found more that 1 AliasOutput proposed by F+1 or more nodes.
				return nil
			}
			found = values[h]
		}
	}
	return found
}

// Take requests proposed by at least F+1 nodes. Then the request is proposed at least by 1 fair node.
func (bps batchProposalSet) decidedRequestRefs(f int) []*isc.RequestRef {
	minNumberMentioned := f + 1
	requestsByKey := map[isc.RequestRefKey]*isc.RequestRef{}
	numMentioned := map[isc.RequestRefKey]int{}
	//
	// Count number of nodes proposing a request.
	maxLen := 0
	for _, bp := range bps {
		for _, reqRef := range bp.requestRefs {
			reqRefFey := reqRef.AsKey()
			numMentioned[reqRefFey]++
			if _, ok := requestsByKey[reqRefFey]; !ok {
				requestsByKey[reqRefFey] = reqRef
			}
		}
		if len(bp.requestRefs) > maxLen {
			maxLen = len(bp.requestRefs)
		}
	}
	//
	// Select the requests proposed by F+1 nodes.
	decided := make([]*isc.RequestRef, 0, maxLen)
	for key, num := range numMentioned {
		if num < minNumberMentioned {
			continue
		}
		decided = append(decided, requestsByKey[key])
	}
	return decided
}

// Returns zero time, if fails to aggregate the time.
func (bps batchProposalSet) aggregatedTime(f int) time.Time {
	ts := make([]time.Time, 0, len(bps))
	for _, bp := range bps {
		ts = append(ts, bp.timeData)
	}
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Before(ts[j])
	})

	proposalCount := len(bps) // |acsProposals| >= N-F by ACS logic.
	if proposalCount <= f {
		return time.Time{} // Zero time marks a failure.
	}
	return ts[proposalCount-f-1] // Max(|acsProposals|-F Lowest) ~= 66 percentile.
}

func (bps batchProposalSet) selectedProposal(aggregatedTime time.Time) gpa.NodeID {
	peers := make([]gpa.NodeID, 0, len(bps))
	for nid := range bps {
		peers = append(peers, nid)
	}
	rnd := util.NewPseudoRand(aggregatedTime.UnixNano())
	return peers[rnd.Intn(len(bps))]
}

func (bps batchProposalSet) selectedFeeDestination(aggregatedTime time.Time) isc.AgentID {
	bp := bps[bps.selectedProposal(aggregatedTime)]
	return bp.feeDestination
}
