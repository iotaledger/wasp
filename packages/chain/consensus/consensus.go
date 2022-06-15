// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/chain"
	mempool_pkg "github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"
	"go.uber.org/atomic"
)

type consensus struct {
	isReady                          atomic.Bool
	chain                            chain.ChainCore
	committee                        chain.Committee
	committeePeerGroup               peering.GroupProvider
	mempool                          mempool_pkg.Mempool
	nodeConn                         chain.ChainNodeConnection
	vmRunner                         vm.VMRunner
	currentState                     state.VirtualStateAccess
	stateOutput                      *iscp.AliasOutputWithID
	stateTimestamp                   time.Time
	timeData                         *iscp.TimeData
	acsSessionID                     uint64
	consensusBatch                   *BatchProposal
	consensusEntropy                 hashing.HashValue
	iAmContributor                   bool
	myContributionSeqNumber          uint16
	contributors                     []uint16
	workflow                         *workflowStatus
	delayBatchProposalUntil          time.Time
	delayRunVMUntil                  time.Time
	delaySendingSignedResult         time.Time
	resultTxEssence                  *iotago.TransactionEssence
	resultState                      state.VirtualStateAccess
	resultSignatures                 []*messages.SignedResultMsgIn
	resultSigAck                     []uint16
	finalTx                          *iotago.Transaction
	postTxDeadline                   time.Time
	pullInclusionStateDeadline       time.Time
	lastTimerTick                    atomic.Int64
	consensusInfoSnapshot            atomic.Value
	timers                           ConsensusTimers
	log                              *logger.Logger
	eventStateTransitionMsgPipe      pipe.Pipe
	eventSignedResultMsgPipe         pipe.Pipe
	eventSignedResultAckMsgPipe      pipe.Pipe
	eventInclusionStateMsgPipe       pipe.Pipe
	eventACSMsgPipe                  pipe.Pipe
	eventVMResultMsgPipe             pipe.Pipe
	eventTimerMsgPipe                pipe.Pipe
	assert                           *assert.Assert
	missingRequestsFromBatch         map[iscp.RequestID][32]byte
	missingRequestsMutex             sync.Mutex
	pullMissingRequestsFromCommittee bool
	receivePeerMessagesAttachID      interface{}
	consensusMetrics                 metrics.ConsensusMetrics
	wal                              chain.WAL
}

var _ chain.Consensus = &consensus{}

const (
	peerMsgTypeSignedResult = iota
	peerMsgTypeSignedResultAck

	maxMsgBuffer = 1000
)

func New(
	chainCore chain.ChainCore,
	mempool mempool_pkg.Mempool,
	committee chain.Committee,
	peerGroup peering.GroupProvider,
	nodeConn chain.ChainNodeConnection,
	pullMissingRequestsFromCommittee bool,
	consensusMetrics metrics.ConsensusMetrics,
	wal chain.WAL,
	timersOpt ...ConsensusTimers,
) chain.Consensus {
	var timers ConsensusTimers
	if len(timersOpt) > 0 {
		timers = timersOpt[0]
	} else {
		timers = NewConsensusTimers()
	}
	log := chainCore.Log().Named("c")
	ret := &consensus{
		chain:                            chainCore,
		committee:                        committee,
		committeePeerGroup:               peerGroup,
		mempool:                          mempool,
		nodeConn:                         nodeConn,
		vmRunner:                         runvm.NewVMRunner(),
		workflow:                         newWorkflowStatus(false),
		resultSignatures:                 make([]*messages.SignedResultMsgIn, committee.Size()),
		resultSigAck:                     make([]uint16, 0, committee.Size()),
		timers:                           timers,
		log:                              log,
		eventStateTransitionMsgPipe:      pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventSignedResultMsgPipe:         pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventSignedResultAckMsgPipe:      pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventInclusionStateMsgPipe:       pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventACSMsgPipe:                  pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventVMResultMsgPipe:             pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventTimerMsgPipe:                pipe.NewLimitInfinitePipe(1),
		assert:                           assert.NewAssert(log),
		pullMissingRequestsFromCommittee: pullMissingRequestsFromCommittee,
		consensusMetrics:                 consensusMetrics,
		wal:                              wal,
	}
	ret.receivePeerMessagesAttachID = ret.committeePeerGroup.Attach(peering.PeerMessageReceiverConsensus, ret.receiveCommitteePeerMessages)
	ret.nodeConn.AttachToMilestones(func(milestonePointer *nodeclient.MilestoneInfo) {
		ret.timeData = &iscp.TimeData{
			MilestoneIndex: milestonePointer.Index,
			Time:           time.Unix(int64(milestonePointer.Timestamp), 0),
		}
	})
	ret.nodeConn.AttachToTxInclusionState(func(txID iotago.TransactionID, inclusionState string) {
		ret.EnqueueTxInclusionsStateMsg(txID, inclusionState)
	})
	ret.refreshConsensusInfo()
	go ret.recvLoop()
	return ret
}

func (c *consensus) receiveCommitteePeerMessages(peerMsg *peering.PeerMessageGroupIn) {
	switch peerMsg.MsgType {
	case peerMsgTypeSignedResult:
		msg, err := messages.NewSignedResultMsg(peerMsg.MsgData)
		if err != nil {
			c.log.Error(err)
			return
		}
		c.EnqueueSignedResultMsg(&messages.SignedResultMsgIn{
			SignedResultMsg: *msg,
			SenderIndex:     peerMsg.SenderIndex,
		})
	case peerMsgTypeSignedResultAck:
		msg, err := messages.NewSignedResultAckMsg(peerMsg.MsgData)
		if err != nil {
			c.log.Error(err)
			return
		}
		c.EnqueueSignedResultAckMsg(&messages.SignedResultAckMsgIn{
			SignedResultAckMsg: *msg,
			SenderIndex:        peerMsg.SenderIndex,
		})
	default:
		c.log.Warnf("Wrong type of consensus message: %v, ignoring it", peerMsg.MsgType)
	}
}

func (c *consensus) IsReady() bool {
	return c.isReady.Load()
}

func (c *consensus) Close() {
	c.nodeConn.DetachFromTxInclusionState()
	c.committeePeerGroup.Detach(c.receivePeerMessagesAttachID)

	c.eventStateTransitionMsgPipe.Close()
	c.eventSignedResultMsgPipe.Close()
	c.eventSignedResultAckMsgPipe.Close()
	c.eventInclusionStateMsgPipe.Close()
	c.eventACSMsgPipe.Close()
	c.eventVMResultMsgPipe.Close()
	c.eventTimerMsgPipe.Close()
}

func (c *consensus) recvLoop() {
	eventStateTransitionMsgCh := c.eventStateTransitionMsgPipe.Out()
	eventSignedResultMsgCh := c.eventSignedResultMsgPipe.Out()
	eventSignedResultAckMsgCh := c.eventSignedResultAckMsgPipe.Out()
	eventInclusionStateMsgCh := c.eventInclusionStateMsgPipe.Out()
	eventACSMsgCh := c.eventACSMsgPipe.Out()
	eventVMResultMsgCh := c.eventVMResultMsgPipe.Out()
	eventTimerMsgCh := c.eventTimerMsgPipe.Out()
	isClosedFun := func() bool {
		return eventStateTransitionMsgCh == nil &&
			eventSignedResultMsgCh == nil &&
			eventSignedResultAckMsgCh == nil &&
			eventInclusionStateMsgCh == nil &&
			eventACSMsgCh == nil &&
			eventVMResultMsgCh == nil &&
			eventTimerMsgCh == nil
	}

	// wait at startup
	for !c.committee.IsReady() {
		time.Sleep(100 * time.Millisecond)
		if isClosedFun() {
			return
		}
	}
	c.log.Debugf("consensus object is ready")
	c.isReady.Store(true)
	for {
		select {
		case msg, ok := <-eventStateTransitionMsgCh:
			if ok {
				c.log.Debugf("Consensus::recvLoop, eventStateTransitionMsg...")
				c.handleStateTransitionMsg(msg.(*messages.StateTransitionMsg))
				c.log.Debugf("Consensus::recvLoop, eventStateTransitionMsg... Done")
			} else {
				eventStateTransitionMsgCh = nil
			}
		case msg, ok := <-eventSignedResultMsgCh:
			if ok {
				c.log.Debugf("Consensus::recvLoop, handleSignedResultMsg...")
				c.handleSignedResultMsg(msg.(*messages.SignedResultMsgIn))
				c.log.Debugf("Consensus::recvLoop, handleSignedResultMsg... Done")
			} else {
				eventSignedResultMsgCh = nil
			}
		case msg, ok := <-eventSignedResultAckMsgCh:
			if ok {
				c.log.Debugf("Consensus::recvLoop, handleSignedResultAckMsg...")
				c.handleSignedResultAckMsg(msg.(*messages.SignedResultAckMsgIn))
				c.log.Debugf("Consensus::recvLoop, handleSignedResultAckMsg... Done")
			} else {
				eventSignedResultAckMsgCh = nil
			}
		case msg, ok := <-eventInclusionStateMsgCh:
			if ok {
				c.log.Debugf("Consensus::recvLoop, eventTxInclusionState...")
				c.handleTxInclusionState(msg.(*messages.TxInclusionStateMsg))
				c.log.Debugf("Consensus::recvLoop, eventTxInclusionState... Done")
			} else {
				eventInclusionStateMsgCh = nil
			}
		case msg, ok := <-eventACSMsgCh:
			if ok {
				c.log.Debugf("Consensus::recvLoop, eventAsynchronousCommonSubset...")
				c.handleAsynchronousCommonSubset(msg.(*messages.AsynchronousCommonSubsetMsg))
				c.log.Debugf("Consensus::recvLoop, eventAsynchronousCommonSubset... Done")
			} else {
				eventACSMsgCh = nil
			}
		case msg, ok := <-eventVMResultMsgCh:
			if ok {
				c.log.Debugf("Consensus::recvLoop, eventVMResultMsg...")
				c.handleVMResultMsg(msg.(*messages.VMResultMsg))
				c.log.Debugf("Consensus::recvLoop, eventVMResultMsg... Done")
			} else {
				eventVMResultMsgCh = nil
			}
		case msg, ok := <-eventTimerMsgCh:
			if ok {
				c.log.Debugf("Consensus::recvLoop, eventTimerMsg...")
				c.handleTimerMsg(msg.(messages.TimerTick))
				c.log.Debugf("Consensus::recvLoop, eventTimerMsg... Done")
			} else {
				eventTimerMsgCh = nil
			}
		}
		if isClosedFun() {
			return
		}
	}
}

func (c *consensus) refreshConsensusInfo() {
	index := uint32(0)
	if c.currentState != nil {
		index = c.currentState.BlockIndex()
	}
	var timeData iscp.TimeData
	if c.timeData == nil {
		timeData = iscp.TimeData{Time: time.Now()}
	} else {
		timeData = *c.timeData
	}
	consensusInfo := &chain.ConsensusInfo{
		StateIndex: index,
		Mempool:    c.mempool.Info(timeData),
		TimerTick:  int(c.lastTimerTick.Load()),
		TimeData:   timeData,
	}
	c.log.Debugf("Refreshing consensus info: index=%v, timerTick=%v, "+
		"totalPool=%v, mempoolReady=%v, inBufCounter=%v, outBufCounter=%v, "+
		"inPoolCounter=%v, outPoolCounter=%v",
		consensusInfo.StateIndex, consensusInfo.TimerTick,
		consensusInfo.Mempool.TotalPool, consensusInfo.Mempool.ReadyCounter,
		consensusInfo.Mempool.InBufCounter, consensusInfo.Mempool.OutBufCounter,
		consensusInfo.Mempool.InPoolCounter, consensusInfo.Mempool.OutPoolCounter,
	)
	c.consensusInfoSnapshot.Store(consensusInfo)
}

func (c *consensus) GetStatusSnapshot() *chain.ConsensusInfo {
	ret := c.consensusInfoSnapshot.Load()
	if ret == nil {
		return nil
	}
	return ret.(*chain.ConsensusInfo)
}

func (c *consensus) GetWorkflowStatus() chain.ConsensusWorkflowStatus {
	return c.workflow
}

func (c *consensus) GetPipeMetrics() chain.ConsensusPipeMetrics {
	return &pipeMetrics{
		eventStateTransitionMsgPipeSize: c.eventStateTransitionMsgPipe.Len(),
		eventSignedResultMsgPipeSize:    c.eventSignedResultMsgPipe.Len(),
		eventSignedResultAckMsgPipeSize: c.eventSignedResultAckMsgPipe.Len(),
		eventInclusionStateMsgPipeSize:  c.eventInclusionStateMsgPipe.Len(),
		eventTimerMsgPipeSize:           c.eventTimerMsgPipe.Len(),
		eventVMResultMsgPipeSize:        c.eventVMResultMsgPipe.Len(),
		eventACSMsgPipeSize:             c.eventACSMsgPipe.Len(),
	}
}
