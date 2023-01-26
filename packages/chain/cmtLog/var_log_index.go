// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type VarLogIndex interface {
	Value() (LogIndex, *isc.AliasOutputWithID)
	// StartReceived() gpa.OutMessages // TODO: Add the BaseAO parameter, or maybe it is enough to have the L1ReplacedBaseAliasOutput?
	ConsensusOutputReceived(consensusLI LogIndex, nextBaseAO *isc.AliasOutputWithID) gpa.OutMessages
	ConsensusTimeoutReceived(consensusLI LogIndex) gpa.OutMessages
	L1ReplacedBaseAliasOutput(nextBaseAO *isc.AliasOutputWithID) gpa.OutMessages
	MsgNextLogIndexReceived(msg *msgNextLogIndex) gpa.OutMessages
}

// Models the current logIndex variable. The LogIndex advances each time
// a consensus is completed or an unexpected AliasOutput is received from
// the ledger or if nodes agree to proceed to next LogIndex.
//
// There is non-commutative race condition on a lagging nodes.
// In the case of successful consensus round:
//   - It receives NextLI messages from the non-lagging nodes.
//     That makes the lagging node to advance its LI.
//   - Additionally the lagging node receives a ConfirmedAO from L1.
//     That makes the node to advance the LI again, if the ConfirmedAO
//     is received after a quorum of NextLI.
//
// To cope with this the NextLI messages carry the next baseAO (optionally),
// and we have to consider this information upon reception of
// L1ReplacedBaseAliasOutput and other events. This makes all the operations
// commutative, thus convergent.
//
// > UPON Reception of Start(baseAO):
// >    Send ⟨NextLI, LogIndex, baseAO⟩, if not sent for LogIndex or later.
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
//
// TODO: The above algorithm is outdated, the code is modified a bit ad hoc. Some redesign/modeling is needed here.
// The initial idea is that we have to run several consensus instances in parallel until they converge.
type varLogIndexImpl struct {
	nodeIDs    []gpa.NodeID // All the peers in this committee.
	n          int          // Total number of nodes.
	f          int          // Maximal number of faulty nodes to tolerate.
	logIndex   LogIndex
	sentNextLI LogIndex                        // LogIndex for which the MsgNextLogIndex was sent.
	maxPeerLIs map[gpa.NodeID]*msgNextLogIndex // Latest peer indexes received from peers.
	latestAO   *isc.AliasOutputWithID          // TODO: ...
	log        *logger.Logger
}

func NewVarLogIndex(
	nodeIDs []gpa.NodeID,
	n int,
	f int,
	initLI LogIndex, // That's not yet increased after restart.
	log *logger.Logger,
) VarLogIndex {
	log.Debugf("NewVarLogIndex, n=%v, f=%v, initLI=%v", n, f, initLI)
	return &varLogIndexImpl{
		nodeIDs:    nodeIDs,
		n:          n,
		f:          f,
		logIndex:   initLI,
		sentNextLI: NilLogIndex(),
		maxPeerLIs: map[gpa.NodeID]*msgNextLogIndex{},
		log:        log,
	}
}

func (v *varLogIndexImpl) Value() (LogIndex, *isc.AliasOutputWithID) {
	return v.logIndex, v.latestAO
}

// > UPON Reception of Start:
// >    Send ⟨NextLI, LogIndex⟩, if not sent for LogIndex or later.
// func (v *varLogIndexImpl) StartReceived() gpa.OutMessages {
// 	v.log.Debugf("StartReceived, logIndex=%v", v.logIndex)
// 	return v.maybeSendNextLogIndex(v.logIndex.Next())
// }

// > UPON Reception of ConsensusOutput{DONE | SKIP}:
// >    Ignore if outdated.
// > 	LogIndex <- max(LogIndex, ConsensusOutput.LogIndex) + 1
func (v *varLogIndexImpl) ConsensusOutputReceived(consensusLI LogIndex, nextBaseAO *isc.AliasOutputWithID) gpa.OutMessages {
	if consensusLI < v.logIndex {
		return nil
	}
	v.logIndex = consensusLI.Next()
	v.latestAO = nextBaseAO
	v.log.Debugf("ConsensusOutputReceived, next logIndex=%v, latestAO=%v", v.logIndex, v.latestAO)
	return v.maybeSendNextLogIndex(v.logIndex, v.latestAO)
}

// > UPON Reception of ConsensusTimeout:
// >    Ignore if outdated.
// >    Let ProposalToSkip = max(LogIndex, ConsensusOutput.LogIndex) + 1
// >    Send ⟨NextLI, ProposalToSkip⟩, if not sent for ProposalToSkip or later.
func (v *varLogIndexImpl) ConsensusTimeoutReceived(consensusLI LogIndex) gpa.OutMessages {
	if consensusLI < v.logIndex {
		return nil
	}
	v.log.Debugf("ConsensusTimeoutReceived, consensusLI=%v", consensusLI)
	// NOTE: v.latestAO remains the same.
	return v.maybeSendNextLogIndex(consensusLI.Next(), v.latestAO)
}

// Only call this if the LocalView returns new BaseAO after AliasOutput Confirmed/Rejected.
//
// > UPON Reception of L1ReplacedBaseAliasOutput:
// >	LogIndex <- LogIndex + 1
func (v *varLogIndexImpl) L1ReplacedBaseAliasOutput(nextBaseAO *isc.AliasOutputWithID) gpa.OutMessages {
	if v.latestAO == nil || !v.latestAO.Equals(nextBaseAO) {
		// NOTE: Hope this condition will keep algorithm in the stable
		// condition, if the stable condition is already reached.
		// TODO: In general, this should be reviewed, modeled, etc.
		v.logIndex++
		v.latestAO = nextBaseAO
	}
	v.log.Debugf("L1ReplacedBaseAliasOutput, next logIndex=%v, ao=%v", v.logIndex, v.latestAO)
	return v.maybeSendNextLogIndex(v.logIndex, v.latestAO)
}

// > UPON Reception of F+1 ⟨NextLI, li⟩ for li > LogIndex:
// >	Send ⟨NextLI, li⟩, if not sent for li or later.
// > UPON Reception of N-F ⟨NextLI, li⟩ for li > LogIndex:
// >	LogIndex <- li
func (v *varLogIndexImpl) MsgNextLogIndexReceived(msg *msgNextLogIndex) gpa.OutMessages {
	v.log.Debugf("MsgNextLogIndexReceived, me=%v, sender=%v, logIndex=%v, ao=%v", msg.Recipient().ShortString(), msg.Sender().ShortString(), msg.nextLogIndex, msg.nextBaseAO)
	msgs := gpa.NoMessages()
	sender := msg.Sender()
	//
	// Validate and record the vote.
	if !v.knownNodeID(sender) {
		v.log.Warnf("MsgNextLogIndex from unknown sender: %+v", msg)
		return nil
	}
	var prevPeerLI LogIndex
	if prevPeerNLI, ok := v.maxPeerLIs[sender]; ok {
		prevPeerLI = prevPeerNLI.nextLogIndex
	} else {
		prevPeerLI = NilLogIndex()
	}
	if prevPeerLI.AsUint32() >= msg.nextLogIndex.AsUint32() {
		return nil
	}
	v.maxPeerLIs[sender] = msg
	//
	// Support log indexes, if there are F+1 votes for that log index.
	supportLogIndex, supportAO := v.votedFor(v.f + 1)
	// v.log.Debugf("supportLogIndex=%v, logIndex=%v", supportLogIndex, v.logIndex)
	if supportLogIndex > v.logIndex && supportAO != nil {
		// if supportLogIndex > v.logIndex && v.latestAO != nil && supportAO != nil && v.latestAO.Equals(supportAO) {
		msgs.AddAll(v.maybeSendNextLogIndex(supportLogIndex, supportAO))
	}
	//
	// Proceed to the next log index, if needed.
	newLogIndex, newAO := v.votedFor(v.n - v.f)
	v.log.Debugf("supportLogIndex=%v, newLogIndex=%v, logIndex=%v", supportLogIndex, newLogIndex, v.logIndex)
	if newLogIndex > v.logIndex {
		v.logIndex = newLogIndex
		v.latestAO = newAO
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
func (v *varLogIndexImpl) votedFor(quorum int) (LogIndex, *isc.AliasOutputWithID) {
	// TODO: Enough to have an AO from F+1 nodes probably.
	//   - If that's because of consensus -> N-F nodes will use that, and we might get N-2F ~= F+1 messages from correct nodes.
	//   - If that's because we are not yet in sync with L1, then any choice is bad, so we just take any or retain the existing.
	countsLI := map[LogIndex]int{}
	for _, msg := range v.maxPeerLIs {
		countsLI[msg.nextLogIndex]++
	}
	maxLI := NilLogIndex()
	for li, c := range countsLI {
		if c >= quorum && li.AsUint32() > maxLI.AsUint32() {
			maxLI = li
		}
	}

	countsAOMap := map[iotago.OutputID]*isc.AliasOutputWithID{}
	countsAO := map[iotago.OutputID]int{}
	for _, msg := range v.maxPeerLIs {
		if msg.nextLogIndex != maxLI {
			continue
		}
		countsAOMap[msg.nextBaseAO.OutputID()] = msg.nextBaseAO
		countsAO[msg.nextBaseAO.OutputID()]++
	}

	var q1fAO *isc.AliasOutputWithID
	for aoID, c := range countsAO {
		if c >= v.f+1 {
			// It is possible to have 2 AOs with quorum of F+1 in general,
			// but in that case if is not clear which to select anyway.
			// If the selection is unlucky, the consensus can decide SKIP and go to the next attempt.
			q1fAO = countsAOMap[aoID]
			break
		}
	}
	return maxLI, q1fAO
}

func (v *varLogIndexImpl) maybeSendNextLogIndex(logIndex LogIndex, baseAO *isc.AliasOutputWithID) gpa.OutMessages {
	if logIndex < v.logIndex {
		return nil
	}
	if v.sentNextLI.AsUint32() >= logIndex.AsUint32() {
		return nil
	}
	v.sentNextLI = logIndex
	msgs := gpa.NoMessages()
	for i := range v.nodeIDs {
		msgs.Add(newMsgNextLogIndex(v.nodeIDs[i], logIndex, baseAO))
	}
	v.log.Debugf("Sending NextLogIndex=%v, latestAO=%v", logIndex, baseAO)
	return msgs
}
