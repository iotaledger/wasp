// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"time"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	iotagob "github.com/iotaledger/iota.go/v3/builder"
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

func (ncc *ncChain) PublishTransaction(tx *iotago.Transaction) error {
	txID, err := tx.ID()
	if err != nil {
		return xerrors.Errorf("failed to get a tx ID: %w", err)
	}
	txMsg, err := iotagob.NewMessageBuilder().Payload(tx).Build()
	if err != nil {
		return xerrors.Errorf("failed to build a tx message: %w", err)
	}
	txMsg, err = ncc.nc.nodeAPIClient.SubmitMessage(ncc.nc.ctx, txMsg, ncc.nc.l1params.DeSerializationParameters)
	if err != nil {
		return xerrors.Errorf("failed to submit a tx message: %w", err)
	}
	txMsgID, err := txMsg.ID()
	if err != nil {
		return xerrors.Errorf("failed to extract a tx message ID: %w", err)
	}
	ncc.log.Infof("Posted TX Message: messageID=%v", txMsgID)

	//
	// TODO: Move it to `nc_transaction.go`
	msgMetaChanges, subInfo := ncc.nc.mqttClient.MessageMetadataChange(*txMsgID)
	if subInfo.Error() != nil {
		return xerrors.Errorf("failed to subscribe: %w", subInfo.Error())
	}
	go func() {
		for msgMetaChange := range msgMetaChanges {
			if msgMetaChange.LedgerInclusionState != nil {
				ncc.inclusionStates.Trigger(*txID, *msgMetaChange.LedgerInclusionState)
			}
		}
	}()

	return ncc.nc.waitUntilConfirmed(ncc.nc.ctx, txMsg)
}

func (ncc *ncChain) queryChainUTXOs() {
	bech32Addr := ncc.chainID.AsAddress().Bech32(ncc.nc.l1params.Bech32Prefix)
	queries := []nodeclient.IndexerQuery{
		&nodeclient.BasicOutputsQuery{AddressBech32: bech32Addr},
		&nodeclient.FoundriesQuery{AddressBech32: bech32Addr},
		&nodeclient.NFTsQuery{AddressBech32: bech32Addr},
		// &nodeclient.AliasesQuery{GovernorBech32: bech32Addr}, // TODO chains can't own alias outputs for now
	}
	for _, query := range queries {
		res, err := ncc.nc.indexerClient.Outputs(ncc.nc.ctx, query)
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
				ncc.outputHandler(oid, out)
			}
		}
	}
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
				ncc.outputHandler(iotago.OutputIDFromTransactionIDAndIndex(*tid, outResponse.OutputIndex), out)
			case <-ncc.nc.ctx.Done():
				return
			}
		}
	}
}

func (ncc *ncChain) subscribeToChainStateUpdates() {
	//
	// Subscribe to the new outputs first.
	// TODO

	//
	// Then fetch all the existing unspent outputs owned by the chain.
	stateOutputID, stateOutput, err := ncc.nc.indexerClient.Alias(ncc.nc.ctx, *ncc.chainID.AsAliasID())
	if err != nil {
		ncc.log.Panicf("error while fetching chain state output: %v", err)
	}
	ncc.stateOutputHandler(*stateOutputID, stateOutput)

	//
	// Then receive all the subscribed new outputs.
	// TODO
}

func (ncc *ncChain) run() {
	go ncc.subscribeToChainStateUpdates()
	go ncc.subscribeToChainOwnedUTXOs()
}
