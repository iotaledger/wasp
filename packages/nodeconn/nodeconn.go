// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// nodeconn package provides an interface to the L1 node (Hornet).
// This component is responsible for:
//   - Protocol details.
//   - Block reattachments and promotions.
package nodeconn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/core/app/pkg/shutdown"
	"github.com/iotaledger/hive.go/core/generics/lo"
	"github.com/iotaledger/hive.go/core/generics/shrinkingmap"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/hive.go/core/timeutil"
	"github.com/iotaledger/hive.go/core/workerpool"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	inx "github.com/iotaledger/inx/go"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/common"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/parameters"
)

const (
	indexerPluginAvailableTimeout = 30 * time.Second
	l1NodeSyncWaitTimeout         = 2 * time.Minute
	inxTimeoutInfo                = 500 * time.Millisecond
	inxTimeoutBlockMetadata       = 500 * time.Millisecond
	inxTimeoutSubmitBlock         = 60 * time.Second
	inxTimeoutPublishTransaction  = 120 * time.Second
	inxTimeoutIndexerQuery        = 2 * time.Second
	inxTimeoutMilestone           = 2 * time.Second
	inxTimeoutOutput              = 2 * time.Second
	inxTimeoutGetPeers            = 2 * time.Second
	reattachWorkerPoolQueueSize   = 100

	chainsCleanupThresholdRatio              = 50.0
	chainsCleanupThresholdCount              = 10
	pendingTransactionsCleanupThresholdRatio = 50.0
	pendingTransactionsCleanupThresholdCount = 1000
)

type LedgerUpdateHandler func(*nodebridge.LedgerUpdate)

// nodeConnection implements chain.NodeConnection.
// Single Wasp node is expected to connect to a single L1 node, thus
// we expect to have a single instance of this structure.
type nodeConnection struct {
	*logger.WrappedLogger

	ctx                   context.Context
	chainsLock            sync.RWMutex
	chainsMap             *shrinkingmap.ShrinkingMap[isc.ChainID, *ncChain]
	indexerClient         nodeclient.IndexerClient
	nodeBridge            *nodebridge.NodeBridge
	nodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics
	nodeClient            *nodeclient.Client

	// pendingTransactionsMap is a map of sent transactions that are pending.
	pendingTransactionsMap  *shrinkingmap.ShrinkingMap[iotago.TransactionID, *pendingTransaction]
	pendingTransactionsLock sync.Mutex
	reattachWorkerPool      *workerpool.WorkerPool

	shutdownHandler *shutdown.ShutdownHandler
}

func setL1ProtocolParams(protocolParameters *iotago.ProtocolParameters, baseToken *nodeclient.InfoResBaseToken) {
	parameters.InitL1(&parameters.L1Params{
		// There are no limits on how big from a size perspective an essence can be,
		// so it is just derived from 32KB - Block fields without payload = max size of the payload
		MaxPayloadSize: parameters.MaxPayloadSize,
		Protocol:       protocolParameters,
		BaseToken:      (*parameters.BaseToken)(baseToken),
	})
}

func newCtxWithTimeout(ctx context.Context, defaultTimeout time.Duration, timeout ...time.Duration) (context.Context, context.CancelFunc) {
	t := defaultTimeout
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return context.WithTimeout(ctx, t)
}

func waitForL1ToBeConnected(ctx context.Context, log *logger.Logger, nodeBridge *nodebridge.NodeBridge) error {
	inxGetPeers := func(ctx context.Context) ([]*nodeclient.PeerResponse, error) {
		ctx, cancel := context.WithTimeout(ctx, inxTimeoutGetPeers)
		defer cancel()

		return nodeBridge.INXNodeClient().Peers(ctx)
	}

	getNodeConnected := func() (bool, error) {
		peers, err := inxGetPeers(ctx)
		if err != nil {
			return false, fmt.Errorf("failed to get peers: %w", err)
		}

		// check for at least one connected peer
		for _, peer := range peers {
			if peer.Connected {
				return true, nil
			}
		}

		log.Info("waiting for L1 to be connected to other peers...")
		return false, nil
	}

	ticker := time.NewTicker(1 * time.Second)
	defer timeutil.CleanupTicker(ticker)

	for {
		select {
		case <-ticker.C:
			nodeConnected, err := getNodeConnected()
			if err != nil {
				return err
			}

			if nodeConnected {
				// node is connected to other peers
				return nil
			}

		case <-ctx.Done():
			// context was canceled
			return common.ErrOperationAborted
		}
	}
}

func waitForL1ToBeSynced(ctx context.Context, log *logger.Logger, nodeBridge *nodebridge.NodeBridge) error {
	getMilestoneIndex := func() (uint32, uint32) {
		nodeStatus := nodeBridge.NodeStatus()

		var lsmi, cmi uint32
		if nodeStatus.GetLatestMilestone() != nil && nodeStatus.GetLatestMilestone().GetMilestoneInfo() != nil {
			lsmi = nodeStatus.GetLatestMilestone().GetMilestoneInfo().MilestoneIndex
		}
		if nodeStatus.GetConfirmedMilestone() != nil && nodeStatus.GetConfirmedMilestone().GetMilestoneInfo() != nil {
			cmi = nodeStatus.GetConfirmedMilestone().GetMilestoneInfo().MilestoneIndex
		}

		return lsmi, cmi
	}

	getNodeSynced := func() bool {
		// we can't use the "GetIsSynced()", because this flag also checks if the node has seen recent milestones.
		// in case the L1 network is halted, even if the node is "synced", we would wait forever.
		// we need another indicator if the L1 saw milestones from other peers recently.
		// lsmi != 0 indicates that the L1 saw heartbeat messages from peers after bootup, and len(peers) > 0

		lsmi, cmi := getMilestoneIndex()
		if lsmi > 0 && lsmi == cmi {
			// node seems to be synced
			return true
		}

		log.Infof("waiting for L1 to be fully synced... (%d/%d)", cmi, lsmi)

		return false
	}

	ticker := time.NewTicker(1 * time.Second)
	defer timeutil.CleanupTicker(ticker)

	for {
		select {
		case <-ticker.C:
			if getNodeSynced() {
				// node is synced
				return nil
			}
		case <-ctx.Done():
			// context was canceled
			return common.ErrOperationAborted
		}
	}
}

func New(
	ctx context.Context,
	log *logger.Logger,
	nodeBridge *nodebridge.NodeBridge,
	nodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics,
	shutdownHandler *shutdown.ShutdownHandler,
) (chain.NodeConnection, error) {
	ctxIndexer, cancelIndexer := context.WithTimeout(ctx, indexerPluginAvailableTimeout)
	defer cancelIndexer()

	indexerClient, err := nodeBridge.Indexer(ctxIndexer)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodeclient indexer: %w", err)
	}

	nc := &nodeConnection{
		WrappedLogger: logger.NewWrappedLogger(log),
		ctx:           nil,
		chainsMap: shrinkingmap.New[isc.ChainID, *ncChain](
			shrinkingmap.WithShrinkingThresholdRatio(chainsCleanupThresholdRatio),
			shrinkingmap.WithShrinkingThresholdCount(chainsCleanupThresholdCount),
		),
		chainsLock:            sync.RWMutex{},
		indexerClient:         indexerClient,
		nodeBridge:            nodeBridge,
		nodeConnectionMetrics: nodeConnectionMetrics,
		nodeClient:            nodeBridge.INXNodeClient(),
		pendingTransactionsMap: shrinkingmap.New[iotago.TransactionID, *pendingTransaction](
			shrinkingmap.WithShrinkingThresholdRatio(pendingTransactionsCleanupThresholdRatio),
			shrinkingmap.WithShrinkingThresholdCount(pendingTransactionsCleanupThresholdCount),
		),
		pendingTransactionsLock: sync.Mutex{},
		shutdownHandler:         shutdownHandler,
	}

	nc.reattachWorkerPool = workerpool.New(nc.reattachWorkerpoolFunc, workerpool.WorkerCount(1), workerpool.QueueSize(reattachWorkerPoolQueueSize))

	return nc, nil
}

func (nc *nodeConnection) Run(ctx context.Context) error {
	nc.ctx = ctx

	syncAndSetProtocolParameters := func() error {
		ctxWaitNodeSynced, cancelWaitNodeSynced := context.WithTimeout(ctx, l1NodeSyncWaitTimeout)
		defer cancelWaitNodeSynced()

		// make sure the node is connected to at least one other peer
		// otherwise the node status may not reflect the network status
		if err := waitForL1ToBeConnected(ctxWaitNodeSynced, nc.WrappedLogger.Logger(), nc.nodeBridge); err != nil {
			return err
		}

		if err := waitForL1ToBeSynced(ctxWaitNodeSynced, nc.WrappedLogger.Logger(), nc.nodeBridge); err != nil {
			return err
		}

		ctxInfo, cancelInfo := context.WithTimeout(ctx, inxTimeoutInfo)
		defer cancelInfo()

		nodeInfo, err := nc.nodeClient.Info(ctxInfo)
		if err != nil {
			return fmt.Errorf("error getting node info: %w", err)
		}
		setL1ProtocolParams(nc.nodeBridge.ProtocolParameters(), nodeInfo.BaseToken)

		return nil
	}

	if err := syncAndSetProtocolParameters(); err != nil {
		return fmt.Errorf("Getting latest L1 protocol parameters failed, error: %w", err)
	}

	nc.reattachWorkerPool.Start()
	go nc.subscribeToLedgerUpdates()
	<-ctx.Done()
	nc.reattachWorkerPool.StopAndWait()

	return nil
}

func (nc *nodeConnection) subscribeToLedgerUpdates() {
	err := nc.nodeBridge.ListenToLedgerUpdates(nc.ctx, 0, 0, nc.handleLedgerUpdate)
	if err != nil && !errors.Is(err, io.EOF) {
		nc.LogError(err)
		nc.shutdownHandler.SelfShutdown(
			fmt.Sprintf("INX connection unexpected error: %s", err.Error()),
			true)
		return
	}
	if nc.ctx.Err() == nil {
		// shutdown in case there isn't a shutdown already in progress
		nc.shutdownHandler.SelfShutdown("INX connection closed", true)
	}
}

func (nc *nodeConnection) getMilestoneTimestamp(ctx context.Context, msIndex iotago.MilestoneIndex) (time.Time, error) {
	ctx, cancel := newCtxWithTimeout(ctx, inxTimeoutMilestone)
	defer cancel()

	milestone, err := nc.nodeBridge.Milestone(ctx, msIndex)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(milestone.Milestone.Timestamp), 0), nil
}

func (nc *nodeConnection) outputForOutputID(ctx context.Context, outputID iotago.OutputID) (iotago.Output, error) {
	ctx, cancel := newCtxWithTimeout(ctx, inxTimeoutOutput)
	defer cancel()

	resp, err := nc.nodeBridge.Client().ReadOutput(ctx, inx.NewOutputId(outputID))
	if err != nil {
		return nil, err
	}

	switch resp.GetPayload().(type) {
	//nolint:nosnakecase // grpc uses underscores
	case *inx.OutputResponse_Output:
		iotaOutput, err := resp.GetOutput().UnwrapOutput(serializer.DeSeriModeNoValidation, nil)
		if err != nil {
			return nil, err
		}
		return iotaOutput, nil

	//nolint:nosnakecase // grpc uses underscores
	case *inx.OutputResponse_Spent:
		iotaOutput, err := resp.GetSpent().GetOutput().UnwrapOutput(serializer.DeSeriModeNoValidation, nil)
		if err != nil {
			return nil, err
		}
		return iotaOutput, nil

	default:
		return nil, errors.New("invalid inx.OutputResponse payload type")
	}
}

func (nc *nodeConnection) checkPendingTransactions(ledgerUpdate *ledgerUpdate) {
	nc.pendingTransactionsLock.Lock()
	defer nc.pendingTransactionsLock.Unlock()

	nc.pendingTransactionsMap.ForEach(func(txID iotago.TransactionID, pendingTx *pendingTransaction) bool {
		inputWasConsumed := false
		for _, consumedInput := range pendingTx.ConsumedInputs() {
			if _, exists := ledgerUpdate.outputsConsumedMap[consumedInput]; exists {
				inputWasConsumed = true

				break
			}
		}

		if !inputWasConsumed {
			// check if the transaction needs to be reattached
			nc.reattachWorkerPool.TrySubmit(pendingTx)
			return true
		}

		// a referenced input of this transaction was consumed, so the pending transaction is affected by this ledger update.
		// => we need to check if the outputs were created, otherwise this is a conflicting transaction.

		// we can easily check this by searching for output index 0.
		// if this was created, the rest was created as well because transactions are atomic.
		txOutputIDIndexZero := iotago.OutputIDFromTransactionIDAndIndex(pendingTx.ID(), 0)

		// mark waiting for pending transaction as done
		nc.clearPendingTransactionWithoutLocking(pendingTx.ID())

		if _, created := ledgerUpdate.outputsCreatedMap[txOutputIDIndexZero]; !created {
			// transaction was conflicting
			pendingTx.SetConflicting(errors.New("input was used in another transaction"))
		} else {
			// transaction was confirmed
			pendingTx.SetConfirmed()
		}

		return true
	})
}

func (nc *nodeConnection) triggerChainCallbacks(ledgerUpdate *ledgerUpdate) error {
	nc.chainsLock.RLock()
	defer nc.chainsLock.RUnlock()

	trackedAliasOutputsCreatedSortedMapByChainID, trackedAliasOutputsCreatedMapByOutputID, err := filterAndSortAliasOutputs(nc.chainsMap, ledgerUpdate)
	if err != nil {
		return err
	}

	otherOutputsCreatedMapByChainID := filterOtherOutputs(nc.chainsMap, ledgerUpdate.outputsCreatedMap, trackedAliasOutputsCreatedMapByOutputID)

	// fire milestone events for every chain
	nc.chainsMap.ForEach(func(_ isc.ChainID, chain *ncChain) bool {
		// the callbacks have to be fired synchronously, we can't guarantee the order of execution of go routines
		chain.HandleMilestone(ledgerUpdate.milestoneIndex, ledgerUpdate.milestoneTimestamp)
		return true
	})

	// fire the alias output events in order
	for chainID, aliasOutputsSorted := range trackedAliasOutputsCreatedSortedMapByChainID {
		ncChain, exists := nc.chainsMap.Get(chainID)
		if !exists {
			continue
		}

		for _, aliasOutputInfo := range aliasOutputsSorted {
			// the callbacks have to be fired synchronously, we can't guarantee the order of execution of go routines
			ncChain.HandleAliasOutput(ledgerUpdate.milestoneIndex, aliasOutputInfo)
		}
	}

	// fire events for all other outputs that were received by the chains
	for chainID, outputs := range otherOutputsCreatedMapByChainID {
		ncChain, exists := nc.chainsMap.Get(chainID)
		if !exists {
			continue
		}

		for _, outputInfo := range outputs {
			// the callbacks have to be fired synchronously, we can't guarantee the order of execution of go routines
			ncChain.HandleRequestOutput(ledgerUpdate.milestoneIndex, outputInfo)
		}
	}

	return nil
}

type ledgerUpdate struct {
	milestoneIndex     iotago.MilestoneIndex
	milestoneTimestamp time.Time
	outputsCreatedMap  map[iotago.OutputID]*isc.OutputInfo
	outputsConsumedMap map[iotago.OutputID]*isc.OutputInfo
}

func (nc *nodeConnection) unwrapLedgerUpdate(update *nodebridge.LedgerUpdate) (*ledgerUpdate, error) {
	var err error

	// we need to get the timestamp of the milestone from the node
	milestoneTimestamp, err := nc.getMilestoneTimestamp(nc.ctx, update.MilestoneIndex)
	if err != nil {
		return nil, err
	}

	outputsConsumed, err := unwrapSpents(update.Consumed)
	if err != nil {
		return nil, err
	}

	outputsCreated, err := unwrapOutputs(update.Created)
	if err != nil {
		return nil, err
	}

	// create maps for faster lookup
	// outputs that are created and consumed in the same milestone exist in both maps
	outputsConsumedMap := lo.KeyBy(outputsConsumed, func(output *isc.OutputInfo) iotago.OutputID {
		return output.OutputID
	})

	outputsCreatedMap := make(map[iotago.OutputID]*isc.OutputInfo, len(outputsCreated))
	lo.ForEach(outputsCreated, func(outputInfo *isc.OutputInfo) {
		// update info in case created outputs were also consumed
		if outputInfoConsumed, exists := outputsConsumedMap[outputInfo.OutputID]; exists {
			outputInfo.TransactionIDSpent = outputInfoConsumed.TransactionIDSpent
		}

		outputsCreatedMap[outputInfo.OutputID] = outputInfo
	})

	return &ledgerUpdate{
		milestoneIndex:     update.MilestoneIndex,
		milestoneTimestamp: milestoneTimestamp,
		outputsCreatedMap:  outputsCreatedMap,
		outputsConsumedMap: outputsConsumedMap,
	}, nil
}

func (nc *nodeConnection) handleLedgerUpdate(update *nodebridge.LedgerUpdate) error {
	// unwrap the ledger update into wasp structs
	ledgerUpdate, err := nc.unwrapLedgerUpdate(update)
	if err != nil {
		return err
	}

	// trigger the callbacks of all affected chains
	if err := nc.triggerChainCallbacks(ledgerUpdate); err != nil {
		return err
	}

	// check if pending transactions were affected by the ledger update
	nc.checkPendingTransactions(ledgerUpdate)

	return nil
}

// GetChain returns the chain if it was registered, otherwise it returns an error.
func (nc *nodeConnection) GetChain(chainID isc.ChainID) (*ncChain, error) {
	nc.chainsLock.RLock()
	defer nc.chainsLock.RUnlock()

	ncc, exists := nc.chainsMap.Get(chainID)
	if !exists {
		return nil, fmt.Errorf("chain %v is not connected", chainID.String())
	}

	return ncc, nil
}

func (nc *nodeConnection) GetMetrics() nodeconnmetrics.NodeConnectionMetrics {
	return nc.nodeConnectionMetrics
}

// doPostTx posts the transaction on layer 1 including tipselection and proof of work.
// this function does not wait until the transaction gets confirmed on L1.
func (nc *nodeConnection) doPostTx(ctx context.Context, tx *iotago.Transaction, tipsAdditional ...iotago.BlockID) (iotago.BlockID, error) {
	if ctx.Err() != nil {
		// context may have already been canceled
		return iotago.EmptyBlockID(), ctx.Err()
	}

	var parents iotago.BlockIDs

	// if no tips are given, we use empty parents, the node will do the tipselection in that case.
	// if tips are given, we should have at least BlockMaxParents/2 parents to have a healthy tangle.
	if len(tipsAdditional) > 0 {
		// add the tips to parents we want to reference.
		parents = append(parents, tipsAdditional...)

		if len(parents) < (iotago.BlockMaxParents / 2) {
			// not enough tips for a healthy tangle, request more from the node
			tips, err := nc.nodeBridge.RequestTips(ctx, uint32((iotago.BlockMaxParents/2)-len(parents)), false)
			if err != nil {
				return iotago.EmptyBlockID(), fmt.Errorf("failed to fetch tips: %w", err)
			}

			parents = append(parents, tips...)
		}
	}

	// Build a Block and post it.
	block, err := builder.NewBlockBuilder().
		Parents(parents).
		Payload(tx).
		Build()
	if err != nil {
		return iotago.EmptyBlockID(), fmt.Errorf("failed to build a tx: %w", err)
	}

	blockID, err := nc.nodeBridge.SubmitBlock(ctx, block)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			// context was canceled
			return iotago.EmptyBlockID(), ctx.Err()
		}
		return iotago.EmptyBlockID(), fmt.Errorf("failed to submit a tx: %w", err)
	}

	return blockID, nil
}

// addPendingTransaction tracks a pending transaction.
func (nc *nodeConnection) addPendingTransaction(pending *pendingTransaction) {
	nc.pendingTransactionsLock.Lock()
	defer nc.pendingTransactionsLock.Unlock()

	nc.pendingTransactionsMap.Set(pending.ID(), pending)
}

// hasPendingTransaction returns true if a pending transaction exists.
func (nc *nodeConnection) hasPendingTransaction(txID iotago.TransactionID) bool {
	nc.pendingTransactionsLock.Lock()
	defer nc.pendingTransactionsLock.Unlock()

	return nc.pendingTransactionsMap.Has(txID)
}

// clearPendingTransactionWithoutLocking removes tracking of a pending transaction.
// write lock must be acquired outside.
func (nc *nodeConnection) clearPendingTransactionWithoutLocking(transactionID iotago.TransactionID) {
	nc.pendingTransactionsMap.Delete(transactionID)
}

// clearPendingTransaction removes tracking of a pending transaction.
func (nc *nodeConnection) clearPendingTransaction(transactionID iotago.TransactionID) {
	nc.pendingTransactionsLock.Lock()
	defer nc.pendingTransactionsLock.Unlock()

	nc.clearPendingTransactionWithoutLocking(transactionID)
}

// reattachWorkerpoolFunc is triggered by handleLedgerUpdate for every pending transaction,
// if the inputs of the pending transaction were not consumed in the ledger update.
func (nc *nodeConnection) reattachWorkerpoolFunc(task workerpool.Task) {
	defer task.Return(nil)

	pendingTx := task.Param(0).(*pendingTransaction)
	if pendingTx.Conflicting() || pendingTx.Confirmed() {
		// no need to reattach
		// we can remove the tracking of the pending transaction
		nc.clearPendingTransaction(pendingTx.transactionID)
		return
	}

	blockID := pendingTx.BlockID()
	if blockID.Empty() {
		// no need to check because no block was posted by this node yet (maybe busy with doPostTx)
		return
	}

	ctxMetadata, cancelCtxMetadata := context.WithTimeout(nc.ctx, inxTimeoutBlockMetadata)
	defer cancelCtxMetadata()

	blockMetadata, err := nc.nodeBridge.BlockMetadata(ctxMetadata, blockID)
	if err != nil {
		// block not found
		nc.LogDebugf("reattaching transaction %s failed, error: block not found", pendingTx.ID().ToHex(), blockID.ToHex())
		return
	}

	// check confirmation while we are at it anyway
	if blockMetadata.ReferencedByMilestoneIndex != 0 {
		// block was referenced

		if blockMetadata.LedgerInclusionState == inx.BlockMetadata_LEDGER_INCLUSION_STATE_INCLUDED {
			// block was included => confirmed
			pendingTx.SetConfirmed()

			return
		}

		// block was referenced, but not included in the ledger
		pendingTx.SetConflicting(fmt.Errorf("tx was not included in the ledger. LedgerInclusionState: %s, ConflictReason: %d", blockMetadata.LedgerInclusionState, blockMetadata.ConflictReason))

		return
	}

	if blockMetadata.ShouldReattach {
		pendingTx.Reattach()

		return
	}

	// reattach or promote if needed
	if blockMetadata.ShouldPromote {
		nc.LogDebugf("promoting transaction %s", pendingTx.ID().ToHex())

		ctxSubmitBlock, cancelSubmitBlock := context.WithTimeout(nc.ctx, inxTimeoutSubmitBlock)
		defer cancelSubmitBlock()

		if err := nc.promoteBlock(ctxSubmitBlock, blockID); err != nil {
			nc.LogDebugf("promoting transaction %s failed, error: %w", pendingTx.ID().ToHex(), err)
			return
		}
	}
}

func (nc *nodeConnection) promoteBlock(ctx context.Context, blockID iotago.BlockID) error {
	tips, err := nc.nodeBridge.RequestTips(ctx, iotago.BlockMaxParents/2, false)
	if err != nil {
		return fmt.Errorf("failed to fetch tips: %w", err)
	}

	// add the blockID we want to promote
	tips = append(tips, blockID)

	block, err := builder.NewBlockBuilder().Parents(tips).Build()
	if err != nil {
		return fmt.Errorf("failed to build promotion block: %w", err)
	}

	if _, err = nc.nodeBridge.SubmitBlock(ctx, block); err != nil {
		return fmt.Errorf("failed to submit promotion block: %w", err)
	}

	return nil
}

// PublishTX handles promoting and reattachments until the tx is confirmed or the context is canceled.
// Publishing can be canceled via the context.
// The result must be returned via the callback, unless ctx is canceled first.
// It is fine to call the callback, even if the ctx is already canceled.
// PublishTX could be called multiple times in parallel, but only once per chain.
func (nc *nodeConnection) PublishTX(
	ctx context.Context,
	chainID isc.ChainID,
	tx *iotago.Transaction,
	callback chain.TxPostHandler,
) error {
	// check if the chain exists
	ncc, err := nc.GetChain(chainID)
	if err != nil {
		return err
	}

	pendingTx, err := ncc.createPendingTransaction(ctx, tx)
	if err != nil {
		return err
	}

	// transactions are published asynchronously
	go func() {
		err = ncc.publishTX(pendingTx)
		if err != nil {
			nc.LogDebug(err.Error())
		}

		// transaction was confirmed if err is nil
		callback(tx, err == nil)
	}()

	return nil
}

// Alias outputs are expected to be returned in order. Considering the Hornet node, the rules are:
//   - Upon Attach -- existing unspent alias output is returned FIRST.
//   - Upon receiving a spent/unspent AO from L1 they are returned in
//     the same order, as the milestones are issued.
//   - If a single milestone has several alias outputs, they have to be ordered
//     according to the chain of TXes.
//
// NOTE: Any out-of-order AO will be considered as a rollback or AO by the chain impl.
func (nc *nodeConnection) AttachChain(
	ctx context.Context, // ctx is the context given by a backgroundworker with PriorityChains
	chainID isc.ChainID,
	recvRequestCB chain.RequestOutputHandler,
	recvAliasOutput chain.AliasOutputHandler,
	recvMilestone chain.MilestoneHandler,
) {
	chain := func() *ncChain {
		// we need to lock until the chain init is done,
		// otherwise there could be race conditions with new ledger updates in parallel
		nc.chainsLock.Lock()
		defer nc.chainsLock.Unlock()

		chain := newNCChain(ctx, nc, chainID, recvRequestCB, recvAliasOutput, recvMilestone)

		// the chain is added to the map, even if not synchronzied yet,
		// so we can track all pending ledger updates until the chain is synchronized.
		nc.chainsMap.Set(chainID, chain)
		nc.nodeConnectionMetrics.SetRegistered(chainID)
		nc.LogDebugf("chain registered: %s = %s", chainID.ShortString(), chainID)

		return chain
	}()

	if err := chain.SyncChainStateWithL1(ctx); err != nil {
		nc.LogError(fmt.Sprintf("synchronizing chain state %s failed: %s", chainID, err.Error()))
		nc.shutdownHandler.SelfShutdown(
			fmt.Sprintf("Cannot sync chain %s with L1, %s", chain.chainID, err.Error()),
			true)
	}

	// disconnect the chain after the context is done
	go func() {
		<-ctx.Done()
		chain.WaitUntilStopped()

		nc.chainsLock.Lock()
		defer nc.chainsLock.Unlock()

		nc.chainsMap.Delete(chainID)
		nc.nodeConnectionMetrics.SetUnregistered(chainID)
		nc.LogDebugf("chain unregistered: %s = %s, |remaining|=%v", chainID.ShortString(), chainID, nc.chainsMap.Size())
	}()
}
