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
	"sync"
	"time"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/events"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/hive.go/core/workerpool"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/inx-app/nodebridge"
	inx "github.com/iotaledger/inx/go"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	indexerPluginAvailableTimeout = 30 * time.Second
	inxTimeoutInfo                = 500 * time.Millisecond
	inxTimeoutBlockMetadata       = 500 * time.Millisecond
	inxTimeoutSubmitBlock         = 60 * time.Second
	inxTimeoutPublishTransaction  = 120 * time.Second
	reattachWorkerPoolQueueSize   = 100
)

type LedgerUpdateHandler func(*nodebridge.LedgerUpdate)

// nodeconn implements chain.NodeConnection.
// Single Wasp node is expected to connect to a single L1 node, thus
// we expect to have a single instance of this structure.
type nodeConn struct {
	ctx           context.Context
	chains        map[string]*ncChain // key = iotago.Address.Key()
	chainsLock    sync.RWMutex
	indexerClient nodeclient.IndexerClient
	metrics       nodeconnmetrics.NodeConnectionMetrics
	log           *logger.Logger
	nodeBridge    *nodebridge.NodeBridge
	nodeClient    *nodeclient.Client

	// pendingTransactionsMap is a map of sent transactions that are pending.
	pendingTransactionsMap  map[iotago.TransactionID]*PendingTransaction
	pendingTransactionsLock sync.Mutex
	reattachWorkerPool      *workerpool.WorkerPool
}

func setL1ProtocolParams(protocolParameters *iotago.ProtocolParameters, baseToken *nodeclient.InfoResBaseToken) {
	parameters.InitL1(&parameters.L1Params{
		// There are no limits on how big from a size perspective an essence can be, so it is just derived from 32KB - Block fields without payload = max size of the payload
		MaxPayloadSize: parameters.MaxPayloadSize,
		Protocol:       protocolParameters,
		BaseToken:      (*parameters.BaseToken)(baseToken),
	})
}

const defaultTimeout = 1 * time.Minute

func newCtxWithTimeout(ctx context.Context, timeout ...time.Duration) (context.Context, context.CancelFunc) {
	t := defaultTimeout
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return context.WithTimeout(ctx, t)
}

func New(ctx context.Context, log *logger.Logger, nodeBridge *nodebridge.NodeBridge) chain.NodeConnection {
	inxNodeClient := nodeBridge.INXNodeClient()

	ctxInfo, cancelInfo := context.WithTimeout(ctx, inxTimeoutInfo)
	defer cancelInfo()

	nodeInfo, err := inxNodeClient.Info(ctxInfo)
	if err != nil {
		panic(xerrors.Errorf("error getting node info: %w", err))
	}
	setL1ProtocolParams(nodeBridge.ProtocolParameters(), nodeInfo.BaseToken)

	ctxIndexer, cancelIndexer := context.WithTimeout(ctx, indexerPluginAvailableTimeout)
	defer cancelIndexer()

	indexerClient, err := nodeBridge.Indexer(ctxIndexer)
	if err != nil {
		panic(xerrors.Errorf("failed to get nodeclient indexer: %v", err))
	}

	nc := nodeConn{
		ctx:                     ctx,
		chains:                  make(map[string]*ncChain),
		chainsLock:              sync.RWMutex{},
		indexerClient:           indexerClient,
		metrics:                 nodeconnmetrics.NewEmptyNodeConnectionMetrics(),
		log:                     log.Named("nc"),
		nodeBridge:              nodeBridge,
		nodeClient:              inxNodeClient,
		pendingTransactionsMap:  make(map[iotago.TransactionID]*PendingTransaction),
		pendingTransactionsLock: sync.Mutex{},
	}

	nc.reattachWorkerPool = workerpool.New(nc.reattachWorkerpoolFunc, workerpool.WorkerCount(1), workerpool.QueueSize(reattachWorkerPoolQueueSize))
	nc.reattachWorkerPool.Start()

	go nc.subscribeToLedgerUpdates()

	return &nc
}

func (nc *nodeConn) subscribeToLedgerUpdates() {
	err := nc.nodeBridge.ListenToLedgerUpdates(nc.ctx, 0, 0, nc.handleLedgerUpdate)
	if err != nil {
		nc.log.Panic(err)
	}
}

func (nc *nodeConn) handleLedgerUpdate(update *nodebridge.LedgerUpdate) error {
	nc.chainsLock.RLock()
	defer nc.chainsLock.RUnlock()

	// inline function used to release the lock with defer
	go func() {
		// create maps for faster lookup.
		// outputs that are created and consumed in the same milestone exist in both maps.
		newSpentsMap := make(map[iotago.OutputID]struct{})
		for _, spent := range update.Consumed {
			newSpentsMap[spent.GetOutput().GetOutputId().Unwrap()] = struct{}{}
		}

		newOutputsMap := make(map[iotago.OutputID]struct{})
		for _, output := range update.Created {
			newOutputsMap[output.GetOutputId().Unwrap()] = struct{}{}
		}

		nc.pendingTransactionsLock.Lock()
		defer nc.pendingTransactionsLock.Unlock()

		// check if pending transactions were affected by the ledger update.
		for _, pendingTx := range nc.pendingTransactionsMap {
			inputWasSpent := false
			for _, consumedInput := range pendingTx.ConsumedInputs() {
				if _, spent := newSpentsMap[consumedInput]; spent {
					inputWasSpent = true

					break
				}
			}

			if inputWasSpent {
				// a referenced input of this transaction was spent, so the pending transaction is affected by this ledger update.
				// => we need to check if the outputs were created, otherwise this is a conflicting transaction.

				// we can easily check this by searching for output index 0.
				// if this was created, the rest was created as well because transactions are atomic.
				txOutputIndexZero := iotago.UTXOInput{
					TransactionID:          pendingTx.ID(),
					TransactionOutputIndex: 0,
				}

				// mark waiting for pending transaction as done
				nc.clearPendingTransactionWithoutLocking(pendingTx.ID())

				if _, created := newOutputsMap[txOutputIndexZero.ID()]; !created {
					// transaction was conflicting
					pendingTx.SetConflicting(xerrors.New("input was used in another transaction"))
				} else {
					// transaction was confirmed
					pendingTx.SetConfirmed()
				}
			} else {
				// check if the transaction needs to be reattached
				nc.reattachWorkerPool.TrySubmit(pendingTx)
			}
		}
	}()

	for _, ledgerOutput := range update.Created {
		output, err := ledgerOutput.UnwrapOutput(serializer.DeSeriModeNoValidation, parameters.L1().Protocol)
		if err != nil {
			return err
		}

		// notify chains about state updates
		if aliasOutput, ok := output.(*iotago.AliasOutput); ok {
			outputID := ledgerOutput.GetOutputId().Unwrap()
			aliasID := util.AliasIDFromAliasOutput(aliasOutput, outputID)
			chainID := isc.ChainIDFromAliasID(aliasID)
			ncChain := nc.chains[chainID.Key()]
			if ncChain != nil {
				go ncChain.HandleStateUpdate(outputID, aliasOutput)
			}
		}

		// notify chains about new UTXOS owned by them
		unlockAddr := output.UnlockConditionSet().Address()
		if unlockAddr == nil {
			continue
		}
		if unlockAliasAddr, ok := unlockAddr.Address.(*iotago.AliasAddress); ok {
			chainID := isc.ChainIDFromAliasID(unlockAliasAddr.AliasID())
			ncChain := nc.chains[chainID.Key()]
			if ncChain != nil {
				go ncChain.HandleUnlockableOutput(ledgerOutput.GetOutputId().Unwrap(), output)
			}
		}
	}

	return nil
}

func (nc *nodeConn) enableMilestoneTrigger() {
	nc.nodeBridge.Events.LatestMilestoneChanged.Hook(hivecore.NewClosure(func(metadata *nodebridge.Milestone) {
		milestone := nodeclient.MilestoneInfo{
			MilestoneID: metadata.MilestoneID.String(),
			Index:       metadata.Milestone.Index,
			Timestamp:   metadata.Milestone.Timestamp,
		}

		nc.log.Debugf("Milestone received, index=%v, timestamp=%v", milestone.Index, milestone.Timestamp)

		nc.metrics.GetInMilestone().CountLastMessage(&milestone)
		nc.milestones.Trigger(&milestone)
	}))
}

func (nc *nodeConn) SetMetrics(metrics nodeconnmetrics.NodeConnectionMetrics) {
	nc.metrics = metrics
}

// RegisterChain implements chain.NodeConnection.
func (nc *nodeConn) RegisterChain(
	chainID *isc.ChainID,
	stateOutputHandler,
	outputHandler func(iotago.OutputID, iotago.Output),
	milestoneHandler func(*nodebridge.Milestone),
) {
	nc.metrics.SetRegistered(chainID)
	ncc := newNCChain(nc, chainID, stateOutputHandler, outputHandler, milestoneHandler)
	nc.chainsLock.Lock()
	defer nc.chainsLock.Unlock()
	nc.chains[chainID.Key()] = ncc
	nc.log.Debugf("nodeconn: chain registered: %s", chainID)
}

// UnregisterChain implements chain.NodeConnection.
func (nc *nodeConn) UnregisterChain(chainID *isc.ChainID) {
	nc.metrics.SetUnregistered(chainID)
	nccKey := chainID.Key()
	nc.chainsLock.Lock()
	defer nc.chainsLock.Unlock()
	if ncc, ok := nc.chains[nccKey]; ok {
		ncc.Close()
		delete(nc.chains, nccKey)
	}
	nc.log.Debugf("nodeconn: chain unregistered: %s", chainID)
}

// GetChain returns the chain if it was registered, otherwise it returns an error.
func (nc *nodeConn) GetChain(chainID *isc.ChainID) (*ncChain, error) {
	nc.chainsLock.RLock()
	defer nc.chainsLock.RUnlock()

	ncc, exists := nc.chains[chainID.Key()]
	if !exists {
		return nil, xerrors.Errorf("Chain %v is not connected.", chainID.String())
	}

	return ncc, nil
}

// PublishStateTransaction implements chain.NodeConnection.
func (nc *nodeConn) PublishTransaction(chainID *isc.ChainID, tx *iotago.Transaction) error {
	ncc, err := nc.GetChain(chainID)
	if err != nil {
		return err
	}

	return ncc.PublishTransaction(tx, inxTimeoutPublishTransaction)
}

// AttachMilestones implements chain.NodeConnection.
func (nc *nodeConn) AttachMilestones(handler func(*nodebridge.Milestone)) *events.Closure {
	closure := events.NewClosure(handler)
	nc.nodeBridge.Events.LatestMilestoneChanged.Hook(closure)
	return closure
}

// DetachMilestones implements chain.NodeConnection.
func (nc *nodeConn) DetachMilestones(attachID *events.Closure) {
	nc.nodeBridge.Events.LatestMilestoneChanged.Detach(attachID)
}

func (nc *nodeConn) PullLatestOutput(chainID *isc.ChainID) {
	ncc := nc.chains[chainID.Key()]
	if ncc == nil {
		nc.log.Errorf("PullLatestOutput: NCChain not  found for chainID %s", chainID)
		return
	}
	ncc.queryLatestChainStateUTXO()
}

func (nc *nodeConn) PullStateOutputByID(chainID *isc.ChainID, id *iotago.UTXOInput) {
	ncc := nc.chains[chainID.Key()]
	if ncc == nil {
		nc.log.Errorf("PullOutputByID: NCChain not  found for chainID %s", chainID)
		return
	}
	ncc.PullStateOutputByID(id.ID())
}

func (nc *nodeConn) GetMetrics() nodeconnmetrics.NodeConnectionMetrics {
	return nc.metrics
}

func (nc *nodeConn) doPostTx(ctx context.Context, tx *iotago.Transaction) (iotago.BlockID, error) {
	// Build a Block and post it.
	block, err := builder.NewBlockBuilder().
		Payload(tx).
		Build()
	if err != nil {
		return iotago.EmptyBlockID(), xerrors.Errorf("failed to build a tx: %w", err)
	}

	blockID, err := nc.nodeBridge.SubmitBlock(ctx, block)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			// context was canceled
			return iotago.EmptyBlockID(), ctx.Err()
		}
		return iotago.EmptyBlockID(), xerrors.Errorf("failed to submit a tx: %w", err)
	}

	return blockID, nil
}

// addPendingTransaction tracks a pending transaction.
func (nc *nodeConn) addPendingTransaction(pending *PendingTransaction) {
	nc.pendingTransactionsLock.Lock()
	defer nc.pendingTransactionsLock.Unlock()

	nc.pendingTransactionsMap[pending.ID()] = pending
}

// clearPendingTransactionWithoutLocking removes tracking of a pending transaction.
// write lock must be acquired outside.
func (nc *nodeConn) clearPendingTransactionWithoutLocking(transactionID iotago.TransactionID) {
	delete(nc.pendingTransactionsMap, transactionID)
}

func (nc *nodeConn) reattachWorkerpoolFunc(task workerpool.Task) {
	defer task.Return(nil)

	pendingTx := task.Param(0).(*PendingTransaction)

	if pendingTx.Conflicting() || pendingTx.Confirmed() {
		// no need to reattach
		return
	}

	blockID := pendingTx.BlockID()
	if blockID == iotago.EmptyBlockID() {
		// no need to check because no block was posted by this node
		return
	}

	ctxMetadata, cancelCtxMetadata := context.WithTimeout(nc.ctx, inxTimeoutBlockMetadata)
	defer cancelCtxMetadata()

	blockMetadata, err := nc.nodeBridge.BlockMetadata(ctxMetadata, blockID)
	if err != nil {
		// block not found
		nc.log.Debugf("reattaching transaction %s failed, error: block not found", pendingTx.ID().ToHex(), blockID.ToHex())
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
		pendingTx.SetConflicting(xerrors.Errorf("tx was not included in the ledger. LedgerInclusionState: %s, ConflictReason: %d", blockMetadata.LedgerInclusionState, blockMetadata.ConflictReason))

		return
	}

	if blockMetadata.ShouldReattach {
		nc.log.Debugf("reattaching transaction %s", pendingTx.ID().ToHex())

		ctxSubmitBlock, cancelSubmitBlock := context.WithTimeout(nc.ctx, inxTimeoutSubmitBlock)
		defer cancelSubmitBlock()

		newBlockID, err := nc.doPostTx(ctxSubmitBlock, pendingTx.Transaction())
		if err != nil {
			nc.log.Debugf("reattaching transaction %s failed, error: %w", pendingTx.ID().ToHex(), err)
			return
		}

		// set the new blockID for promote/reattach checks
		pendingTx.SetBlockID(newBlockID)

		return
	}

	// reattach or promote if needed
	if blockMetadata.ShouldPromote {
		nc.log.Debugf("promoting transaction %s", pendingTx.ID().ToHex())

		ctxSubmitBlock, cancelSubmitBlock := context.WithTimeout(nc.ctx, inxTimeoutSubmitBlock)
		defer cancelSubmitBlock()

		if err := nc.promoteBlock(ctxSubmitBlock, blockID); err != nil {
			nc.log.Debugf("promoting transaction %s failed, error: %w", pendingTx.ID().ToHex(), err)
			return
		}
	}
}

func (nc *nodeConn) promoteBlock(ctx context.Context, blockID iotago.BlockID) error {
	tips, err := nc.nodeBridge.RequestTips(ctx, iotago.BlockMaxParents/2, false)
	if err != nil {
		return xerrors.Errorf("failed to fetch tips: %w", err)
	}

	// add the blockID we want to promote
	tips = append(tips, blockID)

	block, err := builder.NewBlockBuilder().Parents(tips).Build()
	if err != nil {
		return xerrors.Errorf("failed to build promotion block: %w", err)
	}

	if _, err = nc.nodeBridge.SubmitBlock(ctx, block); err != nil {
		return xerrors.Errorf("failed to submit promotion block: %w", err)
	}

	return nil
}
