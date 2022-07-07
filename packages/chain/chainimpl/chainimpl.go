// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	mempool_pkg "github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/chain/nodeconnchain"
	"github.com/iotaledger/wasp/packages/chain/statemgr"
	"github.com/iotaledger/wasp/packages/cryptolib"
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
	_ chain.Chain                     = &chainObj{}
	_ map[cryptolib.PublicKeyKey]bool // We rely on value comparison on the pubkeys, just assert that here.
)

type chainObj struct {
	committee                          atomic.Value
	mempool                            mempool_pkg.Mempool
	mempoolLastCleanedIndex            uint32
	dismissed                          atomic.Bool
	dismissOnce                        sync.Once
	chainID                            *iscp.ChainID
	chainStateSync                     coreutil.ChainStateSync
	stateReader                        state.OptimisticStateReader
	procset                            *processors.Cache
	lastSeenOutputStateIndex           *uint32
	stateMgr                           chain.StateManager
	consensus                          chain.Consensus
	log                                *logger.Logger
	nodeConn                           chain.ChainNodeConnection
	db                                 kvstore.KVStore
	netProvider                        peering.NetworkProvider
	dksProvider                        registry.DKShareRegistryProvider
	eventRequestProcessed              *events.Event
	eventChainTransition               *events.Event
	eventChainTransitionClosure        *events.Closure
	receiveChainPeerMessagesAttachID   interface{}
	detachFromCommitteePeerMessagesFun func()
	chainPeers                         peering.PeerDomainProvider
	candidateNodes                     []*governance.AccessNodeInfo
	offLedgerReqsAcksMutex             sync.RWMutex
	offLedgerReqsAcks                  map[iscp.RequestID][]*cryptolib.PublicKey
	offledgerBroadcastUpToNPeers       int
	offledgerBroadcastInterval         time.Duration
	pullMissingRequestsFromCommittee   bool
	lastSeenVirtualState               state.VirtualStateAccess
	chainMetrics                       metrics.ChainMetrics
	dismissChainMsgPipe                pipe.Pipe
	aliasOutputPipe                    pipe.Pipe
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
	processorConfig *processors.Config,
	offledgerBroadcastUpToNPeers int,
	offledgerBroadcastInterval time.Duration,
	pullMissingRequestsFromCommittee bool,
	chainMetrics metrics.ChainMetrics,
	wal chain.WAL,
) chain.Chain {
	var err error
	log.Debugf("creating chain object for %s", chainID.String())

	chainLog := log.Named("c-" + chainID.AsAddress().String()[2:8])
	chainStateSync := coreutil.NewChainStateSync()
	ret := &chainObj{
		mempool:        mempool_pkg.New(chainID.AsAddress(), state.NewOptimisticStateReader(db, chainStateSync), chainLog, chainMetrics),
		procset:        processors.MustNew(processorConfig),
		chainID:        chainID,
		log:            chainLog,
		db:             db,
		chainStateSync: chainStateSync,
		stateReader:    state.NewOptimisticStateReader(db, chainStateSync),
		netProvider:    netProvider,
		dksProvider:    dksProvider,
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ iscp.RequestID))(params[0].(iscp.RequestID))
		}),
		eventChainTransition: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ *chain.ChainTransitionEventData))(params[0].(*chain.ChainTransitionEventData))
		}),
		candidateNodes:                   make([]*governance.AccessNodeInfo, 0),
		offLedgerReqsAcks:                make(map[iscp.RequestID][]*cryptolib.PublicKey),
		offledgerBroadcastUpToNPeers:     offledgerBroadcastUpToNPeers,
		offledgerBroadcastInterval:       offledgerBroadcastInterval,
		pullMissingRequestsFromCommittee: pullMissingRequestsFromCommittee,
		chainMetrics:                     chainMetrics,
		dismissChainMsgPipe:              pipe.NewLimitInfinitePipe(1),
		aliasOutputPipe:                  pipe.NewLimitInfinitePipe(maxMsgBuffer),
		offLedgerRequestPeerMsgPipe:      pipe.NewLimitInfinitePipe(maxMsgBuffer),
		requestAckPeerMsgPipe:            pipe.NewLimitInfinitePipe(maxMsgBuffer),
		missingRequestIDsPeerMsgPipe:     pipe.NewLimitInfinitePipe(maxMsgBuffer),
		missingRequestPeerMsgPipe:        pipe.NewLimitInfinitePipe(maxMsgBuffer),
		timerTickMsgPipe:                 pipe.NewLimitInfinitePipe(1),
		wal:                              wal,
	}
	ret.nodeConn, err = nodeconnchain.NewChainNodeConnection(chainID, nc, chainLog)
	if err != nil {
		ret.log.Errorf("NewChain: unable to create chain node connection: v", err)
		return nil
	}

	ret.committee.Store(&committeeStruct{})

	var peeringID peering.PeeringID
	copy(peeringID[:], chainID.Bytes())

	chainPeerNodes := []*cryptolib.PublicKey{netProvider.Self().PubKey()}
	ret.chainPeers, err = netProvider.PeerDomain(peeringID, chainPeerNodes)
	if err != nil {
		log.Errorf("NewChain: unable to create chainPeers domain: %v", err)
		return nil
	}
	stateMgrDomain, err := statemgr.NewDomainWithFallback(peeringID, netProvider, log)
	if err != nil {
		log.Errorf("NewChain: unable to create stateMgr.fallbackPeers domain: %v", err)
		return nil
	}

	ret.stateMgr = statemgr.New(db, ret, stateMgrDomain, ret.nodeConn, chainMetrics, wal)
	ret.stateMgr.SetChainPeers(chainPeerNodes)

	ret.eventChainTransitionClosure = events.NewClosure(ret.processChainTransition)
	ret.eventChainTransition.Attach(ret.eventChainTransitionClosure)
	ret.nodeConn.AttachToOnLedgerRequest(ret.receiveOnLedgerRequest)
	ret.nodeConn.AttachToAliasOutput(ret.EnqueueAliasOutput)
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

func (c *chainObj) receiveOnLedgerRequest(request iscp.OnLedgerRequest) {
	c.log.Debugf("receiveOnLedgerRequest: %s", request.ID())
	c.mempool.ReceiveRequest(request)
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
		msg, err := messages.OffLedgerRequestMsgFromBytes(peerMsg.MsgData)
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
	if !msg.IsGovernance {
		// save last received from normal state transition
		c.lastSeenVirtualState = msg.VirtualState
	}
	if c.lastSeenVirtualState == nil {
		c.log.Warnf("processChainTransition: virtual state hasn't been received yet; ignoring chain transition event")
		return
	}
	stateIndex := c.lastSeenVirtualState.BlockIndex()
	oidStr := iscp.OID(msg.ChainOutput.ID())
	rootCommitment := state.RootCommitment(c.lastSeenVirtualState.TrieNodeStore())
	if msg.IsGovernance {
		c.log.Debugf("processChainTransition: processing governance transition at state %d, output %s, state hash %s",
			stateIndex, oidStr, rootCommitment)
		chain.LogGovernanceTransition(stateIndex, oidStr, rootCommitment, c.log)
		chain.PublishGovernanceTransition(msg.ChainOutput)
	} else {
		// normal state update:
		c.log.Debugf("processChainTransition: processing state %d transition, output %s; state hash %s; last cleaned state is %d", stateIndex, iscp.OID(msg.ChainOutput.ID()), rootCommitment, c.mempoolLastCleanedIndex)
		c.stateReader.SetBaseline()
		chainID := iscp.ChainIDFromAliasID(msg.ChainOutput.GetAliasID())
		var reqids []iscp.RequestID
		for i := c.mempoolLastCleanedIndex + 1; i <= c.lastSeenVirtualState.BlockIndex(); i++ {
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
			chain.PublishRequestsSettled(&chainID, i, reqids)
			// publish events
			for _, reqid := range reqids {
				c.eventRequestProcessed.Trigger(reqid)
			}
			c.publishNewBlockEvents(stateIndex)

			c.log.Debugf("processChainTransition state %d: state %d cleaned, deleted requests: %+v",
				stateIndex, i, iscp.ShortRequestIDs(reqids))
		}
		chain.PublishStateTransition(&chainID, msg.ChainOutput, len(reqids))
		chain.LogStateTransition(stateIndex, oidStr, rootCommitment, reqids, c.log)

		c.mempoolLastCleanedIndex = stateIndex
		c.updateChainNodes(stateIndex)
		c.chainMetrics.CurrentStateIndex(stateIndex)
	}
	if c.consensus == nil {
		c.log.Warnf("processChainTransition: skipping notifying consensus as it is not initiated")
	} else {
		c.consensus.EnqueueStateTransitionMsg(msg.IsGovernance, c.lastSeenVirtualState, msg.ChainOutput, msg.OutputTimestamp)
	}
	c.log.Debugf("processChainTransition completed: state index: %d, state hash: %s", stateIndex, rootCommitment)
}

func (c *chainObj) updateChainNodes(stateIndex uint32) {
	c.log.Debugf("updateChainNodes, stateIndex=%v", stateIndex)
	govAccessNodes := make([]*cryptolib.PublicKey, 0)
	govCandidateNodes := make([]*governance.AccessNodeInfo, 0)
	if stateIndex > 0 {
		res, err := viewcontext.New(c).CallViewExternal(
			governance.Contract.Hname(),
			governance.ViewGetChainNodes.Hname(),
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
	newMembers := make(map[cryptolib.PublicKeyKey]*cryptolib.PublicKey)
	selfPubKey := c.netProvider.Self().PubKey()
	newMembers[selfPubKey.AsKey()] = selfPubKey
	cmt := c.getCommittee()
	if cmt != nil {
		for _, cm := range cmt.DKShare().GetNodePubKeys() {
			newMembers[cm.AsKey()] = cm
		}
	}
	for _, newAccessNode := range govAccessNodes {
		newMembers[newAccessNode.AsKey()] = newAccessNode
	}

	//
	// Pass it to the underlying domain to make a graceful update.
	newMemberList := make([]*cryptolib.PublicKey, 0)
	for _, pubKey := range newMembers {
		pubKeyCopy := pubKey
		newMemberList = append(newMemberList, pubKeyCopy)
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
			publisher.Publish("vmmsg", c.chainID.String(), msg)
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
