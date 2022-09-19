// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	dss_node "github.com/iotaledger/wasp/packages/chain/dss/node"
	mempool_pkg "github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/assert"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"
)

type consensus struct {
	isReady                          atomic.Bool
	chain                            chain.ChainCore
	committee                        chain.Committee
	committeePeerGroup               peering.GroupProvider
	mempool                          mempool_pkg.Mempool
	vmRunner                         vm.VMRunner
	currentState                     state.VirtualStateAccess
	stateOutput                      *isc.AliasOutputWithID
	stateTimestamp                   time.Time
	timeData                         time.Time
	acsSessionID                     uint64
	consensusBatch                   *BatchProposal
	consensusEntropy                 hashing.HashValue
	iAmContributor                   bool
	myContributionSeqNumber          uint16
	contributors                     []uint16
	workflow                         *workflowStatus
	delayBatchProposalUntil          time.Time
	delayRunVMUntil                  time.Time
	resultTxEssence                  *iotago.TransactionEssence
	resultState                      state.VirtualStateAccess
	finalTx                          *iotago.Transaction
	postTxDeadline                   time.Time
	pullInclusionStateDeadline       time.Time
	lastTimerTick                    atomic.Int64
	consensusInfoSnapshot            atomic.Value
	timers                           ConsensusTimers
	log                              *logger.Logger
	eventStateTransitionMsgPipe      pipe.Pipe
	eventDssIndexProposalMsgPipe     pipe.Pipe
	eventDssSignatureMsgPipe         pipe.Pipe
	eventPeerLogIndexMsgPipe         pipe.Pipe
	eventACSMsgPipe                  pipe.Pipe
	eventVMResultMsgPipe             pipe.Pipe
	eventTimerMsgPipe                pipe.Pipe
	assert                           *assert.Assert
	missingRequestsFromBatch         map[isc.RequestID][32]byte
	missingRequestsMutex             sync.Mutex
	pullMissingRequestsFromCommittee bool
	receivePeerMessagesAttachID      interface{}
	consensusMetrics                 metrics.ConsensusMetrics
	dssNode                          dss_node.DSSNode
	dssIndexProposal                 []int
	dssIndexProposalsDecided         [][]int
	dssSignature                     []byte
	markedForReset                   bool
	consensusJournal                 journal.ConsensusJournal
	consensusJournalLogIndex         journal.LogIndex // Index of the currently running log index.
	wal                              chain.WAL
	publishTx                        func(chainID *isc.ChainID, tx *iotago.Transaction) error
}

var _ chain.Consensus = &consensus{}

const (
	peerMsgTypePeerLogIndexMsg = iota

	maxMsgBuffer = 1000
)

func New(
	chainCore chain.ChainCore,
	mempool mempool_pkg.Mempool,
	committee chain.Committee,
	peerGroup peering.GroupProvider,
	pullMissingRequestsFromCommittee bool,
	consensusMetrics metrics.ConsensusMetrics,
	dssNode dss_node.DSSNode,
	consensusJournal journal.ConsensusJournal,
	wal chain.WAL,
	publishTx func(chainID *isc.ChainID, tx *iotago.Transaction) error,
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
		vmRunner:                         runvm.NewVMRunner(),
		workflow:                         newWorkflowStatus(false),
		timers:                           timers,
		log:                              log,
		eventStateTransitionMsgPipe:      pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventDssIndexProposalMsgPipe:     pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventDssSignatureMsgPipe:         pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventPeerLogIndexMsgPipe:         pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventACSMsgPipe:                  pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventVMResultMsgPipe:             pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventTimerMsgPipe:                pipe.NewLimitInfinitePipe(1),
		assert:                           assert.NewAssert(log),
		pullMissingRequestsFromCommittee: pullMissingRequestsFromCommittee,
		consensusMetrics:                 consensusMetrics,
		dssNode:                          dssNode,
		consensusJournal:                 consensusJournal,
		wal:                              wal,
		publishTx:                        publishTx,
	}
	ret.receivePeerMessagesAttachID = ret.committeePeerGroup.Attach(peering.PeerMessageReceiverConsensus, ret.receiveCommitteePeerMessages) // TODO: Don't need to attach here at all.
	ret.refreshConsensusInfo()
	go ret.recvLoop()
	return ret
}

func (c *consensus) receiveCommitteePeerMessages(peerMsg *peering.PeerMessageGroupIn) {
	switch peerMsg.MsgType {
	case peerMsgTypePeerLogIndexMsg:
		msg, err := messages.NewPeerLogIndexMsg(peerMsg.MsgData)
		if err != nil {
			c.log.Error(err)
			return
		}
		c.EnqueuePeerLogIndexMsg(&messages.PeerLogIndexMsgIn{
			PeerLogIndexMsg: *msg,
			SenderIndex:     peerMsg.SenderIndex,
		})
	default:
		c.log.Warnf("Wrong type of consensus message: %v, ignoring it", peerMsg.MsgType)
	}
}

func (c *consensus) IsReady() bool {
	return c.isReady.Load()
}

func (c *consensus) Close() {
	c.committeePeerGroup.Detach(c.receivePeerMessagesAttachID)

	c.eventStateTransitionMsgPipe.Close()

	tmpPipe := c.eventDssIndexProposalMsgPipe // TODO: Temporary workaround.
	c.eventDssIndexProposalMsgPipe = nil
	tmpPipe.Close()

	c.eventDssSignatureMsgPipe.Close()
	c.eventPeerLogIndexMsgPipe.Close()
	c.eventACSMsgPipe.Close()
	c.eventVMResultMsgPipe.Close()
	c.eventTimerMsgPipe.Close()
}

//nolint:gocyclo
func (c *consensus) recvLoop() {
	eventStateTransitionMsgCh := c.eventStateTransitionMsgPipe.Out()
	eventDssIndexProposalMsgCh := c.eventDssIndexProposalMsgPipe.Out()
	eventDssSignatureMsgCh := c.eventDssSignatureMsgPipe.Out()
	eventPeerLogIndexMsgCh := c.eventPeerLogIndexMsgPipe.Out()
	eventACSMsgCh := c.eventACSMsgPipe.Out()
	eventVMResultMsgCh := c.eventVMResultMsgPipe.Out()
	eventTimerMsgCh := c.eventTimerMsgPipe.Out()
	isClosedFun := func() bool {
		return eventStateTransitionMsgCh == nil &&
			eventDssIndexProposalMsgCh == nil &&
			eventDssSignatureMsgCh == nil &&
			eventPeerLogIndexMsgCh == nil &&
			// eventInclusionStateMsgCh == nil &&
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
		case msg, ok := <-eventDssIndexProposalMsgCh:
			if ok {
				c.log.Debugf("Consensus::recvLoop, handleDssIndexProposalMsg...")
				c.handleDssIndexProposalMsg(msg.(*messages.DssIndexProposalMsg))
				c.log.Debugf("Consensus::recvLoop, handleDssIndexProposalMsg... Done")
			} else {
				eventDssIndexProposalMsgCh = nil
			}
		case msg, ok := <-eventDssSignatureMsgCh:
			if ok {
				c.log.Debugf("Consensus::recvLoop, handleDssSignatureMsg...")
				c.handleDssSignatureMsg(msg.(*messages.DssSignatureMsg))
				c.log.Debugf("Consensus::recvLoop, handleDssSignatureMsg... Done")
			} else {
				eventDssSignatureMsgCh = nil
			}
		case msg, ok := <-eventPeerLogIndexMsgCh:
			if ok {
				c.log.Debugf("Consensus::recvLoop, handlePeerLogIndexMsg...")
				c.handlePeerLogIndexMsg(msg.(*messages.PeerLogIndexMsgIn))
				c.log.Debugf("Consensus::recvLoop, handlePeerLogIndexMsg... Done")
			} else {
				eventPeerLogIndexMsgCh = nil
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

func (c *consensus) SetTimeData(t time.Time) {
	c.timeData = t
}

func (c *consensus) refreshConsensusInfo() {
	index := uint32(0)
	if c.currentState != nil {
		index = c.currentState.BlockIndex()
	}
	var timeData time.Time
	if c.timeData.IsZero() {
		timeData = time.Now()
	} else {
		timeData = c.timeData
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
		eventPeerLogIndexMsgPipeSize:    c.eventPeerLogIndexMsgPipe.Len(),
		eventTimerMsgPipeSize:           c.eventTimerMsgPipe.Len(),
		eventVMResultMsgPipeSize:        c.eventVMResultMsgPipe.Len(),
		eventACSMsgPipeSize:             c.eventACSMsgPipe.Len(),
	}
}
