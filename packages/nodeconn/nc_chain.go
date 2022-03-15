// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"time"

	"github.com/iotaledger/hive.go/events"
	iotago "github.com/iotaledger/iota.go/v3"
	iotagob "github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	iotagox "github.com/iotaledger/iota.go/v3/x"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

// nodeconn_chain is responsible for maintaining the information related to a single chain.
type ncChain struct {
	nc              *nodeConn
	chainAddr       iotago.Address
	msgs            map[hashing.HashValue]*ncTransaction
	outputHandler   func(iotago.OutputID, *iotago.Output)
	inclusionStates *events.Event
}

func newNCChain(nc *nodeConn, chainAddr iotago.Address, outputHandler func(iotago.OutputID, *iotago.Output)) *ncChain {
	inclusionStates := events.NewEvent(func(handler interface{}, params ...interface{}) {
		handler.(func(iotago.TransactionID, string))(params[0].(iotago.TransactionID), params[1].(string))
	})
	ncc := ncChain{
		nc:              nc,
		chainAddr:       chainAddr,
		msgs:            make(map[hashing.HashValue]*ncTransaction),
		outputHandler:   outputHandler,
		inclusionStates: inclusionStates,
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
	txMsg, err = ncc.nc.nodeClient.SubmitMessage(ncc.nc.ctx, txMsg)
	if err != nil {
		return xerrors.Errorf("failed to submit a tx message: %w", err)
	}
	txMsgID, err := txMsg.ID()
	if err != nil {
		return xerrors.Errorf("failed to extract a tx message ID: %w", err)
	}
	ncc.nc.log.Infof("Posted TX Message: messageID=%v", txMsgID)

	//
	// TODO: Move it to `nc_transaction.go`
	msgMetaChanges := ncc.nc.nodeEvents.MessageMetadataChange(*txMsgID)
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
			ncc.nc.log.Infof("Retrying output subscription for chainAddr=%v", ncc.chainAddr.String())
			time.Sleep(500 * time.Millisecond) // Delay between retries.
		}

		//
		// Subscribe to the new outputs first.
		eventsCh := ncc.nc.nodeEvents.OutputsByUnlockConditionAndAddress(
			ncc.chainAddr,
			iscp.NetworkPrefix,
			iotagox.UnlockConditionAny,
		)

		//
		// Then fetch all the existing unspent outputs.
		res, err := ncc.nc.nodeClient.Indexer().Outputs(ncc.nc.ctx, &nodeclient.OutputsQuery{
			AddressBech32: ncc.chainAddr.Bech32(iscp.NetworkPrefix),
		})
		if err != nil {
			ncc.nc.log.Warnf("failed to query address outputs: %v", err)
			continue
		}
		for res.Next() {
			outs, err := res.Outputs()
			if err != nil {
				ncc.nc.log.Warnf("failed to fetch address outputs: %v", err)
			}
			oids := res.Response.Items.MustOutputIDs()
			for i, o := range outs {
				out := o
				oid := oids[i]
				ncc.outputHandler(oid, &out)
			}
		}

		//
		// Then receive all the subscrived new outputs.
		for {
			select {
			case outResponse := <-eventsCh:
				out, err := outResponse.Output()
				if err != nil {
					ncc.nc.log.Warnf("error while receiving unspent output: %v", err)
					continue
				}
				tid, err := outResponse.TxID()
				if err != nil {
					ncc.nc.log.Warnf("error while receiving unspent output tx id: %v", err)
					continue
				}
				ncc.outputHandler(iotago.OutputIDFromTransactionIDAndIndex(*tid, outResponse.OutputIndex), &out)
			case <-ncc.nc.ctx.Done():
				return
			}
		}
	}
}
