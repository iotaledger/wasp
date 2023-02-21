// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

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
// Upon reception of the proposal query (ConsensusProposalsAsync) for NextAO
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
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/core/logger"
	consGR "github.com/iotaledger/wasp/packages/chain/cons/gr"
	"github.com/iotaledger/wasp/packages/chain/mempool/distSync"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

const (
	distShareDebugTick      = 10 * time.Second
	distShareTimeTick       = 3 * time.Second
	distShareMaxMsgsPerTick = 100
	distShareRePublishTick  = 5 * time.Second
	waitRequestCleanupEvery = 10
)

// Partial interface for providing chain events to the outside.
// This interface is in the mempool part only because it tracks
// the actual state for checking the consumed requests.
type ChainListener interface {
	// This function is called by the chain when new block is applied to the
	// state. This block might be not confirmed yet, but the chain is going
	// to build the next block on top of this one.
	BlockApplied(chainID isc.ChainID, block state.Block)
}

type Mempool interface {
	consGR.Mempool
	// Invoked by the chain, when new alias output is considered as a tip/head
	// of the chain. Mempool can reorganize its state by removing/rejecting
	// or re-adding some requests, depending on how the head has changed.
	// It can mean simple advance of the chain, or a rollback or a reorg.
	// This function is guaranteed to be called in the order, which is
	// considered the chain block order by the ChainMgr.
	TrackNewChainHead(st state.State, from, till *isc.AliasOutputWithID, added, removed []state.Block)
	// Invoked by the chain when a new off-ledger request is received from a node user.
	// Inter-node off-ledger dissemination is NOT performed via this function.
	ReceiveOnLedgerRequest(request isc.OnLedgerRequest)
	// This is called when this node receives an off-ledger request from a user directly.
	// I.e. when this node is an entry point of the off-ledger request.
	ReceiveOffLedgerRequest(request isc.OffLedgerRequest)
	// Invoked by the ChainMgr when a time of a tangle changes.
	TangleTimeUpdated(tangleTime time.Time)
	// Invoked by the chain when a set of server nodes has changed.
	// These nodes should be used to disseminate the off-ledger requests.
	ServerNodesUpdated(committeePubKeys []*cryptolib.PublicKey, serverNodePubKeys []*cryptolib.PublicKey)
	AccessNodesUpdated(committeePubKeys []*cryptolib.PublicKey, accessNodePubKeys []*cryptolib.PublicKey)
}

type RequestPool[V isc.Request] interface {
	Has(reqRef *isc.RequestRef) bool
	Get(reqRef *isc.RequestRef) V
	Add(request V)
	Remove(request V)
	Filter(predicate func(request V, ts time.Time) bool)
	StatusString() string
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
	timePool                       TimePool
	onLedgerPool                   RequestPool[isc.OnLedgerRequest]
	offLedgerPool                  RequestPool[isc.OffLedgerRequest]
	distSync                       gpa.GPA
	chainHeadAO                    *isc.AliasOutputWithID
	chainHeadState                 state.State
	serverNodesUpdatedPipe         pipe.Pipe[*reqServerNodesUpdated]
	serverNodes                    []*cryptolib.PublicKey
	accessNodesUpdatedPipe         pipe.Pipe[*reqAccessNodesUpdated]
	accessNodes                    []*cryptolib.PublicKey
	committeeNodes                 []*cryptolib.PublicKey
	waitReq                        WaitReq
	waitChainHead                  []*reqConsensusProposals
	reqConsensusProposalsPipe      pipe.Pipe[*reqConsensusProposals]
	reqConsensusRequestsPipe       pipe.Pipe[*reqConsensusRequests]
	reqReceiveOnLedgerRequestPipe  pipe.Pipe[isc.OnLedgerRequest]
	reqReceiveOffLedgerRequestPipe pipe.Pipe[isc.OffLedgerRequest]
	reqTangleTimeUpdatedPipe       pipe.Pipe[time.Time]
	reqTrackNewChainHeadPipe       pipe.Pipe[*reqTrackNewChainHead]
	netRecvPipe                    pipe.Pipe[*peering.PeerMessageIn]
	netPeeringID                   peering.PeeringID
	netPeerPubs                    map[gpa.NodeID]*cryptolib.PublicKey
	net                            peering.NetworkProvider
	log                            *logger.Logger
	metrics                        metrics.MempoolMetrics
	listener                       ChainListener
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

type reqConsensusProposals struct {
	ctx         context.Context
	aliasOutput *isc.AliasOutputWithID
	responseCh  chan<- []*isc.RequestRef
}

type reqConsensusRequests struct {
	ctx         context.Context
	requestRefs []*isc.RequestRef
	responseCh  chan<- []isc.Request
}

type reqTrackNewChainHead struct {
	st      state.State
	from    *isc.AliasOutputWithID
	till    *isc.AliasOutputWithID
	added   []state.Block
	removed []state.Block
}

func New(
	ctx context.Context,
	chainID isc.ChainID,
	nodeIdentity *cryptolib.KeyPair,
	net peering.NetworkProvider,
	log *logger.Logger,
	metrics metrics.MempoolMetrics,
	listener ChainListener,
) Mempool {
	netPeeringID := peering.HashPeeringIDFromBytes(chainID.Bytes(), []byte("Mempool")) // ChainID × Mempool
	waitReq := NewWaitReq(waitRequestCleanupEvery)
	mpi := &mempoolImpl{
		chainID:                        chainID,
		tangleTime:                     time.Time{},
		timePool:                       NewTimePool(),
		onLedgerPool:                   NewTypedPool[isc.OnLedgerRequest](waitReq),
		offLedgerPool:                  NewTypedPool[isc.OffLedgerRequest](waitReq),
		chainHeadAO:                    nil,
		serverNodesUpdatedPipe:         pipe.NewInfinitePipe[*reqServerNodesUpdated](),
		serverNodes:                    []*cryptolib.PublicKey{},
		accessNodesUpdatedPipe:         pipe.NewInfinitePipe[*reqAccessNodesUpdated](),
		accessNodes:                    []*cryptolib.PublicKey{},
		committeeNodes:                 []*cryptolib.PublicKey{},
		waitReq:                        waitReq,
		waitChainHead:                  []*reqConsensusProposals{},
		reqConsensusProposalsPipe:      pipe.NewInfinitePipe[*reqConsensusProposals](),
		reqConsensusRequestsPipe:       pipe.NewInfinitePipe[*reqConsensusRequests](),
		reqReceiveOnLedgerRequestPipe:  pipe.NewInfinitePipe[isc.OnLedgerRequest](),
		reqReceiveOffLedgerRequestPipe: pipe.NewInfinitePipe[isc.OffLedgerRequest](),
		reqTangleTimeUpdatedPipe:       pipe.NewInfinitePipe[time.Time](),
		reqTrackNewChainHeadPipe:       pipe.NewInfinitePipe[*reqTrackNewChainHead](),
		netRecvPipe:                    pipe.NewInfinitePipe[*peering.PeerMessageIn](),
		netPeeringID:                   netPeeringID,
		netPeerPubs:                    map[gpa.NodeID]*cryptolib.PublicKey{},
		net:                            net,
		log:                            log,
		metrics:                        metrics,
		listener:                       listener,
	}
	mpi.distSync = distSync.New(
		mpi.pubKeyAsNodeID(nodeIdentity.GetPublicKey()),
		mpi.distSyncRequestNeededCB,
		mpi.distSyncRequestReceivedCB,
		distShareMaxMsgsPerTick,
		log,
	)
	netRecvPipeInCh := mpi.netRecvPipe.In()
	netAttachID := net.Attach(&netPeeringID, peering.PeerMessageReceiverMempool, func(recv *peering.PeerMessageIn) {
		if recv.MsgType != msgTypeMempool {
			mpi.log.Warnf("Unexpected message, type=%v", recv.MsgType)
			return
		}
		netRecvPipeInCh <- recv
	})
	go mpi.run(ctx, netAttachID)
	return mpi
}

func (mpi *mempoolImpl) TangleTimeUpdated(tangleTime time.Time) {
	mpi.reqTangleTimeUpdatedPipe.In() <- tangleTime
}

func (mpi *mempoolImpl) TrackNewChainHead(st state.State, from, till *isc.AliasOutputWithID, added, removed []state.Block) {
	mpi.reqTrackNewChainHeadPipe.In() <- &reqTrackNewChainHead{st, from, till, added, removed}
}

func (mpi *mempoolImpl) ReceiveOnLedgerRequest(request isc.OnLedgerRequest) {
	mpi.reqReceiveOnLedgerRequestPipe.In() <- request
}

func (mpi *mempoolImpl) ReceiveOffLedgerRequest(request isc.OffLedgerRequest) {
	mpi.reqReceiveOffLedgerRequestPipe.In() <- request
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

func (mpi *mempoolImpl) ConsensusProposalsAsync(ctx context.Context, aliasOutput *isc.AliasOutputWithID) <-chan []*isc.RequestRef {
	res := make(chan []*isc.RequestRef, 1)
	req := &reqConsensusProposals{
		ctx:         ctx,
		aliasOutput: aliasOutput,
		responseCh:  res,
	}
	mpi.reqConsensusProposalsPipe.In() <- req
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

func (mpi *mempoolImpl) run(ctx context.Context, netAttachID interface{}) { //nolint: gocyclo
	serverNodesUpdatedPipeOutCh := mpi.serverNodesUpdatedPipe.Out()
	accessNodesUpdatedPipeOutCh := mpi.accessNodesUpdatedPipe.Out()
	reqConsensusProposalsPipeOutCh := mpi.reqConsensusProposalsPipe.Out()
	reqConsensusRequestsPipeOutCh := mpi.reqConsensusRequestsPipe.Out()
	reqReceiveOnLedgerRequestPipeOutCh := mpi.reqReceiveOnLedgerRequestPipe.Out()
	reqReceiveOffLedgerRequestPipeOutCh := mpi.reqReceiveOffLedgerRequestPipe.Out()
	reqTangleTimeUpdatedPipeOutCh := mpi.reqTangleTimeUpdatedPipe.Out()
	reqTrackNewChainHeadPipeOutCh := mpi.reqTrackNewChainHeadPipe.Out()
	netRecvPipeOutCh := mpi.netRecvPipe.Out()
	debugTicker := time.NewTicker(distShareDebugTick)
	timeTicker := time.NewTicker(distShareTimeTick)
	rePublishTicker := time.NewTicker(distShareRePublishTick)
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
		case recv, ok := <-reqConsensusProposalsPipeOutCh:
			if !ok {
				reqConsensusProposalsPipeOutCh = nil
				break
			}
			mpi.handleConsensusProposals(recv)
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
		case <-ctx.Done():
			// mpi.serverNodesUpdatedPipe.Close() // TODO: Causes panic: send on closed channel
			// mpi.accessNodesUpdatedPipe.Close()
			// mpi.reqConsensusProposalsPipe.Close()
			// mpi.reqConsensusRequestsPipe.Close()
			// mpi.reqReceiveOnLedgerRequestPipe.Close()
			// mpi.reqReceiveOffLedgerRequestPipe.Close()
			// mpi.reqTangleTimeUpdatedPipe.Close()
			// mpi.reqTrackNewChainHeadPipe.Close()
			// mpi.netRecvPipe.Close()
			debugTicker.Stop()
			timeTicker.Stop()
			mpi.net.Detach(netAttachID)
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
		return req
	}
	if mpi.chainHeadState != nil {
		requestID := requestRef.ID
		receipt, err := blocklog.GetRequestReceipt(mpi.chainHeadState, requestID)
		if err == nil && receipt != nil && receipt.Request.IsOffLedger() {
			return receipt.Request
		}
		return nil
	}
	return nil
}

// A callback for distSync.
func (mpi *mempoolImpl) distSyncRequestReceivedCB(request isc.Request) {
	offLedgerReq, ok := request.(isc.OffLedgerRequest)
	if !ok {
		mpi.log.Warn("Dropping non-OffLedger request form dist %T: %+v", request, request)
		return
	}
	mpi.addOffLedgerRequestIfUnseen(offLedgerReq)
}

func (mpi *mempoolImpl) addOffLedgerRequestIfUnseen(request isc.OffLedgerRequest) bool {
	mpi.log.Debugf("trying to add to mempool, requestID: %s", request.ID().String())
	if mpi.chainHeadState != nil {
		requestID := request.ID()
		processed, err := blocklog.IsRequestProcessed(mpi.chainHeadState, requestID)
		if err != nil {
			panic(fmt.Errorf(
				"cannot check if request.ID=%v is processed in the blocklog at state=%v: %w",
				requestID,
				mpi.chainHeadState,
				err,
			))
		}
		if processed {
			return false // Already processed.
		}
	}
	if !mpi.offLedgerPool.Has(isc.RequestRefFromRequest(request)) {
		mpi.offLedgerPool.Add(request)
		mpi.metrics.IncRequestsReceived(request)
		mpi.log.Debugf("accepted by the mempool, requestID: %s", request.ID().String())
		return true
	}
	return false
}

func (mpi *mempoolImpl) handleServerNodesUpdated(recv *reqServerNodesUpdated) {
	mpi.serverNodes = recv.serverNodePubKeys
	mpi.committeeNodes = recv.committeePubKeys
	mpi.sendMessages(mpi.distSync.Input(distSync.NewInputServerNodes(
		lo.Map(mpi.serverNodes, mpi.pubKeyAsNodeIDMap),
		lo.Map(mpi.committeeNodes, mpi.pubKeyAsNodeIDMap),
	)))
}

func (mpi *mempoolImpl) handleAccessNodesUpdated(recv *reqAccessNodesUpdated) {
	mpi.accessNodes = recv.accessNodePubKeys
	mpi.committeeNodes = recv.committeePubKeys
	mpi.sendMessages(mpi.distSync.Input(distSync.NewInputAccessNodes(
		lo.Map(mpi.accessNodes, mpi.pubKeyAsNodeIDMap),
		lo.Map(mpi.committeeNodes, mpi.pubKeyAsNodeIDMap),
	)))
}

// This implementation only tracks a single branch. So, we will only respond
// to the request matching the TrackNewChainHead call.
func (mpi *mempoolImpl) handleConsensusProposals(recv *reqConsensusProposals) {
	if mpi.chainHeadAO == nil || !recv.aliasOutput.Equals(mpi.chainHeadAO) {
		mpi.log.Debugf("handleConsensusProposals, have to wait for chain head to become %v", recv.aliasOutput)
		mpi.waitChainHead = append(mpi.waitChainHead, recv)
		return
	}
	mpi.log.Debugf("handleConsensusProposals, already have the chain head %v", recv.aliasOutput)
	mpi.handleConsensusProposalsForChainHead(recv)
}

func (mpi *mempoolImpl) handleConsensusProposalsForChainHead(recv *reqConsensusProposals) {
	//
	// The case for matching ChainHeadAO and request BaseAO
	reqRefs := []*isc.RequestRef{}
	if !mpi.tangleTime.IsZero() { // Wait for tangle-time to process the on ledger requests.
		mpi.onLedgerPool.Filter(func(request isc.OnLedgerRequest, ts time.Time) bool {
			if isc.RequestIsExpired(request, mpi.tangleTime) {
				return false // Drop it from the mempool
			}
			if isc.RequestIsUnlockable(request, mpi.chainID.AsAddress(), mpi.tangleTime) {
				reqRefs = append(reqRefs, isc.RequestRefFromRequest(request))
			}
			return true // Keep them for now
		})
	}
	mpi.offLedgerPool.Filter(func(request isc.OffLedgerRequest, ts time.Time) bool {
		reqRefs = append(reqRefs, isc.RequestRefFromRequest(request))
		return true // Keep them for now
	})
	if len(reqRefs) > 0 {
		recv.responseCh <- reqRefs
		close(recv.responseCh)
		return
	}
	//
	// Wait for any request.
	mpi.waitReq.WaitAny(recv.ctx, func(req isc.Request) {
		recv.responseCh <- []*isc.RequestRef{isc.RequestRefFromRequest(req)}
		close(recv.responseCh)
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
		mpi.sendMessages(mpi.distSync.Input(distSync.NewInputRequestNeeded(missing[i], true)))
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
	if _, ok := request.Features().ReturnAmount(); ok {
		mpi.log.Warnf("dropping request, because it has ReturnAmount, ID=%v", requestID)
		return
	}
	//
	// Check, maybe mempool already has it.
	if mpi.onLedgerPool.Has(requestRef) || mpi.timePool.Has(requestRef) {
		return
	}
	//
	// Maybe it has been processed before?
	if mpi.chainHeadState != nil {
		processed, err := blocklog.IsRequestProcessed(mpi.chainHeadState, requestID)
		if err != nil {
			panic(fmt.Errorf("cannot check if request was processed: %w", err))
		}
		if processed {
			return
		}
	}
	//
	// Add the request either to the onLedger request pool or time-locked request pool.
	reqUnlockCondSet := request.Output().UnlockConditionSet()
	timeLock := reqUnlockCondSet.Timelock()
	if timeLock != nil {
		expiration := reqUnlockCondSet.Expiration()
		if expiration != nil && timeLock.UnixTime >= expiration.UnixTime {
			// can never be processed, just reject
			return
		}
		if mpi.tangleTime.IsZero() || timeLock.UnixTime > uint32(mpi.tangleTime.Unix()) {
			mpi.timePool.AddRequest(time.Unix(int64(timeLock.UnixTime), 0), request)
			return
		}
	}
	mpi.onLedgerPool.Add(request)
	mpi.metrics.IncRequestsReceived(request)
}

func (mpi *mempoolImpl) handleReceiveOffLedgerRequest(request isc.OffLedgerRequest) {
	if mpi.addOffLedgerRequestIfUnseen(request) {
		mpi.sendMessages(mpi.distSync.Input(distSync.NewInputPublishRequest(request)))
	}
}

func (mpi *mempoolImpl) handleTangleTimeUpdated(tangleTime time.Time) {
	oldTangleTime := mpi.tangleTime
	mpi.tangleTime = tangleTime
	//
	// Add requests from time locked pool.
	reqs := mpi.timePool.TakeTill(tangleTime)
	for i := range reqs {
		switch req := reqs[i].(type) {
		case isc.OnLedgerRequest:
			mpi.onLedgerPool.Add(req)
			mpi.metrics.IncRequestsReceived(req)
		case isc.OffLedgerRequest:
			mpi.offLedgerPool.Add(req)
			mpi.metrics.IncRequestsReceived(req)
		default:
			panic(fmt.Errorf("unexpected request type: %T, %+v", req, req))
		}
	}
	//
	// Notify existing on-ledger requests if that's first time update.
	if oldTangleTime.IsZero() {
		mpi.onLedgerPool.Filter(func(request isc.OnLedgerRequest, ts time.Time) bool {
			mpi.waitReq.Have(request)
			return true
		})
	}
}

// - Re-add all the request from the reverted blocks.
// - Cleanup requests from the blocks that were added.
func (mpi *mempoolImpl) handleTrackNewChainHead(req *reqTrackNewChainHead) {
	mpi.log.Debugf("handleTrackNewChainHead, %v from %v, current=%v", req.till, req.from, mpi.chainHeadAO)
	if len(req.removed) != 0 {
		mpi.log.Infof("Reorg detected, removing %v blocks, adding %v blocks", len(req.removed), len(req.added))
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
		mpi.listener.BlockApplied(mpi.chainID, block)
		for _, receipt := range blockReceipts {
			mpi.metrics.IncRequestsProcessed()
			mpi.tryRemoveRequest(receipt.Request)
			mpi.log.Debugf("removed from mempool: %s", receipt.Request.ID().String())
		}
	}
	//
	// Cleanup processed requests, if that's the first time we received the state.
	if mpi.chainHeadState == nil {
		mpi.log.Debugf("Cleanup processed requests based on the received state...")
		mpi.tryCleanupProcessed(req.st)
		mpi.log.Debugf("Cleanup processed requests based on the received state... Done")
	}
	//
	// Record the head state.
	mpi.chainHeadState = req.st
	mpi.chainHeadAO = req.till
	//
	// Process the pending consensus proposal requests if any.
	if len(mpi.waitChainHead) != 0 {
		newWaitChainHead := []*reqConsensusProposals{}
		for i, waiting := range mpi.waitChainHead {
			if waiting.ctx.Err() != nil {
				continue // Drop it.
			}
			if waiting.aliasOutput.Equals(mpi.chainHeadAO) {
				mpi.handleConsensusProposalsForChainHead(waiting)
				continue // Drop it from wait queue.
			}
			newWaitChainHead = append(newWaitChainHead, mpi.waitChainHead[i])
		}
		mpi.waitChainHead = newWaitChainHead
	}
}

func (mpi *mempoolImpl) handleNetMessage(recv *peering.PeerMessageIn) {
	msg, err := mpi.distSync.UnmarshalMessage(recv.MsgData)
	if err != nil {
		mpi.log.Warnf("cannot parse message: %v", err)
		return
	}
	msg.SetSender(mpi.pubKeyAsNodeID(recv.SenderPubKey))
	outMsgs := mpi.distSync.Message(msg) // Output is handled via callbacks in this case.
	mpi.sendMessages(outMsgs)
}

func (mpi *mempoolImpl) handleDistSyncDebugTick() {
	mpi.log.Debugf(
		"Mempool onLedger=%v, offLedger=%v distSync=%v",
		mpi.onLedgerPool.StatusString(),
		mpi.offLedgerPool.StatusString(),
		mpi.distSync.StatusString(),
	)
}

func (mpi *mempoolImpl) handleDistSyncTimeTick() {
	mpi.sendMessages(mpi.distSync.Input(distSync.NewInputTimeTick()))
}

// Re-send off-ledger messages that are hanging here for a long time.
// Probably not a lot of nodes have them.
func (mpi *mempoolImpl) handleRePublishTimeTick() {
	retryOlder := time.Now().Add(-distShareRePublishTick)
	mpi.offLedgerPool.Filter(func(request isc.OffLedgerRequest, ts time.Time) bool {
		if ts.Before(retryOlder) {
			mpi.sendMessages(mpi.distSync.Input(distSync.NewInputPublishRequest(request)))
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
		mpi.onLedgerPool.Add(req)
	case isc.OffLedgerRequest:
		mpi.offLedgerPool.Add(req)
	default:
		panic(fmt.Errorf("unexpected request type: %T", req))
	}
}

func (mpi *mempoolImpl) tryRemoveRequest(req isc.Request) {
	switch req := req.(type) {
	case isc.OnLedgerRequest:
		mpi.onLedgerPool.Remove(req)
	case isc.OffLedgerRequest:
		mpi.offLedgerPool.Remove(req)
	default:
		mpi.log.Warn("Trying to remove request of unexpected type %T: %+v", req, req)
	}
}

func (mpi *mempoolImpl) tryCleanupProcessed(chainState state.State) {
	mpi.onLedgerPool.Filter(unprocessedPredicate[isc.OnLedgerRequest](chainState, mpi.log))
	mpi.offLedgerPool.Filter(unprocessedPredicate[isc.OffLedgerRequest](chainState, mpi.log))
	mpi.timePool.Filter(unprocessedPredicate[isc.Request](chainState, mpi.log))
}

func (mpi *mempoolImpl) sendMessages(outMsgs gpa.OutMessages) {
	if outMsgs == nil {
		return
	}
	outMsgs.MustIterate(func(m gpa.Message) {
		msgData, err := m.MarshalBinary()
		if err != nil {
			mpi.log.Warnf("Failed to send a message: %v", err)
			return
		}
		pm := &peering.PeerMessageData{
			PeeringID:   mpi.netPeeringID,
			MsgReceiver: peering.PeerMessageReceiverMempool,
			MsgType:     msgTypeMempool,
			MsgData:     msgData,
		}
		mpi.net.SendMsgByPubKey(mpi.netPeerPubs[m.Recipient()], pm)
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
func unprocessedPredicate[V isc.Request](chainState state.State, log *logger.Logger) func(V, time.Time) bool {
	return func(request V, ts time.Time) bool {
		requestID := request.ID()

		processed, err := blocklog.IsRequestProcessed(chainState, requestID)
		if err != nil {
			log.Warn("Cannot check if request %v is processed at state.TrieRoot=%v, err=%v", requestID, chainState.TrieRoot(), err)
			return false
		}

		if processed {
			log.Debugf("Request already processed %v at state.TrieRoot=%v", requestID, chainState.TrieRoot())
			return false
		}

		return true
	}
}
