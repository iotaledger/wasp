// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"bytes"
	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"
	"github.com/iotaledger/wasp/packages/chain/consensus"
	"sync"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/vm/processors"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/statemgr"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util"
	"go.uber.org/atomic"
)

type chainObj struct {
	committee                    chain.Committee
	mempool                      chain.Mempool
	isReadyStateManager          bool
	isReadyConsensus             bool
	isConnectPeriodOver          bool
	isQuorumOfConnectionsReached bool
	mutexIsReady                 sync.Mutex
	isOpenQueue                  atomic.Bool
	dismissed                    atomic.Bool
	dismissOnce                  sync.Once
	onActivation                 func()
	chainID                      coretypes.ChainID
	procset                      *processors.ProcessorCache
	chMsg                        chan interface{}
	stateMgr                     chain.StateManager
	consensus                    chain.Consensus
	eventRequestProcessed        *events.Event
	log                          *logger.Logger
	nodeConn                     *txstream.Client
	netProvider                  peering.NetworkProvider
	dksProvider                  tcrypto.RegistryProvider
	blobProvider                 coretypes.BlobCache
}

func newChainObj(
	chr *registry.ChainRecord,
	log *logger.Logger,
	nodeConn *txstream.Client,
	netProvider peering.NetworkProvider,
	dksProvider tcrypto.RegistryProvider,
	blobProvider coretypes.BlobCache,
	onActivation func(),
) chain.Chain {
	log.Debugf("creating chain: %s", chr)

	chainLog := log.Named(util.Short(chr.ChainID.String()))
	ret := &chainObj{
		mempool:      chain.NewMempool(blobProvider),
		procset:      processors.MustNew(),
		chMsg:        make(chan interface{}, 100),
		chainID:      chr.ChainID,
		onActivation: onActivation,
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ coretypes.RequestID))(params[0].(coretypes.RequestID))
		}),
		log:          chainLog,
		nodeConn:     nodeConn,
		netProvider:  netProvider,
		dksProvider:  dksProvider,
		blobProvider: blobProvider,
	}
	ret.stateMgr = statemgr.New(ret, newPeers(nil), ret.log)
	go func() {
		for msg := range ret.chMsg {
			ret.dispatchMessage(msg)
		}
	}()
	return ret
}

// iAmInTheCommittee checks if NetIDs makes sense
func iAmInTheCommittee(committeeNodes []string, n, index uint16, netProvider peering.NetworkProvider) bool {
	if len(committeeNodes) != int(n) {
		return false
	}
	return committeeNodes[index] == netProvider.Self().NetID()
}

func (c *chainObj) dispatchMessage(msg interface{}) {
	if !c.isOpenQueue.Load() {
		return
	}

	switch msgt := msg.(type) {
	case *peering.PeerMessage:
		c.processPeerMessage(msgt)
	case *chain.StateUpdateMsg:
		c.stateMgr.EventStateUpdateMsg(msgt)
	case *chain.StateTransitionMsg:
		if c.consensus != nil {
			c.consensus.EventStateTransitionMsg(msgt)
		}
	case chain.PendingBlockMsg:
		c.stateMgr.EventPendingBlockMsg(msgt)
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
			if c.stateMgr != nil {
				c.stateMgr.EventTimerMsg(msgt / 2)
			}
		} else {
			if c.consensus != nil {
				c.consensus.EventTimerMsg(msgt / 2)
			}
		}
	}
}

func (c *chainObj) processPeerMessage(msg *peering.PeerMessage) {

	rdr := bytes.NewReader(msg.MsgData)

	switch msg.MsgType {

	case chain.MsgStateIndexPingPong:
		msgt := &chain.StateIndexPingPongMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex

		c.stateMgr.EvidenceStateIndex(msgt.BlockIndex)
		c.stateMgr.EventStateIndexPingPongMsg(msgt)

	case chain.MsgNotifyRequests:
		msgt := &chain.NotifyReqMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.EvidenceStateIndex(msgt.BlockIndex)

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
		c.stateMgr.EvidenceStateIndex(msgt.BlockIndex)

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
		c.stateMgr.EvidenceStateIndex(msgt.BlockIndex)

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
		c.stateMgr.EvidenceStateIndex(msgt.BlockIndex)

		msgt.SenderIndex = msg.SenderIndex
		msgt.Timestamp = msg.Timestamp

		if c.consensus != nil {
			c.consensus.EventSignedHashMsg(msgt)
		}

	case chain.MsgGetBatch:
		msgt := &chain.GetBlockMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}

		msgt.SenderIndex = msg.SenderIndex

		c.stateMgr.EventGetBlockMsg(msgt)

	case chain.MsgBatchHeader:
		msgt := &chain.BlockHeaderMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.EvidenceStateIndex(msgt.BlockIndex)

		msgt.SenderIndex = msg.SenderIndex
		c.stateMgr.EventBlockHeaderMsg(msgt)

	case chain.MsgStateUpdate:
		msgt := &chain.StateUpdateMsg{}
		if err := msgt.Read(rdr); err != nil {
			c.log.Error(err)
			return
		}
		c.stateMgr.EvidenceStateIndex(msgt.BlockIndex)

		msgt.SenderIndex = msg.SenderIndex
		c.stateMgr.EventStateUpdateMsg(msgt)

	default:
		c.log.Errorf("processPeerMessage: wrong msg type")
	}
}

// processStateMessage processes chain output
// If necessary, it creates/changes committee object and sends new peers to the stateManager
func (c *chainObj) processStateMessage(msg *chain.StateMsg) {
	if c.committee != nil && c.committee.DKShare().Address.Equals(msg.ChainOutput.GetStateAddress()) {
		// nothing changed in the committee
		c.stateMgr.EventStateOutputMsg(msg)
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
	var err error
	if c.committee, err = NewCommittee(c, msg.ChainOutput.GetStateAddress(), c.netProvider, c.dksProvider, c.log); err != nil {
		c.committee = nil
		c.log.Errorf("failed to create committee object for address %s. ChainID: %s",
			msg.ChainOutput.GetStateAddress(), c.chainID)
		return
	}
	c.committee.OnPeerMessage(func(recv *peering.RecvEvent) {
		c.ReceiveMessage(recv.Msg)
	})
	c.consensus = consensus.New(c.mempool, c.committee, c.nodeConn, c.log)
	c.stateMgr.SetPeers(newPeers(c.committee))
}
