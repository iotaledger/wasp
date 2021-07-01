// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"bytes"
	"sync"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/committee"
	"github.com/iotaledger/wasp/packages/chain/consensus"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/nodeconnimpl"
	"github.com/iotaledger/wasp/packages/chain/statemgr"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry/committee_record"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
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
	procset               *processors.Cache
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

type committeeStruct struct {
	valid bool
	cmt   chain.Committee
}

func NewChain(
	chainID *chainid.ChainID,
	log *logger.Logger,
	txstreamClient *txstream.Client,
	peerNetConfig coretypes.PeerNetworkConfigProvider,
	db kvstore.KVStore,
	netProvider peering.NetworkProvider,
	dksProvider coretypes.DKShareRegistryProvider,
	committeeRegistry coretypes.CommitteeRegistryProvider,
	blobProvider coretypes.BlobCache,
	processorConfig *processors.Config,
) chain.Chain {
	log.Debugf("creating chain object for %s", chainID.String())

	chainLog := log.Named(chainID.Base58()[:6] + ".")
	chainStateSync := coreutil.NewChainStateSync()
	ret := &chainObj{
		mempool:           mempool.New(state.NewOptimisticStateReader(db, chainStateSync), blobProvider, chainLog),
		procset:           processors.MustNew(processorConfig),
		chMsg:             make(chan interface{}, 100),
		chainID:           *chainID,
		log:               chainLog,
		nodeConn:          nodeconnimpl.New(txstreamClient, chainLog),
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
	ret.committee.Store(&committeeStruct{})
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
		} else if c.consensus != nil {
			c.consensus.EventTimerMsg(msgt / 2)
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
		msgt, err := chain.OffLedgerRequestMsgFromBytes(msg.MsgData)
		if err != nil {
			c.log.Error(err)
			return
		}
		c.ReceiveOffLedgerRequest(msgt.Req)
	case chain.MsgMissingRequestIDs:
		if !parameters.GetBool(parameters.PullMissingRequestsFromCommittee) {
			return
		}
		msgt, err := chain.MissingRequestIDsMsgFromBytes(msg.MsgData)
		if err != nil {
			c.log.Error(err)
			return
		}
		c.SendMissingRequestsToPeer(msgt, msg.SenderNetID)
	case chain.MsgMissingRequest:
		if !parameters.GetBool(parameters.PullMissingRequestsFromCommittee) {
			return
		}
		msgt, err := chain.MissingRequestMsgFromBytes(msg.MsgData)
		if err != nil {
			c.log.Error(err)
			return
		}
		if c.consensus.ShouldReceiveMissingRequest(msgt.Request) {
			c.mempool.ReceiveRequest(msgt.Request)
		}
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
		err = c.rotateCommitteeIfNeeded(msg.ChainOutput, cmt)
	} else {
		err = c.createCommitteeIfNeeded(msg.ChainOutput)
	}
	if err != nil {
		c.log.Errorf("processStateMessage: %v", err)
		return
	}
	c.stateMgr.EventStateMsg(msg)
}

func (c *chainObj) rotateCommitteeIfNeeded(anchorOutput *ledgerstate.AliasOutput, currentCmt chain.Committee) error {
	if currentCmt.Address().Equals(anchorOutput.GetStateAddress()) {
		// nothing changed. no rotation
		return nil
	}
	// address changed
	if !anchorOutput.GetIsGovernanceUpdated() {
		return xerrors.Errorf("rotateCommitteeIfNeeded: inconsistency. Governance transition expected... New output: %s", anchorOutput.String())
	}
	rec, err := c.getOwnCommitteeRecord(anchorOutput.GetStateAddress())
	if err != nil {
		return xerrors.Errorf("rotateCommitteeIfNeeded: %w", err)
	}
	// rotation needed
	// close current in any case
	c.log.Infof("CLOSING COMMITTEE for %s", currentCmt.Address().Base58())

	currentCmt.Close()
	c.consensus.Close()
	c.setCommittee(nil)
	c.consensus = nil
	if rec != nil {
		// create new if committee record is available
		if err = c.createNewCommitteeAndConsensus(rec); err != nil {
			return xerrors.Errorf("rotateCommitteeIfNeeded: creating committee and consensus: %v", err)
		}
	}
	return nil
}

func (c *chainObj) createCommitteeIfNeeded(anchorOutput *ledgerstate.AliasOutput) error {
	// check if I am in the committee
	rec, err := c.getOwnCommitteeRecord(anchorOutput.GetStateAddress())
	if err != nil {
		return xerrors.Errorf("rotateCommitteeIfNeeded: %w", err)
	}
	if rec != nil {
		// create if record is present
		if err = c.createNewCommitteeAndConsensus(rec); err != nil {
			return xerrors.Errorf("rotateCommitteeIfNeeded: creating committee and consensus: %v", err)
		}
	}
	return nil
}

func (c *chainObj) getOwnCommitteeRecord(addr ledgerstate.Address) (*committee_record.CommitteeRecord, error) {
	rec, err := c.committeeRegistry.GetCommitteeRecord(addr)
	if err != nil {
		return nil, xerrors.Errorf("createCommitteeIfNeeded: reading committee record: %v", err)
	}
	if rec == nil {
		// committee record wasn't found in th registry, I am not the part of the committee
		return nil, nil
	}
	// just in case check if I am among committee nodes
	// should not happen
	if !util.StringInList(c.peerNetworkConfig.OwnNetID(), rec.Nodes) {
		return nil, xerrors.Errorf("createCommitteeIfNeeded: I am not among nodes of the committee record. Inconsistency")
	}
	return rec, nil
}

func (c *chainObj) createNewCommitteeAndConsensus(cmtRec *committee_record.CommitteeRecord) error {
	c.log.Debugf("createNewCommitteeAndConsensus: creating a new committee...")
	cmt, err := committee.New(
		cmtRec,
		&c.chainID,
		c.netProvider,
		c.peerNetworkConfig,
		c.dksProvider,
		c.log,
	)
	if err != nil {
		c.setCommittee(nil)
		return xerrors.Errorf("createNewCommitteeAndConsensus: failed to create committee object for state address %s: %w",
			cmtRec.Address.Base58(), err)
	}
	cmt.Attach(c)
	c.log.Debugf("creating new consensus object...")
	c.consensus = consensus.New(c, c.mempool, cmt, c.nodeConn)
	c.setCommittee(cmt)

	c.log.Infof("NEW COMMITTEE OF VALIDATORS has been initialized for the state address %s", cmtRec.Address.Base58())
	return nil
}

func (c *chainObj) getCommittee() chain.Committee {
	ret := c.committee.Load().(*committeeStruct)
	if !ret.valid {
		return nil
	}
	return ret.cmt
}

func (c *chainObj) setCommittee(cmt chain.Committee) {
	if cmt == nil {
		c.committee.Store(&committeeStruct{})
	} else {
		c.committee.Store(&committeeStruct{
			valid: true,
			cmt:   cmt,
		})
	}
}
