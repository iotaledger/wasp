package cmt_log

import (
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type QuorumCounter struct {
	maxPeerVotes map[gpa.NodeID]*MsgNextLogIndex // Latest peer indexes received from peers.
	myLastVoteLI LogIndex
	myLastVoteAO *isc.AliasOutputWithID
	log          *logger.Logger
}

func NewQuorumCounter(log *logger.Logger) *QuorumCounter {
	return &QuorumCounter{
		maxPeerVotes: map[gpa.NodeID]*MsgNextLogIndex{},
		log:          log,
	}
}

// func (qc *QuorumCounter) CastVote() MsgNextLogIndex {
// 	return *NewMsgNextLogIndex(rec)
// }

func (qc *QuorumCounter) VoteReceived(vote *MsgNextLogIndex) {
	sender := vote.Sender()
	var prevPeerLI LogIndex
	if prevPeerNLI, ok := qc.maxPeerVotes[sender]; ok {
		prevPeerLI = prevPeerNLI.NextLogIndex
	} else {
		prevPeerLI = NilLogIndex()
	}
	if prevPeerLI.AsUint32() >= vote.NextLogIndex.AsUint32() {
		return
	}
	qc.maxPeerVotes[sender] = vote
}

func (qc *QuorumCounter) HaveVoteFrom(from gpa.NodeID) bool {
	_, have := qc.maxPeerVotes[from]
	return have
}

func (qc *QuorumCounter) EnoughVotes(quorum int, countEqual bool) LogIndex {
	countsLI := map[iotago.OutputID]map[LogIndex]int{}
	for _, vote := range qc.maxPeerVotes {
		var oid iotago.OutputID // Will keep it nil, if !countEqual.
		if countEqual && vote.NextBaseAO != nil {
			oid = vote.NextBaseAO.OutputID()
		}
		if _, ok := countsLI[oid]; !ok {
			countsLI[oid] = map[LogIndex]int{}
		}
		countsLI[oid][vote.NextLogIndex]++
	}
	maxLI := NilLogIndex()
	for _, oidCounts := range countsLI {
		for li := range oidCounts {
			// Count votes: all vote for this LI, if votes for it or higher LI.
			c := 0
			for li2, c2 := range oidCounts {
				if li2 >= li {
					c += c2
				}
			}
			// If quorum reached and it is higher than we had before, take it.
			if c >= quorum && li > maxLI {
				maxLI = li
			}
		}
	}
	return maxLI
}
