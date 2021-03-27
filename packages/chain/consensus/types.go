// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/goshimmer/packages/txstream"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/tcrypto/tbdn"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

type operator struct {
	chain chain.Chain

	dkshare *tcrypto.DKShare
	//currentState
	currentState state.VirtualState
	stateTx      *sctransaction_old.TransactionEssence
	balances     map[valuetransaction.ID][]*balance.Balance

	// consensus stage
	consensusStage         int
	consensusStageDeadline time.Time
	//
	requestBalancesDeadline time.Time

	// notifications with future currentState indices
	notificationsBacklog []*chain.NotifyReqMsg

	// backlog of requests with all information
	requests map[coretypes.RequestID]*request

	peerPermutation *util.Permutation16

	leaderStatus            *leaderStatus
	sentResultToLeaderIndex uint16
	sentResultToLeader      *sctransaction_old.TransactionEssence

	postedResultTxid       *valuetransaction.ID
	nextPullInclusionLevel time.Time // if postedResultTxid != nil

	nextArgSolidificationDeadline time.Time

	log *logger.Logger

	// data for concurrent access, from APIs mostly
	concurrentAccessMutex sync.RWMutex
	requestIdsProtected   map[coretypes.RequestID]bool

	// Channels for accepting external events.
	eventStateTransitionMsgCh           chan *chain.StateTransitionMsg
	eventBalancesMsgCh                  chan chain.BalancesMsg
	eventRequestMsgCh                   chan *chain.RequestMsg
	eventNotifyReqMsgCh                 chan *chain.NotifyReqMsg
	eventStartProcessingBatchMsgCh      chan *chain.StartProcessingBatchMsg
	eventResultCalculatedCh             chan *chain.VMResultMsg
	eventSignedHashMsgCh                chan *chain.SignedHashMsg
	eventNotifyFinalResultPostedMsgCh   chan *chain.NotifyFinalResultPostedMsg
	eventTransactionInclusionLevelMsgCh chan *txstream.MsgTxInclusionState
	eventTimerMsgCh                     chan chain.TimerTick
	closeCh                             chan bool
}

type leaderStatus struct {
	reqs          []*request
	batch         state.Block
	batchHash     hashing.HashValue
	timestamp     int64
	balances      map[valuetransaction.ID][]*balance.Balance
	resultTx      *sctransaction_old.TransactionEssence
	finalized     bool
	signedResults []*signedResult
}

type signedResult struct {
	essenceHash hashing.HashValue
	sigShare    tbdn.SigShare
}

// backlog entry. Keeps stateTx of the request
type request struct {
	// id of the hash of request tx id and request block index
	reqId coretypes.RequestID
	// from request message. nil if request message wasn't received yet
	reqTx *sctransaction_old.TransactionEssence
	// from request message. Not nil only if free tokens were attached to the request
	freeTokens coretypes.ColoredBalancesOld
	// time when request message was received by the operator
	whenMsgReceived time.Time
	// notification vector for the current currentState
	notifications []bool
	// true if arguments were decoded/solidified already. If not, the request in not eligible for the batch
	argsSolid bool

	log *logger.Logger
}

func NewOperator(committee chain.Chain, dkshare *tcrypto.DKShare, log *logger.Logger) *operator {
	defer committee.SetReadyConsensus()

	ret := &operator{
		chain:                               committee,
		dkshare:                             dkshare,
		requests:                            make(map[coretypes.RequestID]*request),
		requestIdsProtected:                 make(map[coretypes.RequestID]bool),
		peerPermutation:                     util.NewPermutation16(committee.Size(), nil),
		log:                                 log.Named("c"),
		eventStateTransitionMsgCh:           make(chan *chain.StateTransitionMsg),
		eventBalancesMsgCh:                  make(chan chain.BalancesMsg),
		eventRequestMsgCh:                   make(chan *chain.RequestMsg),
		eventNotifyReqMsgCh:                 make(chan *chain.NotifyReqMsg),
		eventStartProcessingBatchMsgCh:      make(chan *chain.StartProcessingBatchMsg),
		eventResultCalculatedCh:             make(chan *chain.VMResultMsg),
		eventSignedHashMsgCh:                make(chan *chain.SignedHashMsg),
		eventNotifyFinalResultPostedMsgCh:   make(chan *chain.NotifyFinalResultPostedMsg),
		eventTransactionInclusionLevelMsgCh: make(chan *chain.TransactionInclusionLevelMsg),
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
		select {
		case msg, ok := <-op.eventStateTransitionMsgCh:
			if ok {
				op.eventStateTransitionMsg(msg)
			}
		case msg, ok := <-op.eventBalancesMsgCh:
			if ok {
				op.eventBalancesMsg(msg)
			}
		case msg, ok := <-op.eventRequestMsgCh:
			if ok {
				op.eventRequestMsg(msg)
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
				op.eventTransactionInclusionLevelMsg(msg)
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
	return *op.dkshare.Index
}

func (op *operator) quorum() uint16 {
	return op.dkshare.T
}

func (op *operator) size() uint16 {
	return op.dkshare.N
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
	// TODO
	// temporary to the chain owner's account
	return *coretypes.NewAgentID(op.chain.ID().AsAddress(), accounts.Interface.Hname())
}
