// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/iotaledger/hive.go/core/contextutils"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
)

const (
	inxInitialStateRetries = 3
)

type pendingLedgerUpdateType int

const (
	pendingLedgerUpdateTypeRequest pendingLedgerUpdateType = iota
	pendingLedgerUpdateTypeAlias
	pendingLedgerUpdateTypeMilestone
)

type pendingLedgerUpdate struct {
	Type        pendingLedgerUpdateType
	LedgerIndex iotago.MilestoneIndex
	Update      any
}

func shouldBeProcessed(out iotago.Output) bool {
	// only outputs without SDRC should be processed.
	return !out.UnlockConditionSet().HasStorageDepositReturnCondition()
}

// ncChain is responsible for maintaining the information related to a single chain.
type ncChain struct {
	*logger.WrappedLogger

	nodeConn             *nodeConnection
	chainID              isc.ChainID
	requestOutputHandler chain.RequestOutputHandler
	aliasOutputHandler   chain.AliasOutputHandler
	milestoneHandler     chain.MilestoneHandler

	pendingLedgerUpdatesLock sync.Mutex
	pendingLedgerUpdates     []*pendingLedgerUpdate
	synchronized             *atomic.Bool
}

func newNCChain(
	nodeConn *nodeConnection,
	chainID isc.ChainID,
	requestOutputHandler chain.RequestOutputHandler,
	aliasOutputHandler chain.AliasOutputHandler,
	milestoneHandler chain.MilestoneHandler,
) *ncChain {
	chain := &ncChain{
		WrappedLogger:            logger.NewWrappedLogger(nodeConn.Logger()),
		nodeConn:                 nodeConn,
		chainID:                  chainID,
		requestOutputHandler:     requestOutputHandler,
		aliasOutputHandler:       aliasOutputHandler,
		milestoneHandler:         milestoneHandler,
		pendingLedgerUpdatesLock: sync.Mutex{},
		pendingLedgerUpdates:     make([]*pendingLedgerUpdate, 0),
		synchronized:             &atomic.Bool{},
	}

	return chain
}

func (ncc *ncChain) addPendingLedgerUpdate(updateType pendingLedgerUpdateType, ledgerIndex iotago.MilestoneIndex, update any) bool {
	if ncc.synchronized.Load() {
		// chain is already synchronized, ledger updates must be applied directly
		return false
	}

	ncc.pendingLedgerUpdatesLock.Lock()
	defer ncc.pendingLedgerUpdatesLock.Unlock()

	// after acquiring the pendingLedgerUpdatesLock, we need to check again if the chain was synchronized in the meantime.
	if ncc.synchronized.Load() {
		// chain is already synchronized, ledger updates must be applied directly
		return false
	}

	ncc.pendingLedgerUpdates = append(ncc.pendingLedgerUpdates, &pendingLedgerUpdate{
		Type:        updateType,
		LedgerIndex: ledgerIndex,
		Update:      update,
	})

	return true
}

// applyPendingLedgerUpdates applies all pending ledger updates to the chain.
// we assume the initial alias output and the owned outputs were already applied to the chain.
// the given ledgerIndex is the index the alias output was valid for.
// HINT: requests might be applied twice, if they are part of a pendingLedgerUpdate that overlaps with
// querying of the initial chain outputs.
func (ncc *ncChain) applyPendingLedgerUpdates(ledgerIndex iotago.MilestoneIndex) error {
	ncc.pendingLedgerUpdatesLock.Lock()
	defer ncc.pendingLedgerUpdatesLock.Unlock()

	for _, update := range ncc.pendingLedgerUpdates {
		if update.LedgerIndex <= ledgerIndex {
			// we can safely skip that pending ledger update, no information will be lost.
			continue
		}

		switch update.Type {
		case pendingLedgerUpdateTypeRequest:
			ncc.requestOutputHandler(update.Update.(*isc.OutputInfo))
		case pendingLedgerUpdateTypeAlias:
			ncc.aliasOutputHandler(update.Update.(*isc.OutputInfo))
		case pendingLedgerUpdateTypeMilestone:
			ncc.milestoneHandler(update.Update.(time.Time))
		default:
			panic("unknown pending ledger update type")
		}
	}

	// delete all pending ledger updates
	ncc.pendingLedgerUpdates = make([]*pendingLedgerUpdate, 0)

	// mark the chain as synchronized
	ncc.synchronized.Store(true)

	return nil
}

func (ncc *ncChain) HandleRequestOutput(ledgerIndex iotago.MilestoneIndex, outputInfo *isc.OutputInfo) {
	if !shouldBeProcessed(outputInfo.Output) {
		// only process outputs that match the filter criteria
		return
	}

	if added := ncc.addPendingLedgerUpdate(pendingLedgerUpdateTypeRequest, ledgerIndex, outputInfo); added {
		// ledger update was added as pending because the chain is not synchronized yet
		return
	}

	ncc.requestOutputHandler(outputInfo)
}

func (ncc *ncChain) HandleAliasOutput(ledgerIndex iotago.MilestoneIndex, outputInfo *isc.OutputInfo) {
	if added := ncc.addPendingLedgerUpdate(pendingLedgerUpdateTypeAlias, ledgerIndex, outputInfo); added {
		// ledger update was added as pending because the chain is not synchronized yet
		return
	}

	ncc.aliasOutputHandler(outputInfo)
}

func (ncc *ncChain) HandleMilestone(milestoneIndex iotago.MilestoneIndex, milestoneTimestamp time.Time) {
	if added := ncc.addPendingLedgerUpdate(pendingLedgerUpdateTypeMilestone, milestoneIndex, milestoneTimestamp); added {
		// ledger update was added as pending because the chain is not synchronized yet
		return
	}

	ncc.milestoneHandler(milestoneTimestamp)
}

func (ncc *ncChain) publishTX(ctx context.Context, tx *iotago.Transaction) error {
	mergedCtx, mergedCancel := contextutils.MergeContexts(ncc.nodeConn.ctx, ctx)
	defer mergedCancel()

	ctxWithTimeout, cancelContext := newCtxWithTimeout(mergedCtx, inxTimeoutPublishTransaction)
	defer cancelContext()

	pendingTx, err := newPendingTransaction(ctxWithTimeout, cancelContext, tx)
	if err != nil {
		return fmt.Errorf("publishing transaction failed: %w", err)
	}

	// track pending tx before publishing the transaction
	ncc.nodeConn.addPendingTransaction(pendingTx)

	ncc.LogDebugf("publishing transaction %v (chainID: %s)...", pendingTx.ID().ToHex(), ncc.chainID)

	// we use the context of the pending transaction to post the transaction. this way
	// the proof of work will be canceled if the transaction already got confirmed on L1.
	// (e.g. another validator finished PoW and tx was confirmed)
	// the given context will be canceled by the pending transaction checks.
	blockID, err := ncc.nodeConn.doPostTx(ctxWithTimeout, tx)
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	// check if the transaction was already included (race condition with other validators)
	if _, err := ncc.nodeConn.nodeClient.TransactionIncludedBlock(ctxWithTimeout, pendingTx.ID(), parameters.L1().Protocol); err == nil {
		// transaction was already included
		pendingTx.SetConfirmed()
	} else {
		// set the current blockID for promote/reattach checks
		pendingTx.SetBlockID(blockID)
	}

	return pendingTx.WaitUntilConfirmed()
}

func (ncc *ncChain) queryLatestChainStateAliasOutput(ctx context.Context) (iotago.MilestoneIndex, *isc.OutputInfo, error) {
	ctx, cancel := newCtxWithTimeout(ctx, inxTimeoutIndexerQuery)
	defer cancel()

	outputID, output, ledgerIndex, err := ncc.nodeConn.indexerClient.Alias(ctx, ncc.chainID.AsAliasID())
	if err != nil {
		return 0, nil, fmt.Errorf("error while fetching chain state output: %w", err)
	}

	ncc.LogDebugf("received chain state update, chainID: %s, outputID: %s", ncc.chainID, outputID.ToHex())

	return ledgerIndex, isc.NewOutputInfo(*outputID, output, iotago.TransactionID{}), nil
}

func (ncc *ncChain) queryChainOutputIDs(ctx context.Context) ([]iotago.OutputID, error) {
	bech32Addr := ncc.chainID.AsAddress().Bech32(parameters.L1().Protocol.Bech32HRP)

	falseCondition := false
	queries := []nodeclient.IndexerQuery{
		&nodeclient.BasicOutputsQuery{AddressBech32: bech32Addr, IndexerStorageDepositParas: nodeclient.IndexerStorageDepositParas{
			HasStorageDepositReturn: &falseCondition,
		}},
		&nodeclient.FoundriesQuery{AliasAddressBech32: bech32Addr},
		&nodeclient.NFTsQuery{AddressBech32: bech32Addr, IndexerStorageDepositParas: nodeclient.IndexerStorageDepositParas{
			HasStorageDepositReturn: &falseCondition,
		}},
		// &nodeclient.AliasesQuery{GovernorBech32: bech32Addr}, // TODO chains can't own alias outputs for now
	}

	// we cache the outputIDs for faster indexer queries, outputs are fetched afterwards
	outputIDs := make([]iotago.OutputID, 0)

	processChainUTXOQuery := func(query nodeclient.IndexerQuery) error {
		ctxQuery, cancelQuery := newCtxWithTimeout(ctx, inxTimeoutIndexerQuery)
		defer cancelQuery()

		res, err := ncc.nodeConn.indexerClient.Outputs(ctxQuery, query)
		if err != nil {
			return fmt.Errorf("failed to query address outputs: %w", err)
		}

		for res.Next() {
			if res.Error != nil {
				return fmt.Errorf("error iterating indexer results: %w", err)
			}

			respOutputIDs, err := res.Response.Items.OutputIDs()
			if err != nil {
				return fmt.Errorf("failed to get outputIDs from response items: %w", err)
			}

			outputIDs = append(outputIDs, respOutputIDs...)
		}

		return nil
	}

	for _, query := range queries {
		if err := processChainUTXOQuery(query); err != nil {
			return nil, err
		}
	}

	return outputIDs, nil
}

func (ncc *ncChain) queryChainState(ctx context.Context) (iotago.MilestoneIndex, time.Time, *isc.OutputInfo, error) {
	cmi := ncc.nodeConn.nodeBridge.ConfirmedMilestoneIndex()

	ledgerIndexAlias, aliasOutput, err := ncc.queryLatestChainStateAliasOutput(ctx)
	if err != nil {
		return 0, time.Time{}, nil, fmt.Errorf("failed to get latest chain state alias output: %w", err)
	}

	if cmi != ledgerIndexAlias {
		return 0, time.Time{}, nil, fmt.Errorf("indexer ledger index does not match confirmed milestone index: (%d!=%d)", ledgerIndexAlias, cmi)
	}

	// we need to get the timestamp of the milestone from the node
	milestoneTimestamp, err := ncc.nodeConn.getMilestoneTimestamp(ctx, cmi)
	if err != nil {
		return 0, time.Time{}, nil, fmt.Errorf("failed to get milestone timestamp: %w", err)
	}

	return ledgerIndexAlias, milestoneTimestamp, aliasOutput, nil
}

// SyncChainStateWithL1 synchronizes the chain state by applying the current confirmed milestone and alias output of the chain.
// It takes care that milestone index of the current confirmed milestone and the ledger index of the alias output match.
// Afterwards all owned outputs of the chain are applied.
// The chain is marked as synchronized after all pending ledger updates were applied,
// which could have beed added in parallel by handleLedgerUpdate.
func (ncc *ncChain) SyncChainStateWithL1(ctx context.Context) error {
	ncc.LogInfof("Synchronizing chain state and owned outputs for %s...", ncc.chainID)

	queryChainStateLoop := func() (iotago.MilestoneIndex, time.Time, *isc.OutputInfo, error) {
		// there is a potential race condition if the ledger index on L1 changes during querying of the initial chain state
		// => we need to retry in that case
		for i := 0; i < inxInitialStateRetries; i++ {
			ledgerIndex, milestoneTimestamp, aliasOutput, err := ncc.queryChainState(ctx)
			if err != nil {
				if i == inxInitialStateRetries-1 {
					// last try, return the error
					return 0, time.Time{}, nil, fmt.Errorf("failed to query initial chain state: %w", err)
				}

				ncc.LogDebugf("failed to query initial chain state: %w, retrying...", err.Error())
				continue
			}
			return ledgerIndex, milestoneTimestamp, aliasOutput, nil
		}

		return 0, time.Time{}, nil, errors.New("failed to query initial chain state")
	}

	ledgerIndex, milestoneTimestamp, aliasOutput, err := queryChainStateLoop()
	if err != nil {
		return err
	}

	// we can safely forward the state to the chain.
	// ledger updates won't be applied in parallel as long as synchronized is not set to true.
	ncc.milestoneHandler(milestoneTimestamp)
	ncc.aliasOutputHandler(aliasOutput)

	// the indexer returns the outputs in sorted order by timestampBooked,
	// so we don't miss newly added outputs if the ledgerIndex increases during the query.
	// HINT: requests might be applied twice, if they are part of a pendingLedgerUpdate that overlaps with
	// querying of the initial chain outputs.
	outputIDs, err := ncc.queryChainOutputIDs(ctx)
	if err != nil {
		return err
	}

	// fetch and apply all owned outputs
	for _, outputID := range outputIDs {
		output, err := ncc.nodeConn.outputForOutputID(ctx, outputID)
		if err != nil {
			return fmt.Errorf("failed to fetch output (%s): %w", outputID.ToHex(), err)
		}

		ncc.LogDebugf("received output, chainID: %s, outputID: %s", ncc.chainID, outputID.ToHex())
		ncc.requestOutputHandler(isc.NewOutputInfo(outputID, output, iotago.TransactionID{}))
	}

	if err := ncc.applyPendingLedgerUpdates(ledgerIndex); err != nil {
		return err
	}

	ncc.LogInfof("Synchronizing chain state and owned outputs for %s... done. (LedgerIndex: %d)", ncc.chainID, ledgerIndex)
	return nil
}
