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
	"sync"
	"time"

	"golang.org/x/exp/slices"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/chainMgr"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/chain/cons"
	consGR "github.com/iotaledger/wasp/packages/chain/cons/gr"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/statemanager"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

const (
	recoveryTimeout         time.Duration = 15 * time.Minute // TODO: Make it configurable?
	redeliveryPeriod        time.Duration = 3 * time.Second  // TODO: Make it configurable?
	printStatusPeriod       time.Duration = 10 * time.Second // TODO: Make it configurable?
	consensusInstsInAdvance int           = 3                // TODO: Make it configurable?

	msgTypeChainMgr byte = iota
)

type ChainRequests interface {
	ChainReader
	ReceiveOffLedgerRequest(request isc.OffLedgerRequest, sender *cryptolib.PublicKey)
	AwaitRequestProcessed(ctx context.Context, requestID isc.RequestID) <-chan *blocklog.RequestReceipt
}

type Chain interface {
	ChainCore
	ChainRequests
	GetConsensusPipeMetrics() ConsensusPipeMetrics // TODO: Review this.
	GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMetrics
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
	Index     uint16
	PubKey    *cryptolib.PublicKey
	NetID     string
	Connected bool
}

type RequestOutputHandler = func(outputID iotago.OutputID, output iotago.Output)

// The Alias Outputs must be passed here in-order. The last alias output in the list
// is the unspent one (if there is a chain of outputs confirmed in a milestone).
type AliasOutputHandler = func(outputIDs []iotago.OutputID, outputs []*iotago.AliasOutput)

type TxPostHandler = func(tx *iotago.Transaction, confirmed bool)

type MilestoneHandler = func(timestamp time.Time)

type ChainNodeConn interface {
	// Publishing can be canceled via the context.
	// The result must be returned via the callback, unless ctx is canceled first.
	PublishTX(
		ctx context.Context,
		chainID *isc.ChainID,
		tx *iotago.Transaction,
		callback TxPostHandler,
	)
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
		chainID *isc.ChainID,
		recvRequestCB RequestOutputHandler,
		recvAliasOutput AliasOutputHandler,
		recvMilestone MilestoneHandler,
	)
}

type chainNodeImpl struct {
	me                     gpa.NodeID
	nodeIdentity           *cryptolib.KeyPair
	chainID                *isc.ChainID
	chainMgr               chainMgr.ChainMgr
	chainStore             state.Store
	nodeConn               NodeConnection
	mempool                mempool.Mempool
	stateMgr               statemanager.StateMgr
	recvAliasOutputPipe    pipe.Pipe
	recvTxPublishedPipe    pipe.Pipe
	recvMilestonePipe      pipe.Pipe
	consensusInsts         map[iotago.Ed25519Address]map[cmtLog.LogIndex]*consensusInst // Running consensus instances.
	consOutputPipe         pipe.Pipe
	consRecoverPipe        pipe.Pipe
	publishingTXes         map[iotago.TransactionID]context.CancelFunc // TX'es now being published.
	procCache              *processors.Cache                           // Cache for the SC processors.
	activeAccessLock       *sync.RWMutex                               // Mutex for accessing the active* fields from other threads.
	activeCommitteeDKShare tcrypto.DKShare                             // DKShare of the current active committee.
	activeCommitteeNodes   []*cryptolib.PublicKey                      // The nodes acting as a committee for the latest consensus.
	activeAccessNodes      []*cryptolib.PublicKey                      // All the nodes authorized for being access nodes (for the ActiveAO).
	netRecvPipe            pipe.Pipe
	netPeeringID           peering.PeeringID
	netPeerPubs            map[gpa.NodeID]*cryptolib.PublicKey
	net                    peering.NetworkProvider
	log                    *logger.Logger
}

type consensusInst struct {
	request    *chainMgr.NeedConsensus
	cancelFunc context.CancelFunc
	consensus  *consGR.ConsGr
	committee  []*cryptolib.PublicKey
}

// Used to correlate consensus request with its output.
type consOutput struct {
	request *chainMgr.NeedConsensus
	output  *consGR.Output
}

// Used to correlate consensus request with its output.
type consRecover struct {
	request *chainMgr.NeedConsensus
}

// This is event received from the NodeConn as response to PublishTX
type txPublished struct {
	committeeAddr   iotago.Ed25519Address
	txID            iotago.TransactionID
	nextAliasOutput *isc.AliasOutputWithID
	confirmed       bool
}

var _ Chain = &chainNodeImpl{}

func New(
	ctx context.Context,
	chainID *isc.ChainID,
	chainStore state.Store,
	nodeConn NodeConnection,
	nodeIdentity *cryptolib.KeyPair,
	processorConfig *processors.Config,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	cmtLogStore cmtLog.Store,
	blockWAL smGPAUtils.BlockWAL,
	net peering.NetworkProvider,
	log *logger.Logger,
) (Chain, error) {
	netPeeringID := peering.PeeringIDFromBytes(append(chainID.Bytes(), []byte("ChainMgr")...))
	cni := &chainNodeImpl{
		nodeIdentity:           nodeIdentity,
		chainID:                chainID,
		chainStore:             chainStore,
		nodeConn:               nodeConn,
		recvAliasOutputPipe:    pipe.NewDefaultInfinitePipe(),
		recvTxPublishedPipe:    pipe.NewDefaultInfinitePipe(),
		recvMilestonePipe:      pipe.NewDefaultInfinitePipe(),
		consensusInsts:         map[iotago.Ed25519Address]map[cmtLog.LogIndex]*consensusInst{},
		consOutputPipe:         pipe.NewDefaultInfinitePipe(),
		consRecoverPipe:        pipe.NewDefaultInfinitePipe(),
		publishingTXes:         map[iotago.TransactionID]context.CancelFunc{},
		procCache:              processors.MustNew(processorConfig),
		activeAccessLock:       &sync.RWMutex{},
		activeCommitteeDKShare: nil,
		activeCommitteeNodes:   []*cryptolib.PublicKey{},
		activeAccessNodes:      []*cryptolib.PublicKey{},
		netRecvPipe:            pipe.NewDefaultInfinitePipe(),
		netPeeringID:           netPeeringID,
		netPeerPubs:            map[gpa.NodeID]*cryptolib.PublicKey{},
		net:                    net,
		log:                    log,
	}
	cni.me = cni.pubKeyAsNodeID(nodeIdentity.GetPublicKey())
	//
	// Create sub-components.
	chainMetrics := metrics.EmptyChainMetrics()
	chainMgr, err := chainMgr.New(cni.me, *cni.chainID, cmtLogStore, dkShareRegistryProvider, cni.pubKeyAsNodeID, cni.handleAccessNodesCB, cni.log)
	if err != nil {
		return nil, xerrors.Errorf("cannot create chainMgr: %w", err)
	}
	stateMgr, err := statemanager.New(
		ctx,
		cni.chainID,
		nodeIdentity.GetPublicKey(),
		[]*cryptolib.PublicKey{nodeIdentity.GetPublicKey()},
		net,
		blockWAL,
		chainStore,
		log,
	)
	if err != nil {
		return nil, xerrors.Errorf("cannot create stateMgr: %w", err)
	}
	// TODO: Review, if all needed functions are called for mempool.
	mempool := mempool.New(ctx, chainID, nodeIdentity, stateMgr, net, log, chainMetrics)
	cni.chainMgr = chainMgr
	cni.stateMgr = stateMgr
	cni.mempool = mempool
	//
	// Connect to the peering network.
	netRecvPipeInCh := cni.netRecvPipe.In()
	netAttachID := net.Attach(&netPeeringID, peering.PeerMessageReceiverChain, func(recv *peering.PeerMessageIn) {
		if recv.MsgType != msgTypeChainMgr {
			cni.log.Warnf("Unexpected message, type=%v", recv.MsgType)
			return
		}
		netRecvPipeInCh <- recv
	})
	//
	// Attach to the L1.
	recvAliasOutputPipeInCh := cni.recvAliasOutputPipe.In()
	recvAliasOutputCB := func(outputIDs []iotago.OutputID, outputs []*iotago.AliasOutput) {
		if len(outputIDs) == 0 {
			return
		}
		for i := range outputIDs {
			aliasOutput := isc.NewAliasOutputWithID(outputs[i], outputIDs[i])
			cni.stateMgr.ReceiveConfirmedAliasOutput(aliasOutput)
		}
		last := len(outputIDs) - 1
		lastAliasOutput := isc.NewAliasOutputWithID(outputs[last], outputIDs[last])
		recvAliasOutputPipeInCh <- lastAliasOutput
	}
	recvRequestCB := func(outputID iotago.OutputID, output iotago.Output) {
		req, err := isc.OnLedgerFromUTXO(output, outputID)
		if err != nil {
			cni.log.Warnf("Cannot create OnLedgerRequest from output: %v", err)
			return
		}
		cni.mempool.ReceiveOnLedgerRequest(req)
	}
	recvMilestonePipeInCh := cni.recvMilestonePipe.In()
	recvMilestoneCB := func(timestamp time.Time) {
		recvMilestonePipeInCh <- timestamp
	}
	nodeConn.AttachChain(ctx, chainID, recvRequestCB, recvAliasOutputCB, recvMilestoneCB)
	//
	// Run the main thread.
	go cni.run(ctx, netAttachID)
	return cni, nil
}

func (cni *chainNodeImpl) ReceiveOffLedgerRequest(request isc.OffLedgerRequest, sender *cryptolib.PublicKey) {
	// TODO: What to do with the sender's pub key?
	cni.mempool.ReceiveOffLedgerRequest(request)
}

func (cni *chainNodeImpl) AwaitRequestProcessed(ctx context.Context, requestID isc.RequestID) <-chan *blocklog.RequestReceipt {
	return cni.mempool.AwaitRequestProcessed(ctx, requestID)
}

func (cni *chainNodeImpl) run(ctx context.Context, netAttachID interface{}) {
	ctxDone := ctx.Done()
	recvAliasOutputPipeOutCh := cni.recvAliasOutputPipe.Out()
	recvTxPublishedPipeOutCh := cni.recvTxPublishedPipe.Out()
	recvMilestonePipeOutCh := cni.recvMilestonePipe.Out()
	netRecvPipeOutCh := cni.netRecvPipe.Out()
	consOutputPipeOutCh := cni.consOutputPipe.Out()
	consRecoverPipeOutCh := cni.consRecoverPipe.Out()
	for {
		select {
		case txPublishResult, ok := <-recvTxPublishedPipeOutCh:
			if !ok {
				recvTxPublishedPipeOutCh = nil
				continue
			}
			cni.handleTxPublished(ctx, txPublishResult.(*txPublished))
		case aliasOutput, ok := <-recvAliasOutputPipeOutCh:
			if !ok {
				recvAliasOutputPipeOutCh = nil
				continue
			}
			cni.handleAliasOutput(ctx, aliasOutput.(*isc.AliasOutputWithID))
		case timestamp, ok := <-recvMilestonePipeOutCh:
			if !ok {
				recvMilestonePipeOutCh = nil
			}
			cni.handleMilestoneTimestamp(timestamp.(time.Time))
		case recv, ok := <-netRecvPipeOutCh:
			if !ok {
				netRecvPipeOutCh = nil
				continue
			}
			cni.handleNetMessage(ctx, recv.(*peering.PeerMessageIn))
		case recv, ok := <-consOutputPipeOutCh:
			if !ok {
				consOutputPipeOutCh = nil
				continue
			}
			cni.handleConsensusOutput(ctx, recv.(*consOutput))
		case recv, ok := <-consRecoverPipeOutCh:
			if !ok {
				consRecoverPipeOutCh = nil
				continue
			}
			cni.handleConsensusRecover(ctx, recv.(*consRecover))
		case <-ctxDone:
			cni.net.Detach(netAttachID)
			return
		}
	}
}

// This will always run in the main thread, because that's a callback for the chainMgr.
func (cni *chainNodeImpl) handleAccessNodesCB(accessNodes []*cryptolib.PublicKey) {
	cni.activeAccessLock.Lock()
	cni.activeAccessNodes = accessNodes
	activeCommitteeNodes := cni.activeCommitteeNodes
	cni.activeAccessLock.Unlock()
	cni.log.Infof("Access nodes updated: %+v", accessNodes)
	cni.mempool.AccessNodesUpdated(activeCommitteeNodes, accessNodes)
	cni.stateMgr.AccessNodesUpdated(accessNodes)
}

func (cni *chainNodeImpl) handleTxPublished(ctx context.Context, txPubResult *txPublished) {
	if _, ok := cni.publishingTXes[txPubResult.txID]; !ok {
		return
	}
	delete(cni.publishingTXes, txPubResult.txID)
	outMsgs := cni.chainMgr.AsGPA().Input(
		chainMgr.NewInputChainTxPublishResult(txPubResult.committeeAddr, txPubResult.txID, txPubResult.nextAliasOutput, txPubResult.confirmed),
	)
	cni.sendMessages(outMsgs)
	cni.handleChainMgrOutput(ctx, cni.chainMgr.AsGPA().Output())
}

func (cni *chainNodeImpl) handleAliasOutput(ctx context.Context, aliasOutput *isc.AliasOutputWithID) {
	outMsgs := cni.chainMgr.AsGPA().Input(
		chainMgr.NewInputAliasOutputConfirmed(aliasOutput),
	)
	cni.sendMessages(outMsgs)
	cni.handleChainMgrOutput(ctx, cni.chainMgr.AsGPA().Output())
}

func (cni *chainNodeImpl) handleMilestoneTimestamp(timestamp time.Time) {
	for ji := range cni.consensusInsts {
		for li := range cni.consensusInsts[ji] {
			ci := cni.consensusInsts[ji][li]
			if ci.cancelFunc != nil {
				ci.consensus.Time(timestamp)
			}
		}
	}
}

func (cni *chainNodeImpl) handleNetMessage(ctx context.Context, recv *peering.PeerMessageIn) {
	msg, err := cni.chainMgr.AsGPA().UnmarshalMessage(recv.MsgData)
	if err != nil {
		cni.log.Warnf("cannot parse message: %v", err)
		return
	}
	msg.SetSender(cni.pubKeyAsNodeID(recv.SenderPubKey))
	outMsgs := cni.chainMgr.AsGPA().Message(msg)
	cni.sendMessages(outMsgs)
	cni.handleChainMgrOutput(ctx, cni.chainMgr.AsGPA().Output())
}

func (cni *chainNodeImpl) handleChainMgrOutput(ctx context.Context, outputUntyped gpa.Output) {
	if outputUntyped == nil {
		cni.cleanupConsensusInsts(nil, nil)
		cni.cleanupPublishingTXes(map[iotago.TransactionID]*chainMgr.NeedPublishTX{})
		return
	}
	output := outputUntyped.(*chainMgr.Output)
	//
	// Start new consensus instances, if needed.
	outputNeedConsensus := output.NeedConsensus()
	if outputNeedConsensus != nil {
		cni.ensureConsensusInput(ctx, outputNeedConsensus)
	}
	//
	// Start publishing TX'es, if there not being posted already.
	outputNeedPostTXes := output.NeedPublishTX()
	for i := range outputNeedPostTXes {
		txToPost := outputNeedPostTXes[i] // Have to take a copy to be used in callback.
		if _, ok := cni.publishingTXes[txToPost.TxID]; !ok {
			subCtx, subCancel := context.WithCancel(ctx)
			cni.publishingTXes[txToPost.TxID] = subCancel
			cni.nodeConn.PublishTX(subCtx, cni.chainID, txToPost.Tx, func(tx *iotago.Transaction, confirmed bool) {
				cni.recvTxPublishedPipe.In() <- &txPublished{
					committeeAddr:   txToPost.CommitteeAddr,
					txID:            txToPost.TxID,
					nextAliasOutput: txToPost.NextAliasOutput,
					confirmed:       confirmed,
				}
			})
		}
	}
	cni.cleanupPublishingTXes(outputNeedPostTXes)
}

func (cni *chainNodeImpl) handleConsensusOutput(ctx context.Context, out *consOutput) {
	var chainMgrInput gpa.Input
	switch out.output.Status {
	case cons.Completed:
		stateAnchor, aliasOutput, err := transaction.GetAnchorFromTransaction(out.output.TX)
		if err != nil {
			panic(xerrors.Errorf("cannot extract next AliasOutput from TX: %w", err))
		}
		nextAO := isc.NewAliasOutputWithID(aliasOutput, stateAnchor.OutputID)
		chainMgrInput = chainMgr.NewInputConsensusOutputDone(
			out.request.CommitteeAddr,
			out.request.LogIndex,
			out.request.BaseAliasOutput.OutputID(),
			nextAO,
			out.output.NextState,
			out.output.TX,
		)
	case cons.Skipped:
		chainMgrInput = chainMgr.NewInputConsensusOutputSkip(
			out.request.CommitteeAddr,
			out.request.LogIndex,
			out.request.BaseAliasOutput.OutputID(),
		)
	default:
		panic(fmt.Errorf("unexpected output state from consensus: %+v", out))
	}
	cni.sendMessages(cni.chainMgr.AsGPA().Input(chainMgrInput))
	cni.handleChainMgrOutput(ctx, cni.chainMgr.AsGPA().Output())
}

func (cni *chainNodeImpl) handleConsensusRecover(ctx context.Context, out *consRecover) {
	chainMgrInput := chainMgr.NewInputConsensusTimeout(
		out.request.CommitteeAddr,
		out.request.LogIndex,
	)
	cni.sendMessages(cni.chainMgr.AsGPA().Input(chainMgrInput))
	cni.handleChainMgrOutput(ctx, cni.chainMgr.AsGPA().Output())
}

func (cni *chainNodeImpl) ensureConsensusInput(ctx context.Context, needConsensus *chainMgr.NeedConsensus) {
	ci := cni.ensureConsensusInst(ctx, needConsensus)
	if ci.request == nil {
		outputCB := func(o *consGR.Output) {
			cni.consOutputPipe.In() <- &consOutput{request: needConsensus, output: o}
		}
		recoverCB := func() {
			cni.consRecoverPipe.In() <- &consRecover{request: needConsensus}
		}
		ci.request = needConsensus
		ci.consensus.Input(needConsensus.BaseAliasOutput, outputCB, recoverCB)
		//
		// Update committee nodes, if changed.
		cni.activeAccessLock.Lock()
		cni.activeCommitteeDKShare = needConsensus.DKShare
		cni.activeAccessLock.Unlock()
		if !util.Same(ci.committee, cni.activeCommitteeNodes) {
			cni.activeAccessLock.Lock()
			cni.activeCommitteeNodes = ci.committee
			cni.activeAccessLock.Unlock()
			cni.log.Infof("Committee nodes updated: %+v", cni.activeCommitteeNodes)
			cni.mempool.AccessNodesUpdated(cni.activeCommitteeNodes, cni.activeAccessNodes)
		}
	}
}

func (cni *chainNodeImpl) ensureConsensusInst(ctx context.Context, needConsensus *chainMgr.NeedConsensus) *consensusInst {
	committeeAddr := needConsensus.CommitteeAddr
	logIndex := needConsensus.LogIndex
	dkShare := needConsensus.DKShare
	if _, ok := cni.consensusInsts[committeeAddr]; !ok {
		cni.consensusInsts[committeeAddr] = map[cmtLog.LogIndex]*consensusInst{}
	}
	addLogIndex := logIndex
	added := false
	for i := 0; i < consensusInstsInAdvance; i++ {
		if _, ok := cni.consensusInsts[committeeAddr][addLogIndex]; !ok {
			consGrCtx, consGrCancel := context.WithCancel(ctx)
			logIndexCopy := addLogIndex
			cgr := consGR.New(
				consGrCtx, cni.chainID, cni.chainStore, dkShare, &logIndexCopy, cni.nodeIdentity,
				cni.procCache, cni.mempool, cni.stateMgr, cni.net,
				recoveryTimeout, redeliveryPeriod, printStatusPeriod, cni.log,
			)
			cni.consensusInsts[committeeAddr][addLogIndex] = &consensusInst{ // TODO: Handle terminations somehow.
				cancelFunc: consGrCancel,
				consensus:  cgr,
				committee:  dkShare.GetNodePubKeys(),
			}
			added = true
		}
		addLogIndex = addLogIndex.Next()
	}
	if added {
		cni.cleanupConsensusInsts(&committeeAddr, &logIndex)
	}
	return cni.consensusInsts[committeeAddr][logIndex]
}

func (cni *chainNodeImpl) cleanupConsensusInsts(keepCommitteeAddr *iotago.Ed25519Address, keepLogIndex *cmtLog.LogIndex) {
	for cmtAddr := range cni.consensusInsts {
		for li := range cni.consensusInsts[cmtAddr] {
			if keepCommitteeAddr != nil && keepLogIndex != nil && cmtAddr.Equal(keepCommitteeAddr) && li >= *keepLogIndex {
				continue
			}
			ci := cni.consensusInsts[cmtAddr][li]
			if ci.request == nil && ci.cancelFunc != nil {
				// We can cancel an instance, if input was not yet provided.
				ci.cancelFunc() // TODO: Somehow cancel hanging instances, maybe with old LogIndex, etc.
				ci.cancelFunc = nil
			}
		}
	}
}

// Cleanup TX'es that are not needed to be posted anymore.
func (cni *chainNodeImpl) cleanupPublishingTXes(neededPostTXes map[iotago.TransactionID]*chainMgr.NeedPublishTX) {
	for txID, cancelFunc := range cni.publishingTXes {
		found := false
		for _, npt := range neededPostTXes {
			if npt.TxID == txID {
				found = true
				break
			}
		}
		if !found {
			cancelFunc()
			delete(cni.publishingTXes, txID)
		}
	}
}

func (cni *chainNodeImpl) sendMessages(outMsgs gpa.OutMessages) {
	if outMsgs == nil {
		return
	}
	outMsgs.MustIterate(func(m gpa.Message) {
		msgData, err := m.MarshalBinary()
		if err != nil {
			cni.log.Warnf("Failed to send a message: %v", err)
			return
		}
		pm := &peering.PeerMessageData{
			PeeringID:   cni.netPeeringID,
			MsgReceiver: peering.PeerMessageReceiverChain,
			MsgType:     msgTypeChainMgr,
			MsgData:     msgData,
		}
		cni.net.SendMsgByPubKey(cni.netPeerPubs[m.Recipient()], pm)
	})
}

func (cni *chainNodeImpl) pubKeyAsNodeID(pubKey *cryptolib.PublicKey) gpa.NodeID {
	nodeID := gpa.NodeID(pubKey.String())
	if _, ok := cni.netPeerPubs[nodeID]; !ok {
		cni.netPeerPubs[nodeID] = pubKey
	}
	return nodeID
}

////////////////////////////////////////////////////////////////////////////////
// Support functions.

func (cni *chainNodeImpl) ID() *isc.ChainID {
	return cni.chainID
}

func (cni *chainNodeImpl) GetStateReader() state.Store {
	return cni.chainStore
}

func (cni *chainNodeImpl) Processors() *processors.Cache {
	return cni.procCache
}

func (cni *chainNodeImpl) Log() *logger.Logger {
	return cni.log
}

func (cni *chainNodeImpl) GetCommitteeInfo() *CommitteeInfo {
	cni.activeAccessLock.RLock()
	dkShare := cni.activeCommitteeDKShare
	cni.activeAccessLock.RUnlock()
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
				Index:     uint16(i),
				PubKey:    nodePubKey,
				NetID:     "",
				Connected: false,
			}
			continue
		}
		peerStatus[i] = &PeerStatus{
			Index:     uint16(i),
			PubKey:    nodePubKey,
			NetID:     netPeerStatus[index].NetID(),
			Connected: netPeerStatus[index].IsAlive(),
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
	cni.activeAccessLock.RLock()
	dkShare := cni.activeCommitteeDKShare
	acNodes := cni.activeAccessNodes
	cni.activeAccessLock.RUnlock()
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
	return governance.NewStateAccess(state).GetCandidateNodes()
}

func (cni *chainNodeImpl) GetConsensusPipeMetrics() ConsensusPipeMetrics {
	return &consensusPipeMetricsImpl{}
}

func (cni *chainNodeImpl) GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMetrics {
	return cni.nodeConn.GetMetrics()
}

func (cni *chainNodeImpl) GetConsensusWorkflowStatus() ConsensusWorkflowStatus {
	return &consensusWorkflowStatusImpl{}
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
