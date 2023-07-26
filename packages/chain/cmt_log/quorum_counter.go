package cmt_log

import (
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type QuorumCounter struct {
	msgCause     MsgNextLogIndexCause
	nodeIDs      []gpa.NodeID
	maxPeerVotes map[gpa.NodeID]*MsgNextLogIndex // Latest peer indexes received from peers.
	lastSentMsgs map[gpa.NodeID]*MsgNextLogIndex // Latest messages sent to peers.
	myLastVoteLI LogIndex
	myLastVoteAO *isc.AliasOutputWithID
	log          *logger.Logger
}

func NewQuorumCounter(msgCause MsgNextLogIndexCause, nodeIDs []gpa.NodeID, log *logger.Logger) *QuorumCounter {
	return &QuorumCounter{
		msgCause:     msgCause,
		nodeIDs:      nodeIDs,
		maxPeerVotes: map[gpa.NodeID]*MsgNextLogIndex{},
		log:          log,
	}
}

func (qc *QuorumCounter) MaybeSendVote(li LogIndex, ao *isc.AliasOutputWithID) gpa.OutMessages {
	if li <= qc.myLastVoteLI {
		return nil
	}
	qc.myLastVoteLI = li
	qc.myLastVoteAO = ao
	msgs := gpa.NoMessages()
	for _, nodeID := range qc.nodeIDs {
		_, haveMsgFrom := qc.maxPeerVotes[nodeID] // It might happen, that we rebooted and lost the state.
		msg := NewMsgNextLogIndex(nodeID, li, ao, qc.msgCause, !haveMsgFrom)
		qc.lastSentMsgs[nodeID] = msg
		msgs.Add(msg)
	}
	return msgs
}

func (qc *QuorumCounter) MyLastVote() (LogIndex, *isc.AliasOutputWithID) {
	return qc.myLastVoteLI, qc.myLastVoteAO
}

func (qc *QuorumCounter) LastMessageForPeer(peer gpa.NodeID, msgs gpa.OutMessages) gpa.OutMessages {
	if msg, ok := qc.lastSentMsgs[peer]; ok {
		msgs.Add(msg.AsResent())
	}
	return msgs
}

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

func (qc *QuorumCounter) EnoughVotes(quorum int, countEqual bool) (LogIndex, *isc.AliasOutputWithID) {
	countsLI := map[iotago.OutputID]map[LogIndex]int{}
	aos := map[iotago.OutputID]*isc.AliasOutputWithID{}
	for _, vote := range qc.maxPeerVotes {
		var oid iotago.OutputID // Will keep it nil, if !countEqual.
		if countEqual && vote.NextBaseAO != nil {
			oid = vote.NextBaseAO.OutputID()
			if _, ok := aos[oid]; !ok {
				aos[oid] = vote.NextBaseAO
			}
		}
		if _, ok := countsLI[oid]; !ok {
			countsLI[oid] = map[LogIndex]int{}
		}
		countsLI[oid][vote.NextLogIndex]++
	}
	maxLI := NilLogIndex()
	var maxAO *isc.AliasOutputWithID
	for oid, oidCounts := range countsLI {
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
				maxAO = aos[oid] // Might be nil.
			}
		}
	}
	return maxLI, maxAO
}
