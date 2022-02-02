// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized and validated
package statemgr

import (
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/util/ready"
	"go.uber.org/atomic"
)

type stateManager struct {
	ready                  *ready.Ready
	store                  kvstore.KVStore
	chain                  chain.ChainCore
	domain                      *DomainWithFallback
	nodeConn                    chain.ChainNodeConnection
	pullStateRetryTime     time.Time
	solidState             state.VirtualStateAccess
	stateOutput            *iotago.AliasOutput
	stateOutputTimestamp   time.Time
	currentSyncData        atomic.Value
	notifiedAnchorOutputID iotago.OutputID
	syncingBlocks          *syncingBlocks
	receivePeerMessagesAttachID interface{}
	timers                 StateManagerTimers
	log                    *logger.Logger

	// Channels for accepting external events.
	eventGetBlockMsgPipe       pipe.Pipe
	eventBlockMsgPipe          pipe.Pipe
	eventStateOutputMsgPipe    pipe.Pipe
	eventOutputMsgPipe         pipe.Pipe
	eventStateCandidateMsgPipe pipe.Pipe
	eventTimerMsgPipe          pipe.Pipe
	stateManagerMetrics        metrics.StateManagerMetrics
	wal                        chain.WAL
}

var _ chain.StateManager = &stateManager{}

const (
	numberOfNodesToRequestBlockFromConst = 5
	maxBlocksToCommitConst               = 10000 // 10k
	maxMsgBuffer                         = 1000

	peerMsgTypeGetBlock = iota
	peerMsgTypeBlock
)

func New(
	store kvstore.KVStore,
	c chain.ChainCore,
	domain *DomainWithFallback,
	nodeconn chain.ChainNodeConnection,
	stateManagerMetrics metrics.StateManagerMetrics,
	wal chain.WAL,
	timersOpt ...StateManagerTimers,
) chain.StateManager {
	var timers StateManagerTimers
	if len(timersOpt) > 0 {
		timers = timersOpt[0]
	} else {
		timers = NewStateManagerTimers()
	}
	ret := &stateManager{
		ready:                      ready.New(fmt.Sprintf("state manager %s", c.ID().String()[:6]+"..")),
		store:                      store,
		chain:                      c,
		nodeConn:                   nodeconn,
		domain:                     domain,
		syncingBlocks:              newSyncingBlocks(c.Log(), timers.GetBlockRetry),
		timers:                     timers,
		log:                        c.Log().Named("s"),
		pullStateRetryTime:         time.Now(),
		eventGetBlockMsgPipe:       pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventBlockMsgPipe:          pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventStateOutputMsgPipe:    pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventOutputMsgPipe:         pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventStateCandidateMsgPipe: pipe.NewLimitInfinitePipe(maxMsgBuffer),
		eventTimerMsgPipe:          pipe.NewLimitInfinitePipe(1),
		stateManagerMetrics:        stateManagerMetrics,
		wal:                        wal,
	}
	ret.receivePeerMessagesAttachID = ret.domain.Attach(peering.PeerMessageReceiverStateManager, ret.receiveChainPeerMessages)
	ret.nodeConn.AttachToOutputReceived(ret.EnqueueOutputMsg)
	go ret.initLoadState()

	return ret
}

func (sm *stateManager) receiveChainPeerMessages(peerMsg *peering.PeerMessageIn) {
	switch peerMsg.MsgType {
	case peerMsgTypeGetBlock:
		msg, err := messages.NewGetBlockMsg(peerMsg.MsgData)
		if err != nil {
			sm.log.Error(err)
			return
		}
		sm.EnqueueGetBlockMsg(&messages.GetBlockMsgIn{
			GetBlockMsg: *msg,
			SenderPubKey: peerMsg.SenderPubKey,
		})
	case peerMsgTypeBlock:
		msg, err := messages.NewBlockMsg(peerMsg.MsgData)
		if err != nil {
			sm.log.Error(err)
			return
		}
		sm.EnqueueBlockMsg(&messages.BlockMsgIn{
			BlockMsg:    *msg,
			SenderPubKey: peerMsg.SenderPubKey,
		})
	default:
		sm.log.Warnf("Wrong type of state manager message: %v, ignoring it", peerMsg.MsgType)
	}
}

func (sm *stateManager) SetChainPeers(peers []*ed25519.PublicKey) {
	sm.domain.SetMainPeers(peers)
}

func (sm *stateManager) Close() {
	sm.nodeConn.DetachFromOutputReceived()
	sm.domain.Detach(sm.receivePeerMessagesAttachID)
	sm.domain.Close()

	sm.eventGetBlockMsgPipe.Close()
	sm.eventBlockMsgPipe.Close()
	sm.eventStateOutputMsgPipe.Close()
	sm.eventOutputMsgPipe.Close()
	sm.eventStateCandidateMsgPipe.Close()
	sm.eventTimerMsgPipe.Close()
}

// initial loading of the solid state
func (sm *stateManager) initLoadState() {
	solidState, stateExists, err := state.LoadSolidState(sm.store, sm.chain.ID())
	if err != nil {
		sm.chain.EnqueueDismissChain(fmt.Sprintf("StateManager.initLoadState: %v", err))
		return
	}
	if stateExists {
		sm.solidState = solidState
		sm.chain.GlobalStateSync().SetSolidIndex(solidState.BlockIndex())
		sm.log.Infof("SOLID STATE has been loaded. Block index: #%d, State hash: %s",
			solidState.BlockIndex(), solidState.StateCommitment().String())
	} else if err := sm.createOriginState(); err != nil {
		// create origin state in DB
		sm.chain.EnqueueDismissChain(fmt.Sprintf("StateManager.initLoadState. Failed to create origin state: %v", err))
		return
	}
	sm.recvLoop() // Check to process external events.
}

func (sm *stateManager) createOriginState() error {
	var err error

	sm.chain.GlobalStateSync().InvalidateSolidIndex()
	sm.solidState, err = state.CreateOriginState(sm.store, sm.chain.ID())
	sm.chain.GlobalStateSync().SetSolidIndex(0)

	if err != nil {
		sm.chain.EnqueueDismissChain(fmt.Sprintf("StateManager.initLoadState. Failed to create origin state: %v", err))
		return err
	}
	sm.log.Infof("ORIGIN STATE has been created")
	return nil
}

func (sm *stateManager) Ready() *ready.Ready {
	return sm.ready
}

func (sm *stateManager) GetStatusSnapshot() *chain.SyncInfo {
	v := sm.currentSyncData.Load()
	if v == nil {
		return nil
	}
	return v.(*chain.SyncInfo)
}

func (sm *stateManager) recvLoop() {
	sm.ready.SetReady()
	eventGetBlockMsgCh := sm.eventGetBlockMsgPipe.Out()
	eventBlockMsgCh := sm.eventBlockMsgPipe.Out()
	eventStateOutputMsgCh := sm.eventStateOutputMsgPipe.Out()
	eventOutputMsgCh := sm.eventOutputMsgPipe.Out()
	eventStateCandidateMsgCh := sm.eventStateCandidateMsgPipe.Out()
	eventTimerMsgCh := sm.eventTimerMsgPipe.Out()
	for {
		select {
		case msg, ok := <-eventGetBlockMsgCh:
			if ok {
				sm.handleGetBlockMsg(msg.(*messages.GetBlockMsgIn))
			} else {
				eventGetBlockMsgCh = nil
			}
		case msg, ok := <-eventBlockMsgCh:
			if ok {
				sm.handleBlockMsg(msg.(*messages.BlockMsgIn))
			} else {
				eventBlockMsgCh = nil
			}
		case msg, ok := <-eventStateOutputMsgCh:
			if ok {
				sm.handleStateMsg(msg.(*messages.StateMsg))
			} else {
				eventStateOutputMsgCh = nil
			}
		case msg, ok := <-eventOutputMsgCh:
			if ok {
				sm.handleOutputMsg(msg.(iotago.Output))
			} else {
				eventOutputMsgCh = nil
			}
		case msg, ok := <-eventStateCandidateMsgCh:
			if ok {
				sm.handleStateCandidateMsg(msg.(*messages.StateCandidateMsg))
			} else {
				eventStateCandidateMsgCh = nil
			}
		case _, ok := <-eventTimerMsgCh:
			if ok {
				sm.handleTimerMsg()
			} else {
				eventTimerMsgCh = nil
			}
		}
		if eventGetBlockMsgCh == nil &&
			eventBlockMsgCh == nil &&
			eventStateOutputMsgCh == nil &&
			eventOutputMsgCh == nil &&
			eventStateCandidateMsgCh == nil &&
			eventTimerMsgCh == nil {
			return
		}
	}
}
