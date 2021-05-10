package consensus1imp

import (
	"bytes"
	"sort"
	"time"

	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
)

func calcBatchParameters(opt []*batchProposal) (time.Time, identity.ID, identity.ID, *coretypes.AgentID) {
	var retTS time.Time

	ts := make([]time.Time, len(opt))
	for i := range ts {
		ts[i] = opt[i].Timestamp
	}
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Before(ts[j])
	})
	retTS = ts[len(opt)/2]

	indices := make([]uint16, len(opt))
	for i := range indices {
		indices[i] = opt[i].ValidatorIndex
	}
	selectedIndex := util.SelectRandomUint16(indices, retTS.UnixNano())

	return retTS, opt[selectedIndex].AccessManaPledge, opt[selectedIndex].ConsensusManaPledge, opt[selectedIndex].FeeDestination
}

// deterministically calculate intersection. It may not be optimal
func calcIntersection(opt []*batchProposal, n, t uint16) []coretypes.RequestID {
	matrix := make(map[coretypes.RequestID][]bool)
	for _, b := range opt {
		for _, reqid := range b.RequestIDs {
			_, ok := matrix[reqid]
			if !ok {
				matrix[reqid] = make([]bool, n)
			}
			matrix[reqid][b.ValidatorIndex] = true
		}
	}
	// collect those which are seen by more nodes than quorum. The rest is not interesting
	seenByQuorum := make([]coretypes.RequestID, 0)
	maxSeen := t
	for reqid, seenVect := range matrix {
		numSeen := countTrue(seenVect)
		if numSeen >= t {
			seenByQuorum = append(seenByQuorum, reqid)
			if numSeen > maxSeen {
				maxSeen = numSeen
			}
		}
	}
	// seenByQuorum may be empty. sort for determinism
	sort.Slice(seenByQuorum, func(i, j int) bool {
		return bytes.Compare(seenByQuorum[i][:], seenByQuorum[j][:]) < 0
	})
	inBatchSet := make(map[coretypes.RequestID]struct{})
	var inBatchIntersection []bool
	for numSeen := maxSeen; numSeen >= t; numSeen-- {
		for _, reqid := range seenByQuorum {
			if _, ok := inBatchSet[reqid]; ok {
				continue
			}
			if countTrue(matrix[reqid]) != numSeen {
				continue
			}
			if inBatchIntersection == nil {
				// starting from the largest number seen, so at least one is guaranteed in the batch
				inBatchIntersection = matrix[reqid]
				inBatchSet[reqid] = struct{}{}
			} else {
				sect := intersect(inBatchIntersection, matrix[reqid])
				if countTrue(sect) >= t {
					inBatchIntersection = sect
					inBatchSet[reqid] = struct{}{}
				}
			}
		}
	}
	ret := make([]coretypes.RequestID, 0, len(inBatchSet))
	for reqid := range inBatchSet {
		ret = append(ret, reqid)
	}
	return ret
}

func countTrue(arr []bool) uint16 {
	var ret uint16
	for _, v := range arr {
		if v {
			ret++
		}
	}
	return ret
}

func intersect(arr1, arr2 []bool) []bool {
	if len(arr1) != len(arr2) {
		panic("len(arr1) != len(arr2)")
	}
	ret := make([]bool, len(arr1))
	for i := range ret {
		ret[i] = arr1[i] && arr2[i]
	}
	return ret
}
