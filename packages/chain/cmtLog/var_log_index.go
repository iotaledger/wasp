// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/gpa"
)

type VarLogIndex interface {
	Value() LogIndex
	StartReceived() gpa.OutMessages
	ConsensusOutputReceived(consensusLI LogIndex)
	ConsensusTimeoutReceived(consensusLI LogIndex) gpa.OutMessages
	L1ReplacedBaseAliasOutput()
	MsgNextLogIndexReceived(msg *msgNextLogIndex) gpa.OutMessages
}

// Models the current logIndex variable. The LogIndex advances each time
// a consensus is completed or an unexpected AliasOutput is received from
// the ledger or if nodes agree to proceed to next LogIndex.
//
// > UPON Reception of Start:
// >    Send ⟨NextLI, LogIndex⟩, if not sent for LogIndex or later.
// > UPON Reception of ConsensusOutput{DONE | SKIP}:
// >    Ignore if outdated.
// > 	LogIndex <- max(LogIndex, ConsensusOutput.LogIndex) + 1
// > UPON Reception of ConsensusTimeout:
// >    Ignore if outdated.
// >    Let ProposalToSkip = max(LogIndex, ConsensusOutput.LogIndex) + 1
// >    Send ⟨NextLI, ProposalToSkip⟩, if not sent for ProposalToSkip or later.
// > UPON Reception of L1ReplacedBaseAliasOutput:
// >	LogIndex <- LogIndex + 1
// > UPON Reception of F+1 ⟨NextLI, li⟩ for li > LogIndex:
// >	Send ⟨NextLI, li⟩, if not sent for li or later.
// > UPON Reception of N-F ⟨NextLI, li⟩ for li > LogIndex:
// >	LogIndex <- li
type varLogIndexImpl struct {
	nodeIDs    []gpa.NodeID // All the peers in this committee.
	n          int          // Total number of nodes.
	f          int          // Maximal number of faulty nodes to tolerate.
	logIndex   LogIndex
	sentNextLI LogIndex                // LogIndex for which the MsgNextLogIndex was sent.
	maxPeerLIs map[gpa.NodeID]LogIndex // Latest peer indexes received from peers.
	log        *logger.Logger
}

func NewVarLogIndex(
	nodeIDs []gpa.NodeID,
	n int,
	f int,
	initLI LogIndex,
	log *logger.Logger,
) VarLogIndex {
	return &varLogIndexImpl{
		nodeIDs:    nodeIDs,
		n:          n,
		f:          f,
		logIndex:   initLI,
		sentNextLI: NilLogIndex(),
		maxPeerLIs: map[gpa.NodeID]LogIndex{},
		log:        log,
	}
}

func (v *varLogIndexImpl) Value() LogIndex {
	return v.logIndex
}

// > UPON Reception of Start:
// >    Send ⟨NextLI, LogIndex⟩, if not sent for LogIndex or later.
func (v *varLogIndexImpl) StartReceived() gpa.OutMessages {
	return v.maybeSendNextLogIndex(v.logIndex)
}

// > UPON Reception of ConsensusOutput{DONE | SKIP}:
// >    Ignore if outdated.
// > 	LogIndex <- max(LogIndex, ConsensusOutput.LogIndex) + 1
func (v *varLogIndexImpl) ConsensusOutputReceived(consensusLI LogIndex) {
	if consensusLI < v.logIndex {
		return
	}
	v.logIndex = consensusLI.Next()
}

// > UPON Reception of ConsensusTimeout:
// >    Ignore if outdated.
// >    Let ProposalToSkip = max(LogIndex, ConsensusOutput.LogIndex) + 1
// >    Send ⟨NextLI, ProposalToSkip⟩, if not sent for ProposalToSkip or later.
func (v *varLogIndexImpl) ConsensusTimeoutReceived(consensusLI LogIndex) gpa.OutMessages {
	if consensusLI < v.logIndex {
		return nil
	}
	return v.maybeSendNextLogIndex(consensusLI.Next())
}

// Only call this if the LocalView returns new BaseAO after AliasOutput Confirmed/Rejected.
//
// > UPON Reception of L1ReplacedBaseAliasOutput:
// >	LogIndex <- LogIndex + 1
func (v *varLogIndexImpl) L1ReplacedBaseAliasOutput() {
	v.logIndex++
}

// > UPON Reception of F+1 ⟨NextLI, li⟩ for li > LogIndex:
// >	Send ⟨NextLI, li⟩, if not sent for li or later.
// > UPON Reception of N-F ⟨NextLI, li⟩ for li > LogIndex:
// >	LogIndex <- li
func (v *varLogIndexImpl) MsgNextLogIndexReceived(msg *msgNextLogIndex) gpa.OutMessages {
	msgs := gpa.NoMessages()
	sender := msg.Sender()
	//
	// Validate and record the vote.
	if !v.knownNodeID(sender) {
		v.log.Warnf("MsgNextLogIndex from unknown sender: %+v", msg)
		return nil
	}
	var prevPeerLogIndex LogIndex
	var found bool
	if prevPeerLogIndex, found = v.maxPeerLIs[sender]; !found {
		prevPeerLogIndex = NilLogIndex()
	}
	if prevPeerLogIndex.AsUint32() >= msg.nextLogIndex.AsUint32() {
		return nil
	}
	v.maxPeerLIs[sender] = msg.nextLogIndex
	//
	// Support log indexes, if there are F+1 votes for that log index.
	supportLogIndex := v.votedFor(v.f + 1)
	if supportLogIndex > v.logIndex {
		msgs.AddAll(v.maybeSendNextLogIndex(supportLogIndex))
	}
	//
	// Proceed to the next log index, if needed.
	newLogIndex := v.votedFor(v.n - v.f)
	if newLogIndex > v.logIndex {
		v.logIndex = newLogIndex
	}
	return msgs
}

func (v *varLogIndexImpl) knownNodeID(nodeID gpa.NodeID) bool {
	for i := range v.nodeIDs {
		if v.nodeIDs[i] == nodeID {
			return true
		}
	}
	return false
}

// Find highest LogIndex for which N-F nodes have voted.
// Returns 0, if not found.
func (v *varLogIndexImpl) votedFor(quorum int) LogIndex {
	counts := map[LogIndex]int{}
	for _, li := range v.maxPeerLIs {
		counts[li]++
	}
	max := NilLogIndex()
	for li, c := range counts {
		if c >= quorum && li.AsUint32() > max.AsUint32() {
			max = li
		}
	}
	return max
}

func (v *varLogIndexImpl) maybeSendNextLogIndex(logIndex LogIndex) gpa.OutMessages {
	if logIndex < v.logIndex {
		return nil
	}
	if v.sentNextLI.AsUint32() >= logIndex.AsUint32() {
		return nil
	}
	v.sentNextLI = logIndex
	msgs := gpa.NoMessages()
	for i := range v.nodeIDs {
		msgs.Add(newMsgNextLogIndex(v.nodeIDs[i], logIndex))
	}
	return msgs
}
