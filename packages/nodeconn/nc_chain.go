// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"context"
	"encoding/json"
	"time"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

// nodeconn_chain is responsible for maintaining the information related to a single chain.
type ncChain struct {
	nc                 *nodeConn
	chainID            *iscp.ChainID
	msgs               map[hashing.HashValue]*ncTransaction
	outputHandler      func(iotago.OutputID, iotago.Output)
	stateOutputHandler func(iotago.OutputID, iotago.Output)
	inclusionStates    *events.Event
	log                *logger.Logger
}

func newNCChain(
	nc *nodeConn,
	chainID *iscp.ChainID,
	stateOutputHandler,
	outputHandler func(iotago.OutputID, iotago.Output),
) *ncChain {
	inclusionStates := events.NewEvent(func(handler interface{}, params ...interface{}) {
		handler.(chain.NodeConnectionInclusionStateHandlerFun)(params[0].(iotago.TransactionID), params[1].(string))
	})
	ncc := ncChain{
		nc:                 nc,
		chainID:            chainID,
		msgs:               make(map[hashing.HashValue]*ncTransaction),
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

	txMsg, err := ncc.nc.doPostTx(ctxWithTimeout, tx)
	if err != nil {
		return err
	}
	txID, err := tx.ID()
	if err != nil {
		return xerrors.Errorf("failed to get a tx ID: %w", err)
	}
	txMsgID, err := txMsg.ID()
	if err != nil {
		return xerrors.Errorf("failed to extract a tx message ID: %w", err)
	}

	//
	// TODO: Move it to `nc_transaction.go`
	msgMetaChanges, subInfo := ncc.nc.mqttClient.MessageMetadataChange(*txMsgID)
	if subInfo.Error() != nil {
		return xerrors.Errorf("failed to subscribe: %w", subInfo.Error())
	}
	go func() {
		for msgMetaChange := range msgMetaChanges {
			if msgMetaChange.LedgerInclusionState != nil {
				str, err := json.Marshal(msgMetaChange)
				if err != nil {
					ncc.nc.log.Errorf("Unexpected error trying to marshal msgMetadataChange: %s", err)
				} else {
					ncc.nc.log.Debugf("msgMetadataChange: %s", str)
				}
				ncc.inclusionStates.Trigger(*txID, *msgMetaChange.LedgerInclusionState)
			}
		}
	}()

	// TODO should promote/re-attach logic not be blocking?
	return ncc.nc.waitUntilConfirmed(ctxWithTimeout, txMsg)
}

func (ncc *ncChain) queryChainUTXOs() {
	bech32Addr := ncc.chainID.AsAddress().Bech32(ncc.nc.l1params.Bech32Prefix)
	queries := []nodeclient.IndexerQuery{
		&nodeclient.BasicOutputsQuery{AddressBech32: bech32Addr},
		&nodeclient.FoundriesQuery{AddressBech32: bech32Addr},
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

func (ncc *ncChain) subscribeToChainOwnedUTXOs() {
	init := true
	for {
		if init {
			init = false
		} else {
			ncc.log.Infof("Retrying output subscription for chainAddr=%v", ncc.chainID.String())
			time.Sleep(500 * time.Millisecond) // Delay between retries.
		}

		//
		// Subscribe to the new outputs first.
		eventsCh, subInfo := ncc.nc.mqttClient.OutputsByUnlockConditionAndAddress(
			ncc.chainID.AsAddress(),
			ncc.nc.l1params.Bech32Prefix,
			nodeclient.UnlockConditionAny,
		)
		if subInfo.Error() != nil {
			ncc.log.Panicf("failed to subscribe: %w", subInfo.Error())
		}
		//
		// Then fetch all the existing unspent outputs owned by the chain.
		ncc.queryChainUTXOs()

		//
		// Then receive all the subscribed new outputs.
		for {
			select {
			case outResponse := <-eventsCh:
				out, err := outResponse.Output()
				if err != nil {
					ncc.log.Warnf("error while receiving unspent output: %v", err)
					continue
				}
				tid, err := outResponse.TxID()
				if err != nil {
					ncc.log.Warnf("error while receiving unspent output tx id: %v", err)
					continue
				}
				outID := iotago.OutputIDFromTransactionIDAndIndex(*tid, outResponse.OutputIndex)
				ncc.log.Debugf("received UTXO, outputID: %s", outID.ToHex())
				ncc.outputHandler(outID, out)
			case <-ncc.nc.ctx.Done():
				return
			}
		}
	}
}

func (ncc *ncChain) subscribeToChainStateUpdates() {
	//
	// Subscribe to the new outputs first.
	eventsCh, subInfo := ncc.nc.mqttClient.AliasOutputsByID(*ncc.chainID.AsAliasID())
	if subInfo.Error() != nil {
		ncc.log.Panicf("failed to subscribe: %w", subInfo.Error())
	}

	//
	// Then fetch all the existing unspent outputs owned by the chain.
	// TODO what should be an adequate timeout for this query?
	ctxWithTimeout, cancelContext := newCtx(ncc.nc.ctx)
	stateOutputID, stateOutput, err := ncc.nc.indexerClient.Alias(ctxWithTimeout, *ncc.chainID.AsAliasID())
	cancelContext()
	if err != nil {
		ncc.log.Panicf("error while fetching chain state output: %v", err)
	}
	ncc.log.Debugf("received chain state update, outputID: %s", stateOutputID.ToHex())
	ncc.stateOutputHandler(*stateOutputID, stateOutput)

	//
	// Then receive all the subscribed new outputs.
	for {
		select {
		case outResponse := <-eventsCh:
			out, err := outResponse.Output()
			if err != nil {
				ncc.log.Warnf("error while receiving chain state unspent output: %v", err)
				continue
			}
			tid, err := outResponse.TxID()
			if err != nil {
				ncc.log.Warnf("error while receiving chain state unspent output tx id: %v", err)
				continue
			}
			outID := iotago.OutputIDFromTransactionIDAndIndex(*tid, outResponse.OutputIndex)
			ncc.log.Debugf("received chain state update, outputID: %s", outID.ToHex())
			ncc.stateOutputHandler(outID, out)
		case <-ncc.nc.ctx.Done():
			return
		}
	}
}

func (ncc *ncChain) run() {
	go ncc.subscribeToChainStateUpdates()
	go ncc.subscribeToChainOwnedUTXOs()
}
