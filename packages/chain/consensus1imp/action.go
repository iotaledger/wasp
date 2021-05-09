package consensus1imp

import (
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
	c.sendProposalForConsensus(c.batchProposal(reqs))
	c.stage = stageConsensus
	c.stageStarted = time.Now()
}

func (c *consensusImpl) batchProposal(reqs []coretypes.Request) *batchProposal {
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

}

func (c *consensusImpl) receiveOptionsFromConsensus(opt []*batchProposal) {
	for _, prop := range opt {
		if prop.StateOutputID != c.stateOutput.ID() {
			//
			c.log.Warnf("consensus options out of context or consensus failure")
			return
		}
	}
	//medianTs, accessPledge, consensusPledge, feeDestination := c.calcBatchParameters(opt)

}

func (c *consensusImpl) calcBatchParameters(opt []*batchProposal) (time.Time, identity.ID, identity.ID, *coretypes.AgentID) {
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
