// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized and validated
package statemgr

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
)

type stateManager struct {
	chain    chain.Chain
	peers    chain.PeerGroupProvider
	nodeConn *txstream.Client

	// flag pingPong[idx] if ping-pong message was received from the peer idx
	pingPong              []bool
	deadlineForPongQuorum time.Time

	pullStateDeadline time.Time

	// pending batches of state updates are candidates to confirmation by the state transaction
	// which leads to the state transition
	// the map key is hash of the variable state which is a result of applying the
	// block of state updates to the solid variable state
	pendingBlocks map[hashing.HashValue]*pendingBlock

	// last variable state stored in the database
	// it may be nil at bootstrap when origin variable state is calculated
	solidState state.VirtualState

	// state transaction with +1 state index from the state index of solid variable state
	// it may be nil if does not exist or not fetched yet
	stateOutput          *ledgerstate.AliasOutput
	stateOutputTimestamp time.Time

	syncedBlocks map[uint32]*syncingBlock

	// was state transition message of the current state sent to the consensus operator
	consensusNotifiedOnStateTransition bool

	// logger
	log *logger.Logger

	// Channels for accepting external events.
	eventStateIndexPingPongMsgCh chan *chain.BlockIndexPingPongMsg
	eventGetBlockMsgCh           chan *chain.GetBlockMsg
	eventBlockHeaderMsgCh        chan *chain.BlockHeaderMsg
	eventStateUpdateMsgCh        chan *chain.StateUpdateMsg
	eventStateOutputMsgCh        chan *chain.StateMsg
	eventPendingBlockMsgCh       chan chain.PendingBlockMsg
	eventTimerMsgCh              chan chain.TimerTick
	closeCh                      chan bool
}

const (
	pullStateTimeout          = 5 * time.Second
	periodBetweenSyncMessages = 1 * time.Second
)

type syncingBlock struct {
	msgCounter    uint16
	stateUpdates  []state.StateUpdate
	stateOutputID ledgerstate.OutputID
	pullDeadline  time.Time
	block         state.Block
}

type pendingBlock struct {
	// block of state updates, not validated yet
	block state.Block
	// resulting variable state after applied the block to the solidState
	nextState state.VirtualState
}

func New(c chain.Chain, peers chain.PeerGroupProvider, nodeconn *txstream.Client, log *logger.Logger) chain.StateManager {
	ret := &stateManager{
		chain:                        c,
		nodeConn:                     nodeconn,
		syncedBlocks:                 make(map[uint32]*syncingBlock),
		pendingBlocks:                make(map[hashing.HashValue]*pendingBlock),
		log:                          log.Named("s"),
		eventStateIndexPingPongMsgCh: make(chan *chain.BlockIndexPingPongMsg),
		eventGetBlockMsgCh:           make(chan *chain.GetBlockMsg),
		eventBlockHeaderMsgCh:        make(chan *chain.BlockHeaderMsg),
		eventStateUpdateMsgCh:        make(chan *chain.StateUpdateMsg),
		eventStateOutputMsgCh:        make(chan *chain.StateMsg),
		eventPendingBlockMsgCh:       make(chan chain.PendingBlockMsg),
		eventTimerMsgCh:              make(chan chain.TimerTick),
		closeCh:                      make(chan bool),
	}
	ret.SetPeers(peers)
	go ret.initLoadState()

	return ret
}

func (sm *stateManager) SetPeers(p chain.PeerGroupProvider) {
	if p != nil {
		sm.log.Debugf("SetPeers: num = %d", p.NumPeers())
	}
	sm.peers = p
	sm.pingPong = make([]bool, p.NumPeers())
}

func (sm *stateManager) Close() {
	close(sm.closeCh)
}

// initial loading of the solid state
func (sm *stateManager) initLoadState() {
	var err error
	var batch state.Block
	var stateExists bool

	sm.solidState, batch, stateExists, err = state.LoadSolidState(sm.chain.ID())
	if err != nil {
		sm.log.Errorf("initLoadState: %v", err)
		sm.chain.Dismiss()
		return
	}
	if stateExists {
		h := sm.solidState.Hash()
		txh := batch.ApprovingOutputID()
		sm.log.Infof("solid state has been loaded. Block index: $%d, State hash: %s, ancor tx: %s",
			sm.solidState.BlockIndex(), h.String(), txh.String())
	} else {
		sm.solidState = nil
		sm.addPendingBlock(state.MustNewOriginBlock(ledgerstate.OutputID{}))
		sm.log.Info("solid state does not exist: WAITING FOR THE ORIGIN TRANSACTION")
	}
	sm.recvLoop() // Start to process external events.
}

func (sm *stateManager) recvLoop() {
	for {
		select {
		case msg, ok := <-sm.eventStateIndexPingPongMsgCh:
			if ok {
				sm.eventStateIndexPingPongMsg(msg)
			}
		case msg, ok := <-sm.eventGetBlockMsgCh:
			if ok {
				sm.eventGetBlockMsg(msg)
			}
		case msg, ok := <-sm.eventBlockHeaderMsgCh:
			if ok {
				sm.eventBlockHeaderMsg(msg)
			}
		case msg, ok := <-sm.eventStateUpdateMsgCh:
			if ok {
				sm.eventStateUpdateMsg(msg)
			}
		case msg, ok := <-sm.eventStateOutputMsgCh:
			if ok {
				sm.eventStateMsg(msg)
			}
		case msg, ok := <-sm.eventPendingBlockMsgCh:
			if ok {
				sm.eventPendingBlockMsg(msg)
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
