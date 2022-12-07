// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"context"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/core/contextutils"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
)

func shouldBeProcessed(out iotago.Output) bool {
	// only outputs without SDRC should be processed.
	return !out.UnlockConditionSet().HasStorageDepositReturnCondition()
}

// ncChain is responsible for maintaining the information related to a single chain.
type ncChain struct {
	*logger.WrappedLogger

	nodeConn             *nodeConnection
	chainID              *isc.ChainID
	requestOutputHandler chain.RequestOutputHandler
	aliasOutputHandler   chain.AliasOutputHandler
	milestoneHandler     chain.MilestoneHandler
}

func newNCChain(
	nodeConn *nodeConnection,
	chainID *isc.ChainID,
	requestOutputHandler chain.RequestOutputHandler,
	aliasOutputHandler chain.AliasOutputHandler,
	milestoneHandler chain.MilestoneHandler,
) (*ncChain, error) {
	chain := &ncChain{
		WrappedLogger: logger.NewWrappedLogger(nodeConn.WrappedLogger.LoggerNamed(chainID.String()[:6])),
		nodeConn:      nodeConn,
		chainID:       chainID,
		requestOutputHandler: func(outputInfo *isc.OutputInfo) {
			// only process outputs that match the filter criteria
			if shouldBeProcessed(outputInfo.Output) {
				requestOutputHandler(outputInfo)
			}
		},
		aliasOutputHandler: aliasOutputHandler,
		milestoneHandler:   milestoneHandler,
	}
	if err := chain.queryInititalState(); err != nil {
		return nil, err
	}

	return chain, nil
}

func (ncc *ncChain) queryInititalState() error {
	ncc.LogInfo("Querying initial state and owned outputs...")

	// TODO: there is a potential race condition if the milestone index
	// on L1 changes during querying of the initial chain state
	cmi := ncc.nodeConn.nodeBridge.ConfirmedMilestoneIndex()

	// we need to get the timestamp of the milestone from the node
	milestoneTimestamp, err := ncc.nodeConn.getMilestoneTimestamp(cmi)
	if err != nil {
		return err
	}

	ncc.milestoneHandler(milestoneTimestamp)

	if err := ncc.queryLatestChainStateUTXO(); err != nil {
		return err
	}
	if err := ncc.queryChainUTXOs(); err != nil {
		return err
	}

	ncc.LogInfof("Querying initial state and owned outputs... done. (MilestoneIndex: %d", cmi)
	return nil
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

	ncc.LogDebugf("publishing transaction %v...", pendingTx.ID().ToHex())

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

func (ncc *ncChain) queryLatestChainStateUTXO() error {
	ctx, cancel := newCtxWithTimeout(ncc.nodeConn.ctx, inxTimeoutIndexerOutputs)
	defer cancel()

	outputID, output, err := ncc.nodeConn.indexerClient.Alias(ctx, *ncc.chainID.AsAliasID())
	if err != nil {
		return fmt.Errorf("error while fetching chain state output: %w", err)
	}

	ncc.LogDebugf("received chain state update, outputID: %s", outputID.ToHex())
	ncc.aliasOutputHandler(isc.NewOutputInfo(*outputID, output, iotago.TransactionID{}))

	return nil
}

func (ncc *ncChain) queryChainUTXOs() error {
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

	processChainUTXOQuery := func(query nodeclient.IndexerQuery) error {
		ctx, cancel := newCtxWithTimeout(ncc.nodeConn.ctx, inxTimeoutIndexerOutputs)
		defer cancel()

		res, err := ncc.nodeConn.indexerClient.Outputs(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to query address outputs: %w", err)
		}

		for res.Next() {
			if res.Error != nil {
				return fmt.Errorf("error iterating indexer results: %w", err)
			}

			outputs, err := res.Outputs()
			if err != nil {
				return fmt.Errorf("failed to fetch address outputs: %w", err)
			}

			outputIDs, err := res.Response.Items.OutputIDs()
			if err != nil {
				return fmt.Errorf("failed to get outputIDs from response items: %w", err)
			}

			for i := range outputs {
				outputID := outputIDs[i]
				ncc.LogDebugf("received UTXO, outputID: %s", outputID.ToHex())
				ncc.requestOutputHandler(isc.NewOutputInfo(outputID, outputs[i], iotago.TransactionID{}))
			}
		}

		return nil
	}

	for _, query := range queries {
		if err := processChainUTXOQuery(query); err != nil {
			return err
		}
	}

	return nil
}
