// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package mempool manages transaction pools for IOTA Smart Contracts.
// A mempool basically does these functions:
//   - Provide a proposed set of requests (refs) for the consensus.
//   - Provide a set of requests for a TX as decided by the consensus.
//   - Share Off-Ledger requests between the committee and the server nodes.
//
// When the consensus asks for a proposal set, the mempool has to determine,
// if a reorg or a rollback has happened and adjust the request set accordingly.
// For this to work the mempool has to maintain not only the requests, but also
// the latest state for which it has provided the proposal. Let's say the mempool
// has provided proposals for PrevAO (AO≡AliasOutput).
//
// Upon reception of the proposal query (ConsensusProposalAsync) for NextAO
// from the consensus, it asks the StateMgr for the virtual state VS(NextAO)
// corresponding to the NextAO and a list of blocks that has to be reverted.
// The state manager collects this information by finding a common ancestor of
// the NextAO and PrevAO, say CommonAO = NextAO ⊓ PrevAO. The blocks to be
// reverted are those in the range (CommonAO, PrevAO].
//
// When the mempool gets VS(NextAO) and RevertBlocks = (CommonAO, PrevAO] it
// re-adds the requests from RevertBlocks to the mempool and then drops the
// requests that are already processed in VS(NextAO). If the RevertBlocks set
// is not empty, it has to drop all the on-ledger requests and re-read them from
// the L1. In the normal execution, we'll have RevertBlocks=∅ and VS(NextAO)
// will differ from VS(PrevAO) in a single block.
//
// The response to the requests decided by the consensus (ConsensusRequestsAsync)
// should be unconditional and should ignore the current state of the requests.
// This call should not modify nor the NextAO not the PrevAO. The state will be
// updated later with the proposal query, because then the chain will know, which
// branch to work on.
//
// Time-locked requests are maintained in the mempool as well. They are provided
// to the proposal based on a tangle time. The tangle time is received from the
// L1 with the milestones.
//
// NOTE: A node looses its off-ledger requests on restart. The on-ledger requests
// will be added back to the mempool by reading them from the L1 node.
//
// TODO: Propose subset of the requests. That's for the next release.
package mempool

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/log"

	consGR "github.com/iotaledger/wasp/packages/chain/cons/gr"
	"github.com/iotaledger/wasp/packages/chain/mempool/distsync"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

const (
	distShareDebugTick      = 10 * time.Second
	distShareTimeTick       = 3 * time.Second
	distShareMaxMsgsPerTick = 100
	distShareRePublishTick  = 5 * time.Second
	waitRequestCleanupEvery = 10
	forceCleanMempoolTick   = 1 * time.Minute
)

// ChainListener is a partial interface for providing chain events to the outside.
// This interface is in the mempool part only because it tracks
// the actual state for checking the consumed requests.
type ChainListener interface {
	// This function is called by the chain when new block is applied to the
	// state. This block might be not confirmed yet, but the chain is going
	// to build the next block on top of this one.
	BlockApplied(chainID isc.ChainID, block state.Block, latestState kv.KVStoreReader)
}

type Mempool interface {
	consGR.Mempool
	// Invoked by the chain, when new anchor is considered as a tip/head
	// of the chain. Mempool can reorganize its state by removing/rejecting
	// or re-adding some requests, depending on how the head has changed.
	// It can mean simple advance of the chain, or a rollback or a reorg.
	// This function is guaranteed to be called in the order, which is
	// considered the chain block order by the ChainMgr.
	TrackNewChainHead(st state.State, from, till *isc.StateAnchor, added, removed []state.Block) <-chan bool
	// Invoked by the chain when a new off-ledger request is received from a node user.
	// Inter-node off-ledger dissemination is NOT performed via this function.
	ReceiveOnLedgerRequest(request isc.OnLedgerRequest)
	// This is called when this node receives an off-ledger request from a user directly.
	// I.e. when this node is an entry point of the off-ledger request.
	ReceiveOffLedgerRequest(request isc.OffLedgerRequest) error
	// Invoked by the ChainMgr when a time of a tangle changes.
	TangleTimeUpdated(tangleTime time.Time)
	// Invoked by the chain when a set of server nodes has changed.
	// These nodes should be used to disseminate the off-ledger requests.
	ServerNodesUpdated(committeePubKeys []*cryptolib.PublicKey, serverNodePubKeys []*cryptolib.PublicKey)
	AccessNodesUpdated(committeePubKeys []*cryptolib.PublicKey, accessNodePubKeys []*cryptolib.PublicKey)

	Iterate(f func(req isc.Request) bool)
}

type Settings struct {
	TTL                        time.Duration // time to live (how much time requests are allowed to sit in the pool without being processed)
	OnLedgerRefreshMinInterval time.Duration
	MaxOffledgerInPool         int
	MaxOnledgerInPool          int
	MaxTimedInPool             int
	MaxOnledgerToPropose       int // (including timed-requests)
	MaxOffledgerToPropose      int
	MaxOffledgerPerAccount     int
}

// This implementation tracks single branch of the chain only. I.e. all the consensus
// instances that are asking for requests for alias outputs different than the current
// head (as considered by the ChainMgr) will get empty proposal sets.
//
// In general we can track several branches, but then we have to remember, which
// requests are available in which branches. Can be implemented later, if needed.
type mempoolImpl struct {
	chainID                        isc.ChainID
	tangleTime                     time.Time
	onLedgerPool                   RequestPool[isc.OnLedgerRequest] // TODO limit this pool
	offLedgerPool                  *OffLedgerPool                   // TODO maybe use `RequestPool` too?
	distSync                       gpa.GPA
	chainHeadAnchor                *isc.StateAnchor
	chainHeadState                 state.State
	serverNodesUpdatedPipe         pipe.Pipe[*reqServerNodesUpdated]
	serverNodes                    []*cryptolib.PublicKey
	accessNodesUpdatedPipe         pipe.Pipe[*reqAccessNodesUpdated]
	accessNodes                    []*cryptolib.PublicKey
	committeeNodes                 []*cryptolib.PublicKey
	consensusInstances             []consGR.ConsensusID
	waitReq                        WaitReq
	waitChainHead                  []*reqConsensusProposal
	reqConsensusProposalPipe       pipe.Pipe[*reqConsensusProposal]
	reqConsensusRequestsPipe       pipe.Pipe[*reqConsensusRequests]
	reqReceiveOnLedgerRequestPipe  pipe.Pipe[isc.OnLedgerRequest]
	reqReceiveOffLedgerRequestPipe pipe.Pipe[isc.OffLedgerRequest]
	reqTangleTimeUpdatedPipe       pipe.Pipe[time.Time]
	reqTrackNewChainHeadPipe       pipe.Pipe[*reqTrackNewChainHead]
	netRecvPipe                    pipe.Pipe[*peering.PeerMessageIn]
	netPeeringID                   peering.PeeringID
	netPeerPubs                    map[gpa.NodeID]*cryptolib.PublicKey
	net                            peering.NetworkProvider
	settings                       Settings
	broadcastInterval              time.Duration // how often requests should be rebroadcasted
	log                            log.Logger
	metrics                        *metrics.ChainMempoolMetrics
	listener                       ChainListener
	refreshOnLedgerRequests        func()
	lastRefreshTimestamp           time.Time
}

var _ Mempool = &mempoolImpl{}

const (
	msgTypeMempool byte = iota
)

type reqServerNodesUpdated struct {
	committeePubKeys  []*cryptolib.PublicKey
	serverNodePubKeys []*cryptolib.PublicKey
}

type reqAccessNodesUpdated struct {
	committeePubKeys  []*cryptolib.PublicKey
	accessNodePubKeys []*cryptolib.PublicKey
}

type reqConsensusProposal struct {
	ctx         context.Context
	anchor      *isc.StateAnchor
	consensusID consGR.ConsensusID
	responseCh  chan<- []*isc.RequestRef
}

func (r *reqConsensusProposal) Respond(reqRefs []*isc.RequestRef) {
	r.responseCh <- reqRefs
	close(r.responseCh)
}

type reqConsensusRequests struct {
	ctx         context.Context
	requestRefs []*isc.RequestRef
	responseCh  chan<- []isc.Request
}

type reqTrackNewChainHead struct {
	st         state.State
	from       *isc.StateAnchor
	till       *isc.StateAnchor
	added      []state.Block
	removed    []state.Block
	responseCh chan<- bool // only for tests, shouldn't be used in the chain package
}

func New(
	ctx context.Context,
	chainID isc.ChainID,
	nodeIdentity *cryptolib.KeyPair,
	net peering.NetworkProvider,
	log log.Logger,
	metrics *metrics.ChainMempoolMetrics,
	pipeMetrics *metrics.ChainPipeMetrics,
	listener ChainListener,
	settings Settings,
	broadcastInterval time.Duration,
	refreshOnLedgerRequests func(),
) Mempool {
	netPeeringID := peering.HashPeeringIDFromBytes(chainID.Bytes(), []byte("Mempool")) // ChainID × Mempool
	waitReq := NewWaitReq(waitRequestCleanupEvery)
	mpi := &mempoolImpl{
		chainID:                        chainID,
		tangleTime:                     time.Time{},
		onLedgerPool:                   NewTypedPool[isc.OnLedgerRequest](settings.MaxOnledgerInPool, waitReq, metrics.SetOnLedgerPoolSize, metrics.SetOnLedgerReqTime, log.NewChildLogger("ONL")),
		offLedgerPool:                  NewOffledgerPool(settings.MaxOffledgerInPool, settings.MaxOffledgerPerAccount, waitReq, metrics.SetOffLedgerPoolSize, metrics.SetOffLedgerReqTime, log.NewChildLogger("OFF")),
		chainHeadAnchor:                nil,
		serverNodesUpdatedPipe:         pipe.NewInfinitePipe[*reqServerNodesUpdated](),
		serverNodes:                    []*cryptolib.PublicKey{},
		accessNodesUpdatedPipe:         pipe.NewInfinitePipe[*reqAccessNodesUpdated](),
		accessNodes:                    []*cryptolib.PublicKey{},
		committeeNodes:                 []*cryptolib.PublicKey{},
		waitReq:                        waitReq,
		waitChainHead:                  []*reqConsensusProposal{},
		reqConsensusProposalPipe:       pipe.NewInfinitePipe[*reqConsensusProposal](),
		reqConsensusRequestsPipe:       pipe.NewInfinitePipe[*reqConsensusRequests](),
		reqReceiveOnLedgerRequestPipe:  pipe.NewInfinitePipe[isc.OnLedgerRequest](),
		reqReceiveOffLedgerRequestPipe: pipe.NewInfinitePipe[isc.OffLedgerRequest](),
		reqTangleTimeUpdatedPipe:       pipe.NewInfinitePipe[time.Time](),
		reqTrackNewChainHeadPipe:       pipe.NewInfinitePipe[*reqTrackNewChainHead](),
		netRecvPipe:                    pipe.NewInfinitePipe[*peering.PeerMessageIn](),
		netPeeringID:                   netPeeringID,
		netPeerPubs:                    map[gpa.NodeID]*cryptolib.PublicKey{},
		net:                            net,
		settings:                       settings,
		broadcastInterval:              broadcastInterval,
		log:                            log,
		metrics:                        metrics,
		listener:                       listener,
		refreshOnLedgerRequests:        refreshOnLedgerRequests,
		lastRefreshTimestamp:           time.Now(),
	}

	pipeMetrics.TrackPipeLen("mp-serverNodesUpdatedPipe", mpi.serverNodesUpdatedPipe.Len)
	pipeMetrics.TrackPipeLen("mp-accessNodesUpdatedPipe", mpi.accessNodesUpdatedPipe.Len)
	pipeMetrics.TrackPipeLen("mp-reqConsensusProposalPipe", mpi.reqConsensusProposalPipe.Len)
	pipeMetrics.TrackPipeLen("mp-reqConsensusRequestsPipe", mpi.reqConsensusRequestsPipe.Len)
	pipeMetrics.TrackPipeLen("mp-reqReceiveOnLedgerRequestPipe", mpi.reqReceiveOnLedgerRequestPipe.Len)
	pipeMetrics.TrackPipeLen("mp-reqReceiveOffLedgerRequestPipe", mpi.reqReceiveOffLedgerRequestPipe.Len)
	pipeMetrics.TrackPipeLen("mp-reqTangleTimeUpdatedPipe", mpi.reqTangleTimeUpdatedPipe.Len)
	pipeMetrics.TrackPipeLen("mp-reqTrackNewChainHeadPipe", mpi.reqTrackNewChainHeadPipe.Len)
	pipeMetrics.TrackPipeLen("mp-netRecvPipe", mpi.netRecvPipe.Len)

	mpi.distSync = distsync.New(
		mpi.pubKeyAsNodeID(nodeIdentity.GetPublicKey()),
		mpi.distSyncRequestNeededCB,
		mpi.distSyncRequestReceivedCB,
		distShareMaxMsgsPerTick,
		mpi.metrics.SetMissingReqs,
		log,
	)
	netRecvPipeInCh := mpi.netRecvPipe.In()
	unhook := net.Attach(&netPeeringID, peering.ReceiverMempool, func(recv *peering.PeerMessageIn) {
		if recv.MsgType != msgTypeMempool {
			mpi.log.LogWarnf("Unexpected message, type=%v", recv.MsgType)
			return
		}
		netRecvPipeInCh <- recv
	})
	go mpi.run(ctx, unhook)
	return mpi
}

func (mpi *mempoolImpl) TangleTimeUpdated(tangleTime time.Time) {
	mpi.reqTangleTimeUpdatedPipe.In() <- tangleTime
}

func (mpi *mempoolImpl) TrackNewChainHead(st state.State, from, till *isc.StateAnchor, added, removed []state.Block) <-chan bool {
	responseCh := make(chan bool)
	mpi.reqTrackNewChainHeadPipe.In() <- &reqTrackNewChainHead{st, from, till, added, removed, responseCh}
	return responseCh
}

func (mpi *mempoolImpl) ReceiveOnLedgerRequest(request isc.OnLedgerRequest) {
	mpi.reqReceiveOnLedgerRequestPipe.In() <- request
}

func (mpi *mempoolImpl) ReceiveOffLedgerRequest(request isc.OffLedgerRequest) error {
	if err := mpi.shouldAddOffledgerRequest(request); err != nil {
		return err
	}
	mpi.reqReceiveOffLedgerRequestPipe.In() <- request
	return nil
}

func (mpi *mempoolImpl) ServerNodesUpdated(committeePubKeys, serverNodePubKeys []*cryptolib.PublicKey) {
	mpi.serverNodesUpdatedPipe.In() <- &reqServerNodesUpdated{
		committeePubKeys:  committeePubKeys,
		serverNodePubKeys: serverNodePubKeys,
	}
}

func (mpi *mempoolImpl) AccessNodesUpdated(committeePubKeys, accessNodePubKeys []*cryptolib.PublicKey) {
	mpi.accessNodesUpdatedPipe.In() <- &reqAccessNodesUpdated{
		committeePubKeys:  committeePubKeys,
		accessNodePubKeys: accessNodePubKeys,
	}
}

func (mpi *mempoolImpl) ConsensusProposalAsync(ctx context.Context, anchor *isc.StateAnchor, consensusID consGR.ConsensusID) <-chan []*isc.RequestRef {
	res := make(chan []*isc.RequestRef, 1)
	req := &reqConsensusProposal{
		ctx:         ctx,
		anchor:      anchor,
		consensusID: consensusID,
		responseCh:  res,
	}
	mpi.reqConsensusProposalPipe.In() <- req
	return res
}

func (mpi *mempoolImpl) ConsensusRequestsAsync(ctx context.Context, requestRefs []*isc.RequestRef) <-chan []isc.Request {
	res := make(chan []isc.Request, 1)
	req := &reqConsensusRequests{
		ctx:         ctx,
		requestRefs: requestRefs,
		responseCh:  res,
	}
	mpi.reqConsensusRequestsPipe.In() <- req
	return res
}

func (mpi *mempoolImpl) Iterate(f func(req isc.Request) bool) {
	// TODO: This makes unprotected concurrent access to the MP state.

	continueIter := true

	mpi.offLedgerPool.Iterate(func(account string, requests []*OrderedPoolEntry) bool {
		for _, entry := range requests {
			if !f(entry.req) {
				continueIter = false
				return false
			}
		}
		return true
	})

	if !continueIter {
		return
	}

	mpi.onLedgerPool.Iterate(func(entry *typedPoolEntry[isc.OnLedgerRequest]) bool {
		return f(entry.req)
	})
}

func (mpi *mempoolImpl) run(ctx context.Context, cleanupFunc context.CancelFunc) { //nolint:gocyclo
	serverNodesUpdatedPipeOutCh := mpi.serverNodesUpdatedPipe.Out()
	accessNodesUpdatedPipeOutCh := mpi.accessNodesUpdatedPipe.Out()
	reqConsensusProposalPipeOutCh := mpi.reqConsensusProposalPipe.Out()
	reqConsensusRequestsPipeOutCh := mpi.reqConsensusRequestsPipe.Out()
	reqReceiveOnLedgerRequestPipeOutCh := mpi.reqReceiveOnLedgerRequestPipe.Out()
	reqReceiveOffLedgerRequestPipeOutCh := mpi.reqReceiveOffLedgerRequestPipe.Out()
	reqTangleTimeUpdatedPipeOutCh := mpi.reqTangleTimeUpdatedPipe.Out()
	reqTrackNewChainHeadPipeOutCh := mpi.reqTrackNewChainHeadPipe.Out()
	netRecvPipeOutCh := mpi.netRecvPipe.Out()
	debugTicker := time.NewTicker(distShareDebugTick)
	timeTicker := time.NewTicker(distShareTimeTick)
	rePublishTicker := time.NewTicker(distShareRePublishTick)
	forceCleanMempoolTicker := time.NewTicker(forceCleanMempoolTick) // this exists to force mempool cleanup on access nodes // thought: maybe access nodes shouldn't have a mempool at all
	for {
		select {
		case recv, ok := <-serverNodesUpdatedPipeOutCh:
			if !ok {
				serverNodesUpdatedPipeOutCh = nil
				break
			}
			mpi.handleServerNodesUpdated(recv)
		case recv, ok := <-accessNodesUpdatedPipeOutCh:
			if !ok {
				accessNodesUpdatedPipeOutCh = nil
				break
			}
			mpi.handleAccessNodesUpdated(recv)
		case recv, ok := <-reqConsensusProposalPipeOutCh:
			if !ok {
				reqConsensusProposalPipeOutCh = nil
				break
			}
			mpi.handleConsensusProposal(recv)
			forceCleanMempoolTicker.Reset(forceCleanMempoolTick) // mempool will be forcebly cleanup if this ticker triggers
		case recv, ok := <-reqConsensusRequestsPipeOutCh:
			if !ok {
				reqConsensusRequestsPipeOutCh = nil
				break
			}
			mpi.handleConsensusRequests(recv)
		case recv, ok := <-reqReceiveOnLedgerRequestPipeOutCh:
			if !ok {
				reqReceiveOnLedgerRequestPipeOutCh = nil
				break
			}
			mpi.handleReceiveOnLedgerRequest(recv)
		case recv, ok := <-reqReceiveOffLedgerRequestPipeOutCh:
			if !ok {
				reqReceiveOffLedgerRequestPipeOutCh = nil
				break
			}
			mpi.handleReceiveOffLedgerRequest(recv)
		case recv, ok := <-reqTangleTimeUpdatedPipeOutCh:
			if !ok {
				reqTangleTimeUpdatedPipeOutCh = nil
				break
			}
			mpi.handleTangleTimeUpdated(recv)
		case recv, ok := <-reqTrackNewChainHeadPipeOutCh:
			if !ok {
				reqTrackNewChainHeadPipeOutCh = nil
				break
			}
			mpi.handleTrackNewChainHead(recv)
		case recv, ok := <-netRecvPipeOutCh:
			if !ok {
				netRecvPipeOutCh = nil
				continue
			}
			mpi.handleNetMessage(recv)
		case <-debugTicker.C:
			mpi.handleDistSyncDebugTick()
		case <-timeTicker.C:
			mpi.handleDistSyncTimeTick()
		case <-rePublishTicker.C:
			mpi.handleRePublishTimeTick()
		case <-forceCleanMempoolTicker.C:
			mpi.handleForceCleanMempool()
		case <-ctx.Done():
			// mpi.serverNodesUpdatedPipe.Close() // TODO: Causes panic: send on closed channel
			// mpi.accessNodesUpdatedPipe.Close()
			// mpi.reqConsensusProposalPipe.Close()
			// mpi.reqConsensusRequestsPipe.Close()
			// mpi.reqReceiveOnLedgerRequestPipe.Close()
			// mpi.reqReceiveOffLedgerRequestPipe.Close()
			// mpi.reqTangleTimeUpdatedPipe.Close()
			// mpi.reqTrackNewChainHeadPipe.Close()
			// mpi.netRecvPipe.Close()
			debugTicker.Stop()
			timeTicker.Stop()
			util.ExecuteIfNotNil(cleanupFunc)
			return
		}
	}
}

// A callback for distSync.
//   - We only return off-ledger requests here.
//   - The requests can be in the mempool
//   - Or they are processed already (a lagging node asks
//     for them), then the state has to be accessed.
func (mpi *mempoolImpl) distSyncRequestNeededCB(requestRef *isc.RequestRef) isc.Request {
	if req := mpi.offLedgerPool.Get(requestRef); req != nil {
		mpi.log.LogDebugf("responding to RequestNeeded(ref=%v), found in offLedgerPool", requestRef)
		return req
	}
	if mpi.chainHeadState != nil {
		requestID := requestRef.ID
		receipt, err := blocklog.NewStateReaderFromChainState(mpi.chainHeadState).GetRequestReceipt(requestID)
		if err == nil && receipt != nil && receipt.Request.IsOffLedger() {
			mpi.log.LogDebugf("responding to RequestNeeded(ref=%v), found in blockLog", requestRef)
			return receipt.Request
		}
		return nil
	}
	return nil
}

// A callback for distSync.
func (mpi *mempoolImpl) distSyncRequestReceivedCB(request isc.Request) bool {
	offLedgerReq, ok := request.(isc.OffLedgerRequest)
	if !ok {
		mpi.log.LogWarnf("Dropping non-OffLedger request form dist %T: %+v", request, request)
		return false
	}
	if err := mpi.shouldAddOffledgerRequest(offLedgerReq); err == nil {
		mpi.log.LogWarnf("shouldAddOffledgerRequest: true, trying to add to offledger %T: %+v", request, request)

		return mpi.addOffledger(offLedgerReq)
	}
	return false
}

func (mpi *mempoolImpl) accountsState() *accounts.StateReader {
	return accounts.NewStateReaderFromChainState(
		mpi.chainHeadState.SchemaVersion(),
		mpi.chainHeadState,
	)
}

func (mpi *mempoolImpl) nonce(account isc.AgentID) uint64 {
	evmState := evmimpl.NewStateAccess(mpi.chainHeadState)

	if evmSender, ok := account.(*isc.EthereumAddressAgentID); ok {
		return evmState.Nonce(evmSender.EthAddress())
	}
	return mpi.accountsState().AccountNonce(account)
}

func (mpi *mempoolImpl) shouldAddOffledgerRequest(req isc.OffLedgerRequest) error {
	mpi.log.LogDebugf("trying to add to mempool, requestID: %s", req.ID().String())
	if err := req.VerifySignature(); err != nil {
		return fmt.Errorf("invalid signature")
	}
	if mpi.offLedgerPool.Has(isc.RequestRefFromRequest(req)) {
		return fmt.Errorf("already in mempool")
	}
	if mpi.chainHeadState == nil {
		return fmt.Errorf("chainHeadState is nil")
	}

	accountNonce := mpi.nonce(req.SenderAccount())
	if req.Nonce() < accountNonce {
		return fmt.Errorf("bad nonce, expected: %d", accountNonce)
	}

	// check user has on-chain balance
	governanceState := governance.NewStateReaderFromChainState(mpi.chainHeadState)
	minFee := governanceState.GetGasFeePolicy().MinFee(isc.RequestGasPrice(req), parameters.BaseTokenDecimals)
	balance := mpi.accountsState().GetBaseTokensBalanceDiscardExtraDecimals(req.SenderAccount())
	if balance < minFee {
		// make an exception for gov calls (sender is chain admin and target is gov contract)
		chainAdmin := governanceState.GetChainAdmin()
		isGovRequest := req.SenderAccount().Equals(chainAdmin) && req.Message().Target.Contract == governance.Contract.Hname()
		if !isGovRequest {
			return fmt.Errorf("not enough funds on chain to cover minimum fee")
		}
	}

	// reject txs with gas price too low
	if gp := req.GasPrice(); gp != nil && gp.Cmp(mpi.offLedgerPool.minGasPrice) == -1 {
		return fmt.Errorf("gas price too low. Must be at least %s", mpi.offLedgerPool.minGasPrice.String())
	}

	return nil
}

func (mpi *mempoolImpl) addOffledger(request isc.OffLedgerRequest) bool {
	if !mpi.offLedgerPool.Add(request) {
		return false
	}
	mpi.metrics.IncRequestsReceived(request)
	mpi.log.LogDebugf("accepted by the mempool, requestID: %s", request.ID().String())
	return true
}

func (mpi *mempoolImpl) handleServerNodesUpdated(recv *reqServerNodesUpdated) {
	mpi.serverNodes = recv.serverNodePubKeys
	mpi.committeeNodes = recv.committeePubKeys
	mpi.sendMessages(mpi.distSync.Input(distsync.NewInputServerNodes(
		lo.Map(mpi.serverNodes, mpi.pubKeyAsNodeIDMap),
		lo.Map(mpi.committeeNodes, mpi.pubKeyAsNodeIDMap),
	)))
}

func (mpi *mempoolImpl) handleAccessNodesUpdated(recv *reqAccessNodesUpdated) {
	mpi.accessNodes = recv.accessNodePubKeys
	mpi.committeeNodes = recv.committeePubKeys
	mpi.sendMessages(mpi.distSync.Input(distsync.NewInputAccessNodes(
		lo.Map(mpi.accessNodes, mpi.pubKeyAsNodeIDMap),
		lo.Map(mpi.committeeNodes, mpi.pubKeyAsNodeIDMap),
	)))
}

// This implementation only tracks a single branch. So, we will only respond
// to the request matching the TrackNewChainHead call.
func (mpi *mempoolImpl) handleConsensusProposal(recv *reqConsensusProposal) {
	if mpi.chainHeadAnchor == nil || !recv.anchor.Equals(mpi.chainHeadAnchor) {
		mpi.log.LogDebugf("handleConsensusProposal, have to wait for chain head to become %v", recv.anchor)
		mpi.waitChainHead = append(mpi.waitChainHead, recv)
		return
	}
	mpi.log.LogDebugf("handleConsensusProposal, already have the chain head %v", recv.anchor)
	mpi.handleConsensusProposalForChainHead(recv)
}

func (mpi *mempoolImpl) refsToPropose(consensusID consGR.ConsensusID) []*isc.RequestRef {
	//
	// The case for matching ChainHeadAnchor and request BaseAO
	onLedgerReqs := []*isc.RequestRef{}
	if !mpi.tangleTime.IsZero() { // Wait for tangle-time to process the on ledger requests.
		mpi.onLedgerPool.Iterate(func(e *typedPoolEntry[isc.OnLedgerRequest]) bool {
			onLedgerReqs = append(onLedgerReqs, isc.RequestRefFromRequest(e.req))
			e.proposedFor = append(e.proposedFor, consensusID)
			return len(onLedgerReqs) < mpi.settings.MaxOnledgerToPropose
		})
	}

	//
	// iterate the ordered txs and add the first valid ones (respect nonce) to propose
	// stop iterating when either: got MaxOffledgerToPropose, or no requests were added during last iteration (there are gaps in nonces)
	accNonces := make(map[string]uint64)                             // cache of account nonces so we don't propose gaps
	orderedList := slices.Clone(mpi.offLedgerPool.orderedByGasPrice) // clone the ordered list of references to requests, so we can alter it safely
	offLedgerReqs := []*isc.RequestRef{}
	for {
		addedThisCycle := false
		for i, e := range orderedList {
			if e == nil {
				continue
			}
			//
			// drop tx with expired TTL
			if time.Since(e.ts) > mpi.settings.TTL { // stop proposing after TTL
				if !e.proposedForAny(mpi.consensusInstances) {
					// request not used in active consensus anymore, remove it
					mpi.log.LogDebugf("refsToPropose, request TTL expired, removing: %s", e.req.ID().String())
					mpi.offLedgerPool.Remove(e.req)
					continue
				}
				mpi.log.LogDebugf("refsToPropose, request TTL expired, skipping: %s", e.req.ID().String())
				continue
			}

			if e.old {
				// this request was marked as "old", do not propose it
				mpi.log.LogDebugf("refsToPropose, skipping old request: %s", e.req.ID().String())
				continue
			}

			reqAccount := e.req.SenderAccount()
			reqAccountKey := reqAccount.String()
			accountNonce, ok := accNonces[reqAccountKey]
			if !ok {
				accountNonce = mpi.nonce(reqAccount)
				accNonces[reqAccountKey] = accountNonce
			}

			reqNonce := e.req.Nonce()
			if reqNonce < accountNonce {
				// nonce too old, delete
				mpi.log.LogDebugf("refsToPropose, account: %s, removing request (%s) with old nonce (%d) from the pool", reqAccount, e.req.ID(), e.req.Nonce())
				mpi.offLedgerPool.Remove(e.req)
				continue
			}

			if reqNonce == accountNonce {
				// expected nonce, add it to the list to propose
				mpi.log.LogDebugf("refsToPropose, account: %s, proposing reqID %s with nonce: %d", reqAccount, e.req.ID().String(), e.req.Nonce())
				offLedgerReqs = append(offLedgerReqs, isc.RequestRefFromRequest(e.req))
				e.markProposed(consensusID)
				addedThisCycle = true
				accountNonce++ // increment the account nonce to match the next valid request
				accNonces[reqAccountKey] = accountNonce
				// delete from this list
				orderedList[i] = nil
			}

			if len(offLedgerReqs) >= mpi.settings.MaxOffledgerToPropose {
				break // got enough requests
			}

			if reqNonce > accountNonce {
				mpi.log.LogDebugf("refsToPropose, account: %s, req %s has a nonce %d which is too high (expected %d), won't be proposed", reqAccount, e.req.ID().String(), e.req.Nonce(), accountNonce)
				continue // skip request
			}
		}
		if !addedThisCycle || len(offLedgerReqs) >= mpi.settings.MaxOffledgerToPropose {
			break
		}
	}

	return slices.Concat(onLedgerReqs, offLedgerReqs)
}

func (mpi *mempoolImpl) handleConsensusProposalForChainHead(recv *reqConsensusProposal) {
	refs := mpi.refsToPropose(recv.consensusID)
	mpi.log.LogDebugf("handleConsensusProposalForChainHead, |refs|=%v", len(refs))
	if len(refs) > 0 {
		recv.Respond(refs)
		return
	}

	//
	// Wait for any request.
	mpi.waitReq.WaitAny(recv.ctx, func(_ isc.Request) {
		mpi.handleConsensusProposalForChainHead(recv)
	})
}

func (mpi *mempoolImpl) handleConsensusRequests(recv *reqConsensusRequests) {
	reqs := make([]isc.Request, len(recv.requestRefs))
	missing := []*isc.RequestRef{}
	missingIdx := map[isc.RequestRefKey]int{}
	for i := range reqs {
		reqRef := recv.requestRefs[i]
		reqs[i] = mpi.onLedgerPool.Get(reqRef)
		if reqs[i] == nil {
			reqs[i] = mpi.offLedgerPool.Get(reqRef)
		}
		if reqs[i] == nil && mpi.chainHeadState != nil {
			// Check also the processed backlog, to avoid consensus blocking while waiting for processed request.
			// It will be rejected later (or state branch will change).
			receipt, err := blocklog.NewStateReaderFromChainState(mpi.chainHeadState).GetRequestReceipt(reqRef.ID)
			if err == nil && receipt != nil {
				reqs[i] = receipt.Request
			}
		}
		if reqs[i] == nil {
			missing = append(missing, reqRef)
			missingIdx[reqRef.AsKey()] = i
		}
	}
	if len(missing) == 0 {
		recv.responseCh <- reqs
		close(recv.responseCh)
		return
	}
	//
	// Wait for missing requests.
	for i := range missing {
		mpi.sendMessages(mpi.distSync.Input(distsync.NewInputRequestNeeded(recv.ctx, missing[i])))
	}
	mpi.waitReq.WaitMany(recv.ctx, missing, func(req isc.Request) {
		reqRefKey := isc.RequestRefFromRequest(req).AsKey()
		if idx, ok := missingIdx[reqRefKey]; ok {
			reqs[idx] = req
			delete(missingIdx, reqRefKey)
			if len(missingIdx) == 0 {
				recv.responseCh <- reqs
				close(recv.responseCh)
			}
		}
	})
}

func (mpi *mempoolImpl) handleReceiveOnLedgerRequest(request isc.OnLedgerRequest) {
	requestID := request.ID()
	requestRef := isc.RequestRefFromRequest(request)
	//
	// TODO: Do not process anything with SDRUC for now.
	/*if _, ok := request.Features().ReturnAmount(); ok {
		mpi.log.LogWarnf("dropping request, because it has ReturnAmount, ID=%v", requestID)
		return
	}*/
	if request.SenderAccount() == nil {
		// do not process requests without the sender feature
		mpi.log.LogWarnf("dropping request, because it has no sender feature, ID=%v", requestID)
		return
	}
	//
	// Check, maybe mempool already has it.
	if mpi.onLedgerPool.Has(requestRef) {
		mpi.log.LogWarnf("request already in the mempool, ID=%v", requestID)
		return
	}
	//
	// Maybe it has been processed before?
	if mpi.chainHeadState != nil {
		processed, err := blocklog.NewStateReaderFromChainState(mpi.chainHeadState).IsRequestProcessed(requestID)
		if err != nil {
			panic(fmt.Errorf("cannot check if request was processed: %w", err))
		}
		if processed {
			mpi.log.LogWarnf("dropping request, because it was already processed, ID=%v", requestID)
			return
		}
	}
	//
	// Add the request either to the onLedger request pool or time-locked request pool.
	mpi.onLedgerPool.Add(request)
	mpi.metrics.IncRequestsReceived(request)
}

func (mpi *mempoolImpl) handleReceiveOffLedgerRequest(request isc.OffLedgerRequest) {
	mpi.log.LogDebugf("Received request %v from outside.", request.ID())
	if mpi.addOffledger(request) {
		mpi.sendMessages(mpi.distSync.Input(distsync.NewInputPublishRequest(request)))
	}
}

func (mpi *mempoolImpl) handleTangleTimeUpdated(tangleTime time.Time) {
	oldTangleTime := mpi.tangleTime
	mpi.tangleTime = tangleTime
	//
	// Notify existing on-ledger requests if that's first time update.
	if oldTangleTime.IsZero() {
		mpi.onLedgerPool.Iterate(func(e *typedPoolEntry[isc.OnLedgerRequest]) bool {
			mpi.waitReq.MarkAvailable(e.req)
			return true
		})
	}
}

// - Re-add all the request from the reverted blocks.
// - Cleanup requests from the blocks that were added.
func (mpi *mempoolImpl) handleTrackNewChainHead(req *reqTrackNewChainHead) {
	defer close(req.responseCh)
	mpi.log.LogDebugf("handleTrackNewChainHead, from %v, current=%v", req.from, mpi.chainHeadAnchor)

	if len(req.removed) != 0 {
		mpi.log.LogInfof("Reorg detected, removing %v blocks, adding %v blocks", len(req.removed), len(req.added))
		// TODO: For IOTA 2.0: Maybe re-read the state from L1 (when reorgs will become possible).
	}
	//
	// Re-add requests from the blocks that are reverted now.
	for _, block := range req.removed {
		blockReceipts, err := blocklog.RequestReceiptsFromBlock(block)
		if err != nil {
			panic(fmt.Errorf("cannot extract receipts from block: %w", err))
		}
		for _, receipt := range blockReceipts {
			mpi.tryReAddRequest(receipt.Request)
		}
	}
	//
	// Cleanup the requests that were consumed in the added blocks.
	for _, block := range req.added {
		blockReceipts, err := blocklog.RequestReceiptsFromBlock(block)
		if err != nil {
			panic(fmt.Errorf("cannot extract receipts from block: %w", err))
		}
		mpi.metrics.IncBlocksPerChain()
		mpi.listener.BlockApplied(mpi.chainID, block, mpi.chainHeadState)
		for _, receipt := range blockReceipts {
			mpi.metrics.IncRequestsProcessed()
			mpi.tryRemoveRequest(receipt.Request)
		}
	}

	//
	// Cleanup processed requests, if that's the first time we received the state.
	if mpi.chainHeadState == nil {
		mpi.log.LogDebugf("Cleanup processed requests based on the received state...")
		mpi.tryCleanupProcessed(req.st)
		mpi.log.LogDebugf("Cleanup processed requests based on the received state... Done")
	}
	//
	// Record the head state.
	mpi.chainHeadState = req.st
	mpi.chainHeadAnchor = req.till
	//
	// Process the pending consensus proposal requests if any.
	if len(mpi.waitChainHead) != 0 {
		newWaitChainHead := []*reqConsensusProposal{}
		for i, waiting := range mpi.waitChainHead {
			if waiting.ctx.Err() != nil {
				continue // Drop it.
			}
			if waiting.anchor.Equals(mpi.chainHeadAnchor) {
				mpi.handleConsensusProposalForChainHead(waiting)
				continue // Drop it from wait queue.
			}
			newWaitChainHead = append(newWaitChainHead, mpi.waitChainHead[i])
		}
		mpi.waitChainHead = newWaitChainHead
	}

	// update defaultGasPrice for offLedger requests
	mpi.offLedgerPool.SetMinGasPrice(governance.NewStateReaderFromChainState(mpi.chainHeadState).GetDefaultGasPrice())
}

func (mpi *mempoolImpl) handleNetMessage(recv *peering.PeerMessageIn) {
	msg, err := mpi.distSync.UnmarshalMessage(recv.MsgData)
	if err != nil {
		mpi.log.LogWarnf("cannot parse message: %v", err)
		return
	}
	msg.SetSender(mpi.pubKeyAsNodeID(recv.SenderPubKey))
	outMsgs := mpi.distSync.Message(msg) // Output is handled via callbacks in this case.
	mpi.sendMessages(outMsgs)
}

func (mpi *mempoolImpl) handleDistSyncDebugTick() {
	mpi.log.LogDebugf(
		"Mempool onLedger=%v, offLedger=%v distSync=%v",
		mpi.onLedgerPool.StatusString(),
		mpi.offLedgerPool.StatusString(),
		mpi.distSync.StatusString(),
	)
}

func (mpi *mempoolImpl) handleDistSyncTimeTick() {
	mpi.sendMessages(mpi.distSync.Input(distsync.NewInputTimeTick()))
}

// Re-send off-ledger messages that are hanging here for a long time.
// Probably not a lot of nodes have them.
func (mpi *mempoolImpl) handleRePublishTimeTick() {
	if mpi.broadcastInterval == 0 {
		return // re-broadcasting is disabled
	}
	retryOlder := time.Now().Add(-mpi.broadcastInterval)
	mpi.offLedgerPool.Cleanup(func(request isc.OffLedgerRequest, ts time.Time) bool {
		if ts.Before(retryOlder) {
			mpi.sendMessages(mpi.distSync.Input(distsync.NewInputPublishRequest(request)))
		}
		return true
	})

	// periodically try to refresh On-ledger requests that might have been dropped
	if time.Since(mpi.lastRefreshTimestamp) > mpi.settings.OnLedgerRefreshMinInterval {
		if mpi.onLedgerPool.ShouldRefreshRequests() {
			mpi.refreshOnLedgerRequests()
			mpi.lastRefreshTimestamp = time.Now()
		}
	}
}

func (mpi *mempoolImpl) handleForceCleanMempool() {
	mpi.offLedgerPool.Iterate(func(account string, entries []*OrderedPoolEntry) bool {
		for _, e := range entries {
			if time.Since(e.ts) > mpi.settings.TTL && !e.proposedForAny(mpi.consensusInstances) {
				mpi.log.LogDebugf("handleForceCleanMempool, request TTL expired, removing: %s", e.req.ID().String())
				mpi.offLedgerPool.Remove(e.req)
			}
		}

		return true
	})
}

func (mpi *mempoolImpl) tryReAddRequest(req isc.Request) {
	switch req := req.(type) {
	case isc.OnLedgerRequest:
		// TODO: For IOTA 2.0: We will have to check, if the request has been reverted with the reorg.
		//
		// For now, the L1 cannot revert committed outputs and all the on-ledger requests
		// are received, when they are committed. Therefore it is safe now to re-add the
		// requests, because they were consumed in an uncommitted (and now reverted) transactions.
		mpi.log.LogDebugf("re-adding on-ledger request to mempool: %s", req.ID())
		mpi.onLedgerPool.Add(req)
	case isc.OffLedgerRequest:
		mpi.log.LogDebugf("re-adding off-ledger request to mempool: %s", req.ID())
		mpi.offLedgerPool.Add(req)
	default:
		panic(fmt.Errorf("unexpected request type: %T", req))
	}
}

func (mpi *mempoolImpl) tryRemoveRequest(req isc.Request) {
	switch req := req.(type) {
	case isc.OnLedgerRequest:
		mpi.log.LogDebugf("removing on-ledger request from mempool: %s", req.ID())
		mpi.onLedgerPool.Remove(req)
	case isc.OffLedgerRequest:
		mpi.log.LogDebugf("removing off-ledger request from mempool: %s", req.ID())
		mpi.offLedgerPool.Remove(req)
	default:
		mpi.log.LogWarnf("Trying to remove request of unexpected type %T: %+v", req, req)
	}
}

func (mpi *mempoolImpl) tryCleanupProcessed(chainState state.State) {
	mpi.onLedgerPool.Cleanup(unprocessedPredicate[isc.OnLedgerRequest](chainState, mpi.log))
	mpi.offLedgerPool.Cleanup(unprocessedPredicate[isc.OffLedgerRequest](chainState, mpi.log))
}

func (mpi *mempoolImpl) sendMessages(outMsgs gpa.OutMessages) {
	if outMsgs == nil {
		return
	}
	outMsgs.MustIterate(func(msg gpa.Message) {
		msgBytes := lo.Must(gpa.MarshalMessage(msg))
		pm := peering.NewPeerMessageData(mpi.netPeeringID, peering.ReceiverMempool, msgTypeMempool, msgBytes)
		mpi.net.SendMsgByPubKey(mpi.netPeerPubs[msg.Recipient()], pm)
	})
}

func (mpi *mempoolImpl) pubKeyAsNodeID(pubKey *cryptolib.PublicKey) gpa.NodeID {
	nodeID := gpa.NodeIDFromPublicKey(pubKey)
	if _, ok := mpi.netPeerPubs[nodeID]; !ok {
		mpi.netPeerPubs[nodeID] = pubKey
	}
	return nodeID
}

// To use with lo.Map.
func (mpi *mempoolImpl) pubKeyAsNodeIDMap(nodePubKey *cryptolib.PublicKey, _ int) gpa.NodeID {
	return mpi.pubKeyAsNodeID(nodePubKey)
}

// Have to have it as a separate function to be able to use type params.
func unprocessedPredicate[V isc.Request](chainState state.State, log log.Logger) func(V, time.Time) bool {
	return func(request V, ts time.Time) bool {
		requestID := request.ID()

		processed, err := blocklog.NewStateReaderFromChainState(chainState).IsRequestProcessed(requestID)
		if err != nil {
			log.LogWarn("Cannot check if request %v is processed at state.TrieRoot=%v, err=%v", requestID, chainState.TrieRoot(), err)
			return false
		}

		if processed {
			log.LogDebugf("Request already processed %v at state.TrieRoot=%v", requestID, chainState.TrieRoot())
			return false
		}

		return true
	}
}
