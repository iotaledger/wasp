// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized and validated
package statemgr

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/util/ready"
	"go.uber.org/atomic"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
)

type stateManager struct {
	ready                   *ready.Ready
	dbp                     *dbprovider.DBProvider
	chain                   chain.ChainCore
	peers                   chain.PeerGroupProvider
	nodeConn                chain.NodeConnection
	pullStateDeadline       time.Time
	stateCandidates         map[hashing.HashValue]*stateCandidate
	solidState              state.VirtualState // never nil
	stateOutput             *ledgerstate.AliasOutput
	stateOutputTimestamp    time.Time
	currentSyncData         atomic.Value
	notifiedSyncedStateHash hashing.HashValue
	syncingBlocks           map[uint32]*syncingBlock
	log                     *logger.Logger

	// Channels for accepting external events.
	eventGetBlockMsgCh     chan *chain.GetBlockMsg
	eventBlockMsgCh        chan *chain.BlockMsg
	eventStateOutputMsgCh  chan *chain.StateMsg
	eventOutputMsgCh       chan ledgerstate.Output
	eventPendingBlockMsgCh chan chain.StateCandidateMsg
	eventTimerMsgCh        chan chain.TimerTick
	closeCh                chan bool
}

type stateCandidate struct {
	state state.VirtualState
	block state.Block
}

const (
	pullStatePeriod           = 2 * time.Second
	periodBetweenSyncMessages = 1 * time.Second
)

type syncingBlock struct {
	pullDeadline      time.Time
	block             state.Block
	approved          bool
	finalHash         hashing.HashValue
	approvingOutputID ledgerstate.OutputID
}

func New(dbp *dbprovider.DBProvider, c chain.ChainCore, peers chain.PeerGroupProvider, nodeconn chain.NodeConnection, log *logger.Logger) chain.StateManager {
	ret := &stateManager{
		ready:                  ready.New(fmt.Sprintf("state manager %s", c.ID().Base58()[:6]+"..")),
		dbp:                    dbp,
		chain:                  c,
		nodeConn:               nodeconn,
		syncingBlocks:          make(map[uint32]*syncingBlock),
		stateCandidates:        make(map[hashing.HashValue]*stateCandidate),
		log:                    log.Named("s"),
		eventGetBlockMsgCh:     make(chan *chain.GetBlockMsg),
		eventBlockMsgCh:        make(chan *chain.BlockMsg),
		eventStateOutputMsgCh:  make(chan *chain.StateMsg),
		eventOutputMsgCh:       make(chan ledgerstate.Output),
		eventPendingBlockMsgCh: make(chan chain.StateCandidateMsg),
		eventTimerMsgCh:        make(chan chain.TimerTick),
		closeCh:                make(chan bool),
	}
	ret.SetPeers(peers)
	go ret.initLoadState()

	return ret
}

func (sm *stateManager) SetPeers(p chain.PeerGroupProvider) {
	n := uint16(0)
	if p != nil {
		n = p.NumPeers()
		sm.log.Debugf("SetPeers: num = %d", n)
	}
	sm.peers = p
}

func (sm *stateManager) Close() {
	close(sm.closeCh)
}

// initial loading of the solid state
func (sm *stateManager) initLoadState() {
	var err error
	var stateExists bool

	sm.solidState, stateExists, err = state.LoadSolidState(sm.dbp, sm.chain.ID())
	if err != nil {
		go sm.chain.ReceiveMessage(chain.DismissChainMsg{
			Reason: fmt.Sprintf("StateManager.initLoadState: %v", err)},
		)
		return
	}
	if stateExists {
		h := sm.solidState.Hash()
		sm.log.Infof("SOLID STATE has been loaded. Block index: #%d, State hash: %s",
			sm.solidState.BlockIndex(), h.String())
	} else {
		// create origin state in DB
		sm.solidState, err = state.CreateOriginState(sm.dbp, sm.chain.ID())
		if err != nil {
			go sm.chain.ReceiveMessage(chain.DismissChainMsg{
				Reason: fmt.Sprintf("StateManager.initLoadState. Failed to create origin state: %v", err)},
			)
			return
		}
		sm.log.Infof("ORIGIN STATE has been created")
	}
	sm.recvLoop() // Start to process external events.
}

func (sm *stateManager) Ready() *ready.Ready {
	return sm.ready
}

func (sm *stateManager) GetSyncInfo() *chain.SyncInfo {
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
