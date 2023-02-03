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

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/pipe"
)

// ErrChainShutdown gets returned if the chain is shutting down.
var ErrChainShutdown = errors.New("chain is shutting down")

const (
	inxInitialStateRetries = 5
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
	requestOutputHandler func(iotago.MilestoneIndex, *isc.OutputInfo)
	aliasOutputHandler   func(iotago.MilestoneIndex, *isc.OutputInfo)
	milestoneHandler     func(iotago.MilestoneIndex, time.Time)

	pendingLedgerUpdatesLock sync.Mutex
	pendingLedgerUpdates     []*pendingLedgerUpdate
	appliedMilestoneIndex    iotago.MilestoneIndex
	synchronized             *atomic.Bool

	shutdownWaitGroup *sync.WaitGroup

	pendingTxTaskPipe  pipe.Pipe[*transactionTask]
	reattachTxTaskPipe pipe.Pipe[*transactionTask]
	lastPendingTxLock  sync.Mutex
	lastPendingTx      *pendingTransaction
}

func newNCChain(
	ctx context.Context,
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
		requestOutputHandler:     nil,
		aliasOutputHandler:       nil,
		milestoneHandler:         nil,
		pendingLedgerUpdatesLock: sync.Mutex{},
		pendingLedgerUpdates:     make([]*pendingLedgerUpdate, 0),
		appliedMilestoneIndex:    0,
		synchronized:             &atomic.Bool{},
		pendingTxTaskPipe:        pipe.NewInfinitePipe[*transactionTask](),
		reattachTxTaskPipe:       pipe.NewInfinitePipe[*transactionTask](),
		shutdownWaitGroup:        &sync.WaitGroup{},
		lastPendingTxLock:        sync.Mutex{},
		lastPendingTx:            nil,
	}

	chain.requestOutputHandler = func(milestoneIndex iotago.MilestoneIndex, outputInfo *isc.OutputInfo) {
		chain.LogDebugf("applying request output: outputID: %s, milestoneIndex: %d, chainID: %s", outputInfo.OutputID.ToHex(), milestoneIndex, chainID)
		requestOutputHandler(outputInfo)
	}

	chain.aliasOutputHandler = func(milestoneIndex iotago.MilestoneIndex, outputInfo *isc.OutputInfo) {
		chain.LogDebugf("applying alias output: outputID: %s, milestoneIndex: %d, chainID: %s", outputInfo.OutputID.ToHex(), milestoneIndex, chainID)
		aliasOutputHandler(outputInfo)
	}

	chain.milestoneHandler = func(milestoneIndex iotago.MilestoneIndex, milestoneTimestamp time.Time) {
		// we need to check if the milestones are applied in correct order
		chain.LogDebugf("applying milestone: milestoneIndex: %d, chainID: %s", milestoneIndex, chainID)
		chain.applyMilestoneIndex(milestoneIndex)
		milestoneHandler(milestoneTimestamp)
	}

	chain.shutdownWaitGroup.Add(1)
	go chain.postTxLoop(ctx)

	return chain
}

func (ncc *ncChain) WaitUntilStopped() {
	ncc.shutdownWaitGroup.Wait()
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
			ncc.requestOutputHandler(update.LedgerIndex, update.Update.(*isc.OutputInfo))
		case pendingLedgerUpdateTypeAlias:
			ncc.aliasOutputHandler(update.LedgerIndex, update.Update.(*isc.OutputInfo))
		case pendingLedgerUpdateTypeMilestone:
			ncc.milestoneHandler(update.LedgerIndex, update.Update.(time.Time))
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

	ncc.requestOutputHandler(ledgerIndex, outputInfo)
}

func (ncc *ncChain) HandleAliasOutput(ledgerIndex iotago.MilestoneIndex, outputInfo *isc.OutputInfo) {
	if added := ncc.addPendingLedgerUpdate(pendingLedgerUpdateTypeAlias, ledgerIndex, outputInfo); added {
		// ledger update was added as pending because the chain is not synchronized yet
		return
	}

	ncc.aliasOutputHandler(ledgerIndex, outputInfo)
}

func (ncc *ncChain) applyMilestoneIndex(milestoneIndex iotago.MilestoneIndex) {
	if ncc.appliedMilestoneIndex != 0 {
		if ncc.appliedMilestoneIndex+1 != milestoneIndex {
			ncc.LogPanicf("wrong milestone index applied (chainID: %s, expected: %d, got: %d)", ncc.chainID, ncc.appliedMilestoneIndex+1, milestoneIndex)
		}
	}
	ncc.appliedMilestoneIndex = milestoneIndex
}

func (ncc *ncChain) HandleMilestone(milestoneIndex iotago.MilestoneIndex, milestoneTimestamp time.Time) {
	if added := ncc.addPendingLedgerUpdate(pendingLedgerUpdateTypeMilestone, milestoneIndex, milestoneTimestamp); added {
		// ledger update was added as pending because the chain is not synchronized yet
		return
	}

	ncc.milestoneHandler(milestoneIndex, milestoneTimestamp)
}

// getLastPendingTx checks if there is a pending transaction for the chain.
// If the pending transaction was already removed or confirmed, it will be ignored.
func (ncc *ncChain) getLastPendingTx() *pendingTransaction {
	ncc.lastPendingTxLock.Lock()
	defer ncc.lastPendingTxLock.Unlock()

	if ncc.lastPendingTx != nil {
		ncc.nodeConn.pendingTransactionsLock.Lock()
		defer ncc.nodeConn.pendingTransactionsLock.Unlock()

		// check if the transaction is still pending, otherwise reset it
		if ncc.nodeConn.pendingTransactionsMap.Has(ncc.lastPendingTx.transactionID) {
			return ncc.lastPendingTx
		}

		ncc.lastPendingTx = nil
	}

	return nil
}

func (ncc *ncChain) setLastPendingTx(pendingTx *pendingTransaction) {
	ncc.lastPendingTxLock.Lock()
	defer ncc.lastPendingTxLock.Unlock()

	ncc.lastPendingTx = pendingTx
}

//nolint:funlen,gocyclo
func (ncc *ncChain) postTxLoop(ctx context.Context) {
	nodeConn := ncc.nodeConn

	cancelTask := func(task *transactionTask) {
		defer ncc.shutdownWaitGroup.Done()

		if task.back == nil {
			return
		}

		task.back <- &pendingTxResult{
			err: ErrChainShutdown,
		}
	}

	// we need to cancel all pending tasks to stop waiting on the results
	cancelTasks := func() {
		for task := range ncc.reattachTxTaskPipe.Out() {
			cancelTask(task)
		}

		for task := range ncc.pendingTxTaskPipe.Out() {
			cancelTask(task)
		}
	}

	defer func() {
		// cancel all outstanding tasks
		cancelTasks()

		// mark the chain as completely shut down
		ncc.shutdownWaitGroup.Done()
	}()

	checkTransactionIncludedAndSetConfirmed := func(pendingTx *pendingTransaction) bool {
		ctxWithTimeout, ctxCancel := context.WithTimeout(nodeConn.ctx, inxTimeoutBlockMetadata)
		defer ctxCancel()

		// check if the transaction was already included (race condition with other validators)
		if _, err := ncc.nodeConn.nodeClient.TransactionIncludedBlock(ctxWithTimeout, pendingTx.ID(), parameters.L1().Protocol); err == nil {
			// transaction was already included
			pendingTx.SetConfirmed()
			return true
		}

		return false
	}

	checkIsPending := func(pendingTx *pendingTransaction) bool {
		// check if the pending transaction is still being tracked.
		// if not, it must have been confirmed already or canceled.
		if !nodeConn.hasPendingTransaction(pendingTx.transactionID) {
			return false
		}

		// if it was not included, it is still pending
		return !checkTransactionIncludedAndSetConfirmed(pendingTx)
	}

	postTransaction := func(pendingTx *pendingTransaction) error {
		// check if the the transaction should be chained with another pending transaction
		chainedTxBlockIDs := iotago.BlockIDs{}
		if pendingTx.lastPendingTx != nil {
			// check if the chained pending transaction is still being tracked.
			// if yes, use the blockID to chain the transaction in the correct order.
			// if not, it must have been confirmed already or if it was canceled instead, also all chained tx would have been canceled already
			if checkIsPending(pendingTx.lastPendingTx) {
				chainedTxBlockIDs = append(chainedTxBlockIDs, pendingTx.lastPendingTx.BlockID())
			} else {
				// the chained pending transaction is not tracked anymore.
				pendingTx.lastPendingTx = nil
			}
		}

		// we link the ctxAttach to ctxConfirmed. this way the proof of work will be canceled if the transaction already got confirmed on L1.
		// (e.g. another validator finished PoW and tx was confirmed)
		// the given context will be canceled by the pending transaction checks.
		ctxAttach, cancelCtxAttach := context.WithTimeout(pendingTx.ctxConfirmed, inxTimeoutPublishTransaction)
		defer cancelCtxAttach()

		// post the transaction
		blockID, err := nodeConn.doPostTx(ctxAttach, pendingTx.transaction, chainedTxBlockIDs...)
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}

		// check if the context was canceled
		if err == nil {
			// set the current blockID for promote/reattach checks and "transaction chaining"
			pendingTx.SetBlockID(blockID)
		}

		// check if the transaction was already included (race condition with other validators)
		checkTransactionIncludedAndSetConfirmed(pendingTx)

		return nil
	}

	processTransactionTask := func(txTask *transactionTask, isReattachment bool) {
		defer ncc.shutdownWaitGroup.Done()

		pendingTx := txTask.pendingTx

		var txErr error

		// check if the transaction is still pending
		if checkIsPending(pendingTx) {
			// transaction is still pending
			for {
				txErr = postTransaction(pendingTx)
				if txErr != nil {
					action := "publishing"
					if isReattachment {
						action = "reattaching"
					}
					ncc.LogDebugf("%s transaction %s (chainID: %s) failed: %s", action, pendingTx.ID().ToHex(), ncc.chainID, txErr.Error())
					continue
				}

				// loop until it was successfully posted or canceled
				break
			}
		}

		if isReattachment {
			// reattach the chain
			// TODO: how do we cancel the whole chain in case of reattachment?
			txTask.pendingTx.PropagateReattach()
		}

		if txTask.back != nil {
			txTask.back <- &pendingTxResult{
				err: txErr,
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			// cancel all outstanding tasks
			cancelTasks()
			return

		case reattachTxTask := <-ncc.reattachTxTaskPipe.Out():
			processTransactionTask(reattachTxTask, true)

		case pendingTxTask := <-ncc.pendingTxTaskPipe.Out():
			// if both channels are ready while we enter the select case, golang picks a random channel first.
			// since reattachments have higher priority over pending tx, we need to process them first until none are left.
			for done := false; !done; {
				select {
				case <-ctx.Done():
					// done
					return
				case reattachTxTask := <-ncc.reattachTxTaskPipe.Out():
					processTransactionTask(reattachTxTask, true)
				default:
					done = true
				}
			}

			processTransactionTask(pendingTxTask, false)
		}
	}
}

type pendingTxResult struct {
	err error
}

type transactionTask struct {
	pendingTx *pendingTransaction
	back      chan *pendingTxResult
}

// createPendingTransaction creates and tracks a pending transaction that is linked
// to the former pending transaction in case the former was not confirmed yet.
// Linking to the former transaction is used to publish the transaction in
// the correct order for the whiteflag confirmation.
// The given context is used to cancel posting and tracking of the transaction.
func (ncc *ncChain) createPendingTransaction(ctx context.Context, tx *iotago.Transaction) (*pendingTransaction, error) {
	// As long as every validator references its own blocks of the pending transactions they posted,
	// the transactions will confirm eventually.
	// Of course there will be conflicts on L1, but we track confirmation based on transaction IDs,
	// so there will be one "winning subtangle/chain" that confirms the transactions for every validator.
	// If only parts of this chain of transactions are confirmed,
	// the validators will reference their own pending transactions in the next milestone cone.
	pendingTx, err := newPendingTransaction(ctx, ncc, tx, ncc.getLastPendingTx())
	if err != nil {
		return nil, fmt.Errorf("publishing transaction failed: %w", err)
	}
	ncc.setLastPendingTx(pendingTx)

	// track pending tx before publishing the transaction
	ncc.nodeConn.addPendingTransaction(pendingTx)

	return pendingTx, nil
}

func (ncc *ncChain) publishTX(pendingTx *pendingTransaction) error {
	ncc.LogDebugf("publishing transaction %s (chainID: %s)...", pendingTx.ID().ToHex(), ncc.chainID)

	ncc.shutdownWaitGroup.Add(1)
	back := make(chan *pendingTxResult)
	ncc.pendingTxTaskPipe.In() <- &transactionTask{
		pendingTx: pendingTx,
		back:      back,
	}

	pendingTxResult := <-back
	if err := pendingTxResult.err; err != nil && !errors.Is(err, context.Canceled) {
		// remove tracking of the pending transaction if posting the transaction failed
		ncc.nodeConn.clearPendingTransaction(pendingTx.transactionID)

		ncc.LogDebugf("publishing transaction %s (chainID: %s) failed: %s", pendingTx.ID().ToHex(), ncc.chainID, err.Error())
		return err
	}

	if err := pendingTx.WaitUntilConfirmed(); err != nil {
		ncc.LogDebugf("publishing transaction %s (chainID: %s) failed: %s", pendingTx.ID().ToHex(), ncc.chainID, err.Error())
		return err
	}

	return nil
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
	ledgerIndexAlias, aliasOutput, err := ncc.queryLatestChainStateAliasOutput(ctx)
	if err != nil {
		return 0, time.Time{}, nil, fmt.Errorf("failed to get latest chain state alias output: %w", err)
	}

	cmi := ncc.nodeConn.nodeBridge.ConfirmedMilestoneIndex()
	if cmi != ledgerIndexAlias {
		if cmi > ledgerIndexAlias {
			// confirmed milestone index is newer than the ledger index of the indexer
			return 0, time.Time{}, nil, fmt.Errorf("indexer ledger index does not match confirmed milestone index: (%d!=%d)", ledgerIndexAlias, cmi)
		}

		// cmi seems to be older than the ledger index of the indexer.
		// this can happen during startup of the node connection.
		// it is safe to query the node for the timestamp of the ledger index of the indexer instead.
	}

	// we need to get the timestamp of the milestone from the node
	milestoneTimestamp, err := ncc.nodeConn.getMilestoneTimestamp(ctx, ledgerIndexAlias)
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

				ncc.LogDebugf("failed to query initial chain state: %s, retrying...", err.Error())
				time.Sleep(50 * time.Millisecond)
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
	ncc.milestoneHandler(ledgerIndex, milestoneTimestamp)
	ncc.aliasOutputHandler(ledgerIndex, aliasOutput)

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
		ncc.requestOutputHandler(ledgerIndex, isc.NewOutputInfo(outputID, output, iotago.TransactionID{}))
	}

	if err := ncc.applyPendingLedgerUpdates(ledgerIndex); err != nil {
		return err
	}

	ncc.LogInfof("Synchronizing chain state and owned outputs for %s... done. (LedgerIndex: %d)", ncc.chainID, ledgerIndex)
	return nil
}
