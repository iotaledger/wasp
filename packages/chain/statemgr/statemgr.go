// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized and validated
package statemgr

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/ready"
	"go.uber.org/atomic"
)

type stateManager struct {
	ready                  *ready.Ready
	store                  kvstore.KVStore
	chain                  chain.ChainCore
	peers                  peering.PeerDomainProvider
	nodeConn               chain.NodeConnection
	pullStateRetryTime     time.Time
	solidState             state.VirtualState
	stateOutput            *ledgerstate.AliasOutput
	stateOutputTimestamp   time.Time
	currentSyncData        atomic.Value
	notifiedAnchorOutputID ledgerstate.OutputID
	syncingBlocks          *syncingBlocks
	timers                 StateManagerTimers
	log                    *logger.Logger

	// Channels for accepting external events.
	eventGetBlockMsgCh       chan *messages.GetBlockMsg
	eventBlockMsgCh          chan *messages.BlockMsg
	eventStateOutputMsgCh    chan *messages.StateMsg
	eventOutputMsgCh         chan ledgerstate.Output
	eventStateCandidateMsgCh chan *messages.StateCandidateMsg
	eventTimerMsgCh          chan messages.TimerTick
	closeCh                  chan bool
}

const (
	numberOfNodesToRequestBlockFromConst = 5
	maxBlocksToCommitConst               = 10000 // 10k
)

func New(store kvstore.KVStore, c chain.ChainCore, peers peering.PeerDomainProvider, nodeconn chain.NodeConnection, timersOpt ...StateManagerTimers) chain.StateManager {
	var timers StateManagerTimers
	if len(timersOpt) > 0 {
		timers = timersOpt[0]
	} else {
		timers = NewStateManagerTimers()
	}
	ret := &stateManager{
		ready:                    ready.New(fmt.Sprintf("state manager %s", c.ID().Base58()[:6]+"..")),
		store:                    store,
		chain:                    c,
		nodeConn:                 nodeconn,
		peers:                    peers,
		syncingBlocks:            newSyncingBlocks(c.Log(), timers.GetBlockRetry),
		timers:                   timers,
		log:                      c.Log().Named("s"),
		pullStateRetryTime:       time.Now(),
		eventGetBlockMsgCh:       make(chan *messages.GetBlockMsg),
		eventBlockMsgCh:          make(chan *messages.BlockMsg),
		eventStateOutputMsgCh:    make(chan *messages.StateMsg),
		eventOutputMsgCh:         make(chan ledgerstate.Output),
		eventStateCandidateMsgCh: make(chan *messages.StateCandidateMsg),
		eventTimerMsgCh:          make(chan messages.TimerTick),
		closeCh:                  make(chan bool),
	}
	go ret.initLoadState()

	return ret
}

func (sm *stateManager) Close() {
	close(sm.closeCh)
}

// initial loading of the solid state
func (sm *stateManager) initLoadState() {
	solidState, stateExists, err := state.LoadSolidState(sm.store, sm.chain.ID())
	if err != nil {
		go sm.chain.ReceiveMessage(messages.DismissChainMsg{
			Reason: fmt.Sprintf("StateManager.initLoadState: %v", err),
		})
		return
	}
	if stateExists {
		sm.solidState = solidState
		sm.chain.GlobalStateSync().SetSolidIndex(solidState.BlockIndex())
		sm.log.Infof("SOLID STATE has been loaded. Block index: #%d, State hash: %s",
			solidState.BlockIndex(), solidState.StateCommitment().String())
	} else if err := sm.createOriginState(); err != nil {
		// create origin state in DB
		go sm.chain.ReceiveMessage(messages.DismissChainMsg{
			Reason: fmt.Sprintf("StateManager.initLoadState. Failed to create origin state: %v", err),
		})
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
		go sm.chain.ReceiveMessage(messages.DismissChainMsg{
			Reason: fmt.Sprintf("StateManager.initLoadState. Failed to create origin state: %v", err),
		},
		)
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
	for {
		select {
		case msg, ok := <-sm.eventGetBlockMsgCh:
			if ok {
				sm.eventGetBlockMsg(msg)
			}
		case msg, ok := <-sm.eventBlockMsgCh:
			if ok {
				sm.eventBlockMsg(msg)
			}
		case msg, ok := <-sm.eventStateOutputMsgCh:
			if ok {
				sm.eventStateMsg(msg)
			}
		case msg, ok := <-sm.eventOutputMsgCh:
			if ok {
				sm.eventOutputMsg(msg)
			}
		case msg, ok := <-sm.eventStateCandidateMsgCh:
			if ok {
				sm.eventStateCandidateMsg(msg)
			}
		case _, ok := <-sm.eventTimerMsgCh:
			if ok {
				sm.eventTimerMsg()
			}
		case <-sm.closeCh:
			return
		}
	}
}
