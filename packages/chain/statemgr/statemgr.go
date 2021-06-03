// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized and validated
package statemgr

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util/ready"
	"go.uber.org/atomic"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
)

type stateManager struct {
	ready                   *ready.Ready
	store                   kvstore.KVStore
	chain                   chain.ChainCore
	peers                   peering.PeerDomainProvider
	nodeConn                chain.NodeConnection
	pullStateRetryTime      time.Time
	solidState              state.VirtualState
	stateOutput             *ledgerstate.AliasOutput
	stateOutputTimestamp    time.Time
	currentSyncData         atomic.Value
	notifiedSyncedStateHash hashing.HashValue
	syncingBlocks           *syncingBlocks
	timers                  Timers
	log                     *logger.Logger

	// Channels for accepting external events.
	eventGetBlockMsgCh     chan *chain.GetBlockMsg
	eventBlockMsgCh        chan *chain.BlockMsg
	eventStateOutputMsgCh  chan *chain.StateMsg
	eventOutputMsgCh       chan ledgerstate.Output
	eventPendingBlockMsgCh chan *chain.StateCandidateMsg
	eventTimerMsgCh        chan chain.TimerTick
	closeCh                chan bool
}

const (
	numberOfNodesToRequestBlockFromConst = 5
	maxBlocksToCommitConst               = 10000 //10k
)

func New(store kvstore.KVStore, c chain.ChainCore, peers peering.PeerDomainProvider, nodeconn chain.NodeConnection, log *logger.Logger, timersOpt ...Timers) chain.StateManager {
	var timers Timers
	if len(timersOpt) > 0 {
		timers = timersOpt[0]
	} else {
		timers = Timers{}
	}
	ret := &stateManager{
		ready:                  ready.New(fmt.Sprintf("state manager %s", c.ID().Base58()[:6]+"..")),
		store:                  store,
		chain:                  c,
		nodeConn:               nodeconn,
		peers:                  peers,
		syncingBlocks:          newSyncingBlocks(log, timers.getGetBlockRetry()),
		timers:                 timers,
		log:                    log.Named("s"),
		pullStateRetryTime:     time.Now(),
		eventGetBlockMsgCh:     make(chan *chain.GetBlockMsg),
		eventBlockMsgCh:        make(chan *chain.BlockMsg),
		eventStateOutputMsgCh:  make(chan *chain.StateMsg),
		eventOutputMsgCh:       make(chan ledgerstate.Output),
		eventPendingBlockMsgCh: make(chan *chain.StateCandidateMsg),
		eventTimerMsgCh:        make(chan chain.TimerTick),
		closeCh:                make(chan bool),
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
		go sm.chain.ReceiveMessage(chain.DismissChainMsg{
			Reason: fmt.Sprintf("StateManager.initLoadState: %v", err)},
		)
		return
	}
	if stateExists {
		sm.log.Infof("SOLID STATE has been loaded. Block index: #%d, State hash: %s",
			solidState.BlockIndex(), solidState.Hash().String())
	} else {
		// create origin state in DB
		if err := sm.createOriginState(); err != nil {
			go sm.chain.ReceiveMessage(chain.DismissChainMsg{
				Reason: fmt.Sprintf("StateManager.initLoadState. Failed to create origin state: %v", err)},
			)
			return
		}
	}
	sm.recvLoop() // Check to process external events.
}

func (sm *stateManager) createOriginState() error {
	sm.chain.GlobalSync().InvalidateSolidIndex()

	sm.chain.GlobalSync().Mutex().Lock()
	defer sm.chain.GlobalSync().Mutex().Unlock()

	var err error
	sm.solidState, err = state.CreateOriginState(sm.store, sm.chain.ID())
	if err != nil {
		go sm.chain.ReceiveMessage(chain.DismissChainMsg{
			Reason: fmt.Sprintf("StateManager.initLoadState. Failed to create origin state: %v", err)},
		)
		return err
	}
	sm.chain.GlobalSync().SetSolidIndex(0)
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
		case msg, ok := <-sm.eventPendingBlockMsgCh:
			if ok {
				sm.eventStateCandidateMsg(msg)
			}
		case msg, ok := <-sm.eventTimerMsgCh:
			if ok {
				sm.eventTimerMsg(msg)
			}
		case <-sm.closeCh:
			return
		}
	}
}
