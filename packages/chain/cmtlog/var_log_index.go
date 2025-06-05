package cmtlog

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/metrics"
)

type VarLogIndex interface {
	// Summary of the internal state.
	StatusString() string

	// Consensus terminated with either with DONE or SKIP.
	// The logIndex is of the consensus that has been completed.
	ConsensusStarted(consensusLI LogIndex) gpa.OutMessages

	// Messages are exchanged, so this function handles them.
	MsgNextLogIndexReceived(msg *MsgNextLogIndex) gpa.OutMessages
}

type varLogIndexImpl struct {
	nodeIDs   []gpa.NodeID                    // All the peers in this committee.
	n         int                             // Total number of nodes.
	f         int                             // Maximal number of faulty nodes to tolerate.
	minLI     LogIndex                        // Minimal LI at which this node can participate (set on boot).
	agreedLI  LogIndex                        // LI for which we have N-F proposals (when reached, consensus starts, the LI is persisted).
	lastMsgs  map[gpa.NodeID]*MsgNextLogIndex // Latest messages we have sent to other peers.
	qcStarted *QuorumCounter
	outputCB  func(li LogIndex) gpa.OutMessages
	metrics   *metrics.ChainCmtLogMetrics
	log       log.Logger
}

func NewVarLogIndex(
	nodeIDs []gpa.NodeID,
	n int,
	f int,
	persistedLI LogIndex,
	outputCB func(li LogIndex) gpa.OutMessages,
	metrics *metrics.ChainCmtLogMetrics,
	log log.Logger,
) VarLogIndex {
	vli := &varLogIndexImpl{
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

func (vli *varLogIndexImpl) StatusString() string {
	return fmt.Sprintf(
		"{varLogIndex: minLI=%v, agreedLI=%v}",
		vli.minLI, vli.agreedLI,
	)
}

func (vli *varLogIndexImpl) ConsensusStarted(consensusLI LogIndex) gpa.OutMessages {
	vli.log.LogDebugf("ConsensusStarted: consensusLI=%v", consensusLI)
	msgs := gpa.NoMessages()
	msgs.AddAll(vli.qcStarted.MaybeSendVote(consensusLI))
	msgs.AddAll(vli.tryOutputOnStarted())
	return msgs
}

func (vli *varLogIndexImpl) MsgNextLogIndexReceived(msg *MsgNextLogIndex) gpa.OutMessages {
	vli.log.LogDebugf("MsgNextLogIndexReceived, %v", msg)
	sender := msg.Sender()
	if !vli.knownNodeID(sender) {
		vli.log.LogWarnf("⊢ MsgNextLogIndex from unknown sender: %+v", msg)
		return nil
	}

	switch msg.Cause {
	case MsgNextLogIndexCauseStarted:
		return vli.msgNextLogIndexOnStarted(msg)
	default:
		vli.log.LogWarnf("⊢ MsgNextLogIndex with unexpected cause: %+v", msg)
		return nil
	}
}

func (vli *varLogIndexImpl) msgNextLogIndexOnStarted(msg *MsgNextLogIndex) gpa.OutMessages {
	vli.qcStarted.VoteReceived(msg)
	return vli.tryOutputOnStarted()
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
	vli.agreedLI = li
	vli.log.LogDebugf("⊢ Output, li=%v", vli.agreedLI)
	if vli.metrics != nil {
		if cause == MsgNextLogIndexCauseStarted {
			vli.metrics.NextLogIndexCauseStarted()
		}
	}
	return vli.outputCB(vli.agreedLI)
}

func (vli *varLogIndexImpl) knownNodeID(nodeID gpa.NodeID) bool {
	return lo.Contains(vli.nodeIDs, nodeID)
}
