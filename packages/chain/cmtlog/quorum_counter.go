package cmtlog

import (
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/packages/gpa"
)

type QuorumCounter struct {
	msgCause     MsgNextLogIndexCause
	nodeIDs      []gpa.NodeID
	maxPeerVotes map[gpa.NodeID]*MsgNextLogIndex // Latest peer indexes received from peers.
	lastSentMsgs map[gpa.NodeID]*MsgNextLogIndex // Latest messages sent to peers.
	myLastVoteLI LogIndex
	log          log.Logger
}

func NewQuorumCounter(msgCause MsgNextLogIndexCause, nodeIDs []gpa.NodeID, log log.Logger) *QuorumCounter {
	return &QuorumCounter{
		msgCause:     msgCause,
		nodeIDs:      nodeIDs,
		maxPeerVotes: map[gpa.NodeID]*MsgNextLogIndex{},
		lastSentMsgs: map[gpa.NodeID]*MsgNextLogIndex{},
		myLastVoteLI: NilLogIndex(),
		log:          log,
	}
}

func (qc *QuorumCounter) MaybeSendVote(li LogIndex) gpa.OutMessages {
	if li <= qc.myLastVoteLI {
		return nil
	}
	qc.myLastVoteLI = li
	msgs := gpa.NoMessages()
	for _, nodeID := range qc.nodeIDs {
		_, haveMsgFrom := qc.maxPeerVotes[nodeID] // It might happen, that we rebooted and lost the state.
		msg := NewMsgNextLogIndex(nodeID, li, qc.msgCause, !haveMsgFrom)
		qc.lastSentMsgs[nodeID] = msg
		msgs.Add(msg)
	}
	return msgs
}

func (qc *QuorumCounter) MyLastVote() LogIndex {
	return qc.myLastVoteLI
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

func (qc *QuorumCounter) EnoughVotes(quorum int) LogIndex {
	countsLI := map[LogIndex]int{}
	for _, vote := range qc.maxPeerVotes {
		countsLI[vote.NextLogIndex]++
	}
	maxLI := NilLogIndex()
	for li := range countsLI {
		// Count votes: all vote for this LI, if votes for it or higher LI.
		c := 0
		for li2, c2 := range countsLI {
			if li2 >= li {
				c += c2
			}
		}
		// If quorum reached and it is higher than we had before, take it.
		if c >= quorum && li > maxLI {
			maxLI = li
		}
	}
	return maxLI
}
