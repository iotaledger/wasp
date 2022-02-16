// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/chain/nodeconnimpl"
	"github.com/iotaledger/wasp/packages/chain/statemgr"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"go.uber.org/atomic"
)

const maxMsgBuffer = 1000

var (
	_ chain.Chain                = &chainObj{}
	_ chain.ChainCore            = &chainObj{}
	_ chain.ChainEntry           = &chainObj{}
	_ chain.ChainRequests        = &chainObj{}
	_ chain.ChainMetrics         = &chainObj{}
	_ map[ed25519.PublicKey]bool // We rely on value comparison on the pubkeys, just assert that here.
)

type chainObj struct {
	committee                          atomic.Value
	mempool                            chain.Mempool
	mempoolLastCleanedIndex            uint32
	dismissed                          atomic.Bool
	dismissOnce                        sync.Once
	chainID                            *iscp.ChainID
	chainStateSync                     coreutil.ChainStateSync
	stateReader                        state.OptimisticStateReader
	procset                            *processors.Cache
	stateMgr                           chain.StateManager
	consensus                          chain.Consensus
	log                                *logger.Logger
	nodeConn                           chain.ChainNodeConnection
	db                                 kvstore.KVStore
	netProvider                        peering.NetworkProvider
	dksProvider                        registry.DKShareRegistryProvider
	blobProvider                       registry.BlobCache
	committeeRegistry                  registry.CommitteeRegistryProvider
	eventRequestProcessed              *events.Event
	eventChainTransition               *events.Event
	eventChainTransitionClosure        *events.Closure
	receiveChainPeerMessagesAttachID   interface{}
	detachFromCommitteePeerMessagesFun func()
	chainPeers                         peering.PeerDomainProvider
	candidateNodes                     []*governance.AccessNodeInfo
	offLedgerReqsAcksMutex             sync.RWMutex
	offLedgerReqsAcks                  map[iscp.RequestID][]*ed25519.PublicKey
	offledgerBroadcastUpToNPeers       int
	offledgerBroadcastInterval         time.Duration
	pullMissingRequestsFromCommittee   bool
	chainMetrics                       metrics.ChainMetrics
	dismissChainMsgPipe                pipe.Pipe
	stateMsgPipe                       pipe.Pipe
	offLedgerRequestPeerMsgPipe        pipe.Pipe
	requestAckPeerMsgPipe              pipe.Pipe
	missingRequestIDsPeerMsgPipe       pipe.Pipe
	missingRequestPeerMsgPipe          pipe.Pipe
	timerTickMsgPipe                   pipe.Pipe
	wal                                chain.WAL
}

type committeeStruct struct {
	valid bool
	cmt   chain.Committee
}

func NewChain(
	chainID *iscp.ChainID,
	log *logger.Logger,
	nc chain.NodeConnection,
	db kvstore.KVStore,
	netProvider peering.NetworkProvider,
	dksProvider registry.DKShareRegistryProvider,
	blobProvider registry.BlobCache,
	committeeRegistry registry.CommitteeRegistryProvider,
	processorConfig *processors.Config,
	offledgerBroadcastUpToNPeers int,
	offledgerBroadcastInterval time.Duration,
	pullMissingRequestsFromCommittee bool,
	chainMetrics metrics.ChainMetrics,
	wal chain.WAL,
) chain.Chain {
	log.Debugf("creating chain object for %s", chainID.String())

	chainLog := log.Named(chainID.Bech32(iscp.Bech32Prefix)[:6] + ".")
	chainStateSync := coreutil.NewChainStateSync()
	ret := &chainObj{
		mempool:           mempool.New(state.NewOptimisticStateReader(db, chainStateSync), chainLog, chainMetrics),
		procset:           processors.MustNew(processorConfig),
		chainID:           chainID,
		log:               chainLog,
		nodeConn:          nodeconnimpl.NewChainNodeConnection(chainID, nc, chainLog),
		db:                db,
		chainStateSync:    chainStateSync,
		stateReader:       state.NewOptimisticStateReader(db, chainStateSync),
		netProvider:       netProvider,
		dksProvider:       dksProvider,
		blobProvider:      blobProvider,
		committeeRegistry: committeeRegistry,
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ iscp.RequestID))(params[0].(iscp.RequestID))
		}),
		eventChainTransition: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ *chain.ChainTransitionEventData))(params[0].(*chain.ChainTransitionEventData))
		}),
		candidateNodes:                   make([]*governance.AccessNodeInfo, 0),
		offLedgerReqsAcks:                make(map[iscp.RequestID][]*ed25519.PublicKey),
		offledgerBroadcastUpToNPeers:     offledgerBroadcastUpToNPeers,
		offledgerBroadcastInterval:       offledgerBroadcastInterval,
		pullMissingRequestsFromCommittee: pullMissingRequestsFromCommittee,
		chainMetrics:                     chainMetrics,
		dismissChainMsgPipe:              pipe.NewLimitInfinitePipe(1),
		stateMsgPipe:                     pipe.NewLimitInfinitePipe(maxMsgBuffer),
		offLedgerRequestPeerMsgPipe:      pipe.NewLimitInfinitePipe(maxMsgBuffer),
		requestAckPeerMsgPipe:            pipe.NewLimitInfinitePipe(maxMsgBuffer),
		missingRequestIDsPeerMsgPipe:     pipe.NewLimitInfinitePipe(maxMsgBuffer),
		missingRequestPeerMsgPipe:        pipe.NewLimitInfinitePipe(maxMsgBuffer),
		timerTickMsgPipe:                 pipe.NewLimitInfinitePipe(1),
		wal:                              wal,
	}
	ret.committee.Store(&committeeStruct{})

	var err error
	chainPeerNodes := []*ed25519.PublicKey{netProvider.Self().PubKey()}
	ret.chainPeers, err = netProvider.PeerDomain(chainID.Array(), chainPeerNodes)
	if err != nil {
		log.Errorf("NewChain: unable to create chainPeers domain: %v", err)
		return nil
	}
	stateMgrDomain, err := statemgr.NewDomainWithFallback(chainID.Array(), netProvider, log.Named("sm"))
	if err != nil {
		log.Errorf("NewChain: unable to create stateMgr.fallbackPeers domain: %v", err)
		return nil
	}

	ret.stateMgr = statemgr.New(db, ret, stateMgrDomain, ret.nodeConn, chainMetrics, wal)
	ret.stateMgr.SetChainPeers(chainPeerNodes)

	ret.eventChainTransitionClosure = events.NewClosure(ret.processChainTransition)
	ret.eventChainTransition.Attach(ret.eventChainTransitionClosure)
	ret.nodeConn.AttachToTransactionReceived(ret.ReceiveTransaction)
	ret.nodeConn.AttachToUnspentAliasOutputReceived(ret.ReceiveState)
	ret.receiveChainPeerMessagesAttachID = ret.chainPeers.Attach(peering.PeerMessageReceiverChain, ret.receiveChainPeerMessages)
	go ret.recvLoop()
	ret.startTimer()
	return ret
}

func (c *chainObj) startTimer() {
	go func() {
		c.stateMgr.Ready().MustWait()
		tick := 0
		for !c.IsDismissed() {
			c.EnqueueTimerTick(tick)
			tick++
			time.Sleep(chain.TimerTickPeriod)
		}
	}()
}

func (c *chainObj) receiveCommitteePeerMessages(peerMsg *peering.PeerMessageGroupIn) {
	if peerMsg.MsgType != chain.PeerMsgTypeMissingRequestIDs {
		c.log.Warnf("Wrong type of chain message (with committee peering ID): %v, ignoring it", peerMsg.MsgType)
		return
	}
	msg, err := messages.NewMissingRequestIDsMsg(peerMsg.MsgData)
	if err != nil {
		c.log.Error(err)
		return
	}
	c.EnqueueMissingRequestIDsMsg(&messages.MissingRequestIDsMsgIn{
		MissingRequestIDsMsg: *msg,
		SenderPubKey:         peerMsg.SenderPubKey,
	})
}

func (c *chainObj) receiveChainPeerMessages(peerMsg *peering.PeerMessageIn) {
	switch peerMsg.MsgType {
	case chain.PeerMsgTypeOffLedgerRequest:
		msg, err := messages.NewOffLedgerRequestMsg(peerMsg.MsgData)
		if err != nil {
			c.log.Error(err)
			return
		}
		c.EnqueueOffLedgerRequestMsg(&messages.OffLedgerRequestMsgIn{
			OffLedgerRequestMsg: *msg,
			SenderPubKey:        peerMsg.SenderPubKey,
		})
	case chain.PeerMsgTypeRequestAck:
		msg, err := messages.NewRequestAckMsg(peerMsg.MsgData)
		if err != nil {
			c.log.Error(err)
			return
		}
		c.EnqueueRequestAckMsg(&messages.RequestAckMsgIn{
			RequestAckMsg: *msg,
			SenderPubKey:  peerMsg.SenderPubKey,
		})
	case chain.PeerMsgTypeMissingRequest:
		msg, err := messages.NewMissingRequestMsg(peerMsg.MsgData)
		if err != nil {
			c.log.Error(err)
			return
		}
		c.EnqueueMissingRequestMsg(msg)
	default:
		c.log.Warnf("Wrong type of chain message (with chain peering ID): %v, ignoring it", peerMsg.MsgType)
	}
}

// processChainTransition processes the unique chain output which exists on the chain's address
// If necessary, it creates/changes/rotates committee object
func (c *chainObj) processChainTransition(msg *chain.ChainTransitionEventData) {
	stateIndex := msg.VirtualState.BlockIndex()
	c.log.Debugf("processChainTransition: processing state %d", stateIndex)
	if !msg.ChainOutput.GetIsGovernanceUpdated() {
		c.log.Debugf("processChainTransition state %d: output %s is not governance updated; state hash %s; last cleaned state is %d",
			stateIndex, iscp.OID(msg.ChainOutput.ID()), msg.VirtualState.StateCommitment().String(), c.mempoolLastCleanedIndex)
		// normal state update:
		c.stateReader.SetBaseline()
		chainID := iscp.ChainIDFromAliasID(msg.ChainOutput.GetAliasAddress())
		var reqids []iscp.RequestID
		for i := c.mempoolLastCleanedIndex + 1; i <= msg.VirtualState.BlockIndex(); i++ {
			c.log.Debugf("processChainTransition state %d: cleaning state %d", stateIndex, i)
			var err error
			reqids, err = blocklog.GetRequestIDsForBlock(c.stateReader, i)
			if reqids == nil {
				// The error means a database error. The optimistic state read failure can't occur here
				// because the state transition message is only sent only after state is committed and before consensus
				// start new round
				c.log.Panicf("processChainTransition. unexpected error: %v", err)
				return // to avoid "possible nil pointer dereference" in later use of `reqids`
			}
			// remove processed requests from the mempool
			c.log.Debugf("processChainTransition state %d cleaning state %d: removing %d requests", stateIndex, i, len(reqids))
			c.mempool.RemoveRequests(reqids...)
			chain.PublishRequestsSettled(chainID, i, reqids)
			// publish events
			for _, reqid := range reqids {
				c.eventRequestProcessed.Trigger(reqid)
			}
			c.publishNewBlockEvents(stateIndex)

			c.log.Debugf("processChainTransition state %d: state %d cleaned, deleted requests: %+v",
				stateIndex, i, iscp.ShortRequestIDs(reqids))
		}
		chain.PublishStateTransition(chainID, msg.ChainOutput, len(reqids))
		chain.LogStateTransition(msg, reqids, c.log)

		c.mempoolLastCleanedIndex = stateIndex
		c.updateChainNodes(stateIndex)
		c.chainMetrics.CurrentStateIndex(stateIndex)
	} else {
		c.log.Debugf("processChainTransition state %d: output %s is governance updated; state hash %s",
			stateIndex, iscp.OID(msg.ChainOutput.ID()), msg.VirtualState.StateCommitment().String())
		chain.LogGovernanceTransition(msg, c.log)
		chain.PublishGovernanceTransition(msg.ChainOutput)
	}
	if c.consensus == nil {
		c.log.Warnf("processChainTransition: skipping notifying consensus as it is not initiated")
	} else {
		c.consensus.EnqueueStateTransitionMsg(msg.VirtualState, msg.ChainOutput, msg.OutputTimestamp)
	}
	c.log.Debugf("processChainTransition completed: state index: %d, state hash: %s", stateIndex, msg.VirtualState.StateCommitment().String())
}

func (c *chainObj) updateChainNodes(stateIndex uint32) {
	c.log.Debugf("updateChainNodes, stateIndex=%v", stateIndex)
	govAccessNodes := make([]ed25519.PublicKey, 0)
	govCandidateNodes := make([]*governance.AccessNodeInfo, 0)
	if stateIndex > 0 {
		res, err := viewcontext.New(c).CallView(
			governance.Contract.Hname(),
			governance.FuncGetChainNodes.Hname(),
			governance.GetChainNodesRequest{}.AsDict(),
		)
		if err != nil {
			c.log.Panicf("unable to read the governance contract state: %v", err)
		}
		govResponse := governance.NewGetChainNodesResponseFromDict(res)
		govAccessNodes = govResponse.AccessNodes
		govCandidateNodes = govResponse.AccessNodeCandidates
	}

	//
	// Collect the new set of access nodes in the communication domain.
	// They include the committee nodes as well as the explicitly set access nodes.
	newMembers := make(map[ed25519.PublicKey]bool)
	newMembers[*c.netProvider.Self().PubKey()] = true
	cmt := c.getCommittee()
	if cmt != nil {
		for _, cm := range cmt.DKShare().NodePubKeys {
			newMembers[*cm] = true
		}
	}
	for _, newAccessNode := range govAccessNodes {
		newMembers[newAccessNode] = true
	}

	//
	// Pass it to the underlying domain to make a graceful update.
	newMemberList := make([]*ed25519.PublicKey, 0)
	for pubKey := range newMembers {
		pubKeyCopy := pubKey
		newMemberList = append(newMemberList, &pubKeyCopy)
	}
	c.chainPeers.UpdatePeers(newMemberList)
	c.stateMgr.SetChainPeers(newMemberList)

	//
	// Remember the candidate nodes as well (as a cache).
	c.candidateNodes = govCandidateNodes
}

func (c *chainObj) publishNewBlockEvents(blockIndex uint32) {
	if blockIndex == 0 {
		// don't run on state #0, root contracts not initialized yet.
		return
	}

	kvPartition := subrealm.NewReadOnly(c.stateReader.KVStoreReader(), kv.Key(blocklog.Contract.Hname().Bytes()))

	evts, err := blocklog.GetBlockEventsInternal(kvPartition, blockIndex)
	if err != nil {
		c.log.Panicf("publishNewBlockEvents - something went wrong getting events for block. %v", err)
	}

	go func() {
		for _, msg := range evts {
			c.log.Debugf("publishNewBlockEvents: '%s'", msg)
			publisher.Publish("vmmsg", c.chainID.Bech32(iscp.Bech32Prefix), msg)
		}
	}()
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
	if stateIndex, err := c.stateReader.BlockIndex(); err != nil {
		c.updateChainNodes(stateIndex)
	}
}
