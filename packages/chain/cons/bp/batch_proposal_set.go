// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bp

import (
	"bytes"
	"encoding/binary"
	"slices"
	"sort"
	"time"

	"golang.org/x/exp/maps"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
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
func (bps batchProposalSet) decidedBaseAliasOutput(f int) *isc.StateAnchor {
	counts := map[hashing.HashValue]int{}
	values := map[hashing.HashValue]*isc.StateAnchor{}
	for _, bp := range bps {
		if bp.baseAliasOutput == nil {
			continue
		}
		h := bp.baseAliasOutput.Hash()
		counts[h]++
		if _, ok := values[h]; !ok {
			values[h] = bp.baseAliasOutput
		}
	}

	var found *isc.StateAnchor
	var uncertain bool
	for h, count := range counts {
		if count > f {
			if found != nil && found.GetStateIndex() == values[h].GetStateIndex() {
				// Found more that 1 AliasOutput proposed by F+1 or more nodes.
				uncertain = true
				continue
			}
			if found == nil || found.GetStateIndex() < values[h].GetStateIndex() {
				found = values[h]
				uncertain = false
			}
		}
	}
	if uncertain {
		return nil
	}
	return found
}

// Take requests proposed by at least F+1 nodes. Then the request is proposed at least by 1 fair node.
// We should only consider the proposals from the nodes that proposed the decided AO, otherwise we can select already processed requests.
func (bps batchProposalSet) decidedRequestRefs(f int, ao *isc.StateAnchor) []*isc.RequestRef {
	minNumberMentioned := f + 1
	requestsByKey := map[isc.RequestRefKey]*isc.RequestRef{}
	numMentioned := map[isc.RequestRefKey]int{}
	//
	// Count number of nodes proposing a request.
	maxLen := 0
	for _, bp := range bps {
		if bp.baseAliasOutput == nil || !bp.baseAliasOutput.Equals(ao) {
			continue
		}
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

func (bps batchProposalSet) decidedRotateTo(f int) *iotago.Address {
	votes := map[iotago.Address]int{}
	for _, bp := range bps {
		if bp.rotateTo != nil {
			votes[*bp.rotateTo] += 1
		}
	}

	var found *iotago.Address
	for address, count := range votes {
		thisAddr := address
		if count > f {
			if found != nil {
				// 2 values with counter > f, thus a collision.
				return nil
			}
			found = &thisAddr
		}
	}
	return found
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

func (bps batchProposalSet) selectedProposal(aggregatedTime time.Time, randomness hashing.HashValue) gpa.NodeID {
	peers := make([]gpa.NodeID, 0, len(bps))
	for nid := range bps {
		peers = append(peers, nid)
	}
	slices.SortFunc(peers, func(a gpa.NodeID, b gpa.NodeID) int {
		return bytes.Compare(a[:], b[:])
	})
	uint64Bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(uint64Bytes, uint64(aggregatedTime.UnixNano()))
	hashed := hashing.HashDataBlake2b(
		uint64Bytes,
		randomness[:],
	)
	randomUint := binary.BigEndian.Uint64(hashed[:])
	randomPos := int(randomUint % uint64(len(bps)))
	return peers[randomPos]
}

func (bps batchProposalSet) selectedFeeDestination(aggregatedTime time.Time, randomness hashing.HashValue) isc.AgentID {
	bp := bps[bps.selectedProposal(aggregatedTime, randomness)]
	return bp.validatorFeeDestination
}

type l1paramsCounter struct {
	counter  int
	l1params *parameters.L1Params
}

// Take the L1Params which is shared more than f+1 nodes
func (bps batchProposalSet) aggregatedL1Params(f int) *parameters.L1Params {
	proposalCount := len(bps) // |acsProposals| >= N-F by ACS logic.
	ps := make([]*parameters.L1Params, 0, proposalCount)
	for _, bp := range bps {
		if bp.l1params == nil {
			continue
		}
		ps = append(ps, bp.l1params)
	}

	// count the amount of each L1Params
	protocolMap := make(map[string]l1paramsCounter)
	var l1paramsCounterMax l1paramsCounter
	for _, l1params := range ps {
		elt, ok := protocolMap[l1params.Hash().Hex()]
		if ok {
			elt.counter += 1
			protocolMap[l1params.Hash().Hex()] = elt
		} else {
			elt = l1paramsCounter{
				counter:  1,
				l1params: l1params,
			}
			protocolMap[l1params.Hash().Hex()] = elt
		}
		if elt.counter > l1paramsCounterMax.counter {
			l1paramsCounterMax.counter = elt.counter
			l1paramsCounterMax.l1params = elt.l1params
		}
	}

	matchingCount := lo.CountBy(lo.Values(protocolMap), func(elt l1paramsCounter) bool {
		return elt.counter > f
	})
	if matchingCount != 1 {
		return nil
	}

	return l1paramsCounterMax.l1params
}

// Here we return coins that are proposed by at least F+1 peers.
func (bps batchProposalSet) aggregatedGasCoins(f int) []*coin.CoinWithRef {
	coinRefs := map[string]*coin.CoinWithRef{}
	coinFrom := map[string]map[gpa.NodeID]bool{}
	for from, bp := range bps {
		for i := range bp.gasCoins {
			coinRef := bp.gasCoins[i]
			bytesStr := string(coinRef.Ref.Bytes())
			if _, ok := coinRefs[bytesStr]; !ok {
				coinRefs[bytesStr] = coinRef
				coinFrom[bytesStr] = map[gpa.NodeID]bool{}
			}
			coinFrom[bytesStr][from] = true
		}
	}

	// Drop the coins proposed by less than F+1 nodes.
	for i, cf := range coinFrom {
		if len(cf) < f+1 {
			delete(coinFrom, i)
		}
	}

	// Drop older versions of the same coin.
	for i := range coinFrom {
		ci := coinRefs[i].Ref
		haveNewer := lo.ContainsBy(maps.Keys(coinFrom), func(j string) bool {
			cj := coinRefs[i].Ref
			return ci.ObjectID.Equals(*cj.ObjectID) && ci.Version < cj.Version
		})
		if haveNewer {
			delete(coinFrom, i)
		}
	}

	// Sort them by the proposal frequency, then by the bytes.
	coinKeys := maps.Keys(coinFrom)
	sort.Slice(coinKeys, func(i, j int) bool {
		fromI := len(coinFrom[coinKeys[i]])
		fromJ := len(coinFrom[coinKeys[j]])
		return fromI < fromJ || (fromI == fromJ && coinKeys[i] < coinKeys[j])
	})

	// Return the selected coins.
	result := []*coin.CoinWithRef{}
	for _, coinKey := range coinKeys {
		result = append(result, coinRefs[coinKey])
	}
	return result
}
