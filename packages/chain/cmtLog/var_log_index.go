// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cons"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type VarLogIndex interface {
	// Returns the latest agreed LI/AO.
	// There is no output value, if LogIndex=⊥.
	Value() (LogIndex, *isc.AliasOutputWithID)
	// Consensus terminated with either with DONE or SKIP.
	// NextAO is the same as before for the SKIP case.
	// NextAO = nil means we don't know the latest AO from L1 (either startup or reject of a chain is happening).
	ConsensusOutputReceived(consensusLI LogIndex, consensusStatus cons.OutputStatus, nextBaseAO *isc.AliasOutputWithID) gpa.OutMessages
	// Consensus decided, that its maybe time to attempt another run.
	// The timeout-ed consensus will be still running, so they will race for the result.
	ConsensusTimeoutReceived(consensusLI LogIndex) gpa.OutMessages
	// This might get a nil alias output in the case when a TX gets rejected and it was not the latest TX in the active chain.
	// NextAO = nil means we don't know the latest AO from L1 (either startup or reject of a chain is happening).
	L1ReplacedBaseAliasOutput(nextBaseAO *isc.AliasOutputWithID) gpa.OutMessages
	// Messages are exchanged, so this function handles them.
	MsgNextLogIndexReceived(msg *msgNextLogIndex) gpa.OutMessages
	// Summary of the internal state.
	StatusString() string
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
// This algorithm don't has to be consensus with a strict single correct answer,
// but it must converge fast, assuming L1 converges. Additionally, in the good
// conditions, it should not de-synchronize itself (like the race condition above).
//
// > VARIABLES
// >   • latestAO   -- Latest AO, as reported by our LocalView.
// >   • proposedLI -- Max LI for which ⟨NextLI, li, xAO⟩ is already sent.
// >   • agreedLI   -- LI proposed for the consensus.
// >   • minLI      -- Do not participate in LI lower than this.
// >   • maxPeerLIs -- Maximal LIs and their AOs received from peers.
// >
// > ON Init(persistedLI):
// >   latestAO ← ⊥
// >   proposedLI, agreedLI ← 0
// >   minLI ← persistedLI + 1
// >   maxPeerLIs ← ∅
// >
// > UPON Reception of ConsensusOutput<DONE|SKIP>(consensusLI, nextAO):
// >   latestAO ← nextAO
// >   TryPropose(consensusLI + 1)
// >
// > UPON Reception of ConsensusTimeout(consensusLI):
// >   TryPropose(consensusLI + 1)
// >
// > UPON Reception of L1ReplacedBaseAliasOutput(nextAO):
// >   IF nextAO was not agreed for LI > ConsensusOutputDONE THEN
// >     latestAO ← nextAO
// >     TryPropose(max(agreedLI+1, EnoughVotes(agreedLI, N-F)+1, EnoughVotes(agreedLI, F+1), minLI))
// >
// > UPON Reception ⟨NextLI, li, ao⟩ from peer p:
// >   IF maxPeerLIs[p].li < li THEN
// >     maxPeerLIs[p] = ⟨li, ao⟩
// >   IF sli = EnoughVotes(proposedLI, F+1) ∧ sli ≥ minLI THEN
// >     TryPropose(sli)
// >   IF ali = EnoughVotes(agreedLI, N-F) ∧ ali ≥ minLI ∧ DerivedAO(ali) ≠ ⊥ THEN
// >     agreedLI ← ali
// >     OUTPUT(ali, DerivedAO(ali))
// >
// > FUNCTION TryPropose(li)
// >   IF proposedLI < li ∧ DerivedAO(li) ≠ ⊥ THEN
// >     proposedLI ← li
// >     Send ⟨NextLI, proposedLI, DerivedAO(li)⟩
// >
// > FUNCTION DerivedAO(li)
// >   RETURN IF ∃! ao: ∃(F+1) ⟨NextLI, ≥li, ao⟩
// >            THEN ao       // Can't be ⊥.
// >            ELSE latestAO // Can be ⊥, thus no derived AO
// >
// > FUNCTION EnoughVotes(aboveLI, quorum)
// >   IF ∃(max) x > aboveLI: ∃(quorum) j: maxPeerLIs[j].li ≥ x
// >     THEN RETURN x
// >     ELSE RETURN 0
// >
//
// Additionally, to recover after a reboot faster, a ⟨NextLI, -, -⟩ message includes the pleaseRepeat flag,
// which is sent while a node has not heard from a peer a message. The peer should resend its last message, if any.
//
// Moreover, we have to cope with the cases, when our latest AO is not known (is nil). That can happen on a startup,
// when AO is not yet received from the L1, and in the case of rejections, when several AOs are pending and some of them got reverted.
// In that case (the latestAO=⊥) the node should not propose to increase the LI, but should support proposals by others.
// As a consequence, it might happen that this variable will provide output even without getting an input from this node.
type varLogIndexImpl struct {
	nodeIDs    []gpa.NodeID                        // All the peers in this committee.
	n          int                                 // Total number of nodes.
	f          int                                 // Maximal number of faulty nodes to tolerate.
	latestAO   *isc.AliasOutputWithID              // Latest known AO, as reported by the varLocalView, can be nil (means we suspend proposing LIs).
	proposedLI LogIndex                            // Highest LI send by this node with a ⟨NextLI, li⟩ message.
	agreedLI   LogIndex                            // LI for which we have N-F proposals (when reached, consensus starts, the LI is persisted).
	minLI      LogIndex                            // Minimal LI at which this node can participate (set on boot).
	maxPeerLIs map[gpa.NodeID]*msgNextLogIndex     // Latest peer indexes received from peers.
	consAggrAO map[LogIndex]*isc.AliasOutputWithID // Recent outputs, to filter L1ReplacedBaseAliasOutput.
	consDoneLI LogIndex                            // Highest LI for which consensus has completed at this node.
	outputCB   func(li LogIndex, ao *isc.AliasOutputWithID)
	lastMsgs   map[gpa.NodeID]*msgNextLogIndex
	log        *logger.Logger
}

// > ON Init(persistedLI):
// >   latestAO ← ⊥
// >   proposedLI, agreedLI ← 0
// >   minLI ← persistedLI + 1
// >   maxPeerLIs ← ∅
func NewVarLogIndex(
	nodeIDs []gpa.NodeID,
	n int,
	f int,
	persistedLI LogIndex,
	outputCB func(li LogIndex, ao *isc.AliasOutputWithID),
	log *logger.Logger,
) VarLogIndex {
	log.Debugf("NewVarLogIndex, n=%v, f=%v, persistedLI=%v", n, f, persistedLI)
	return &varLogIndexImpl{
		nodeIDs:    nodeIDs,
		n:          n,
		f:          f,
		latestAO:   nil,
		proposedLI: NilLogIndex(),
		agreedLI:   NilLogIndex(),
		minLI:      persistedLI.Next(),
		maxPeerLIs: map[gpa.NodeID]*msgNextLogIndex{},
		consAggrAO: map[LogIndex]*isc.AliasOutputWithID{},
		consDoneLI: NilLogIndex(),
		outputCB:   outputCB,
		lastMsgs:   map[gpa.NodeID]*msgNextLogIndex{},
		log:        log,
	}
}

func (v *varLogIndexImpl) StatusString() string {
	return fmt.Sprintf(
		"{varLogIndex: proposedLI=%v, agreedLI=%v, consDoneLI=%v, minLI=%v}",
		v.proposedLI, v.agreedLI, v.consDoneLI, v.minLI,
	)
}

func (v *varLogIndexImpl) Value() (LogIndex, *isc.AliasOutputWithID) {
	if ao, ok := v.consAggrAO[v.agreedLI]; ok {
		return v.agreedLI, ao
	}
	return NilLogIndex(), nil
}

// > UPON Reception of ConsensusOutput<DONE|SKIP>(consensusLI, nextAO):
// >   latestAO ← nextAO
// >   TryPropose(consensusLI + 1)
func (v *varLogIndexImpl) ConsensusOutputReceived(consensusLI LogIndex, consensusStatus cons.OutputStatus, nextBaseAO *isc.AliasOutputWithID) gpa.OutMessages {
	v.log.Debugf("ConsensusOutputReceived: consensusLI=%v, nextBaseAO=%v", consensusLI, nextBaseAO)
	if consensusLI < v.agreedLI {
		return nil
	}
	v.latestAO = nextBaseAO // We can set nil here, means we don't know the last AO from our L1.
	if consensusStatus == cons.Completed && consensusLI > v.consDoneLI {
		// Cleanup information before the successful consensus.
		v.consDoneLI = consensusLI
		for li := range v.consAggrAO {
			if li < v.consDoneLI {
				delete(v.consAggrAO, li)
			}
		}
	}
	return v.tryPropose(consensusLI.Next())
}

// > UPON Reception of ConsensusTimeout(consensusLI):
// >   TryPropose(consensusLI + 1)
func (v *varLogIndexImpl) ConsensusTimeoutReceived(consensusLI LogIndex) gpa.OutMessages {
	v.log.Debugf("ConsensusTimeoutReceived: consensusLI=%v", consensusLI)
	if consensusLI < v.agreedLI {
		return nil
	}
	// NOTE: v.latestAO remains the same.
	return v.tryPropose(consensusLI.Next())
}

// > UPON Reception of L1ReplacedBaseAliasOutput(nextAO):
// >   IF nextAO was not agreed for LI > ConsensusOutputDONE THEN
// >     latestAO ← nextAO
// >     TryPropose(max(agreedLI+1, EnoughVotes(agreedLI, N-F)+1, EnoughVotes(agreedLI, F+1), minLI))
func (v *varLogIndexImpl) L1ReplacedBaseAliasOutput(nextBaseAO *isc.AliasOutputWithID) gpa.OutMessages {
	v.log.Debugf("L1ReplacedBaseAliasOutput, nextBaseAO=%v", nextBaseAO)
	if nextBaseAO != nil && v.wasRecentlyAgreed(nextBaseAO) {
		v.log.Debugf("skipping, wasRecentlyAgreed: %v", nextBaseAO)
		return nil
	}
	v.latestAO = nextBaseAO // We can set nil here, means we don't know the last AO from our L1.
	return v.tryPropose(MaxLogIndex(
		v.agreedLI.Next(),                         // Either propose next.
		v.enoughVotes(v.agreedLI, v.n-v.f).Next(), // Or we have skipped some round, and now we propose to go to next.
		v.enoughVotes(v.agreedLI, v.f+1),          // Or support the exiting, maybe we had no latestAO before.
		v.minLI,                                   // And, LI is not smaller than the minimal.
	))
}

// > UPON Reception ⟨NextLI, li, ao⟩ from peer p:
// >   IF maxPeerLIs[p].li < li THEN
// >     maxPeerLIs[p] = ⟨li, ao⟩
// >   IF sli = EnoughVotes(proposedLI, F+1) ∧ sli ≥ minLI THEN
// >     TryPropose(sli)
// >   IF ali = EnoughVotes(agreedLI, N-F) ∧ ali ≥ minLI ∧ DerivedAO(ali) ≠ ⊥ THEN
// >     agreedLI ← ali
// >     OUTPUT(ali, DerivedAO(ali))
func (v *varLogIndexImpl) MsgNextLogIndexReceived(msg *msgNextLogIndex) gpa.OutMessages {
	v.log.Debugf("MsgNextLogIndexReceived, %v", msg)
	sender := msg.Sender()
	//
	// Validate and record the vote.
	if !v.knownNodeID(sender) {
		v.log.Warnf("MsgNextLogIndex from unknown sender: %+v", msg)
		return nil
	}
	msgs := gpa.NoMessages()
	if lastMsg, ok := v.lastMsgs[msg.Sender()]; ok && msg.pleaseRepeat {
		msgs.Add(lastMsg.AsResent())
	}
	var prevPeerLI LogIndex
	if prevPeerNLI, ok := v.maxPeerLIs[sender]; ok {
		prevPeerLI = prevPeerNLI.nextLogIndex
	} else {
		prevPeerLI = NilLogIndex()
	}
	if prevPeerLI.AsUint32() >= msg.nextLogIndex.AsUint32() {
		return msgs
	}
	v.maxPeerLIs[sender] = msg
	if sli := v.enoughVotes(v.proposedLI, v.f+1); sli >= v.minLI {
		msgs.AddAll(v.tryPropose(sli))
	}
	if ali := v.enoughVotes(v.agreedLI, v.n-v.f); ali >= v.minLI {
		if derivedAO := v.deriveAO(ali); derivedAO != nil {
			v.agreedLI = ali
			v.consAggrAO[v.agreedLI] = derivedAO
			v.log.Debugf("Output, agreedLI=%v, derivedAO=%v", v.agreedLI, derivedAO)
			v.outputCB(v.agreedLI, derivedAO)
		}
	}
	return msgs
}

// > FUNCTION TryPropose(li)
// >   IF proposedLI < li ∧ DerivedAO(li) ≠ ⊥ THEN
// >     proposedLI ← li
// >     Send ⟨NextLI, proposedLI, DerivedAO(li)⟩
func (v *varLogIndexImpl) tryPropose(li LogIndex) gpa.OutMessages {
	v.log.Debugf("tryPropose: li=%v", li)
	if v.proposedLI >= li {
		v.log.Debugf("tryPropose: skip, v.proposedLI=%v >= li=%v", v.proposedLI, li)
		return nil
	}
	derivedAO := v.deriveAO(li)
	if derivedAO == nil {
		v.log.Debugf("tryPropose: skip, derivedAO=%v, v.latestAO=%v", derivedAO, v.latestAO)
		return nil
	}
	v.proposedLI = li
	v.log.Debugf("Sending NextLogIndex=%v, baseAO=%v", v.proposedLI, derivedAO)
	msgs := gpa.NoMessages()
	for _, nodeID := range v.nodeIDs {
		_, haveMsgFrom := v.maxPeerLIs[nodeID] // It might happen, that we rebooted and lost the state.
		msg := newMsgNextLogIndex(nodeID, v.proposedLI, derivedAO, !haveMsgFrom)
		v.lastMsgs[nodeID] = msg
		msgs.Add(msg)
	}
	return msgs
}

// > FUNCTION DerivedAO(li)
// >   RETURN IF ∃! ao: ∃(F+1) ⟨NextLI, ≥li, ao⟩
// >            THEN ao       // Can't be ⊥.
// >            ELSE latestAO // Can be ⊥, thus no derived AO
func (v *varLogIndexImpl) deriveAO(li LogIndex) *isc.AliasOutputWithID {
	countsAOMap := map[iotago.OutputID]*isc.AliasOutputWithID{}
	countsAO := map[iotago.OutputID]int{}
	for _, msg := range v.maxPeerLIs {
		if msg.nextLogIndex < li {
			continue
		}
		countsAOMap[msg.nextBaseAO.OutputID()] = msg.nextBaseAO
		countsAO[msg.nextBaseAO.OutputID()]++
	}

	var q1fAO *isc.AliasOutputWithID
	var found bool
	for aoID, c := range countsAO {
		if c >= v.f+1 {
			if found {
				// Non unique, return our value.
				return v.latestAO
			}
			q1fAO = countsAOMap[aoID]
			found = true
		}
	}
	if found {
		return q1fAO
	}
	return v.latestAO
}

// Find highest LogIndex for which N-F nodes have voted.
// Returns 0, if not found.
// > FUNCTION EnoughVotes(aboveLI, quorum)
// >   IF ∃(max) x > aboveLI: ∃(quorum) j: maxPeerLIs[j].li ≥ x
// >     THEN RETURN x
// >     ELSE RETURN 0
func (v *varLogIndexImpl) enoughVotes(aboveLI LogIndex, quorum int) LogIndex {
	countsLI := map[LogIndex]int{}
	for _, msg := range v.maxPeerLIs {
		if msg.nextLogIndex > aboveLI {
			countsLI[msg.nextLogIndex]++
		}
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

func (v *varLogIndexImpl) wasRecentlyAgreed(ao *isc.AliasOutputWithID) bool {
	for _, liAO := range v.consAggrAO {
		if ao.Equals(liAO) {
			return true
		}
	}
	return false
}

func (v *varLogIndexImpl) knownNodeID(nodeID gpa.NodeID) bool {
	return lo.Contains(v.nodeIDs, nodeID)
}
