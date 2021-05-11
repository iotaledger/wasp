package consensus1imp

import (
	"bytes"
	"sort"
	"time"

	"github.com/iotaledger/wasp/packages/util"

	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (c *consensusImpl) takeAction() {
	c.startConsensusIfNeeded()
}

func (c *consensusImpl) startConsensusIfNeeded() {
	if c.stage != stageIdle {
		return
	}
	reqs := c.mempool.GetReadyList()
	if len(reqs) == 0 {
		return
	}
	c.sendProposalForConsensus(c.prepareBatchProposal(reqs))
}

func (c *consensusImpl) prepareBatchProposal(reqs []coretypes.Request) *batchProposal {
	ts := time.Now()
	if !ts.After(c.stateTimestamp) {
		ts = c.stateTimestamp.Add(1 * time.Nanosecond)
	}
	consensusManaPledge := identity.ID{}
	accessManaPledge := identity.ID{}
	feeDestination := coretypes.NewAgentID(c.chain.ID().AsAddress(), 0)
	ret := &batchProposal{
		StateOutputID:       c.stateOutput.ID(),
		RequestIDs:          make([]coretypes.RequestID, len(reqs)),
		Timestamp:           ts,
		ConsensusManaPledge: consensusManaPledge,
		AccessManaPledge:    accessManaPledge,
		FeeDestination:      feeDestination,
	}
	for i := range ret.RequestIDs {
		ret.RequestIDs[i] = reqs[i].ID()
	}
	return ret
}

func (c *consensusImpl) sendProposalForConsensus(proposal *batchProposal) {
	// TODO
	c.stage = stageConsensus
	c.stageStarted = time.Now()
}

func (c *consensusImpl) receiveBatchOptionsFromConsensus(opt []*batchProposal) {
	for _, prop := range opt {
		if prop.StateOutputID != c.stateOutput.ID() {
			//
			c.log.Warnf("receiveOptionsFromConsensus: consensus options out of context or consensus failure")
			return
		}
		if prop.ValidatorIndex >= c.committee.Size() {
			c.log.Warnf("wrong validtor index from consensus")
			return
		}
	}
	inBatchSet := calcIntersection(opt, c.committee.Size(), c.committee.Quorum())
	if len(inBatchSet) == 0 {
		c.log.Warnf("receiveOptionsFromConsensus: intersecection is empty. Consensus failure")
		return
	}
	medianTs, accessPledge, consensusPledge, feeDestination := calcBatchParameters(opt)
	c.consensusBatch = &batchProposal{
		ValidatorIndex:      c.committee.OwnPeerIndex(),
		StateOutputID:       c.stateOutput.ID(),
		RequestIDs:          inBatchSet,
		Timestamp:           medianTs,
		ConsensusManaPledge: consensusPledge,
		AccessManaPledge:    accessPledge,
		FeeDestination:      feeDestination,
	}
	c.stage = stageConsensusCompleted
	c.stageStarted = time.Now()
	c.runVMIfNeeded()
}

func (c *consensusImpl) runVMIfNeeded() {
	if c.stage != stageConsensusCompleted {
		return
	}
	reqs := c.mempool.GetRequestsByIDs(c.consensusBatch.Timestamp, c.consensusBatch.RequestIDs...)
	// check if all ready
	for _, req := range reqs {
		if req == nil {
			// some requests are not ready, so skip VM call this time. Maybe next time will be luckier
			c.log.Debugf("runVMIfNeeded: not all requests ready for processing")
			return
		}
	}
	// sort for determinism and for the reason to arrange off-ledger requests
	// equal Order() request are ordered by ID
	sort.Slice(reqs, func(i, j int) bool {
		switch {
		case reqs[i].Order() < reqs[j].Order():
			return true
		case reqs[i].Order() > reqs[j].Order():
			return false
		default:
			return bytes.Compare(reqs[i].ID().Bytes(), reqs[j].ID().Bytes()) < 0
		}
	})

	// TODO run VM async

}

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
