package cmt_log

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/gpa"
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
	outputCB  func(li LogIndex)
	log       *logger.Logger
}

func NewVarLogIndex(
	nodeIDs []gpa.NodeID,
	n int,
	f int,
	persistedLI LogIndex,
	outputCB func(li LogIndex),
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
		qcL1AORep: NewQuorumCounter(MsgNextLogIndexCauseL1ReplacedAO, nodeIDs, log),
		qcRecover: NewQuorumCounter(MsgNextLogIndexCauseRecover, nodeIDs, log),
		outputCB:  outputCB,
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

func (vli *varLogIndexImpl) LogIndexUsed(li LogIndex) { // TODO: Call it.
	if vli.minLI <= li {
		vli.minLI = li.Next()
	}
}

func (vli *varLogIndexImpl) ConsensusOutputReceived(consensusLI LogIndex) gpa.OutMessages {
	vli.log.Debugf("ConsensusOutputReceived: consensusLI=%v", consensusLI)
	msgs := vli.qcConsOut.MaybeSendVote(consensusLI.Next())
	vli.tryOutputOnConsOut()
	return msgs
}

func (vli *varLogIndexImpl) ConsensusRecoverReceived(consensusLI LogIndex) gpa.OutMessages {
	vli.log.Debugf("ConsensusRecoverReceived: consensusLI=%v", consensusLI)
	msgs := vli.qcRecover.MaybeSendVote(consensusLI.Next())
	vli.tryOutputOnRecover()
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
	vli.tryOutputOnL1ReplacedAO()
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
		vli.msgNextLogIndexOnConsOut(msg)
		return nil
	case MsgNextLogIndexCauseRecover:
		return vli.msgNextLogIndexOnRecover(msg)
	case MsgNextLogIndexCauseL1ReplacedAO:
		vli.msgNextLogIndexOnL1ReplacedAO(msg)
		return nil
	default:
		vli.log.Warnf("⊢ MsgNextLogIndex with unexpected cause: %+v", msg)
		return nil
	}
}

func (vli *varLogIndexImpl) msgNextLogIndexOnConsOut(msg *MsgNextLogIndex) {
	vli.qcConsOut.VoteReceived(msg)
	vli.tryOutputOnConsOut()
}

func (vli *varLogIndexImpl) msgNextLogIndexOnRecover(msg *MsgNextLogIndex) gpa.OutMessages {
	vli.qcRecover.VoteReceived(msg)
	sli, _ := vli.qcRecover.EnoughVotes(vli.f+1, false)
	msgs := vli.qcRecover.MaybeSendVote(sli)
	vli.tryOutputOnRecover()
	return msgs
}

func (vli *varLogIndexImpl) msgNextLogIndexOnL1ReplacedAO(msg *MsgNextLogIndex) {
	vli.qcL1AORep.VoteReceived(msg)
	vli.tryOutputOnL1ReplacedAO()
}

// If we voted for that LI based on consensus output, and there is N-F supporters, then proceed.
func (vli *varLogIndexImpl) tryOutputOnConsOut() {
	ali, _ := vli.qcConsOut.EnoughVotes(vli.n-vli.f, false)
	vli.tryOutput(ali)
}

func (vli *varLogIndexImpl) tryOutputOnRecover() {
	ali, _ := vli.qcRecover.EnoughVotes(vli.n-vli.f, false)
	vli.tryOutput(ali)
}

func (vli *varLogIndexImpl) tryOutputOnL1ReplacedAO() {
	ali, _ := vli.qcL1AORep.EnoughVotes(vli.n-vli.f, false)
	vli.tryOutput(ali)
}

// That's output for the consensus. We will start consensus instances with strictly increasing LIs with non-nil AOs.
func (vli *varLogIndexImpl) tryOutput(li LogIndex) {
	if li <= vli.agreedLI || li < vli.minLI {
		return
	}
	vli.agreedLI = li
	vli.log.Debugf("⊢ Output, li=%v", vli.agreedLI)
	vli.outputCB(vli.agreedLI)
}

func (vli *varLogIndexImpl) knownNodeID(nodeID gpa.NodeID) bool {
	return lo.Contains(vli.nodeIDs, nodeID)
}
