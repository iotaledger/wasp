// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/state"
	"sync"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/committeeimpl"
	"github.com/iotaledger/wasp/packages/chain/consensusimpl"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/nodeconnimpl"
	"github.com/iotaledger/wasp/packages/chain/statemgr"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
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
	procset               *processors.ProcessorCache
	chMsg                 chan interface{}
	stateMgr              chain.StateManager
	consensus             chain.Consensus
	log                   *logger.Logger
	nodeConn              *txstream.Client
	dbProvider            *dbprovider.DBProvider
	netProvider           peering.NetworkProvider
	dksProvider           tcrypto.RegistryProvider
	blobProvider          coretypes.BlobCache
	eventRequestProcessed *events.Event
	eventStateTransition  *events.Event
	eventSynced           *events.Event
}

func NewChain(
	chr *registry.ChainRecord,
	log *logger.Logger,
	nodeConn *txstream.Client,
	dbProvider *dbprovider.DBProvider,
	netProvider peering.NetworkProvider,
	dksProvider tcrypto.RegistryProvider,
	blobProvider coretypes.BlobCache,
) chain.Chain {
	log.Debugf("creating chain object for %s", chr.ChainID.String())

	chainLog := log.Named(chr.ChainID.Base58()[:6] + ".")
	stateReader, err := state.NewStateReader(dbProvider, chr.ChainID)
	if err != nil {
		log.Errorf("NewChain: %v", err)
		return nil
	}
	ret := &chainObj{
		mempool:      mempool.New(stateReader, blobProvider, chainLog),
		procset:      processors.MustNew(),
		chMsg:        make(chan interface{}, 100),
		chainID:      *chr.ChainID,
		log:          chainLog,
		nodeConn:     nodeConn,
		dbProvider:   dbProvider,
		netProvider:  netProvider,
		dksProvider:  dksProvider,
		blobProvider: blobProvider,
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

	ret.stateMgr = statemgr.New(dbProvider, ret, newPeers(nil), nodeconnimpl.New(ret.nodeConn), ret.log)
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
	case chain.StateCandidateMsg:
		c.stateMgr.EventStateCandidateMsg(msgt)
	case *chain.InclusionStateMsg:
		if c.consensus != nil {
			c.consensus.EventTransactionInclusionStateMsg(msgt)
		}
	case *chain.StateMsg:
		c.processStateMessage(msgt)
	case *chain.VMResultMsg:
		// VM finished working
		if c.consensus != nil {
			c.consensus.EventResultCalculated(msgt)
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
			total, withMsg, solid := c.mempool.Stats()
			c.log.Debugf("mempool total = %d, withMsg = %d solid = %d", total, withMsg, solid)
		}
	}
}

func (c *chainObj) processPeerMessage(msg *peering.PeerMessage) {
	rdr := bytes.NewReader(msg.MsgData)

	switch msg.MsgType {

	case chain.MsgNotifyRequests:
		msgt := &chain.NotifyReqMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}

		msgt.SenderIndex = msg.SenderIndex

		if c.consensus != nil {
			c.consensus.EventNotifyReqMsg(msgt)
		}

	case chain.MsgNotifyFinalResultPosted:
		msgt := &chain.NotifyFinalResultPostedMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}

		msgt.SenderIndex = msg.SenderIndex

		if c.consensus != nil {
			c.consensus.EventNotifyFinalResultPostedMsg(msgt)
		}

	case chain.MsgStartProcessingRequest:
		msgt := &chain.StartProcessingBatchMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}

		msgt.SenderIndex = msg.SenderIndex
		msgt.Timestamp = msg.Timestamp

		if c.consensus != nil {
			c.consensus.EventStartProcessingBatchMsg(msgt)
		}

	case chain.MsgSignedHash:
		msgt := &chain.SignedHashMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}

		msgt.SenderIndex = msg.SenderIndex
		msgt.Timestamp = msg.Timestamp

		if c.consensus != nil {
			c.consensus.EventSignedHashMsg(msgt)
		}

	case chain.MsgGetBlock:
		msgt := &chain.GetBlockMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}

		msgt.SenderIndex = msg.SenderIndex

		c.stateMgr.EventGetBlockMsg(msgt)

	case chain.MsgBlock:
		msgt := &chain.BlockMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}

		msgt.SenderIndex = msg.SenderIndex
		c.stateMgr.EventBlockMsg(msgt)

	default:
		c.log.Errorf("processPeerMessage: wrong msg type")
	}
}

// processStateMessage processes chain output
// If necessary, it creates/changes committee object and sends new peers to the stateManager
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
	if c.committee != nil && c.committee.DKShare().Address.Equals(msg.ChainOutput.GetStateAddress()) {
		// nothing changed in the committee
		c.stateMgr.EventStateMsg(msg)
		return
	}
	// create or change committee object
	if c.committee != nil {
		c.committee.Close()
	}
	if c.consensus != nil {
		c.consensus.Close()
	}
	c.committee, c.consensus = nil, nil
	c.log.Debugf("creating new committee...")
	if c.committee, err = committeeimpl.NewCommittee(msg.ChainOutput.GetStateAddress(), c.netProvider, c.dksProvider, c.log); err != nil {
		c.committee = nil
		c.log.Errorf("failed to create committee object for address %s: %v", msg.ChainOutput.GetStateAddress(), err)
		return
	}
	c.committee.OnPeerMessage(func(recv *peering.RecvEvent) {
		c.ReceiveMessage(recv.Msg)
	})
	c.log.Debugf("created new committee for state address %s", msg.ChainOutput.GetStateAddress().Base58())

	c.log.Debugf("creating new consensus object..")
	c.consensus = consensusimpl.New(c, c.mempool, c.committee, nodeconnimpl.New(c.nodeConn), c.log)
	c.stateMgr.SetPeers(newPeers(c.committee))
	c.stateMgr.EventStateMsg(msg)
}

func (c *chainObj) processStateTransition(msg *chain.StateTransitionEventData) {
	chain.LogStateTransition(msg, c.log)
	reqids := chain.PublishStateTransition(msg.VirtualState, msg.ChainOutput)
	for _, reqid := range reqids {
		c.eventRequestProcessed.Trigger(reqid)
	}
	c.mempool.RemoveRequests(reqids...)

	// send to consensus
	c.ReceiveMessage(&chain.StateTransitionMsg{
		VariableState: msg.VirtualState,
		ChainOutput:   msg.ChainOutput,
		Timestamp:     msg.OutputTimestamp,
	})
}

func (c *chainObj) processSynced(outputID ledgerstate.OutputID, blockIndex uint32) {
	chain.LogSyncedEvent(outputID, blockIndex, c.log)
}
