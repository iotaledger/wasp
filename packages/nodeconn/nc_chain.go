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
	nc              *nodeConn
	chainAddr       iotago.Address
	msgs            map[hashing.HashValue]*ncTransaction
	outputHandler   func(iotago.OutputID, iotago.Output)
	inclusionStates *events.Event
	log             *logger.Logger
}

func newNCChain(nc *nodeConn, chainAddr iotago.Address, outputHandler func(iotago.OutputID, iotago.Output)) *ncChain {
	inclusionStates := events.NewEvent(func(handler interface{}, params ...interface{}) {
		handler.(chain.NodeConnectionInclusionStateHandlerFun)(params[0].(iotago.TransactionID), params[1].(string))
	})
	ncc := ncChain{
		nc:              nc,
		chainAddr:       chainAddr,
		msgs:            make(map[hashing.HashValue]*ncTransaction),
		outputHandler:   outputHandler,
		inclusionStates: inclusionStates,
		log:             nc.log.Named(chainAddr.String()[:6]),
	}
	go ncc.run()
	return &ncc
}

func (ncc *ncChain) Key() string {
	return ncc.chainAddr.Key()
}

func (ncc *ncChain) Close() {
	// Nothing. The ncc.nc.ctx is used for that.
}

func (ncc *ncChain) PublishTransaction(stateIndex uint32, tx *iotago.Transaction) error {
	txID, err := tx.ID()
	if err != nil {
		return xerrors.Errorf("failed to get a tx ID: %w", err)
	}
	txMsg, err := iotagob.NewMessageBuilder().Payload(tx).Build()
	if err != nil {
		return xerrors.Errorf("failed to build a tx message: %w", err)
	}
	txMsg, err = ncc.nc.nodeClient.SubmitMessage(ncc.nc.ctx, txMsg, iotago.ZeroRentParas) // TODO change
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
	msgMetaChanges, subInfo := ncc.nc.nodeEvents.MessageMetadataChange(*txMsgID)
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

	return nil
}

func (ncc *ncChain) run() {
	init := true
	for {
		if init {
			init = false
		} else {
			ncc.log.Infof("Retrying output subscription for chainAddr=%v", ncc.chainAddr.String())
			time.Sleep(500 * time.Millisecond) // Delay between retries.
		}

		//
		// Subscribe to the new outputs first.
		eventsCh, subInfo := ncc.nc.nodeEvents.OutputsByUnlockConditionAndAddress(
			ncc.chainAddr,
			iscp.NetworkPrefix,
			nodeclient.UnlockConditionAny,
		)
		if subInfo.Error() != nil {
			ncc.log.Panicf("failed to subscribe: %w", subInfo.Error())
		}
		//
		// Then fetch all the existing unspent outputs.
		indexer, err := ncc.nc.nodeClient.Indexer(ncc.nc.ctx)
		res, err := indexer.Outputs(ncc.nc.ctx, &nodeclient.OutputsQuery{
			AddressBech32: ncc.chainAddr.Bech32(iscp.NetworkPrefix),
		})
		if err != nil {
			ncc.log.Warnf("failed to query address outputs: %v", err)
			continue
		}
		for res.Next() {
			outs, err := res.Outputs()
			if err != nil {
				ncc.log.Warnf("failed to fetch address outputs: %v", err)
			}
			oids := res.Response.Items.MustOutputIDs()
			for i, out := range outs {
				oid := oids[i]
				ncc.outputHandler(oid, out)
			}
		}

		//
		// Then receive all the subscrived new outputs.
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
