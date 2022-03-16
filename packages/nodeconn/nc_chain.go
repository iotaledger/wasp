// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	iotagob "github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	iotagox "github.com/iotaledger/iota.go/v3/x"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"golang.org/x/xerrors"
)

// nodeconn_chain is responsible for maintaining the information related to a single chain.
type ncChain struct {
	nc        *nodeConn
	chainAddr iotago.Address
	msgs      map[hashing.HashValue]*ncTransaction
	log       *logger.Logger

	aliasOutputHandler     chain.NodeConnectionAliasOutputHandlerFun
	onLedgerRequestHandler chain.NodeConnectionOnLedgerRequestHandlerFun
	inclusionStateHandler  chain.NodeConnectionInclusionStateHandlerFun
	milestonesHandlerRef   *events.Closure
	mutex                  sync.RWMutex
}

var _ chain.ChainNodeConnection = &ncChain{}

func newNCChain(nc *nodeConn, chainAddr iotago.Address) *ncChain {
	ncc := ncChain{
		nc:        nc,
		chainAddr: chainAddr,
		msgs:      make(map[hashing.HashValue]*ncTransaction),
		log:       nc.log.Named(chainAddr.String()[:6]),
	}
	ncc.aliasOutputHandler = ncc.defaultAliasOutputHandler
	ncc.onLedgerRequestHandler = ncc.defaultOnLedgerRequestHandler
	ncc.inclusionStateHandler = ncc.defaultInclusionStateHandle
	go ncc.run()
	ncc.log.Debugf("Chain nodeconnection created")
	return &ncc
}

func (ncc *ncChain) Key() string {
	return ncc.chainAddr.Key()
}

func (ncc *ncChain) Close() {
	// Nothing. The ncc.nc.ctx is used for that.
	ncc.DetachFromMilestones()
	ncc.log.Debugf("Chain nodeconnection closed")
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
	ncc.log.Infof("Posted TX Message: messageID=%v", txMsgID)

	//
	// TODO: Move it to `nc_transaction.go`
	msgMetaChanges := ncc.nc.nodeEvents.MessageMetadataChange(*txMsgID)
	go func() {
		for msgMetaChange := range msgMetaChanges {
			if msgMetaChange.LedgerInclusionState != nil {
				ncc.mutex.RLock()
				ncc.inclusionStateHandler(*txID, *msgMetaChange.LedgerInclusionState)
				ncc.mutex.RUnlock()
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

func (ncc *ncChain) outputHandler(outputID iotago.OutputID, output iotago.Output) {
	outputIDUTXO := outputID.UTXOInput()
	outputIDstring := iscp.OID(outputIDUTXO)
	ncc.log.Debugf("handling output ID %v", outputIDstring)
	aliasOutput, ok := output.(*iotago.AliasOutput)
	if ok {
		ncc.log.Debugf("handling output ID %v: calling alias output handler", outputIDstring)
		ncc.mutex.RLock()
		ncc.aliasOutputHandler(iscp.NewAliasOutputWithID(aliasOutput, outputIDUTXO))
		ncc.mutex.RUnlock()
		return
	}
	onLedgerRequest, err := iscp.OnLedgerFromUTXO(output, outputIDUTXO)
	if err != nil {
		ncc.log.Warnf("handling output ID %v: unknown output type; ignoring it", outputIDstring)
		return
	}
	ncc.log.Debugf("handling output ID %v: calling on ledger request handler", outputIDstring)
	ncc.mutex.RLock()
	ncc.onLedgerRequestHandler(onLedgerRequest)
	ncc.mutex.RUnlock()
}

func (ncc *ncChain) defaultAliasOutputHandler(output *iscp.AliasOutputWithID) {
	ncc.log.Debugf("default alias output handler: ignoring alias output with ID %v", iscp.OID(output.ID()))
}

func (ncc *ncChain) AttachToAliasOutput(handler chain.NodeConnectionAliasOutputHandlerFun) {
	ncc.mutex.Lock()
	defer ncc.mutex.Unlock()
	ncc.aliasOutputHandler = handler
}

func (ncc *ncChain) DetachFromAliasOutput() {
	ncc.mutex.Lock()
	defer ncc.mutex.Unlock()
	ncc.aliasOutputHandler = ncc.defaultAliasOutputHandler
}

func (ncc *ncChain) defaultOnLedgerRequestHandler(request *iscp.OnLedgerRequestData) {
	ncc.log.Debugf("default on ledger request handler: ignoring on ledger request with ID %s", request.ID())
}

func (ncc *ncChain) AttachToOnLedgerRequest(handler chain.NodeConnectionOnLedgerRequestHandlerFun) {
	ncc.mutex.Lock()
	defer ncc.mutex.Unlock()
	ncc.onLedgerRequestHandler = handler
}

func (ncc *ncChain) DetachFromOnLedgerRequest() {
	ncc.mutex.Lock()
	defer ncc.mutex.Unlock()
	ncc.onLedgerRequestHandler = ncc.defaultOnLedgerRequestHandler
}

func (ncc *ncChain) defaultInclusionStateHandle(txID iotago.TransactionID, state string) {
	ncc.log.Debugf("default on inclusion state handler: ignoring inclution state %s for transaction ID %v", state, iscp.TxID(&txID))
}

func (ncc *ncChain) AttachToTxInclusionState(handler chain.NodeConnectionInclusionStateHandlerFun) {
	ncc.mutex.Lock()
	defer ncc.mutex.Unlock()
	ncc.inclusionStateHandler = handler
}

func (ncc *ncChain) DetachFromTxInclusionState() {
	ncc.mutex.Lock()
	defer ncc.mutex.Unlock()
	ncc.inclusionStateHandler = ncc.defaultInclusionStateHandle
}

func (ncc *ncChain) AttachToMilestones(handler chain.NodeConnectionMilestonesHandlerFun) {
	ncc.DetachFromMilestones()
	ncc.milestonesHandlerRef = ncc.nc.AttachMilestones(handler)
}

func (ncc *ncChain) DetachFromMilestones() {
	if ncc.milestonesHandlerRef != nil {
		ncc.nc.DetachMilestones(ncc.milestonesHandlerRef)
		ncc.milestonesHandlerRef = nil
	}
}

func (ncc *ncChain) PullLatestOutput() {
	ncc.nc.PullLatestOutput(ncc.chainAddr)
}

func (ncc *ncChain) PullTxInclusionState(txid iotago.TransactionID) {
	ncc.nc.PullTxInclusionState(ncc.chainAddr, txid)
}

func (ncc *ncChain) PullOutputByID(outputID *iotago.UTXOInput) {
	ncc.nc.PullOutputByID(ncc.chainAddr, outputID)
}

func (ncc *ncChain) GetMetrics() nodeconnmetrics.NodeConnectionMessagesMetrics {
	// TODO
	return nil
}
