// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"context"
	"time"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
)

// nodeconn_chain is responsible for maintaining the information related to a single chain.
type ncChain struct {
	nc                 *nodeConn
	chainID            *isc.ChainID
	outputHandler      func(iotago.OutputID, iotago.Output)
	stateOutputHandler func(iotago.OutputID, iotago.Output)
	inclusionStates    *events.Event
	log                *logger.Logger
}

func newNCChain(
	nc *nodeConn,
	chainID *isc.ChainID,
	stateOutputHandler,
	outputHandler func(iotago.OutputID, iotago.Output),
) *ncChain {
	inclusionStates := events.NewEvent(func(handler interface{}, params ...interface{}) {
		handler.(chain.NodeConnectionInclusionStateHandlerFun)(params[0].(iotago.TransactionID), params[1].(string))
	})

	ncc := ncChain{
		nc:                 nc,
		chainID:            chainID,
		outputHandler:      outputHandler,
		stateOutputHandler: stateOutputHandler,
		inclusionStates:    inclusionStates,
		log:                nc.log.Named(chainID.String()[:6]),
	}
	ncc.run()
	return &ncc
}

func (ncc *ncChain) Key() string {
	return ncc.chainID.Key()
}

func (ncc *ncChain) Close() {
	// Nothing. The ncc.nc.ctx is used for that.
}

func (ncc *ncChain) PublishTransaction(tx *iotago.Transaction, timeout ...time.Duration) error {
	ctxWithTimeout, cancelContext := newCtx(ncc.nc.ctx, timeout...)
	defer cancelContext()

	ctxPendingTransaction, cancelPendingTransaction := context.WithCancel(ctxWithTimeout)

	pendingTx, err := NewPendingTransaction(ctxPendingTransaction, cancelPendingTransaction, tx)
	if err != nil {
		return xerrors.Errorf("publishing transaction: %w", err)
	}

	// track pending tx before publishing the transaction
	ncc.nc.addPendingTransactionWithoutLocking(pendingTx)

	ncc.log.Debugf("publishing transaction %v...", isc.TxID(pendingTx.ID()))
	blockID, err := ncc.nc.doPostTx(ctxWithTimeout, tx)
	if err != nil {
		return err
	}

	// check if the transaction was already included (race condition with other validators)
	if _, err := ncc.nc.nodeClient.TransactionIncludedBlock(ctxPendingTransaction, pendingTx.ID(), ncc.nc.nodeBridge.ProtocolParameters()); err == nil {
		// transaction was already included
		pendingTx.Confirmed.Store(true)
		cancelPendingTransaction()
	} else {
		// set the current blockID for promote/reattach checks
		pendingTx.SetBlockID(blockID)
	}

	return pendingTx.waitUntilConfirmed()
}

func (ncc *ncChain) PullStateOutputByID(id iotago.OutputID) {
	ctxWithTimeout, cancelContext := newCtx(ncc.nc.ctx)
	res, err := ncc.nc.nodeClient.OutputByID(ctxWithTimeout, id)
	cancelContext()
	if err != nil {
		ncc.log.Errorf("PullOutputByID: error querying API - chainID %s OutputID %s:  %s", ncc.chainID, id, err)
		return
	}
	out, err := res.Output()
	if err != nil {
		ncc.log.Errorf("PullOutputByID: error getting output from response - chainID %s OutputID %s:  %s", ncc.chainID, id, err)
		return
	}
	ncc.stateOutputHandler(id, out)
}

func shouldBeProcessed(out iotago.Output) bool {
	// only outputs without SDRC should be processed.
	return !out.UnlockConditionSet().HasStorageDepositReturnCondition()
}

func (ncc *ncChain) queryChainUTXOs() {
	bech32Addr := ncc.chainID.AsAddress().Bech32(parameters.L1().Protocol.Bech32HRP)
	queries := []nodeclient.IndexerQuery{
		&nodeclient.BasicOutputsQuery{AddressBech32: bech32Addr},
		&nodeclient.FoundriesQuery{AliasAddressBech32: bech32Addr},
		&nodeclient.NFTsQuery{AddressBech32: bech32Addr},
		// &nodeclient.AliasesQuery{GovernorBech32: bech32Addr}, // TODO chains can't own alias outputs for now
	}

	var ctxWithTimeout context.Context
	var cancelContext context.CancelFunc
	for _, query := range queries {
		if ctxWithTimeout != nil && ctxWithTimeout.Err() == nil {
			// cancel the ctx of the last query
			cancelContext()
		}
		// TODO what should be an adequate timeout for each of these queries?
		ctxWithTimeout, cancelContext = newCtx(ncc.nc.ctx)

		res, err := ncc.nc.indexerClient.Outputs(ctxWithTimeout, query)
		if err != nil {
			ncc.log.Warnf("failed to query address outputs: %v", err)
			continue
		}

		for res.Next() {
			if res.Error != nil {
				ncc.log.Warnf("error iterating indexer results: %v", err)
			}
			outs, err := res.Outputs()
			if err != nil {
				ncc.log.Warnf("failed to fetch address outputs: %v", err)
				continue
			}
			oids, err := res.Response.Items.OutputIDs()
			if err != nil {
				ncc.log.Warnf("failed to get outputIDs from response items: %v", err)
				continue
			}
			for i, out := range outs {
				oid := oids[i]
				ncc.log.Debugf("received UTXO, outputID: %s", oid.ToHex())
				ncc.outputHandler(oid, out)
			}
		}
	}
	cancelContext()
}

func (ncc *ncChain) queryLatestChainStateUTXO() {
	// TODO what should be an adequate timeout for this query?
	ctxWithTimeout, cancelContext := newCtx(ncc.nc.ctx)
	stateOutputID, stateOutput, err := ncc.nc.indexerClient.Alias(ctxWithTimeout, *ncc.chainID.AsAliasID())
	cancelContext()

	if err != nil {
		ncc.log.Panicf("error while fetching chain state output: %v", err)
	}

	ncc.log.Debugf("received chain state update, outputID: %s", stateOutputID.ToHex())
	ncc.stateOutputHandler(*stateOutputID, stateOutput)
}

func (ncc *ncChain) HandleUnlockableOutput(outputID iotago.OutputID, output iotago.Output) {
	if shouldBeProcessed(output) {
		ncc.outputHandler(outputID, output)
	}
}

func (ncc *ncChain) HandleStateUpdate(outputID iotago.OutputID, output iotago.Output) {
	ncc.stateOutputHandler(outputID, output)
}

func (ncc *ncChain) run() {
	ncc.log.Infof("Subscribing to ledger updates")

	go ncc.queryLatestChainStateUTXO()
	go ncc.queryChainUTXOs()
}
