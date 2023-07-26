package cmt_log

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain/cons"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

// TODO: Delay L1 replace events -- to decrease probability of need to wait for N-F ⟨NextOnBoth, ...⟩ events.
// The delay should be implemented in the VarLocalView, to filter-out the AOs that are part of the chain.
//
// The events causing the log index to advance are of the following categories.
//
// NextLI/Recover (Q=F+1):
//   - First L1 AO received after a boot for this committee.
//   - Recover event from the consensus (on a timeout).
//
// NextLI/ConsDone (Q=N-F):
//   - Consensus has completed.
//
// NextLI/L1Replace (Q=N-F):
//   - A reorg in L1 (impossible on Hornet).
//   - Other nodes posted AO, and now it is confirmed (race condition with ConsDone).
//   - Pipelined chain was rejected, and now we go back to some older AO (// TODO: Have to considered, this happens)
//   - Chain was rotated back to this committee (if an AO with other committee was seen, this will come as a recovery case).
//
// Here is a list of rules for sending all kinds of events.
//
// TODO: Incorrect, see the markdown. // Rules for the Recovery events:
//   - If a node enters a recovery mode (First L1 AO or ConsRecover), it sends NextLI/Recovery message.
//     Later it can support other recovery messages, if they carry higher LI.
//   - If any node receives NextLI/Recovery (with any LI), it sends own NextLI/Recovery with Max(Q1F/Recovery, proposedLI).
//     This way lagging nodes will get in sync with the working nodes.
//   - We have to make the recovery finite to avoid interference with the normal operation.
//     The recovery mode ends by receiving ConsensusOutput<DONE|SKIP> event.
//
// Rules for the ConsOut messages:
//   - If a node receives ConsensusOutput, it sends NextLi/ConsOut with LI=ConsensusOutput.LI+1
//     if such message was not sent before with LI >= ConsensusOutput.LI+1.
//
// Rules for the L1Replace events:
//   - // TODO: Maybe we can ignore them for now
//
// > MESSAGES
// >   • ⟨NextOnRecover, li, ao⟩ -- Sent on node boot or consensus recover, supported on Q=F+1, decided on Q=N-F.
// >   • ⟨NextOnL2Cons,  li, ao⟩ -- Sent when the previous consensus is done, decided on Q=N-F.
// >   • ⟨NextOnL1Conf,  li, ao⟩ -- Sent when an AO is replaced by the L1 network, decided on Q=N-F.
// >   • ⟨NextOnBoth,    li, ao⟩ -- Sent when an AO is replaced by the L1 network AND consensus is done, decided on Q=N-F.
// >
// > VARIABLES
// >   • ..
// >
// > ON Init(persistedLI):
// >   latestAO ← ⊥
// >   proposedLI, agreedLI ← 0
// >   minLI ← persistedLI + 1
// >   maxPeerLIs ← ∅
// >   Send ⟨NextOnRecover, minLI⟩
// >
// > UPON Reception of ConsensusOutput<DONE|SKIP>(consensusLI, nextAO):
// >   latestAO ← nextAO
// >   TryPropose(consensusLI + 1)
// >
// > UPON Reception of ConsensusTimeout(consensusLI):
// >   IF consensusLI > proposedLI THEN // TODO: Consider recoverLI
// >       Send ⟨NextLI/Recover, consensusLI+1, latestAO⟩
// >
// > UPON Reception of L1ReplacedBaseAliasOutput(nextAO):
// >   IF isFirstAO and no ConsDone quorum THEN
// >       Send ⟨NextLI/Recover, minLI, nextAO⟩
// >   ELSE
// >       IGNORE(for now)
// >
// > UPON Reception ⟨NextLI/Recover, li, ao⟩ from peer p:
// >   IF recoverLIs[p].li < li THEN
// >     recoverLIs[p] = ⟨li, ao⟩
// >   LET recSupLI = EnoughVotes(recoverLIs, F+1)
// >   LET recAgrLI = EnoughVotes(recoverLIs, N-F)
// >   IF recSupLI > recoverLIs.proposedLI THEN
// >     recoverLIs.proposedLI = recSupLI
// >     Send ⟨NextLI, recSupLI, DerivedAO(li)⟩
// >   IF recAgrLI > agreedLI THEN
// >     agreedLI = recAgrLI
// >     OUTPUT (agreedLI, latestAO)
// >
// > FUNCTION TryPropose(li)
// >   IF proposedLI < li ∧ DerivedAO(li) ≠ ⊥ THEN
// >     proposedLI ← li
// >     Send ⟨NextLI, proposedLI, DerivedAO(li)⟩
// >
// > FUNCTION EnoughVotes(LIs, quorum)
// >   IF ∃(max) x: ∃(quorum) j: LIs[j].li ≥ x
// >     THEN RETURN x
// >     ELSE RETURN 0
// >
type varLogIndexImplV2 struct {
	nodeIDs        []gpa.NodeID                    // All the peers in this committee.
	n              int                             // Total number of nodes.
	f              int                             // Maximal number of faulty nodes to tolerate.
	minLI          LogIndex                        // Minimal LI at which this node can participate (set on boot).
	proposedLI     LogIndex                        // Highest LI send by this node with a ⟨NextLI, li⟩ message.
	agreedLI       LogIndex                        // LI for which we have N-F proposals (when reached, consensus starts, the LI is persisted).
	agreedAO       *isc.AliasOutputWithID          // AO agreed for agreedLI.
	latestAO       *isc.AliasOutputWithID          // Latest known AO, as reported by the varLocalView, can be nil (means we suspend proposing LIs).
	lastMsgs       map[gpa.NodeID]*MsgNextLogIndex // Latest messages we have sent to other peers.
	qcConsOut      *QuorumCounter
	qcL1AOReplaced *QuorumCounter
	qcRecover      *QuorumCounter
	outputCB       func(li LogIndex, ao *isc.AliasOutputWithID)
	log            *logger.Logger
}

func NewVarLogIndexV2(
	nodeIDs []gpa.NodeID,
	n int,
	f int,
	persistedLI LogIndex,
	outputCB func(li LogIndex, ao *isc.AliasOutputWithID),
	deriveAOByQuorum bool,
	log *logger.Logger,
) VarLogIndex {
	vli := &varLogIndexImplV2{
		nodeIDs:        nodeIDs,
		n:              n,
		f:              f,
		minLI:          persistedLI.Next(),
		proposedLI:     NilLogIndex(),
		agreedLI:       NilLogIndex(),
		agreedAO:       nil,
		latestAO:       nil, // TODO: Update it properly.
		lastMsgs:       map[gpa.NodeID]*MsgNextLogIndex{},
		qcConsOut:      NewQuorumCounter(MsgNextLogIndexCauseConsOut, nodeIDs, log),
		qcL1AOReplaced: NewQuorumCounter(MsgNextLogIndexCauseL1ReplacedAO, nodeIDs, log),
		qcRecover:      NewQuorumCounter(MsgNextLogIndexCauseRecover, nodeIDs, log),
		outputCB:       outputCB,
		log:            log,
	}
	return vli
}

func (vli *varLogIndexImplV2) StatusString() string {
	return fmt.Sprintf(
		"{varLogIndexV2: proposedLI=%v, agreedLI=%v, agreedAO=%v, minLI=%v}",
		vli.proposedLI, vli.agreedLI, vli.agreedAO, vli.minLI,
	)
}

func (vli *varLogIndexImplV2) Value() (LogIndex, *isc.AliasOutputWithID) {
	if vli.agreedAO == nil {
		return NilLogIndex(), nil
	}
	return vli.agreedLI, vli.agreedAO
}

// > UPON Reception of ConsensusOutput<DONE|SKIP>(consensusLI, nextAO):
func (vli *varLogIndexImplV2) ConsensusOutputReceived(consensusLI LogIndex, consensusStatus cons.OutputStatus, nextBaseAO *isc.AliasOutputWithID) gpa.OutMessages {
	vli.log.Debugf("ConsensusOutputReceived: consensusLI=%v, nextBaseAO=%v", consensusLI, nextBaseAO)
	vli.tryOutputWithAO(nextBaseAO)
	msgs := vli.qcConsOut.MaybeSendVote(consensusLI.Next(), nil)
	vli.tryOutputOnConsOut()
	return msgs
}

// > UPON Reception of ConsensusTimeout(consensusLI):
func (vli *varLogIndexImplV2) ConsensusRecoverReceived(consensusLI LogIndex) gpa.OutMessages {
	vli.log.Debugf("ConsensusRecoverReceived: consensusLI=%v", consensusLI)
	return vli.qcRecover.MaybeSendVote(consensusLI.Next(), nil)
}

// > UPON Reception of L1ReplacedBaseAliasOutput(nextAO):
func (vli *varLogIndexImplV2) L1ReplacedBaseAliasOutput(nextBaseAO *isc.AliasOutputWithID) gpa.OutMessages {
	vli.log.Debugf("L1ReplacedBaseAliasOutput, nextBaseAO=%v", nextBaseAO)
	msgs := gpa.NoMessages()
	//
	// If we tried to output the LI/AO, but the AO was unknown/nil, complete the output with this latest AO.
	vli.tryOutputWithAO(nextBaseAO)
	//
	// If an initial recover event was not sent after a boot yet, send it.
	if recoverLI, _ := vli.qcRecover.MyLastVote(); recoverLI.IsNil() {
		msgs.AddAll(vli.qcRecover.MaybeSendVote(vli.minLI, nil))
	}

	//
	//
	myConsOutLI, myConsOutAO := vli.qcConsOut.MyLastVote()
	if nextBaseAO.Equals(myConsOutAO) {
		// This is the race condition with the AO. Don't increase the AO.
		msgs.AddAll(vli.qcRecover.MaybeSendVote(myConsOutLI, nil))
	} else {
		// It should be the actual change from the L1.
		msgs.AddAll(vli.qcRecover.MaybeSendVote(myConsOutLI.Next(), nil))
	}

	return msgs
}

// > UPON Reception ⟨NextLI, li, ao⟩ from peer p:
func (vli *varLogIndexImplV2) MsgNextLogIndexReceived(msg *MsgNextLogIndex) gpa.OutMessages {
	vli.log.Debugf("MsgNextLogIndexReceived, %v", msg)
	sender := msg.Sender()
	if !vli.knownNodeID(sender) {
		vli.log.Warnf("⊢ MsgNextLogIndex from unknown sender: %+v", msg)
		return nil
	}

	switch msg.Cause {
	case MsgNextLogIndexCauseConsOut:
		vli.msgNextLogIndexOnConsOut(msg)
		return nil
	case MsgNextLogIndexCauseL1ReplacedAO:
		return vli.msgNextLogIndexOnL1ReplacedAO(msg)
	case MsgNextLogIndexCauseRecover:
		return vli.msgNextLogIndexOnRecover(msg)
	default:
		vli.log.Warnf("⊢ MsgNextLogIndex with unexpected cause: %+v", msg)
		return nil
	}
}

// > UPON Reception ⟨NextLI/ConsOut, li, ao⟩ from peer p:
// >   TODO: Pseudo-code.
func (vli *varLogIndexImplV2) msgNextLogIndexOnConsOut(msg *MsgNextLogIndex) {
	vli.qcConsOut.VoteReceived(msg)
	vli.tryOutputOnConsOut()
}

// > UPON Reception ⟨NextLI/Recover, li, ao⟩ from peer p:
// >   IF recoverLIs[p].li < li THEN
// >     recoverLIs[p] = ⟨li, ao⟩
// >   LET recSupportLI = EnoughVotes(recoverLIs, F+1)
// >   LET recAgreedLI = EnoughVotes(recoverLIs, N-F)
// >   IF recSupportLI > recoverLIs.proposedLI THEN
// >     recoverLIs.proposedLI = recSupportLI
// >     Send ⟨NextLI, recSupportLI, DerivedAO(li)⟩
// >   IF recAgreedLI > agreedLI THEN
// >     agreedLI = recAgreedLI
// >     OUTPUT (agreedLI, latestAO)
func (vli *varLogIndexImplV2) msgNextLogIndexOnRecover(msg *MsgNextLogIndex) gpa.OutMessages {
	vli.qcRecover.VoteReceived(msg)
	sli, _ := vli.qcRecover.EnoughVotes(vli.f+1, false)
	ali, _ := vli.qcRecover.EnoughVotes(vli.n-vli.f, false)
	msgs := vli.qcRecover.MaybeSendVote(sli, nil)
	vli.tryOutput(ali, vli.latestAO)
	return msgs
}

func (vli *varLogIndexImplV2) msgNextLogIndexOnL1ReplacedAO(msg *MsgNextLogIndex) gpa.OutMessages {
	panic("todo")
}

// If we voted for that LI based on consensus output, and there is N-F supporters, then proceed.
// The N-F is counted including this node. That can be optimized, but lets keep it simpler for now.
func (vli *varLogIndexImplV2) tryOutputOnConsOut() {
	agreedConsOutLI, _ := vli.qcConsOut.EnoughVotes(vli.n-vli.f, false)
	myConsOutLI, myConsOutAO := vli.qcConsOut.MyLastVote()
	if agreedConsOutLI >= myConsOutLI {
		vli.tryOutput(myConsOutLI, myConsOutAO)
	}
}

// That's output for the consensus. We will start consensus instances with strictly increasing LIs with non-nil AOs.
func (vli *varLogIndexImplV2) tryOutput(li LogIndex, ao *isc.AliasOutputWithID) {
	if li <= vli.agreedLI || li < vli.minLI {
		return
	}
	vli.agreedLI = li
	vli.agreedAO = ao
	if ao != nil {
		vli.log.Debugf("⊢ Output, li=%v, ao=%v", vli.agreedLI, vli.agreedAO)
		vli.outputCB(vli.agreedLI, vli.agreedAO)
	}
}

// It is possible in tryOutput that AO is not known, so we will output when the latest AO becomes known.
func (vli *varLogIndexImplV2) tryOutputWithAO(ao *isc.AliasOutputWithID) {
	if vli.agreedLI.IsNil() || vli.agreedAO != nil || ao == nil {
		return
	}
	vli.agreedAO = ao
	vli.log.Debugf("⊢ Output, li=%v, ao=%v", vli.agreedLI, vli.agreedAO)
	vli.outputCB(vli.agreedLI, vli.agreedAO)
}

func (vli *varLogIndexImplV2) knownNodeID(nodeID gpa.NodeID) bool {
	return lo.Contains(vli.nodeIDs, nodeID)
}
