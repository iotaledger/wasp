// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"bytes"
	"sync"

	"github.com/iotaledger/wasp/packages/offledger"
	"github.com/iotaledger/wasp/packages/registry/chainrecord"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/plugins/gossip"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"

	"github.com/iotaledger/wasp/packages/coretypes/coreutil"

	"github.com/iotaledger/wasp/packages/chain/consensus"
	"github.com/iotaledger/wasp/packages/chain/mempool"

	"github.com/iotaledger/wasp/packages/chain/statemgr"
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
	committee             atomic.Value
	mempool               chain.Mempool
	dismissed             atomic.Bool
	dismissOnce           sync.Once
	chainID               chainid.ChainID
	chainStateSync        coreutil.ChainStateSync
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
	dksProvider           coretypes.DKShareRegistryProvider
	committeeRegistry     coretypes.CommitteeRegistryProvider
	blobProvider          coretypes.BlobCache
	eventRequestProcessed *events.Event
	eventChainTransition  *events.Event
	eventSynced           *events.Event
	peers                 *peering.PeerDomainProvider
}

func NewChain(
	chr *chainrecord.ChainRecord,
	log *logger.Logger,
	txstream *txstream.Client,
	peerNetConfig coretypes.PeerNetworkConfigProvider,
	db kvstore.KVStore,
	netProvider peering.NetworkProvider,
	dksProvider coretypes.DKShareRegistryProvider,
	committeeRegistry coretypes.CommitteeRegistryProvider,
	blobProvider coretypes.BlobCache,
) chain.Chain {
	log.Debugf("creating chain object for %s", chr.ChainID.String())

	chainLog := log.Named(chr.ChainID.Base58()[:6] + ".")
	chainStateSync := coreutil.NewChainStateSync()
	ret := &chainObj{
		mempool:           mempool.New(state.NewOptimisticStateReader(db, chainStateSync), blobProvider, chainLog),
		procset:           processors.MustNew(),
		chMsg:             make(chan interface{}, 100),
		chainID:           *chr.ChainID,
		log:               chainLog,
		nodeConn:          nodeconnimpl.New(txstream, chainLog),
		db:                db,
		chainStateSync:    chainStateSync,
		stateReader:       state.NewOptimisticStateReader(db, chainStateSync),
		peerNetworkConfig: peerNetConfig,
		netProvider:       netProvider,
		dksProvider:       dksProvider,
		committeeRegistry: committeeRegistry,
		blobProvider:      blobProvider,
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ coretypes.RequestID))(params[0].(coretypes.RequestID))
		}),
		eventChainTransition: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ *chain.ChainTransitionEventData))(params[0].(*chain.ChainTransitionEventData))
		}),
		eventSynced: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(outputID ledgerstate.OutputID, blockIndex uint32))(params[0].(ledgerstate.OutputID), params[1].(uint32))
		}),
	}
	ret.eventChainTransition.Attach(events.NewClosure(ret.processChainTransition))
	ret.eventSynced.Attach(events.NewClosure(ret.processSynced))

	peers, err := netProvider.PeerDomain(peerNetConfig.Neighbors())
	if err != nil {
		log.Errorf("NewChain: %v", err)
		return nil
	}
	ret.stateMgr = statemgr.New(db, ret, peers, ret.nodeConn)
	ret.peers = &peers
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
			c.log.Debugf("mempool total = %d, ready = %d, in = %d, out = %d", stats.TotalPool, stats.Ready, stats.InPoolCounter, stats.OutPoolCounter)
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
	case chain.MsgOffLedgerRequest:
		msgt, err := offledger.OffLedgerRequestMsgFromBytes(msg.MsgData)
		if err != nil {
			c.log.Error(err)
			return
		}
		gossip.Gossip().ProcessOffLedgerRequest(msgt.ChainID, msgt.Req)
		return

	default:
		c.log.Errorf("processPeerMessage: wrong msg type")
	}
}

// processChainTransition processes the unique chain output which exists on the chain's address
// If necessary, it creates/changes/rotates committee object
func (c *chainObj) processChainTransition(msg *chain.ChainTransitionEventData) {
	if !msg.ChainOutput.GetIsGovernanceUpdated() {
		// normal state update:
		c.stateReader.SetBaseline()
		reqids, err := blocklog.GetRequestIDsForLastBlock(c.stateReader)
		if err != nil {
			// The error means a database error. The optimistic state read failure can't occur here
			// because the state transition message is only sent only after state is committed and before consensus
			// start new round
			c.log.Panicf("processChainTransition. unexpected error: %v", err)
		}
		// remove processed requests from the mempool
		c.mempool.RemoveRequests(reqids...)
		// publish events
		chain.LogStateTransition(msg, reqids, c.log)
		chain.PublishStateTransition(msg.ChainOutput, reqids)
		for _, reqid := range reqids {
			c.eventRequestProcessed.Trigger(reqid)
		}

		c.log.Debugf("processChainTransition (state): state index: %d, state hash: %s, requests: %+v",
			msg.VirtualState.BlockIndex(), msg.VirtualState.Hash().String(), coretypes.ShortRequestIDs(reqids))
	} else {
		chain.LogGovernanceTransition(msg, c.log)
		chain.PublishGovernanceTransition(msg.ChainOutput)

		c.log.Debugf("processChainTransition (rotate): state index: %d, state hash: %s",
			msg.VirtualState.BlockIndex(), msg.VirtualState.Hash().String())
	}
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

// processStateMessage processes the only chain output which exists on the chain's address
// If necessary, it creates/changes/rotates committee object
func (c *chainObj) processStateMessage(msg *chain.StateMsg) {
	sh, err := hashing.HashValueFromBytes(msg.ChainOutput.GetStateData())
	if err != nil {
		c.log.Error(xerrors.Errorf("parsing state hash: %w", err))
		return
	}
	c.log.Debugf("processStateMessage. stateIndex: %d, stateHash: %d, stateAddr: %d, state transition: %v",
		msg.ChainOutput.GetStateIndex(), sh.String(),
		msg.ChainOutput.GetStateAddress().Base58(), !msg.ChainOutput.GetIsGovernanceUpdated(),
	)
	cmt := c.getCommittee()
	if cmt != nil {
		// committee already exists
		if msg.ChainOutput.GetIsGovernanceUpdated() &&
			!cmt.Address().Equals(msg.ChainOutput.GetStateAddress()) {
			// governance transition. Committee needs to be rotated
			// close current committee and consensus
			cmt.Close()
			c.consensus.Close()
			c.setCommittee(nil)
			c.consensus = nil
			err = c.createNewCommitteeAndConsensus(msg.ChainOutput.GetStateAddress())
		}
	} else {
		// committee does not exist yet. Must be created
		err = c.createNewCommitteeAndConsensus(msg.ChainOutput.GetStateAddress())
	}
	if err != nil {
		c.log.Errorf("processStateMessage: %v", err)
		return
	}
	c.stateMgr.EventStateMsg(msg)
}

func (c *chainObj) createNewCommitteeAndConsensus(addr ledgerstate.Address) error {
	c.log.Debugf("creating new committee...")
	cmt, err := committee.New(
		addr,
		&c.chainID,
		c.netProvider,
		c.peerNetworkConfig,
		c.dksProvider,
		c.committeeRegistry,
		c.log,
	)
	if err != nil {
		c.setCommittee(nil)
		return xerrors.Errorf("failed to create committee object for state address %s: %w", addr.Base58(), err)
	}
	cmt.Attach(c)
	c.log.Debugf("creating new consensus object...")
	c.consensus = consensus.New(c, c.mempool, cmt, c.nodeConn)
	c.setCommittee(cmt)

	c.log.Infof("NEW COMMITTEE OF VALiDATORS has been initialized for the state address %s", addr.Base58())
	return nil
}

func (c *chainObj) getCommittee() chain.Committee {
	ret := c.committee.Load()
	if ret == nil {
		return nil
	}
	return ret.(chain.Committee)
}

func (c *chainObj) setCommittee(cmt chain.Committee) {
	c.committee.Store(cmt)
}
