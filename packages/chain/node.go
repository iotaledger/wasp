// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// This runs single chain will all the committees, mempool, state mgr etc.
// The main task for this package to run the protocol as in a threaded environment,
// communicate between ChainMgr, Mempool, StateMgr, NodeConn and ConsensusInstances.
//
// The following threads (goroutines) are running for a chain:
//   - ChainMgr (the main synchronizing thread)
//   - Mempool
//   - StateMgr
//   - Consensus (a thread for each instance).
//
// This object interacts with:
//   - NodeConn.
//   - Administrative functions.
package chain

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/chainmanager"
	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/chain/cons"
	consGR "github.com/iotaledger/wasp/packages/chain/cons/cons_gr"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/statemanager"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_snapshots"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

const (
	msgTypeChainMgr byte = iota
)

var (
	RedeliveryPeriod         = 2 * time.Second
	PrintStatusPeriod        = 3 * time.Second
	ConsensusInstsInAdvance  = 3
	AwaitReceiptCleanupEvery = 100
)

type ChainRequests interface {
	ReceiveOffLedgerRequest(request isc.OffLedgerRequest, sender *cryptolib.PublicKey) error
	AwaitRequestProcessed(ctx context.Context, requestID isc.RequestID, confirmed bool) <-chan *blocklog.RequestReceipt
}

type Chain interface {
	ChainCore
	ChainRequests
	// This is invoked when a node owner updates the chain configuration,
	// possibly to update the per-node accessNode list.
	ConfigUpdated(accessNodesPerNode []*cryptolib.PublicKey)
	// This is invoked when the accessMgr determines the nodes which
	// consider this node as an access node for this chain. The chain
	// can query the nodes for blocks, etc. NOTE: servers = access⁻¹
	ServersUpdated(serverNodes []*cryptolib.PublicKey)
	// Metrics and the current descriptive state.
	GetChainMetrics() *metrics.ChainMetrics
	GetConsensusPipeMetrics() ConsensusPipeMetrics // TODO: Review this.
	GetConsensusWorkflowStatus() ConsensusWorkflowStatus
}

type CommitteeInfo struct {
	Address       iotago.Address
	Size          uint16
	Quorum        uint16
	QuorumIsAlive bool
	PeerStatus    []*PeerStatus
}

type PeerStatus struct {
	Name       string
	Index      uint16
	PubKey     *cryptolib.PublicKey
	PeeringURL string
	Connected  bool
}

type RequestOutputHandler = func(outputInfo *isc.OutputInfo)

// The Alias Outputs must be passed here in-order. The last alias output in the list
// is the unspent one (if there is a chain of outputs confirmed in a milestone).
type AliasOutputHandler = func(outputInfo *isc.OutputInfo)

type TxPostHandler = func(tx *iotago.Transaction, confirmed bool)

type MilestoneHandler = func(timestamp time.Time)

type ChainNodeConn interface {
	// Publishing can be canceled via the context.
	// The result must be returned via the callback, unless ctx is canceled first.
	// PublishTX handles promoting and reattachments until the tx is confirmed or the context is canceled.
	PublishTX(
		ctx context.Context,
		chainID isc.ChainID,
		tx *iotago.Transaction,
		callback TxPostHandler,
	) error
	// Alias outputs are expected to be returned in order. Considering the Hornet node, the rules are:
	//   - Upon Attach -- existing unspent alias output is returned FIRST.
	//   - Upon receiving a spent/unspent AO from L1 they are returned in
	//     the same order, as the milestones are issued.
	//   - If a single milestone has several alias outputs, they have to be ordered
	//     according to the chain of TXes.
	//
	// NOTE: Any out-of-order AO will be considered as a rollback or AO by the chain impl.
	AttachChain(
		ctx context.Context,
		chainID isc.ChainID,
		recvRequestCB RequestOutputHandler,
		recvAliasOutput AliasOutputHandler,
		recvMilestone MilestoneHandler,
		onChainConnect func(),
		onChainDisconnect func(),
	)
}

type chainNodeImpl struct {
	me                  gpa.NodeID
	nodeIdentity        *cryptolib.KeyPair
	chainID             isc.ChainID
	chainMgr            gpa.AckHandler
	chainStore          indexedstore.IndexedStore
	nodeConn            NodeConnection
	tangleTime          time.Time
	mempool             mempool.Mempool
	stateMgr            statemanager.StateMgr
	recvAliasOutputPipe pipe.Pipe[*isc.AliasOutputWithID]
	recvTxPublishedPipe pipe.Pipe[*txPublished]
	recvMilestonePipe   pipe.Pipe[time.Time]
	consensusInsts      *shrinkingmap.ShrinkingMap[iotago.Ed25519Address, *shrinkingmap.ShrinkingMap[cmt_log.LogIndex, *consensusInst]] // Running consensus instances.
	consOutputPipe      pipe.Pipe[*consOutput]
	consRecoverPipe     pipe.Pipe[*consRecover]
	publishingTXes      *shrinkingmap.ShrinkingMap[iotago.TransactionID, context.CancelFunc] // TX'es now being published.
	procCache           *processors.Cache                                                    // Cache for the SC processors.
	configUpdatedCh     chan *configUpdate
	serversUpdatedPipe  pipe.Pipe[*serversUpdate]
	awaitReceiptActCh   chan *awaitReceiptReq
	awaitReceiptCnfCh   chan *awaitReceiptReq
	stateTrackerAct     StateTracker
	stateTrackerCnf     StateTracker
	blockWAL            sm_gpa_utils.BlockWAL
	//
	// Configuration values.
	consensusDelay   time.Duration
	recoveryTimeout  time.Duration
	validatorAgentID isc.AgentID
	//
	// Information for other components.
	listener               ChainListener          // Object expecting event notifications.
	accessLock             *sync.RWMutex          // Mutex for accessing informative fields from other threads.
	activeCommitteeDKShare tcrypto.DKShare        // DKShare of the current active committee.
	activeCommitteeNodes   []*cryptolib.PublicKey // The nodes acting as a committee for the latest consensus.
	activeAccessNodes      []*cryptolib.PublicKey // All the nodes authorized for being access nodes (∪{{Self}, accessNodesFromNode, accessNodesFrom{ACT, CNF}}, activeCommitteeNodes}).
	accessNodesFromNode    []*cryptolib.PublicKey // Access nodes, as configured locally by a user in this node.
	accessNodesFromCNF     []*cryptolib.PublicKey // Access nodes, as configured in the governance contract (for the active state).
	accessNodesFromACT     []*cryptolib.PublicKey // Access nodes, as configured in the governance contract (for the confirmed state).
	serverNodes            []*cryptolib.PublicKey // The nodes we can query (because they consider us an access node).
	latestConfirmedAO      *isc.AliasOutputWithID // Confirmed by L1, can be lagging from latestActiveAO.
	latestConfirmedState   state.State            // State corresponding to latestConfirmedAO, for performance reasons.
	latestConfirmedStateAO *isc.AliasOutputWithID // Set only when the corresponding state is retrieved.
	latestActiveAO         *isc.AliasOutputWithID // This is the AO the chain is build on.
	latestActiveState      state.State            // State corresponding to latestActiveAO, for performance reasons.
	latestActiveStateAO    *isc.AliasOutputWithID // Set only when the corresponding state is retrieved.
	//
	// Infrastructure.
	netRecvPipe         pipe.Pipe[*peering.PeerMessageIn]
	netPeeringID        peering.PeeringID
	netPeerPubs         map[gpa.NodeID]*cryptolib.PublicKey
	net                 peering.NetworkProvider
	shutdownCoordinator *shutdown.Coordinator
	chainMetrics        *metrics.ChainMetrics
	log                 *logger.Logger
}

type consensusInst struct {
	request    *chainmanager.NeedConsensus
	cancelFunc context.CancelFunc
	consensus  *consGR.ConsGr
	committee  []*cryptolib.PublicKey
}

func (ci *consensusInst) Cancel() {
	if ci.cancelFunc == nil {
		return
	}
	ci.cancelFunc()
	ci.cancelFunc = nil
}

// Used to correlate consensus request with its output.
type consOutput struct {
	request *chainmanager.NeedConsensus
	output  *consGR.Output
}

func (co *consOutput) String() string {
	return fmt.Sprintf("{cons.consOutput, request=%v, output=%v}", co.request, co.output)
}

// Used to correlate consensus request with its output.
type consRecover struct {
	request *chainmanager.NeedConsensus
}

func (cr *consRecover) String() string {
	return fmt.Sprintf("{cons.consRecover, request=%v}", cr.request)
}

// This is event received from the NodeConn as response to PublishTX
type txPublished struct {
	committeeAddr   iotago.Ed25519Address
	logIndex        cmt_log.LogIndex
	txID            iotago.TransactionID
	nextAliasOutput *isc.AliasOutputWithID
	confirmed       bool
}

// Represents config update event locally on this node.
type configUpdate struct {
	accessNodes []*cryptolib.PublicKey
}

type serversUpdate struct {
	serverNodes []*cryptolib.PublicKey
}

var _ Chain = &chainNodeImpl{}

//nolint:funlen
func New(
	ctx context.Context,
	log *logger.Logger,
	chainID isc.ChainID,
	chainStore indexedstore.IndexedStore,
	nodeConn NodeConnection,
	nodeIdentity *cryptolib.KeyPair,
	processorConfig *processors.Config,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	consensusStateRegistry cmt_log.ConsensusStateRegistry,
	recoverFromWAL bool,
	blockWAL sm_gpa_utils.BlockWAL,
	snapshotManager sm_snapshots.SnapshotManager,
	listener ChainListener,
	accessNodesFromNode []*cryptolib.PublicKey,
	net peering.NetworkProvider,
	chainMetrics *metrics.ChainMetrics,
	shutdownCoordinator *shutdown.Coordinator,
	onChainConnect func(),
	onChainDisconnect func(),
	deriveAliasOutputByQuorum bool,
	pipeliningLimit int,
	consensusDelay time.Duration,
	recoveryTimeout time.Duration,
	validatorAgentID isc.AgentID,
	smParameters sm_gpa.StateManagerParameters,
) (Chain, error) {
	log.Debugf("Starting the chain, chainID=%v", chainID)
	if listener == nil {
		listener = NewEmptyChainListener()
	}
	if accessNodesFromNode == nil {
		accessNodesFromNode = []*cryptolib.PublicKey{}
	}
	netPeeringID := peering.HashPeeringIDFromBytes(chainID.Bytes(), []byte("ChainManager")) // ChainID × ChainManager
	cni := &chainNodeImpl{
		nodeIdentity:           nodeIdentity,
		chainID:                chainID,
		chainStore:             chainStore,
		nodeConn:               nodeConn,
		tangleTime:             time.Time{}, // Zero time, while we haven't received it from the L1.
		recvAliasOutputPipe:    pipe.NewInfinitePipe[*isc.AliasOutputWithID](),
		recvTxPublishedPipe:    pipe.NewInfinitePipe[*txPublished](),
		recvMilestonePipe:      pipe.NewInfinitePipe[time.Time](),
		consensusInsts:         shrinkingmap.New[iotago.Ed25519Address, *shrinkingmap.ShrinkingMap[cmt_log.LogIndex, *consensusInst]](),
		consOutputPipe:         pipe.NewInfinitePipe[*consOutput](),
		consRecoverPipe:        pipe.NewInfinitePipe[*consRecover](),
		publishingTXes:         shrinkingmap.New[iotago.TransactionID, context.CancelFunc](),
		procCache:              processors.MustNew(processorConfig),
		configUpdatedCh:        make(chan *configUpdate, 1),
		serversUpdatedPipe:     pipe.NewInfinitePipe[*serversUpdate](),
		awaitReceiptActCh:      make(chan *awaitReceiptReq, 1),
		awaitReceiptCnfCh:      make(chan *awaitReceiptReq, 1),
		stateTrackerAct:        nil, // Set bellow.
		stateTrackerCnf:        nil, // Set bellow.
		blockWAL:               blockWAL,
		consensusDelay:         consensusDelay,
		recoveryTimeout:        recoveryTimeout,
		validatorAgentID:       validatorAgentID,
		listener:               listener,
		accessLock:             &sync.RWMutex{},
		activeCommitteeDKShare: nil,
		activeCommitteeNodes:   []*cryptolib.PublicKey{},
		activeAccessNodes:      nil, // Set bellow.
		accessNodesFromNode:    nil, // Set bellow.
		accessNodesFromACT:     nil, // Set bellow.
		accessNodesFromCNF:     nil, // Set bellow.
		serverNodes:            nil, // Set bellow.
		latestConfirmedAO:      nil,
		latestConfirmedState:   nil,
		latestConfirmedStateAO: nil,
		latestActiveAO:         nil,
		latestActiveState:      nil,
		latestActiveStateAO:    nil,
		netRecvPipe:            pipe.NewInfinitePipe[*peering.PeerMessageIn](),
		netPeeringID:           netPeeringID,
		netPeerPubs:            map[gpa.NodeID]*cryptolib.PublicKey{},
		net:                    net,
		shutdownCoordinator:    shutdownCoordinator,
		chainMetrics:           chainMetrics,
		log:                    log,
	}

	cni.chainMetrics.Pipe.TrackPipeLen("node-recvAliasOutputPipe", cni.recvAliasOutputPipe.Len)
	cni.chainMetrics.Pipe.TrackPipeLen("node-recvTxPublishedPipe", cni.recvTxPublishedPipe.Len)
	cni.chainMetrics.Pipe.TrackPipeLen("node-recvMilestonePipe", cni.recvMilestonePipe.Len)
	cni.chainMetrics.Pipe.TrackPipeLen("node-consOutputPipe", cni.consOutputPipe.Len)
	cni.chainMetrics.Pipe.TrackPipeLen("node-consRecoverPipe", cni.consRecoverPipe.Len)
	cni.chainMetrics.Pipe.TrackPipeLen("node-serversUpdatedPipe", cni.serversUpdatedPipe.Len)
	cni.chainMetrics.Pipe.TrackPipeLen("node-netRecvPipe", cni.netRecvPipe.Len)

	if recoverFromWAL {
		cni.recoverStoreFromWAL(chainStore, blockWAL)
	}
	cni.me = cni.pubKeyAsNodeID(nodeIdentity.GetPublicKey())
	//
	// Create sub-components.
	chainMgr, err := chainmanager.New(
		cni.me,
		cni.chainID,
		cni.chainStore,
		consensusStateRegistry,
		dkShareRegistryProvider,
		cni.pubKeyAsNodeID,
		func() ([]*cryptolib.PublicKey, []*cryptolib.PublicKey) {
			cni.accessLock.RLock()
			defer cni.accessLock.RUnlock()
			return cni.activeAccessNodes, cni.activeCommitteeNodes
		},
		func(ao *isc.AliasOutputWithID) {
			cni.stateTrackerAct.TrackAliasOutput(ao, true)
		},
		func(block state.Block) {
			if err := cni.stateMgr.PreliminaryBlock(block); err != nil {
				cni.log.Warnf("Failed to save a preliminary block %v: %v", block.L1Commitment(), err)
			}
		},
		func(dkShare tcrypto.DKShare) {
			cni.accessLock.Lock()
			cni.activeCommitteeDKShare = dkShare
			activeCommitteeNodes := cni.activeCommitteeNodes
			cni.accessLock.Unlock()
			var newCommitteeNodes []*cryptolib.PublicKey
			if dkShare == nil {
				newCommitteeNodes = []*cryptolib.PublicKey{}
			} else {
				newCommitteeNodes = dkShare.GetNodePubKeys()
			}
			if !util.Same(newCommitteeNodes, activeCommitteeNodes) {
				cni.log.Infof("Committee nodes updated to %v, was %v", newCommitteeNodes, activeCommitteeNodes)
				cni.updateAccessNodes(func() {
					cni.activeCommitteeNodes = newCommitteeNodes
				})
			}
		},
		deriveAliasOutputByQuorum,
		pipeliningLimit,
		cni.chainMetrics.CmtLog,
		cni.log.Named("CM"),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create chainMgr: %w", err)
	}
	// TODO does it make sense to pass itself (own pub key) here?
	peerPubKeys := []*cryptolib.PublicKey{nodeIdentity.GetPublicKey()}
	peerPubKeys = append(peerPubKeys, cni.accessNodesFromNode...)
	stateMgr, err := statemanager.New(
		ctx,
		cni.chainID,
		nodeIdentity.GetPublicKey(),
		peerPubKeys,
		net,
		blockWAL,
		snapshotManager,
		chainStore,
		shutdownCoordinator.Nested("StateMgr"),
		chainMetrics.StateManager,
		chainMetrics.Pipe,
		cni.log,
		smParameters,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create stateMgr: %w", err)
	}
	mempool := mempool.New(
		ctx,
		chainID,
		nodeIdentity,
		net,
		cni.log.Named("MP"),
		chainMetrics.Mempool,
		chainMetrics.Pipe,
		cni.listener,
	)
	cni.chainMgr = gpa.NewAckHandler(cni.me, chainMgr.AsGPA(), RedeliveryPeriod)
	cni.stateMgr = stateMgr
	cni.mempool = mempool
	cni.stateTrackerAct = NewStateTracker(ctx, stateMgr, cni.handleStateTrackerActCB, chainMetrics.StateManager.SetChainActiveStateWant, chainMetrics.StateManager.SetChainActiveStateHave, cni.log.Named("ST.ACT"))
	cni.stateTrackerCnf = NewStateTracker(ctx, stateMgr, cni.handleStateTrackerCnfCB, chainMetrics.StateManager.SetChainConfirmedStateWant, chainMetrics.StateManager.SetChainConfirmedStateHave, cni.log.Named("ST.CNF"))
	cni.updateAccessNodes(func() {
		cni.accessNodesFromNode = accessNodesFromNode
		cni.accessNodesFromACT = []*cryptolib.PublicKey{}
		cni.accessNodesFromCNF = []*cryptolib.PublicKey{}
	})
	cni.updateServerNodes([]*cryptolib.PublicKey{})
	//
	// Connect to the peering network.
	netRecvPipeInCh := cni.netRecvPipe.In()
	unhook := net.Attach(&netPeeringID, peering.ReceiverChain, func(recv *peering.PeerMessageIn) {
		if recv.MsgType != msgTypeChainMgr {
			cni.log.Warnf("Unexpected message, type=%v", recv.MsgType)
			return
		}
		netRecvPipeInCh <- recv
	})
	//
	// Attach to the L1.
	recvRequestCB := func(outputInfo *isc.OutputInfo) {
		log.Debugf("recvRequestCB[%p], consumed=%v, outputID=%v", cni, outputInfo.Consumed(), outputInfo.OutputID.ToHex())
		cni.chainMetrics.NodeConn.L1RequestReceived()
		req, err := isc.OnLedgerFromUTXO(outputInfo.Output, outputInfo.OutputID)
		if err != nil {
			cni.log.Warnf("Cannot create OnLedgerRequest from output: %v", err)
			return
		}
		if req.IsInternalUTXO(cni.chainID) {
			cni.log.Debugf("Ignoring internal UTXO with ID=%v, will not consider it a request: %v", outputInfo.OutputID.ToHex(), req.String())
			return
		}
		cni.mempool.ReceiveOnLedgerRequest(req)
	}
	recvAliasOutputPipeInCh := cni.recvAliasOutputPipe.In()
	recvAliasOutputCB := func(outputInfo *isc.OutputInfo) {
		log.Debugf("recvAliasOutputCB[%p], %v", cni, outputInfo.OutputID.ToHex())
		cni.chainMetrics.NodeConn.L1AliasOutputReceived()
		if outputInfo.Consumed() {
			// we don't need to send consumed alias outputs to the pipe
			return
		}
		recvAliasOutputPipeInCh <- outputInfo.AliasOutputWithID()
	}
	recvMilestonePipeInCh := cni.recvMilestonePipe.In()
	recvMilestoneCB := func(timestamp time.Time) {
		log.Debugf("recvMilestoneCB[%p], %v", cni, timestamp)
		recvMilestonePipeInCh <- timestamp
	}
	nodeConn.AttachChain(ctx, chainID, recvRequestCB, recvAliasOutputCB, recvMilestoneCB, onChainConnect, onChainDisconnect)
	//
	// Run the main thread.

	go cni.run(ctx, unhook)
	return cni, nil
}

func (cni *chainNodeImpl) ReceiveOffLedgerRequest(request isc.OffLedgerRequest, sender *cryptolib.PublicKey) error {
	cni.log.Debugf("ReceiveOffLedgerRequest: %v from outside.", request.ID())
	// TODO: What to do with the sender's pub key?
	return cni.mempool.ReceiveOffLedgerRequest(request)
}

func (cni *chainNodeImpl) AwaitRequestProcessed(ctx context.Context, requestID isc.RequestID, confirmed bool) <-chan *blocklog.RequestReceipt {
	query, responseCh := newAwaitReceiptReq(ctx, requestID, cni.log)
	if confirmed {
		cni.awaitReceiptCnfCh <- query
	} else {
		cni.awaitReceiptActCh <- query
	}
	return responseCh
}

func (cni *chainNodeImpl) ConfigUpdated(accessNodesPerNode []*cryptolib.PublicKey) {
	cni.configUpdatedCh <- &configUpdate{accessNodes: accessNodesPerNode}
}

func (cni *chainNodeImpl) ServersUpdated(serverNodes []*cryptolib.PublicKey) {
	cni.serversUpdatedPipe.In() <- &serversUpdate{serverNodes: serverNodes}
}

//nolint:gocyclo
func (cni *chainNodeImpl) run(ctx context.Context, cleanupFunc context.CancelFunc) {
	defer util.ExecuteIfNotNil(cleanupFunc)

	recvAliasOutputPipeOutCh := cni.recvAliasOutputPipe.Out()
	recvTxPublishedPipeOutCh := cni.recvTxPublishedPipe.Out()
	recvMilestonePipeOutCh := cni.recvMilestonePipe.Out()
	netRecvPipeOutCh := cni.netRecvPipe.Out()
	consOutputPipeOutCh := cni.consOutputPipe.Out()
	consRecoverPipeOutCh := cni.consRecoverPipe.Out()
	serversUpdatedPipeOutCh := cni.serversUpdatedPipe.Out()
	redeliveryPeriodTicker := time.NewTicker(RedeliveryPeriod)
	consensusDelayTicker := time.NewTicker(cni.consensusDelay)
	for {
		if ctx.Err() != nil {
			if cni.shutdownCoordinator == nil {
				return
			}
			// needs to wait for state mgr and consensusInst
			cni.shutdownCoordinator.WaitNestedWithLogging(1 * time.Second)
			cni.shutdownCoordinator.Done()
			return
		}
		select {
		case txPublishResult, ok := <-recvTxPublishedPipeOutCh:
			if !ok {
				recvTxPublishedPipeOutCh = nil
				continue
			}
			cni.handleTxPublished(ctx, txPublishResult)
		case aliasOutput, ok := <-recvAliasOutputPipeOutCh:
			if !ok {
				recvAliasOutputPipeOutCh = nil
				continue
			}
			cni.handleAliasOutput(ctx, aliasOutput)
		case timestamp, ok := <-recvMilestonePipeOutCh:
			if !ok {
				recvMilestonePipeOutCh = nil
			}
			cni.handleMilestoneTimestamp(timestamp)
		case recv, ok := <-netRecvPipeOutCh:
			if !ok {
				netRecvPipeOutCh = nil
				continue
			}
			cni.handleNetMessage(ctx, recv)
		case recv, ok := <-consOutputPipeOutCh:
			if !ok {
				consOutputPipeOutCh = nil
				continue
			}
			cni.handleConsensusOutput(ctx, recv)
		case recv, ok := <-consRecoverPipeOutCh:
			if !ok {
				consRecoverPipeOutCh = nil
				continue
			}
			cni.handleConsensusRecover(ctx, recv)
		case cfg, ok := <-cni.configUpdatedCh:
			if !ok {
				cni.configUpdatedCh = nil
				continue
			}
			cni.handleAccessNodesConfigUpdated(cfg.accessNodes)
		case srv, ok := <-serversUpdatedPipeOutCh:
			if !ok {
				serversUpdatedPipeOutCh = nil
				continue
			}
			cni.handleServersUpdated(srv.serverNodes)
		case query, ok := <-cni.awaitReceiptActCh:
			if !ok {
				cni.awaitReceiptActCh = nil
				continue
			}
			cni.stateTrackerAct.AwaitRequestReceipt(query) // TODO: Actually have to wait for both: the ACT and CNF, because the node can become access node while waiting for the request.
		case query, ok := <-cni.awaitReceiptCnfCh:
			if !ok {
				cni.awaitReceiptCnfCh = nil
				continue
			}
			cni.stateTrackerCnf.AwaitRequestReceipt(query)
		case resp, ok := <-cni.stateTrackerAct.ChainNodeAwaitStateMgrCh():
			if ok {
				cni.stateTrackerAct.ChainNodeStateMgrResponse(resp)
			}
		case resp, ok := <-cni.stateTrackerCnf.ChainNodeAwaitStateMgrCh():
			if ok {
				cni.stateTrackerCnf.ChainNodeStateMgrResponse(resp)
			}
		case <-consensusDelayTicker.C:
			cni.sendMessages(cni.chainMgr.Input(chainmanager.NewInputCanPropose()))
			cni.handleChainMgrOutput(ctx, cni.chainMgr.Output())
		case t := <-redeliveryPeriodTicker.C:
			cni.sendMessages(cni.chainMgr.Input(cni.chainMgr.MakeTickInput(t)))
			cni.handleChainMgrOutput(ctx, cni.chainMgr.Output())
		case <-ctx.Done():
			continue
		}
	}
}

// The active state is needed by the mempool to cleanup the processed requests, etc.
// The request/receipt awaits are already handled in the StateTracker.
func (cni *chainNodeImpl) handleStateTrackerActCB(st state.State, from, till *isc.AliasOutputWithID, added, removed []state.Block) {
	cni.log.Debugf("handleStateTrackerActCB: till %v from %v", till, from)
	cni.accessLock.Lock()
	cni.latestActiveState = st
	cni.latestActiveStateAO = till
	latestConfirmedAO := cni.latestConfirmedAO
	cni.accessLock.Unlock()

	// Set the state to match the ActiveOrConfirmed state.
	if latestConfirmedAO == nil || till.GetStateIndex() > latestConfirmedAO.GetStateIndex() {
		l1Commitment := transaction.MustL1CommitmentFromAliasOutput(till.GetAliasOutput())
		if err := cni.chainStore.SetLatest(l1Commitment.TrieRoot()); err != nil {
			panic(fmt.Errorf("cannot set L1Commitment=%v as latest: %w", l1Commitment, err))
		}
		cni.log.Debugf("Latest state set to ACT index=%v, trieRoot=%v", till.GetStateIndex(), l1Commitment.TrieRoot())
	}

	newAccessNodes := governance.NewStateAccess(st).AccessNodes()
	if !util.Same(newAccessNodes, cni.accessNodesFromACT) {
		cni.updateAccessNodes(func() {
			cni.accessNodesFromACT = newAccessNodes
		})
	}

	cni.mempool.TrackNewChainHead(st, from, till, added, removed)
}

// The committed state is required here because:
//   - This way we make sure the state has all the blocks till the specified trie root.
//   - We set it as latest here. This is then used in many places to access the latest version of the state.
//
// The request/receipt awaits are already handled in the StateTracker.
func (cni *chainNodeImpl) handleStateTrackerCnfCB(st state.State, from, till *isc.AliasOutputWithID, added, removed []state.Block) {
	cni.log.Debugf("handleStateTrackerCnfCB: till %v from %v", till, from)
	cni.accessLock.Lock()
	cni.latestConfirmedState = st
	cni.latestConfirmedStateAO = till
	latestActiveStateAO := cni.latestActiveStateAO
	cni.accessLock.Unlock()

	newAccessNodes := governance.NewStateAccess(st).AccessNodes()
	if !util.Same(newAccessNodes, cni.accessNodesFromCNF) {
		cni.updateAccessNodes(func() {
			cni.accessNodesFromCNF = newAccessNodes
		})
	}

	// Set the state to match the ActiveOrConfirmed state.
	if latestActiveStateAO == nil || latestActiveStateAO.GetStateIndex() <= till.GetStateIndex() {
		l1Commitment := transaction.MustL1CommitmentFromAliasOutput(till.GetAliasOutput())
		if err := cni.chainStore.SetLatest(l1Commitment.TrieRoot()); err != nil {
			panic(fmt.Errorf("cannot set L1Commitment=%v as latest: %w", l1Commitment, err))
		}
		cni.log.Debugf("Latest state set to CNF index=%v, trieRoot=%v", till.GetStateIndex(), l1Commitment.TrieRoot())
	}
}

func (cni *chainNodeImpl) handleAccessNodesConfigUpdated(accessNodesFromNode []*cryptolib.PublicKey) {
	cni.log.Debugf("handleAccessNodesConfigUpdated")
	cni.updateAccessNodes(func() {
		cni.accessNodesFromNode = accessNodesFromNode
	})
}

func (cni *chainNodeImpl) handleServersUpdated(serverNodes []*cryptolib.PublicKey) {
	cni.log.Debugf("handleServersUpdated")
	cni.updateServerNodes(serverNodes)
}

func (cni *chainNodeImpl) handleTxPublished(ctx context.Context, txPubResult *txPublished) {
	cni.log.Debugf("handleTxPublished")
	if !cni.publishingTXes.Has(txPubResult.txID) {
		return
	}
	cni.publishingTXes.Delete(txPubResult.txID)

	outMsgs := cni.chainMgr.Input(
		chainmanager.NewInputChainTxPublishResult(txPubResult.committeeAddr, txPubResult.logIndex, txPubResult.txID, txPubResult.nextAliasOutput, txPubResult.confirmed),
	)
	cni.sendMessages(outMsgs)
	cni.handleChainMgrOutput(ctx, cni.chainMgr.Output())
}

func (cni *chainNodeImpl) handleAliasOutput(ctx context.Context, aliasOutput *isc.AliasOutputWithID) {
	cni.log.Debugf("handleAliasOutput: %v", aliasOutput)
	if aliasOutput.GetStateIndex() == 0 {
		initBlock, err := origin.InitChainByAliasOutput(cni.chainStore, aliasOutput)
		if err != nil {
			cni.log.Errorf("Ignoring InitialAO for the chain: %v", err)
			return
		}
		if err := cni.blockWAL.Write(initBlock); err != nil {
			panic(fmt.Errorf("cannot write initial block to the WAL: %w", err))
		}
	}

	cni.stateTrackerCnf.TrackAliasOutput(aliasOutput, true)
	cni.stateTrackerAct.TrackAliasOutput(aliasOutput, false) // ACT state will be equal to CNF or ahead of it.
	outMsgs := cni.chainMgr.Input(
		chainmanager.NewInputAliasOutputConfirmed(aliasOutput),
	)
	cni.sendMessages(outMsgs)
	cni.handleChainMgrOutput(ctx, cni.chainMgr.Output())
}

func (cni *chainNodeImpl) handleMilestoneTimestamp(timestamp time.Time) {
	cni.log.Debugf("handleMilestoneTimestamp: %v", timestamp)
	cni.tangleTime = timestamp
	cni.mempool.TangleTimeUpdated(timestamp)
	cni.consensusInsts.ForEach(func(address iotago.Ed25519Address, consensusInstances *shrinkingmap.ShrinkingMap[cmt_log.LogIndex, *consensusInst]) bool {
		consensusInstances.ForEach(func(li cmt_log.LogIndex, consensusInstance *consensusInst) bool {
			if consensusInstance.cancelFunc != nil {
				consensusInstance.consensus.Time(timestamp)
			}
			return true
		})
		return true
	})
}

func (cni *chainNodeImpl) handleNetMessage(ctx context.Context, recv *peering.PeerMessageIn) {
	msg, err := cni.chainMgr.UnmarshalMessage(recv.MsgData)
	if err != nil {
		cni.log.Warnf("cannot parse message: %v", err)
		return
	}
	msg.SetSender(cni.pubKeyAsNodeID(recv.SenderPubKey))
	cni.sendMessages(cni.chainMgr.Message(msg))
	cni.handleChainMgrOutput(ctx, cni.chainMgr.Output())
}

func (cni *chainNodeImpl) handleChainMgrOutput(ctx context.Context, outputUntyped gpa.Output) {
	cni.log.Debugf("handleChainMgrOutput: %v", outputUntyped)
	if outputUntyped == nil { // TODO: Will never be nil, fix it.
		// TODO: Cleanup consensus instances for all the committees after some time.
		// Not sure, if it is OK to terminate them immediately at this point.
		// This is for the case, if the current node is not in a committee of a chain anymore.
		cni.cleanupPublishingTXes(nil)
		return
	}
	output := outputUntyped.(*chainmanager.Output)
	//
	// Start new consensus instances, if needed.
	outputNeedConsensus := output.NeedConsensus()
	if outputNeedConsensus != nil {
		cni.ensureConsensusInput(ctx, outputNeedConsensus)
	}
	//
	// Start publishing TX'es, if there not being posted already.
	outputNeedPostTXes := output.NeedPublishTX()
	outputNeedPostTXes.ForEach(func(ti iotago.TransactionID, needPublishTx *chainmanager.NeedPublishTX) bool {
		txToPost := needPublishTx // Have to take a copy to be used in callback.
		if !cni.publishingTXes.Has(txToPost.TxID) {
			subCtx, subCancel := context.WithCancel(ctx)
			cni.publishingTXes.Set(txToPost.TxID, subCancel)
			publishStart := time.Now()
			if err := cni.nodeConn.PublishTX(subCtx, cni.chainID, txToPost.Tx, func(_ *iotago.Transaction, confirmed bool) {
				cni.chainMetrics.NodeConn.TXPublishResult(confirmed, time.Since(publishStart))
				cni.recvTxPublishedPipe.In() <- &txPublished{
					committeeAddr:   txToPost.CommitteeAddr,
					logIndex:        txToPost.LogIndex,
					txID:            txToPost.TxID,
					nextAliasOutput: txToPost.NextAliasOutput,
					confirmed:       confirmed,
				}
			}); err != nil {
				cni.log.Error(err.Error())
			}
			cni.chainMetrics.NodeConn.TXPublishStarted()
		}

		return true
	})

	cni.cleanupPublishingTXes(outputNeedPostTXes)
	//
	// Update info for access by other components.
	cni.accessLock.Lock()
	cni.latestConfirmedAO = output.LatestConfirmedAliasOutput()
	cni.latestActiveAO = output.LatestActiveAliasOutput()
	// if cni.latestActiveAO == nil {	// TODO: Check, how is this handled in the case of rejections.
	// 	cni.latestActiveState = nil
	// 	cni.latestActiveStateAO = nil
	// }
	cni.accessLock.Unlock()
}

func (cni *chainNodeImpl) handleConsensusOutput(ctx context.Context, out *consOutput) {
	cni.log.Debugf("handleConsensusOutput, %v", out)
	var chainMgrInput gpa.Input
	switch out.output.Status {
	case cons.Completed:
		chainMgrInput = chainmanager.NewInputConsensusOutputDone(
			out.request.CommitteeAddr,
			out.request.LogIndex,
			out.request.BaseAliasOutput.OutputID(),
			out.output.Result,
		)
	case cons.Skipped:
		chainMgrInput = chainmanager.NewInputConsensusOutputSkip(
			out.request.CommitteeAddr,
			out.request.LogIndex,
			out.request.BaseAliasOutput.OutputID(),
		)
	default:
		panic(fmt.Errorf("unexpected output state from consensus: %+v", out))
	}
	// We can cleanup the instances that are BEFORE the instance that produced
	// an output, because all the nodes will eventually get the NextLI messages,
	// and will switch to newer instances.
	cni.cleanupConsensusInsts(out.request.CommitteeAddr, out.request.LogIndex)
	cni.sendMessages(cni.chainMgr.Input(chainMgrInput))
	cni.handleChainMgrOutput(ctx, cni.chainMgr.Output())
}

func (cni *chainNodeImpl) handleConsensusRecover(ctx context.Context, out *consRecover) {
	cni.log.Debugf("handleConsensusRecover: %v", out)
	chainMgrInput := chainmanager.NewInputConsensusTimeout(
		out.request.CommitteeAddr,
		out.request.LogIndex,
	)
	cni.sendMessages(cni.chainMgr.Input(chainMgrInput))
	cni.handleChainMgrOutput(ctx, cni.chainMgr.Output())
}

func (cni *chainNodeImpl) ensureConsensusInput(ctx context.Context, needConsensus *chainmanager.NeedConsensus) {
	ci := cni.ensureConsensusInst(ctx, needConsensus)
	if ci.request == nil {
		outputCB := func(o *consGR.Output) {
			cni.consOutputPipe.In() <- &consOutput{request: needConsensus, output: o}
		}
		recoverCB := func() {
			cni.consRecoverPipe.In() <- &consRecover{request: needConsensus}
		}
		ci.request = needConsensus
		cni.stateTrackerAct.TrackAliasOutput(needConsensus.BaseAliasOutput, true)
		ci.consensus.Input(needConsensus.BaseAliasOutput, outputCB, recoverCB)
	}
}

func (cni *chainNodeImpl) ensureConsensusInst(ctx context.Context, needConsensus *chainmanager.NeedConsensus) *consensusInst {
	committeeAddr := needConsensus.CommitteeAddr
	logIndex := needConsensus.LogIndex
	dkShare := needConsensus.DKShare

	consensusInstances, _ := cni.consensusInsts.GetOrCreate(committeeAddr, func() *shrinkingmap.ShrinkingMap[cmt_log.LogIndex, *consensusInst] {
		return shrinkingmap.New[cmt_log.LogIndex, *consensusInst]()
	})

	addLogIndex := logIndex
	for i := 0; i < ConsensusInstsInAdvance; i++ {
		if !consensusInstances.Has(addLogIndex) {
			consGrCtx, consGrCancel := context.WithCancel(ctx)
			logIndexCopy := addLogIndex
			cgr := consGR.New(
				consGrCtx, cni.chainID, cni.chainStore, dkShare, &logIndexCopy, cni.nodeIdentity,
				cni.procCache, cni.mempool, cni.stateMgr, cni.net,
				cni.validatorAgentID,
				cni.recoveryTimeout, RedeliveryPeriod, PrintStatusPeriod,
				cni.chainMetrics.Consensus,
				cni.chainMetrics.Pipe,
				cni.log.Named(fmt.Sprintf("C-%v.LI-%v", committeeAddr.String()[:10], logIndexCopy)),
			)
			consensusInstances.Set(addLogIndex, &consensusInst{
				cancelFunc: consGrCancel,
				consensus:  cgr,
				committee:  dkShare.GetNodePubKeys(),
			})
			if !cni.tangleTime.IsZero() {
				cgr.Time(cni.tangleTime)
			}
		}
		addLogIndex = addLogIndex.Next()
	}

	consensusInstance, _ := consensusInstances.Get(logIndex)
	return consensusInstance
}

// Cleanup consensus instances, except the instances with LogIndexes above the specified for a particular committee.
// If nils are provided for the keep* variables, all the instances are cleaned up.
func (cni *chainNodeImpl) cleanupConsensusInsts(committeeAddr iotago.Ed25519Address, keepLogIndex cmt_log.LogIndex) {
	consensusInstances, exists := cni.consensusInsts.Get(committeeAddr)
	if !exists {
		return
	}

	consensusInstances.ForEach(func(li cmt_log.LogIndex, consensusInstance *consensusInst) bool {
		if li >= keepLogIndex {
			return true
		}
		cni.log.Debugf("Canceling consensus instance for Committee=%v, LogIndex=%v", committeeAddr.String(), li)

		consensusInstance.Cancel()
		consensusInstances.Delete(li)
		return true
	})

	if consensusInstances.Size() == 0 {
		cni.consensusInsts.Delete(committeeAddr)
	}
}

// Cleanup TX'es that are not needed to be posted anymore.
func (cni *chainNodeImpl) cleanupPublishingTXes(neededPostTXes *shrinkingmap.ShrinkingMap[iotago.TransactionID, *chainmanager.NeedPublishTX]) {
	if neededPostTXes == nil || neededPostTXes.Size() == 0 {
		// just create a new map
		cni.publishingTXes = shrinkingmap.New[iotago.TransactionID, context.CancelFunc]()
		return
	}

	cni.publishingTXes.ForEach(func(txID iotago.TransactionID, cancelFunc context.CancelFunc) bool {
		if !neededPostTXes.Has(txID) { // remove anything that doesn't need a tx to be posted
			cancelFunc()
			cni.publishingTXes.Delete(txID)
		}
		return true
	})
}

func (cni *chainNodeImpl) sendMessages(outMsgs gpa.OutMessages) {
	if outMsgs == nil {
		return
	}
	outMsgs.MustIterate(func(msg gpa.Message) {
		recipientPubKey, ok := cni.netPeerPubs[msg.Recipient()]
		if !ok {
			cni.log.Warnf("Pub key for the recipient not found: %v", msg.Recipient())
			return
		}
		pm := peering.NewPeerMessageData(cni.netPeeringID, peering.ReceiverChain, msgTypeChainMgr, msg)
		cni.net.SendMsgByPubKey(recipientPubKey, pm)
	})
}

// activeAccessNodes = ∪{{Self}, accessNodesFromNode, accessNodesFromACT, accessNodesFromCNF, activeCommitteeNodes}
func (cni *chainNodeImpl) deriveActiveAccessNodes() {
	nodes := []*cryptolib.PublicKey{cni.nodeIdentity.GetPublicKey()}
	index := map[cryptolib.PublicKeyKey]bool{nodes[0].AsKey(): true}
	for _, k := range cni.accessNodesFromNode {
		if _, ok := index[k.AsKey()]; !ok {
			index[k.AsKey()] = true
			nodes = append(nodes, k)
		}
	}
	for _, k := range cni.accessNodesFromACT {
		if _, ok := index[k.AsKey()]; !ok {
			index[k.AsKey()] = true
			nodes = append(nodes, k)
		}
	}
	for _, k := range cni.accessNodesFromCNF {
		if _, ok := index[k.AsKey()]; !ok {
			index[k.AsKey()] = true
			nodes = append(nodes, k)
		}
	}
	for _, k := range cni.activeCommitteeNodes {
		if _, ok := index[k.AsKey()]; !ok {
			index[k.AsKey()] = true
			nodes = append(nodes, k)
		}
	}
	cni.activeAccessNodes = nodes
}

func (cni *chainNodeImpl) updateAccessNodes(update func()) {
	cni.accessLock.Lock()
	oldAccessNodes := cni.activeAccessNodes
	oldCommitteeNodes := cni.activeCommitteeNodes
	update()
	cni.deriveActiveAccessNodes()
	serverNodes := cni.serverNodes
	activeAccessNodes := cni.activeAccessNodes
	activeCommitteeNodes := cni.activeCommitteeNodes
	cni.accessLock.Unlock()
	anSame := util.Same(oldAccessNodes, activeAccessNodes)
	cnSame := util.Same(oldCommitteeNodes, activeCommitteeNodes)
	if !anSame {
		cni.log.Infof("Access nodes updated, active=%+v", activeAccessNodes)
		cni.listener.AccessNodesUpdated(cni.chainID, activeAccessNodes)
	}
	if !anSame || !cnSame {
		if !cnSame {
			cni.log.Infof("Committee nodes updated, active=%+v", activeCommitteeNodes)
		}
		cni.mempool.AccessNodesUpdated(activeCommitteeNodes, activeAccessNodes)
		cni.stateMgr.ChainNodesUpdated(serverNodes, activeAccessNodes, activeCommitteeNodes)
	}
}

func (cni *chainNodeImpl) updateServerNodes(serverNodes []*cryptolib.PublicKey) {
	cni.accessLock.Lock()
	oldServerNodes := cni.serverNodes
	cni.serverNodes = serverNodes
	activeAccessNodes := cni.activeAccessNodes
	activeCommitteeNodes := cni.activeCommitteeNodes
	cni.accessLock.Unlock()
	if oldServerNodes == nil || !util.Same(oldServerNodes, serverNodes) {
		cni.log.Infof("Server nodes updated, servers=%+v", serverNodes)
		cni.mempool.ServerNodesUpdated(activeCommitteeNodes, serverNodes)
		cni.stateMgr.ChainNodesUpdated(serverNodes, activeAccessNodes, activeCommitteeNodes)
		cni.listener.ServerNodesUpdated(cni.chainID, serverNodes)
	}
}

func (cni *chainNodeImpl) pubKeyAsNodeID(pubKey *cryptolib.PublicKey) gpa.NodeID {
	nodeID := gpa.NodeIDFromPublicKey(pubKey)
	if _, ok := cni.netPeerPubs[nodeID]; !ok {
		cni.netPeerPubs[nodeID] = pubKey
	}
	return nodeID
}

////////////////////////////////////////////////////////////////////////////////
// Support functions.

func (cni *chainNodeImpl) ID() isc.ChainID {
	return cni.chainID
}

func (cni *chainNodeImpl) Store() indexedstore.IndexedStore {
	return cni.chainStore
}

func (cni *chainNodeImpl) Processors() *processors.Cache {
	return cni.procCache
}

func (cni *chainNodeImpl) Log() *logger.Logger {
	return cni.log
}

func (cni *chainNodeImpl) LatestAliasOutput(freshness StateFreshness) (*isc.AliasOutputWithID, error) {
	cni.accessLock.RLock()
	latestActiveAO := cni.latestActiveStateAO
	latestConfirmedAO := cni.latestConfirmedStateAO
	cni.accessLock.RUnlock()
	switch freshness {
	case ActiveOrCommittedState:
		if latestActiveAO != nil {
			if latestConfirmedAO == nil || latestActiveAO.GetStateIndex() > latestConfirmedAO.GetStateIndex() {
				cni.log.Debugf("LatestAliasOutput(%v) => active = %v", freshness, latestActiveAO)
				return latestActiveAO, nil
			}
		}
		if latestConfirmedAO != nil {
			cni.log.Debugf("LatestAliasOutput(%v) => confirmed = %v", freshness, latestConfirmedAO)
			return latestConfirmedAO, nil
		}
		return nil, fmt.Errorf("have no active nor confirmed state")
	case ConfirmedState:
		if latestConfirmedAO != nil {
			cni.log.Debugf("LatestAliasOutput(%v) => confirmed = %v", freshness, latestConfirmedAO)
			return latestConfirmedAO, nil
		}
		return nil, fmt.Errorf("have no confirmed state")
	case ActiveState:
		if latestActiveAO != nil {
			cni.log.Debugf("LatestAliasOutput(%v) => active = %v", freshness, latestActiveAO)
			return latestActiveAO, nil
		}
		return nil, fmt.Errorf("have no active state")
	default:
		panic(fmt.Errorf("unexpected StateFreshness: %v", freshness))
	}
}

func (cni *chainNodeImpl) LatestState(freshness StateFreshness) (state.State, error) {
	cni.accessLock.RLock()
	latestActiveState := cni.latestActiveState
	latestConfirmedState := cni.latestConfirmedState
	cni.accessLock.RUnlock()
	switch freshness {
	case ActiveOrCommittedState:
		if latestActiveState != nil {
			if latestConfirmedState == nil || latestActiveState.BlockIndex() > latestConfirmedState.BlockIndex() {
				cni.log.Debugf("LatestState(%v) => active = %v", freshness, latestActiveState)
				return latestActiveState, nil
			}
		}
		if latestConfirmedState != nil {
			cni.log.Debugf("LatestState(%v) => confirmed = %v", freshness, latestConfirmedState)
			return latestConfirmedState, nil
		}
		latestInStore, err := cni.chainStore.LatestState()
		cni.log.Debugf("LatestState(%v) => inStore = %v, %v", freshness, latestInStore, err)
		return latestInStore, err
	case ConfirmedState:
		if latestConfirmedState != nil {
			cni.log.Debugf("LatestState(%v) => confirmed = %v", freshness, latestConfirmedState)
			return latestConfirmedState, nil
		}
		latestInStore, err := cni.chainStore.LatestState()
		cni.log.Debugf("LatestState(%v) => inStore = %v, %v", freshness, latestInStore, err)
		return latestInStore, err
	case ActiveState:
		if latestActiveState != nil {
			cni.log.Debugf("LatestState(%v) => active = %v", freshness, latestActiveState)
			return latestActiveState, nil
		}
		return nil, fmt.Errorf("chain %v has no active state", cni.chainID)
	default:
		panic(fmt.Errorf("unexpected StateFreshness: %v", freshness))
	}
}

func (cni *chainNodeImpl) GetCommitteeInfo() *CommitteeInfo {
	cni.accessLock.RLock()
	dkShare := cni.activeCommitteeDKShare
	cni.accessLock.RUnlock()
	if dkShare == nil {
		return nil // There is no current committee for now.
	}
	committeePubKeys := dkShare.GetNodePubKeys()
	netPeerStatus := cni.net.PeerStatus()
	peerStatus := make([]*PeerStatus, len(committeePubKeys))
	connectedCount := uint16(0)
	for i, nodePubKey := range committeePubKeys {
		index := slices.IndexFunc(netPeerStatus, func(st peering.PeerStatusProvider) bool {
			return st.PubKey().Equals(nodePubKey)
		})
		if index == -1 {
			peerStatus[i] = &PeerStatus{
				Index:      uint16(i),
				PubKey:     nodePubKey,
				PeeringURL: "",
				Connected:  false,
			}
			continue
		}
		peerStatus[i] = &PeerStatus{
			Index:      uint16(i),
			PubKey:     nodePubKey,
			PeeringURL: netPeerStatus[index].PeeringURL(),
			Connected:  netPeerStatus[index].IsAlive(),
		}
		if netPeerStatus[index].IsAlive() {
			connectedCount++
		}
	}
	ci := &CommitteeInfo{
		Address:       dkShare.GetAddress(),
		Size:          dkShare.GetN(),
		Quorum:        dkShare.GetT(),
		QuorumIsAlive: connectedCount >= dkShare.GetT(),
		PeerStatus:    peerStatus,
	}
	return ci
}

func (cni *chainNodeImpl) GetChainNodes() []peering.PeerStatusProvider {
	cni.accessLock.RLock()
	dkShare := cni.activeCommitteeDKShare
	acNodes := cni.activeAccessNodes
	srNodes := cni.serverNodes
	cni.accessLock.RUnlock()
	allNodeKeys := map[cryptolib.PublicKeyKey]*cryptolib.PublicKey{}
	//
	// Add committee nodes.
	if dkShare != nil {
		for _, nodePubKey := range dkShare.GetNodePubKeys() {
			allNodeKeys[nodePubKey.AsKey()] = nodePubKey
		}
	}
	//
	// Add access nodes.
	for _, nodePubKey := range acNodes {
		if _, ok := allNodeKeys[nodePubKey.AsKey()]; !ok {
			allNodeKeys[nodePubKey.AsKey()] = nodePubKey
		}
	}
	//
	// Add server nodes.
	for _, nodePubKey := range srNodes {
		if _, ok := allNodeKeys[nodePubKey.AsKey()]; !ok {
			allNodeKeys[nodePubKey.AsKey()] = nodePubKey
		}
	}
	//
	// Collect the relevant info.
	allNodes := []peering.PeerStatusProvider{}
	netNodes := cni.net.PeerStatus()
	for _, nodeKey := range allNodeKeys {
		index := slices.IndexFunc(netNodes, func(psp peering.PeerStatusProvider) bool {
			return nodeKey.Equals(psp.PubKey())
		})
		if index != -1 {
			allNodes = append(allNodes, netNodes[index])
		}
	}
	return allNodes
}

func (cni *chainNodeImpl) GetCandidateNodes() []*governance.AccessNodeInfo {
	state, err := cni.chainStore.LatestState()
	if err != nil {
		cni.log.Error("Cannot get latest chain state: %v", err)
		return []*governance.AccessNodeInfo{}
	}
	return governance.NewStateAccess(state).CandidateNodes()
}

func (cni *chainNodeImpl) GetChainMetrics() *metrics.ChainMetrics {
	return cni.chainMetrics
}

func (cni *chainNodeImpl) GetConsensusPipeMetrics() ConsensusPipeMetrics {
	return &consensusPipeMetricsImpl{}
}

func (cni *chainNodeImpl) GetConsensusWorkflowStatus() ConsensusWorkflowStatus {
	return &consensusWorkflowStatusImpl{}
}

func (cni *chainNodeImpl) recoverStoreFromWAL(chainStore indexedstore.IndexedStore, chainWAL sm_gpa_utils.BlockWAL) {
	//
	// Load all the existing blocks from the WAL.
	blocksAdded := 0
	err := chainWAL.ReadAllByStateIndex(func(stateIndex uint32, block state.Block) bool {
		cni.log.Debugf("TryRecoverStoreFromWAL: Adding a block to the store, stateIndex=%v, l1Commitment=%v, previousL1Commitment=%v", block.StateIndex(), block.L1Commitment(), block.PreviousL1Commitment())
		var stateDraft state.StateDraft
		if block.StateIndex() == 0 {
			stateDraft = chainStore.NewOriginStateDraft()
		} else {
			var stateErr error
			stateDraft, stateErr = chainStore.NewEmptyStateDraft(block.PreviousL1Commitment())
			if stateErr != nil {
				panic(fmt.Errorf("cannot create new state draft for previousL1Commitment=%v: %w", block.PreviousL1Commitment(), stateErr))
			}
		}
		block.Mutations().ApplyTo(stateDraft)
		chainStore.Commit(stateDraft)
		blocksAdded++
		return true
	})
	if err != nil {
		panic(fmt.Errorf("failed to iterate over WAL blocks: %w", err))
	}
	cni.log.Infof("TryRecoverStoreFromWAL: Done, added %v blocks.", blocksAdded)
}

type consensusPipeMetricsImpl struct{}                                        // TODO: Fake data, for now. Review metrics in general.
func (cpm *consensusPipeMetricsImpl) GetEventStateTransitionMsgPipeSize() int { return 0 }
func (cpm *consensusPipeMetricsImpl) GetEventPeerLogIndexMsgPipeSize() int    { return 0 }
func (cpm *consensusPipeMetricsImpl) GetEventACSMsgPipeSize() int             { return 0 }
func (cpm *consensusPipeMetricsImpl) GetEventVMResultMsgPipeSize() int        { return 0 }
func (cpm *consensusPipeMetricsImpl) GetEventTimerMsgPipeSize() int           { return 0 }

type consensusWorkflowStatusImpl struct{}                                       // TODO: Fake data, for now. Review metrics in general.
func (cws *consensusWorkflowStatusImpl) IsStateReceived() bool                  { return false }
func (cws *consensusWorkflowStatusImpl) IsBatchProposalSent() bool              { return false }
func (cws *consensusWorkflowStatusImpl) IsConsensusBatchKnown() bool            { return false }
func (cws *consensusWorkflowStatusImpl) IsVMStarted() bool                      { return false }
func (cws *consensusWorkflowStatusImpl) IsVMResultSigned() bool                 { return false }
func (cws *consensusWorkflowStatusImpl) IsTransactionFinalized() bool           { return false }
func (cws *consensusWorkflowStatusImpl) IsTransactionPosted() bool              { return false }
func (cws *consensusWorkflowStatusImpl) IsTransactionSeen() bool                { return false }
func (cws *consensusWorkflowStatusImpl) IsInProgress() bool                     { return false }
func (cws *consensusWorkflowStatusImpl) GetBatchProposalSentTime() time.Time    { return time.Time{} }
func (cws *consensusWorkflowStatusImpl) GetConsensusBatchKnownTime() time.Time  { return time.Time{} }
func (cws *consensusWorkflowStatusImpl) GetVMStartedTime() time.Time            { return time.Time{} }
func (cws *consensusWorkflowStatusImpl) GetVMResultSignedTime() time.Time       { return time.Time{} }
func (cws *consensusWorkflowStatusImpl) GetTransactionFinalizedTime() time.Time { return time.Time{} }
func (cws *consensusWorkflowStatusImpl) GetTransactionPostedTime() time.Time    { return time.Time{} }
func (cws *consensusWorkflowStatusImpl) GetTransactionSeenTime() time.Time      { return time.Time{} }
func (cws *consensusWorkflowStatusImpl) GetCompletedTime() time.Time            { return time.Time{} }
func (cws *consensusWorkflowStatusImpl) GetCurrentStateIndex() uint32           { return 0 }
