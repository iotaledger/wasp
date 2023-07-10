package cmt_log

import (
	"fmt"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain/cons"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/samber/lo"
)

// TODO: Delay L1 replace events -- to decrease probability of need to wait for N-F ⟨NextOnBoth, ...⟩ events.
// The delay should be implemented in the VarLocalView, to filter-out the AOs that are part of the chain.
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
// >
// > UPON Reception of L1ReplacedBaseAliasOutput(nextAO):
// >
// > UPON Reception ⟨NextLI, li, ao⟩ from peer p:
// >
// > FUNCTION TryPropose(li)
// >   IF proposedLI < li ∧ DerivedAO(li) ≠ ⊥ THEN
// >     proposedLI ← li
// >     Send ⟨NextLI, proposedLI, DerivedAO(li)⟩
// >
type varLogIndexImplV2 struct {
	nodeIDs         []gpa.NodeID                    // All the peers in this committee.
	n               int                             // Total number of nodes.
	f               int                             // Maximal number of faulty nodes to tolerate.
	minLI           LogIndex                        // Minimal LI at which this node can participate (set on boot).
	proposedLI      LogIndex                        // Highest LI send by this node with a ⟨NextLI, li⟩ message.
	agreedLI        LogIndex                        // LI for which we have N-F proposals (when reached, consensus starts, the LI is persisted).
	agreedAO        *isc.AliasOutputWithID          // AO agreed for agreedLI.
	lastMsgs        map[gpa.NodeID]*MsgNextLogIndex // Latest messages we have sent to other peers.
	qcConsCompleted *QuorumCounter
	qcL1AOReplaced  *QuorumCounter
	qcRecover       *QuorumCounter
	qcCommon        *QuorumCounter
	log             *logger.Logger
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
		nodeIDs:         nodeIDs,
		n:               n,
		f:               f,
		minLI:           persistedLI.Next(),
		proposedLI:      NilLogIndex(),
		agreedLI:        NilLogIndex(),
		agreedAO:        nil,
		lastMsgs:        map[gpa.NodeID]*MsgNextLogIndex{},
		qcConsCompleted: NewQuorumCounter(log),
		qcL1AOReplaced:  NewQuorumCounter(log),
		qcRecover:       NewQuorumCounter(log),
		qcCommon:        NewQuorumCounter(log),
		log:             log,
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
	nextLI := consensusLI.Next()
	if nextLI > vli.proposedLI {
		// TODO: Send.
	}
	if nextLI == vli.proposedLI {

	}
	// if consensusLI < v.agreedLI {
	// 	v.log.Debugf("⊢ Ignoring, received consensusLI=%v < agreedLI=%v", consensusLI, v.agreedLI)
	// 	return nil
	// }
	// v.latestAO = nextBaseAO // We can set nil here, means we don't know the last AO from our L1.
	// return v.tryPropose(consensusLI.Next())
	panic("todo") // TODO: Implement.
}

// > UPON Reception of ConsensusTimeout(consensusLI):
func (vli *varLogIndexImplV2) ConsensusRecoverReceived(consensusLI LogIndex) gpa.OutMessages {
	vli.log.Debugf("ConsensusRecoverReceived: consensusLI=%v", consensusLI)
	// if consensusLI < v.agreedLI {
	// 	return nil
	// }
	// // NOTE: v.latestAO remains the same.
	// return v.tryPropose(consensusLI.Next())
	panic("todo") // TODO: Implement.
}

// > UPON Reception of L1ReplacedBaseAliasOutput(nextAO):
func (vli *varLogIndexImplV2) L1ReplacedBaseAliasOutput(nextBaseAO *isc.AliasOutputWithID) gpa.OutMessages {
	vli.log.Debugf("L1ReplacedBaseAliasOutput, nextBaseAO=%v", nextBaseAO)

	// v.latestAO = nextBaseAO // We can set nil here, means we don't know the last AO from our L1.
	// return v.tryPropose(MaxLogIndex(
	// 	v.agreedLI.Next(),                         // Either propose next.
	// 	v.enoughVotes(v.agreedLI, v.n-v.f).Next(), // Or we have skipped some round, and now we propose to go to next.
	// 	v.enoughVotes(v.agreedLI, v.f+1),          // Or support the exiting, maybe we had no latestAO before.
	// 	v.minLI,                                   // And, LI is not smaller than the minimal.
	// ))
	panic("todo") // TODO: Implement.
}

// > UPON Reception ⟨NextLI, li, ao⟩ from peer p:
func (vli *varLogIndexImplV2) MsgNextLogIndexReceived(msg *MsgNextLogIndex) gpa.OutMessages {
	vli.log.Debugf("MsgNextLogIndexReceived, %v", msg)
	sender := msg.Sender()
	//
	// Validate and record the vote.
	if !vli.knownNodeID(sender) {
		vli.log.Warnf("⊢ MsgNextLogIndex from unknown sender: %+v", msg)
		return nil
	}

	var qcForCause *QuorumCounter
	switch msg.Cause {
	case MsgNextLogIndexCauseConsCompleted:
		qcForCause = vli.qcConsCompleted
	case MsgNextLogIndexCauseL1ReplacedAO:
		qcForCause = vli.qcL1AOReplaced
	case MsgNextLogIndexCauseRecover:
		qcForCause = vli.qcRecover
	default:
		vli.log.Warnf("⊢ MsgNextLogIndex with unexpected cause: %+v", msg)
		return nil
	}

	msgs := gpa.NoMessages()
	if lastMsg, ok := vli.lastMsgs[msg.Sender()]; ok && msg.PleaseRepeat {
		msgs.Add(lastMsg.AsResent())
	}

	qcForCause.VoteReceived(msg)
	vli.qcCommon.VoteReceived(msg)

	if msg.Cause == MsgNextLogIndexCauseRecover {
		if sli := qcForCause.EnoughVotes(vli.f+1, false); sli >= vli.minLI && sli >= vli.proposedLI {
			// In the case of recover, we have to support other nodes on low threshold.
			msgs.AddAll(vli.tryPropose(sli, MsgNextLogIndexCauseRecover))
		}
		if ali := qcForCause.EnoughVotes(vli.n-vli.f, false); ali >= vli.minLI && ali >= vli.agreedLI {

		}
	} else {
		if ali := qcForCause.EnoughVotes(vli.n-vli.f, true); ali >= vli.minLI && ali >= vli.agreedLI {

		}
	}

	// if ali := vli.enoughVotes(vli.agreedLI, vli.n-vli.f); ali >= vli.minLI {
	// 	if derivedAO := vli.deriveAO(ali); derivedAO != nil {
	// 		vli.agreedLI = ali
	// 		vli.agreedAO = derivedAO
	// 		vli.log.Debugf("⊢ Output, agreedLI=%v, derivedAO=%v", vli.agreedLI, derivedAO)
	// 		vli.outputCB(vli.agreedLI, derivedAO)
	// 	}
	// }
	return msgs
}

func (vli *varLogIndexImplV2) tryPropose(li LogIndex, cause MsgNextLogIndexCause) gpa.OutMessages {
	panic("todo") // TODO: ...
}

func (vli *varLogIndexImplV2) knownNodeID(nodeID gpa.NodeID) bool {
	return lo.Contains(vli.nodeIDs, nodeID)
}
