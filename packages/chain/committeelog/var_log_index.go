package committeelog

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/metrics"
)

type VarLogIndex struct {
	nodeIDs   []gpa.NodeID                    // All the peers in this committee.
	n         int                             // Total number of nodes.
	f         int                             // Maximal number of faulty nodes to tolerate.
	minLI     LogIndex                        // Minimal LI at which this node can participate (set on boot).
	agreedLI  LogIndex                        // LI for which we have N-F proposals (when reached, consensus starts, the LI is persisted).
	lastMsgs  map[gpa.NodeID]*MsgNextLogIndex // Latest messages we have sent to other peers.
	qcStarted *QuorumCounter
	outputCB  func(li LogIndex) []gpa.PayloadOut
	metrics   *metrics.ChainCommitteeLogMetrics
	log       log.Logger
}

func NewVarLogIndex(
	nodeIDs []gpa.NodeID,
	n int,
	f int,
	persistedLI LogIndex,
	outputCB func(li LogIndex) []gpa.PayloadOut,
	metrics *metrics.ChainCommitteeLogMetrics,
	log log.Logger,
) *VarLogIndex {
	vli := &VarLogIndex{
		nodeIDs:   nodeIDs,
		n:         n,
		f:         f,
		minLI:     persistedLI.Next(),
		agreedLI:  NilLogIndex(),
		lastMsgs:  map[gpa.NodeID]*MsgNextLogIndex{},
		qcStarted: NewQuorumCounter(MsgNextLogIndexCauseStarted, nodeIDs, log),
		outputCB:  outputCB,
		metrics:   metrics,
		log:       log,
	}
	return vli
}

func (vli *VarLogIndex) StatusString() string {
	return fmt.Sprintf(
		"{varLogIndex: minLI=%v, agreedLI=%v}",
		vli.minLI, vli.agreedLI,
	)
}

func (vli *VarLogIndex) ConsensusStarted(consensusLI LogIndex) []gpa.PayloadOut {
	vli.log.LogDebugf("ConsensusStarted: consensusLI=%v", consensusLI)
	var msgs []gpa.PayloadOut
	msgs = append(msgs, vli.qcStarted.MaybeSendVote(consensusLI)...)
	msgs = append(msgs, vli.tryOutputOnStarted()...)
	return msgs
}

func (vli *VarLogIndex) MsgNextLogIndexReceived(msg gpa.PayloadIn[MsgNextLogIndex]) []gpa.PayloadOut {
	vli.log.LogDebugf("MsgNextLogIndexReceived, %v", msg)
	sender := msg.Sender
	if !vli.knownNodeID(sender) {
		vli.log.LogWarnf("⊢ MsgNextLogIndex from unknown sender: %+v", msg)
		return nil
	}

	switch msg.Payload.Cause {
	case MsgNextLogIndexCauseStarted:
		return vli.msgNextLogIndexOnStarted(msg)
	default:
		vli.log.LogWarnf("⊢ MsgNextLogIndex with unexpected cause: %+v", msg)
		return nil
	}
}

func (vli *VarLogIndex) msgNextLogIndexOnStarted(msg gpa.PayloadIn[MsgNextLogIndex]) []gpa.PayloadOut {
	vli.qcStarted.VoteReceived(msg)
	return vli.tryOutputOnStarted()
}

func (vli *VarLogIndex) tryOutputOnStarted() []gpa.PayloadOut {
	ali := vli.qcStarted.EnoughVotes(vli.f + 1)
	return vli.tryOutput(ali, MsgNextLogIndexCauseStarted)
}

// That's output for the consensus. We will start consensus instances with strictly increasing LIs with non-nil Anchors.
func (vli *VarLogIndex) tryOutput(li LogIndex, cause MsgNextLogIndexCause) []gpa.PayloadOut {
	if li <= vli.agreedLI || li < vli.minLI {
		return nil
	}
	vli.agreedLI = li
	vli.log.LogDebugf("⊢ Output, li=%v", vli.agreedLI)
	if vli.metrics != nil {
		if cause == MsgNextLogIndexCauseStarted {
			vli.metrics.NextLogIndexCauseStarted()
		}
	}
	return vli.outputCB(vli.agreedLI)
}

func (vli *VarLogIndex) knownNodeID(nodeID gpa.NodeID) bool {
	return lo.Contains(vli.nodeIDs, nodeID)
}
