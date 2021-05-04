// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensusimpl

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto/tbdn"
	"github.com/iotaledger/wasp/packages/util"
)

const pullBacklogPeriod = 3 * time.Second

type operator struct {
	chain     chain.ChainCore
	committee chain.Committee
	mempool   chain.Mempool
	nodeConn  chain.NodeConnection
	//currentState
	currentState   state.VirtualState
	stateOutput    *ledgerstate.AliasOutput
	stateTimestamp time.Time

	// consensus stage
	consensusStage         int
	consensusStageDeadline time.Time

	peerPermutation *util.Permutation16

	leaderStatus            *leaderStatus
	sentResultToLeaderIndex uint16
	sentResultToLeader      *ledgerstate.TransactionEssence

	pullBacklogDeadline time.Time

	postedResultTxid       ledgerstate.TransactionID
	nextPullInclusionLevel time.Time // if postedResultTxid != nil

	nextArgSolidificationDeadline time.Time

	log *logger.Logger

	// data for concurrent access, from APIs mostly
	concurrentAccessMutex sync.RWMutex
	requestIdsProtected   map[coretypes.RequestID]bool

	// Channels for accepting external events.
	eventStateTransitionMsgCh           chan *chain.StateTransitionMsg
	eventNotifyReqMsgCh                 chan *chain.NotifyReqMsg
	eventStartProcessingBatchMsgCh      chan *chain.StartProcessingBatchMsg
	eventResultCalculatedCh             chan *chain.VMResultMsg
	eventSignedHashMsgCh                chan *chain.SignedHashMsg
	eventNotifyFinalResultPostedMsgCh   chan *chain.NotifyFinalResultPostedMsg
	eventTransactionInclusionLevelMsgCh chan *chain.InclusionStateMsg
	eventTimerMsgCh                     chan chain.TimerTick
	closeCh                             chan bool
}

type leaderStatus struct {
	reqs            []coretypes.Request
	virtualState    state.VirtualState
	batchHash       hashing.HashValue
	timestamp       time.Time
	resultTxEssence *ledgerstate.TransactionEssence
	finalized       bool
	signedResults   []*signedResult
}

type signedResult struct {
	essenceHash hashing.HashValue
	sigShare    tbdn.SigShare
}

func New(chainCore chain.ChainCore, mempool chain.Mempool, committee chain.Committee, nodeConn chain.NodeConnection, log *logger.Logger) *operator {
	ret := &operator{
		chain:                               chainCore,
		committee:                           committee,
		mempool:                             mempool,
		nodeConn:                            nodeConn,
		requestIdsProtected:                 make(map[coretypes.RequestID]bool),
		peerPermutation:                     util.NewPermutation16(committee.Size(), nil),
		log:                                 log.Named("c"),
		eventStateTransitionMsgCh:           make(chan *chain.StateTransitionMsg),
		eventNotifyReqMsgCh:                 make(chan *chain.NotifyReqMsg),
		eventStartProcessingBatchMsgCh:      make(chan *chain.StartProcessingBatchMsg),
		eventResultCalculatedCh:             make(chan *chain.VMResultMsg),
		eventSignedHashMsgCh:                make(chan *chain.SignedHashMsg),
		eventNotifyFinalResultPostedMsgCh:   make(chan *chain.NotifyFinalResultPostedMsg),
		eventTransactionInclusionLevelMsgCh: make(chan *chain.InclusionStateMsg),
		eventTimerMsgCh:                     make(chan chain.TimerTick),
		closeCh:                             make(chan bool),
	}
	ret.setNextConsensusStage(consensusStageNoSync)
	go ret.recvLoop()
	return ret
}

func (op *operator) Close() {
	close(op.closeCh)
}

func (op *operator) recvLoop() {
	for {
		if !op.committee.IsReady() {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		select {
		case msg, ok := <-op.eventStateTransitionMsgCh:
			if ok {
				op.eventStateTransitionMsg(msg)
			}
		case msg, ok := <-op.eventNotifyReqMsgCh:
			if ok {
				op.eventNotifyReqMsg(msg)
			}
		case msg, ok := <-op.eventStartProcessingBatchMsgCh:
			if ok {
				op.eventStartProcessingBatchMsg(msg)
			}
		case msg, ok := <-op.eventResultCalculatedCh:
			if ok {
				op.eventResultCalculated(msg)
			}
		case msg, ok := <-op.eventSignedHashMsgCh:
			if ok {
				op.eventSignedHashMsg(msg)
			}
		case msg, ok := <-op.eventNotifyFinalResultPostedMsgCh:
			if ok {
				op.eventNotifyFinalResultPostedMsg(msg)
			}
		case msg, ok := <-op.eventTransactionInclusionLevelMsgCh:
			if ok {
				op.eventTransactionInclusionStateMsg(msg)
			}
		case msg, ok := <-op.eventTimerMsgCh:
			if ok {
				op.eventTimerMsg(msg)
			}
		case <-op.closeCh:
			return
		}
	}
}

func (op *operator) peerIndex() uint16 {
	return op.committee.OwnPeerIndex()
}

func (op *operator) quorum() uint16 {
	return op.committee.Quorum()
}

func (op *operator) size() uint16 {
	return op.committee.Size()
}

func (op *operator) blockIndex() (uint32, bool) {
	if op.currentState == nil {
		return 0, false
	}
	return op.currentState.BlockIndex(), true
}

func (op *operator) mustStateIndex() uint32 {
	ret, ok := op.blockIndex()
	if !ok {
		panic("mustStateIndex")
	}
	return ret
}

func (op *operator) getFeeDestination() coretypes.AgentID {
	// TODO temporary
	return *coretypes.NewAgentID(op.chain.ID().AsAddress(), 0)
}
