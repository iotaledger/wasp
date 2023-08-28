package cmt_log

// CASE: On the quorum to start.
// Assume 3/4 committee, nodes A, B, C, D.
// Assume node D send NextLI/ConsDone(y) to A and B and crashed; node C is lagging several instances back (x < y).
// A and B received NextLI/ConsDone(y) from A, B and D, thus proceed to the LogIndex=y, but ACS is stuck, because 3 nodes are needed.
// Node C cannot support A and B in LI=y because, C cannot proceed with LI=x, because other nodes already dropped that old consensus instance.
// Node D cannot support A and B, because its minLI=y and it asks to recover to LI=y+1.
// As a consequence, we need to introduce NextLI/Started with a quorum F+1. Each node sends it after the consensus is started on that node.
// This way any other node knows there exist at least 1 correct node started the consensus.
//   ==> This implies it received NextLI/ConsDone (or others) from N-F nodes, thus at least F+1 correct nodes.
//   ##> We need at least F+1 proposals with the new AO to avoid producing duplicate TXes.
//
// CASE: On the mempool and missing requests.
// Assume 3/4 committee, nodes A, B, C, D.
// Nodes A and B completed the consensus, node C is lagging several instances back, node D died during the consensus, after ACS is done.
// A and B cannot proceed to the next LI, because they have NextLI from 2 nodes only (A and B).
// D is still down.
// C cannot complete the consensus, because it has OnLedger request missing. It cannot be shared between the nodes.
//   ??> WHAT TO DO?
//   ??> Don't count L1RepAO and ConsOut separately. That's not needed anymore with the current approach I guess.
//   ==> Nah, Just send NextLI/L1RepAO for known LIs as well.
//
// TODO CASE: After a recovery, a node joins a round for which the OnLedger request is already consumed.
// Assume 3/4 committee, nodes A, B, C, D.
// Nodes A, B and D are running consensus instance x, node C is lagging and was rebooted.
// Node D crashes before finishing the consensus, but A and B manage to complete it and publish the TX.
// Node C received AO of that TX, and recovers to LI=X with that AO as input.
// In the end, C cannot proceed in LI=X, because don't have an OnLedger request decided by ACS, and consumed with the TX.
// Nodes A and B cannot proceed to LI=X+1, because N-F=3 quorum is needed, but only 2 votes are received.
//   ??> If consensus InputAO=OutputAO, then proceed to the next? How general is that?
//   ??> In general, it is useless to try to complete LI=X, because L1 inputs are already consumed.
//   !!> Too much lagging nodes are effectively faulty nodes. That makes the adversary more powerful, if they are fast.
//   ==> ?
//

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/metrics"
)

type VarLogIndex interface {
	// Summary of the internal state.
	StatusString() string

	// Returns the latest agreed LI.
	// There is no output value, if LogIndex=⊥.
	Value() LogIndex

	// Mark the log index as used (consensus initiated for it already) so
	// that it would not be proposed anymore in future.
	LogIndexUsed(li LogIndex)

	// Consensus terminated with either with DONE or SKIP.
	// The logIndex is of the consensus that has been completed.
	ConsensusOutputReceived(consensusLI LogIndex) gpa.OutMessages

	// Consensus decided, that its maybe time to attempt another run.
	// The timeout-ed consensus will be still running, so they will race for the result.
	ConsensusRecoverReceived(consensusLI LogIndex) gpa.OutMessages

	// This is called when we have to move to the next log index based on the AO received from L1.
	L1ReplacedBaseAliasOutput() gpa.OutMessages

	// This is called, if an AO is confirmed for which we know a log index (was pending).
	L1ConfirmedAliasOutput(li LogIndex) gpa.OutMessages

	// Messages are exchanged, so this function handles them.
	MsgNextLogIndexReceived(msg *MsgNextLogIndex) gpa.OutMessages
}

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
// > MESSAGES // TODO: ...
// >   • ⟨NextOnRecover, li, ao⟩ -- Sent on node boot or consensus recover, supported on Q=F+1, decided on Q=N-F.
// >   • ⟨NextOnL2Cons,  li, ao⟩ -- Sent when the previous consensus is done, decided on Q=N-F.
// >   • ⟨NextOnL1Conf,  li, ao⟩ -- Sent when an AO is replaced by the L1 network, decided on Q=N-F.
// >   • ⟨NextOnBoth,    li, ao⟩ -- Sent when an AO is replaced by the L1 network AND consensus is done, decided on Q=N-F.
// >
// > VARIABLES
// >   • ..
// >
// > ON Init(persistedLI):
// >   // TODO: ...
// >
// > UPON Reception of ConsensusOutput<DONE|SKIP>(consensusLI, nextAO):
// >   // TODO: ...
// >
// > UPON Reception of ConsensusTimeout(consensusLI):
// >   // TODO: ...
// >
// > UPON Reception of L1ReplacedBaseAliasOutput(nextAO):
// >   // TODO: ...
// >
type varLogIndexImpl struct {
	nodeIDs   []gpa.NodeID                    // All the peers in this committee.
	n         int                             // Total number of nodes.
	f         int                             // Maximal number of faulty nodes to tolerate.
	minLI     LogIndex                        // Minimal LI at which this node can participate (set on boot).
	agreedLI  LogIndex                        // LI for which we have N-F proposals (when reached, consensus starts, the LI is persisted).
	lastMsgs  map[gpa.NodeID]*MsgNextLogIndex // Latest messages we have sent to other peers.
	qcConsOut *QuorumCounter
	qcL1AORep *QuorumCounter
	qcRecover *QuorumCounter
	qcStarted *QuorumCounter
	outputCB  func(li LogIndex)
	metrics   *metrics.ChainCmtLogMetrics
	log       *logger.Logger
}

func NewVarLogIndex(
	nodeIDs []gpa.NodeID,
	n int,
	f int,
	persistedLI LogIndex,
	outputCB func(li LogIndex),
	metrics *metrics.ChainCmtLogMetrics,
	log *logger.Logger,
) VarLogIndex {
	vli := &varLogIndexImpl{
		nodeIDs:   nodeIDs,
		n:         n,
		f:         f,
		minLI:     persistedLI.Next(),
		agreedLI:  NilLogIndex(),
		lastMsgs:  map[gpa.NodeID]*MsgNextLogIndex{},
		qcConsOut: NewQuorumCounter(MsgNextLogIndexCauseConsOut, nodeIDs, log),
		qcL1AORep: NewQuorumCounter(MsgNextLogIndexCauseL1RepAO, nodeIDs, log),
		qcRecover: NewQuorumCounter(MsgNextLogIndexCauseRecover, nodeIDs, log),
		qcStarted: NewQuorumCounter(MsgNextLogIndexCauseStarted, nodeIDs, log),
		outputCB:  outputCB,
		metrics:   metrics,
		log:       log,
	}
	return vli
}

func (vli *varLogIndexImpl) StatusString() string {
	return fmt.Sprintf(
		"{varLogIndex: minLI=%v, agreedLI=%v}",
		vli.minLI, vli.agreedLI,
	)
}

func (vli *varLogIndexImpl) Value() LogIndex {
	if vli.agreedLI < vli.minLI {
		return NilLogIndex()
	}
	return vli.agreedLI
}

func (vli *varLogIndexImpl) LogIndexUsed(li LogIndex) { // TODO: Call it. Or remove it.
	if vli.minLI <= li {
		vli.minLI = li.Next()
	}
}

func (vli *varLogIndexImpl) ConsensusOutputReceived(consensusLI LogIndex) gpa.OutMessages {
	vli.log.Debugf("ConsensusOutputReceived: consensusLI=%v", consensusLI)
	msgs := gpa.NoMessages()
	msgs.AddAll(vli.qcConsOut.MaybeSendVote(consensusLI.Next()))
	msgs.AddAll(vli.tryOutputOnConsOut())
	return msgs
}

func (vli *varLogIndexImpl) ConsensusRecoverReceived(consensusLI LogIndex) gpa.OutMessages {
	vli.log.Debugf("ConsensusRecoverReceived: consensusLI=%v", consensusLI)
	msgs := gpa.NoMessages()
	msgs.AddAll(vli.qcRecover.MaybeSendVote(consensusLI.Next()))
	msgs.AddAll(vli.tryOutputOnRecover())
	return msgs
}

func (vli *varLogIndexImpl) L1ReplacedBaseAliasOutput() gpa.OutMessages {
	vli.log.Debugf("L1ReplacedBaseAliasOutput")
	msgs := gpa.NoMessages()
	//
	// Send the boot time recovery, if it was not sent yet.
	if vli.qcRecover.MyLastVote().IsNil() {
		msgs.AddAll(vli.qcRecover.MaybeSendVote(vli.minLI))
	}
	//
	// Vote for the first non-agreed log index.
	voteForLI := vli.minLI
	if vli.agreedLI >= vli.minLI {
		voteForLI = vli.agreedLI.Next()
	}
	msgs.AddAll(vli.qcL1AORep.MaybeSendVote(voteForLI))
	//
	// Report an agreed LI, if any.
	msgs.AddAll(vli.tryOutputOnL1RepAO())
	return msgs
}

func (vli *varLogIndexImpl) L1ConfirmedAliasOutput(li LogIndex) gpa.OutMessages {
	vli.log.Debugf("L1ConfirmedAliasOutput")
	//
	// Vote for this LI, if have not voted for any higher.
	msgs := gpa.NoMessages()
	msgs.AddAll(vli.qcL1AORep.MaybeSendVote(li))
	//
	// Report an agreed LI, if any.
	msgs.AddAll(vli.tryOutputOnL1RepAO())
	return msgs
}

func (vli *varLogIndexImpl) MsgNextLogIndexReceived(msg *MsgNextLogIndex) gpa.OutMessages {
	vli.log.Debugf("MsgNextLogIndexReceived, %v", msg)
	sender := msg.Sender()
	if !vli.knownNodeID(sender) {
		vli.log.Warnf("⊢ MsgNextLogIndex from unknown sender: %+v", msg)
		return nil
	}

	switch msg.Cause {
	case MsgNextLogIndexCauseConsOut:
		return vli.msgNextLogIndexOnConsOut(msg)
	case MsgNextLogIndexCauseRecover:
		return vli.msgNextLogIndexOnRecover(msg)
	case MsgNextLogIndexCauseL1RepAO:
		return vli.msgNextLogIndexOnL1RepAO(msg)
	case MsgNextLogIndexCauseStarted:
		return vli.msgNextLogIndexOnStarted(msg)
	default:
		vli.log.Warnf("⊢ MsgNextLogIndex with unexpected cause: %+v", msg)
		return nil
	}
}

func (vli *varLogIndexImpl) msgNextLogIndexOnConsOut(msg *MsgNextLogIndex) gpa.OutMessages {
	vli.qcConsOut.VoteReceived(msg)
	return vli.tryOutputOnConsOut()
}

func (vli *varLogIndexImpl) msgNextLogIndexOnRecover(msg *MsgNextLogIndex) gpa.OutMessages {
	msgs := gpa.NoMessages()
	vli.qcRecover.VoteReceived(msg)
	sli := vli.qcRecover.EnoughVotes(vli.f + 1)
	msgs.AddAll(vli.qcRecover.MaybeSendVote(sli))
	if msg.PleaseRepeat {
		if msgs.Count() == 0 {
			msgs = vli.qcRecover.LastMessageForPeer(msg.Sender(), msgs)
		}
		msgs = vli.qcConsOut.LastMessageForPeer(msg.Sender(), msgs)
		msgs = vli.qcL1AORep.LastMessageForPeer(msg.Sender(), msgs)
		msgs = vli.qcStarted.LastMessageForPeer(msg.Sender(), msgs)
	}
	msgs.AddAll(vli.tryOutputOnRecover())
	return msgs
}

func (vli *varLogIndexImpl) msgNextLogIndexOnL1RepAO(msg *MsgNextLogIndex) gpa.OutMessages {
	vli.qcL1AORep.VoteReceived(msg)
	return vli.tryOutputOnL1RepAO()
}

func (vli *varLogIndexImpl) msgNextLogIndexOnStarted(msg *MsgNextLogIndex) gpa.OutMessages {
	vli.qcStarted.VoteReceived(msg)
	return vli.tryOutputOnStarted()
}

// If we voted for that LI based on consensus output, and there is N-F supporters, then proceed.
func (vli *varLogIndexImpl) tryOutputOnConsOut() gpa.OutMessages {
	ali := vli.qcConsOut.EnoughVotes(vli.n - vli.f)
	return vli.tryOutput(ali, MsgNextLogIndexCauseConsOut)
}

func (vli *varLogIndexImpl) tryOutputOnRecover() gpa.OutMessages {
	ali := vli.qcRecover.EnoughVotes(vli.n - vli.f)
	return vli.tryOutput(ali, MsgNextLogIndexCauseRecover)
}

func (vli *varLogIndexImpl) tryOutputOnL1RepAO() gpa.OutMessages {
	ali := vli.qcL1AORep.EnoughVotes(vli.n - vli.f)
	return vli.tryOutput(ali, MsgNextLogIndexCauseL1RepAO)
}

func (vli *varLogIndexImpl) tryOutputOnStarted() gpa.OutMessages {
	ali := vli.qcStarted.EnoughVotes(vli.f + 1)
	return vli.tryOutput(ali, MsgNextLogIndexCauseStarted)
}

// That's output for the consensus. We will start consensus instances with strictly increasing LIs with non-nil AOs.
func (vli *varLogIndexImpl) tryOutput(li LogIndex, cause MsgNextLogIndexCause) gpa.OutMessages {
	if li <= vli.agreedLI || li < vli.minLI {
		return nil
	}
	msgs := vli.qcStarted.MaybeSendVote(li)
	vli.agreedLI = li
	vli.log.Debugf("⊢ Output, li=%v", vli.agreedLI)
	vli.outputCB(vli.agreedLI)
	if vli.metrics != nil {
		switch cause {
		case MsgNextLogIndexCauseConsOut:
			vli.metrics.NextLogIndexCauseConsOut()
		case MsgNextLogIndexCauseRecover:
			vli.metrics.NextLogIndexCauseRecover()
		case MsgNextLogIndexCauseL1RepAO:
			vli.metrics.NextLogIndexCauseL1RepAO()
		case MsgNextLogIndexCauseStarted:
			vli.metrics.NextLogIndexCauseStarted()
		}
	}
	return msgs
}

func (vli *varLogIndexImpl) knownNodeID(nodeID gpa.NodeID) bool {
	return lo.Contains(vli.nodeIDs, nodeID)
}
