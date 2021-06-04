// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"bytes"
	"sync"

	"github.com/iotaledger/wasp/packages/coretypes/coreutil"

	"github.com/iotaledger/wasp/packages/chain/consensus"
	"github.com/iotaledger/wasp/packages/registry"

	"github.com/iotaledger/wasp/packages/chain/mempool"

	"github.com/iotaledger/wasp/packages/chain/statemgr"
	"github.com/iotaledger/wasp/packages/registry/chainrecord"

	"github.com/iotaledger/wasp/packages/state"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/committee"
	"github.com/iotaledger/wasp/packages/chain/nodeconnimpl"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"go.uber.org/atomic"
	"golang.org/x/xerrors"
)

type chainObj struct {
	committee             chain.Committee
	mempool               chain.Mempool
	dismissed             atomic.Bool
	dismissOnce           sync.Once
	chainID               coretypes.ChainID
	globalSync            coreutil.GlobalSync
	stateReader           state.OptimisticStateReader
	procset               *processors.ProcessorCache
	chMsg                 chan interface{}
	stateMgr              chain.StateManager
	consensus             chain.Consensus
	log                   *logger.Logger
	nodeConn              chain.NodeConnection
	db                    kvstore.KVStore
	peerNetworkConfig     coretypes.PeerNetworkConfigProvider
	netProvider           peering.NetworkProvider
	dksProvider           registry.DKShareRegistryProvider
	committeeRegistry     registry.CommitteeRegistryProvider
	blobProvider          coretypes.BlobCache
	eventRequestProcessed *events.Event
	eventStateTransition  *events.Event
	eventSynced           *events.Event
}

func NewChain(
	chr *chainrecord.ChainRecord,
	log *logger.Logger,
	txstream *txstream.Client,
	peerNetConfig coretypes.PeerNetworkConfigProvider,
	db kvstore.KVStore,
	netProvider peering.NetworkProvider,
	dksProvider registry.DKShareRegistryProvider,
	committeeRegistry registry.CommitteeRegistryProvider,
	blobProvider coretypes.BlobCache,
) chain.Chain {
	log.Debugf("creating chain object for %s", chr.ChainID.String())

	chainLog := log.Named(chr.ChainID.Base58()[:6] + ".")
	globalSync := coreutil.NewGlobalSync()
	ret := &chainObj{
		mempool:           mempool.New(state.NewOptimisticStateReader(db, globalSync), blobProvider, chainLog),
		procset:           processors.MustNew(),
		chMsg:             make(chan interface{}, 100),
		chainID:           *chr.ChainID,
		log:               chainLog,
		nodeConn:          nodeconnimpl.New(txstream, chainLog),
		db:                db,
		globalSync:        globalSync,
		stateReader:       state.NewOptimisticStateReader(db, globalSync),
		peerNetworkConfig: peerNetConfig,
		netProvider:       netProvider,
		dksProvider:       dksProvider,
		committeeRegistry: committeeRegistry,
		blobProvider:      blobProvider,
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ coretypes.RequestID))(params[0].(coretypes.RequestID))
		}),
		eventStateTransition: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ *chain.StateTransitionEventData))(params[0].(*chain.StateTransitionEventData))
		}),
		eventSynced: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(outputID ledgerstate.OutputID, blockIndex uint32))(params[0].(ledgerstate.OutputID), params[1].(uint32))
		}),
	}
	ret.eventStateTransition.Attach(events.NewClosure(ret.processStateTransition))
	ret.eventSynced.Attach(events.NewClosure(ret.processSynced))

	peers, err := netProvider.PeerDomain(peerNetConfig.Neighbors())
	if err != nil {
		log.Errorf("NewChain: %v", err)
		return nil
	}
	ret.stateMgr = statemgr.New(db, ret, peers, ret.nodeConn, ret.log)
	var peeringID peering.PeeringID = ret.chainID.Array()
	peers.Attach(&peeringID, func(recv *peering.RecvEvent) {
		ret.ReceiveMessage(recv.Msg)
	})
	go func() {
		for msg := range ret.chMsg {
			ret.dispatchMessage(msg)
		}
	}()
	ret.startTimer()
	return ret
}

func (c *chainObj) dispatchMessage(msg interface{}) {
	switch msgt := msg.(type) {
	case *peering.PeerMessage:
		c.processPeerMessage(msgt)
	case *chain.DismissChainMsg:
		c.Dismiss(msgt.Reason)
	case *chain.StateTransitionMsg:
		if c.consensus != nil {
			c.consensus.EventStateTransitionMsg(msgt)
		}
	case *chain.StateCandidateMsg:
		c.stateMgr.EventStateCandidateMsg(msgt)
	case *chain.InclusionStateMsg:
		if c.consensus != nil {
			c.consensus.EventInclusionsStateMsg(msgt)
		}
	case *chain.StateMsg:
		c.processStateMessage(msgt)
	case *chain.VMResultMsg:
		// VM finished working
		if c.consensus != nil {
			c.consensus.EventVMResultMsg(msgt)
		}
	case *chain.AsynchronousCommonSubsetMsg:
		if c.consensus != nil {
			c.consensus.EventAsynchronousCommonSubsetMsg(msgt)
		}
	case chain.TimerTick:
		if msgt%2 == 0 {
			c.stateMgr.EventTimerMsg(msgt / 2)
		} else {
			if c.consensus != nil {
				c.consensus.EventTimerMsg(msgt / 2)
			}
		}
		if msgt%40 == 0 {
			stats := c.mempool.Stats()
			c.log.Debugf("mempool total = %d, ready = %d, in = %d, out = %d", stats.Total, stats.Ready, stats.InCounter, stats.OutCounter)
		}
	}
}

func (c *chainObj) processPeerMessage(msg *peering.PeerMessage) {
	rdr := bytes.NewReader(msg.MsgData)

	switch msg.MsgType {

	case chain.MsgGetBlock:
		msgt := &chain.GetBlockMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		msgt.SenderNetID = msg.SenderNetID
		c.stateMgr.EventGetBlockMsg(msgt)

	case chain.MsgBlock:
		msgt := &chain.BlockMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		msgt.SenderNetID = msg.SenderNetID
		c.stateMgr.EventBlockMsg(msgt)

	case chain.MsgSignedResult:
		msgt := &chain.SignedResultMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex
		if c.consensus != nil {
			c.consensus.EventSignedResultMsg(msgt)
		}

	default:
		c.log.Errorf("processPeerMessage: wrong msg type")
	}
}

// processStateMessage processes the only chain output which exists on the chain's address
// If necessary, it creates/changes/rotates committee object
func (c *chainObj) processStateMessage(msg *chain.StateMsg) {
	sh, err := hashing.HashValueFromBytes(msg.ChainOutput.GetStateData())
	if err != nil {
		c.log.Error(xerrors.Errorf("parsing state hash: %w", err))
		return
	}
	c.log.Debugw("processStateMessage",
		"stateIndex", msg.ChainOutput.GetStateIndex(),
		"stateHash", sh.String(),
		"stateAddr", msg.ChainOutput.GetStateAddress().Base58(),
	)
	if c.committee != nil && c.committee.Address().Equals(msg.ChainOutput.GetStateAddress()) {
		// nothing changed in the committee, just pass the message to state manager
		c.stateMgr.EventStateMsg(msg)
		return
	}
	// create or change committee object
	if c.committee != nil {
		// closes the current committee
		c.committee.Close()
	}
	if c.consensus != nil {
		// closes the current consensus object. All ongoing communications between current validators are interrupted
		c.consensus.Close()
	}
	c.committee, c.consensus = nil, nil
	c.log.Debugf("creating new committee...")

	c.committee, err = committee.New(
		msg.ChainOutput.GetStateAddress(),
		&c.chainID,
		c.netProvider,
		c.peerNetworkConfig,
		c.dksProvider,
		c.committeeRegistry,
		c.log,
	)
	if err != nil {
		c.committee = nil
		c.log.Errorf("failed to create committee object for state address %s: %v", msg.ChainOutput.GetStateAddress().Base58(), err)
		return
	}
	c.committee.Attach(c)
	c.log.Debugf("creating new consensus object...")
	c.consensus = consensus.New(c, c.mempool, c.committee, c.nodeConn, c.log)

	c.log.Infof("NEW COMMITTEE OF VALDATORS initialized for state address %s", msg.ChainOutput.GetStateAddress().Base58())
	c.stateMgr.EventStateMsg(msg)
}

func (c *chainObj) processStateTransition(msg *chain.StateTransitionEventData) {
	chain.LogStateTransition(msg, c.log)
	reqids := chain.PublishStateTransition(msg.VirtualState, msg.ChainOutput, msg.RequestIDs)
	for _, reqid := range reqids {
		c.eventRequestProcessed.Trigger(reqid)
	}
	c.mempool.RemoveRequests(reqids...)

	// send to consensus
	c.ReceiveMessage(&chain.StateTransitionMsg{
		State:          msg.VirtualState,
		StateOutput:    msg.ChainOutput,
		StateTimestamp: msg.OutputTimestamp,
	})
}

func (c *chainObj) processSynced(outputID ledgerstate.OutputID, blockIndex uint32) {
	chain.LogSyncedEvent(outputID, blockIndex, c.log)
}
