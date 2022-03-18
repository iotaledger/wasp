// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconnchain

import (
	"sync"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

// nodeconn_chain is responsible for maintaining the information related to a single chain.
type nodeconnChain struct {
	nc        chain.NodeConnection
	chainAddr iotago.Address
	log       *logger.Logger

	aliasOutputIsHandled       bool
	aliasOutputCh              chan *iscp.AliasOutputWithID
	aliasOutputStopCh          chan bool
	onLedgerRequestIsHandled   bool
	onLedgerRequestCh          chan *iscp.OnLedgerRequestData
	onLedgerRequestStopCh      chan bool
	txInclusionStateIsHandled  bool
	txInclusionStateCh         chan *txInclusionStateMsg
	txInclusionStateStopCh     chan bool
	txInclusionStateHandlerRef *events.Closure
	milestonesHandlerRef       *events.Closure
	mutex                      sync.Mutex // NOTE: mutexes might also be separated for aliasOutput, onLedgerRequest and txInclusionState; however, it is not going to be used heavily, so the common one is used.
}

type txInclusionStateMsg struct {
	txID  iotago.TransactionID
	state string
}

var _ chain.ChainNodeConnection = &nodeconnChain{}

func NewChainNodeConnection(nc chain.NodeConnection, chainAddr iotago.Address, log *logger.Logger) (chain.ChainNodeConnection, error) {
	var err error
	result := nodeconnChain{
		nc:                     nc,
		chainAddr:              chainAddr,
		log:                    log.Named("ncc-" + chainAddr.String()[2:8]),
		aliasOutputCh:          make(chan *iscp.AliasOutputWithID),
		aliasOutputStopCh:      make(chan bool),
		onLedgerRequestCh:      make(chan *iscp.OnLedgerRequestData),
		onLedgerRequestStopCh:  make(chan bool),
		txInclusionStateCh:     make(chan *txInclusionStateMsg),
		txInclusionStateStopCh: make(chan bool),
	}
	result.nc.RegisterChain(result.chainAddr, result.outputHandler)
	result.txInclusionStateHandlerRef, err = result.nc.AttachTxInclusionStateEvents(result.chainAddr, result.txInclusionStateHandler)
	if err != nil {
		result.log.Errorf("cannot create chain nodeconnection: %v", err)
		return nil, err
	}
	result.log.Debugf("chain nodeconnection created")
	return &result, nil
}

func (nccT *nodeconnChain) outputHandler(outputID iotago.OutputID, output iotago.Output) {
	outputIDUTXO := outputID.UTXOInput()
	outputIDstring := iscp.OID(outputIDUTXO)
	nccT.log.Debugf("handling output ID %v", outputIDstring)
	aliasOutput, ok := output.(*iotago.AliasOutput)
	if ok {
		nccT.log.Debugf("handling output ID %v: writing alias output to channel", outputIDstring)
		nccT.aliasOutputCh <- iscp.NewAliasOutputWithID(aliasOutput, outputIDUTXO)
		nccT.log.Debugf("handling output ID %v: alias output handled", outputIDstring)
		return
	}
	onLedgerRequest, err := iscp.OnLedgerFromUTXO(output, outputIDUTXO)
	if err != nil {
		nccT.log.Warnf("handling output ID %v: unknown output type; ignoring it", outputIDstring)
		return
	}
	nccT.log.Debugf("handling output ID %v: writing on ledger request to channel", outputIDstring)
	nccT.onLedgerRequestCh <- onLedgerRequest
	nccT.log.Debugf("handling output ID %v: on ledger request handled", outputIDstring)
}

func (nccT *nodeconnChain) txInclusionStateHandler(txID iotago.TransactionID, state string) {
	txIDStr := iscp.TxID(&txID)
	nccT.log.Debugf("handling inclusion state of tx ID %v: %v", txIDStr, state)
	nccT.txInclusionStateCh <- &txInclusionStateMsg{
		txID:  txID,
		state: state,
	}
	nccT.log.Debugf("handling inclusion state of tx ID %v: inclusion state %v handled", txIDStr, state)
}

func (nccT *nodeconnChain) AttachToAliasOutput(handler chain.NodeConnectionAliasOutputHandlerFun) {
	nccT.mutex.Lock()
	defer nccT.mutex.Unlock()
	if nccT.aliasOutputIsHandled {
		nccT.log.Errorf("alias output handler already started!") // NOTE: this should not happen; maybe panic?
		return
	}
	nccT.aliasOutputIsHandled = true
	go func() {
		for {
			select {
			case aliasOutput := <-nccT.aliasOutputCh:
				handler(aliasOutput)
			case <-nccT.aliasOutputStopCh:
				nccT.log.Debugf("alias output handler stopped")
				return
			}
		}
	}()
	nccT.log.Debugf("alias output handler started")
}

func (nccT *nodeconnChain) DetachFromAliasOutput() {
	nccT.mutex.Lock()
	defer nccT.mutex.Unlock()
	if nccT.aliasOutputIsHandled {
		nccT.aliasOutputStopCh <- true
		nccT.aliasOutputIsHandled = false
	}
}

func (nccT *nodeconnChain) AttachToOnLedgerRequest(handler chain.NodeConnectionOnLedgerRequestHandlerFun) {
	nccT.mutex.Lock()
	defer nccT.mutex.Unlock()
	if nccT.onLedgerRequestIsHandled {
		nccT.log.Errorf("on ledger request handler already started!") // NOTE: this should not happen; maybe panic?
		return
	}
	nccT.onLedgerRequestIsHandled = true
	go func() {
		for {
			select {
			case onLedgerRequest := <-nccT.onLedgerRequestCh:
				handler(onLedgerRequest)
			case <-nccT.onLedgerRequestStopCh:
				nccT.log.Debugf("on ledger request handler stopped")
				return
			}
		}
	}()
	nccT.log.Debugf("on ledger request handler started")
}

func (nccT *nodeconnChain) DetachFromOnLedgerRequest() {
	nccT.mutex.Lock()
	defer nccT.mutex.Unlock()
	if nccT.onLedgerRequestIsHandled {
		nccT.onLedgerRequestStopCh <- true
		nccT.onLedgerRequestIsHandled = false
	}
}

func (nccT *nodeconnChain) AttachToTxInclusionState(handler chain.NodeConnectionInclusionStateHandlerFun) {
	nccT.mutex.Lock()
	defer nccT.mutex.Unlock()
	if nccT.txInclusionStateIsHandled {
		nccT.log.Errorf("transaction inclusion state handler already started!")
		return
	}
	nccT.txInclusionStateIsHandled = true
	go func() {
		for {
			select {
			case msg := <-nccT.txInclusionStateCh:
				handler(msg.txID, msg.state)
			case <-nccT.txInclusionStateStopCh:
				nccT.log.Debugf("transaction inclusion state handler stopped")
				return
			}
		}
	}()
	nccT.log.Debugf("transaction inclusion state handler started")
}

func (nccT *nodeconnChain) DetachFromTxInclusionState() {
	nccT.mutex.Lock()
	defer nccT.mutex.Unlock()
	if nccT.txInclusionStateIsHandled {
		nccT.txInclusionStateStopCh <- true
		nccT.txInclusionStateIsHandled = false
	}
}

func (nccT *nodeconnChain) AttachToMilestones(handler chain.NodeConnectionMilestonesHandlerFun) {
	nccT.mutex.Lock()
	defer nccT.mutex.Unlock()
	nccT.DetachFromMilestones()
	nccT.milestonesHandlerRef = nccT.nc.AttachMilestones(handler)
}

func (nccT *nodeconnChain) DetachFromMilestones() {
	nccT.mutex.Lock()
	defer nccT.mutex.Unlock()
	if nccT.milestonesHandlerRef != nil {
		nccT.nc.DetachMilestones(nccT.milestonesHandlerRef)
		nccT.milestonesHandlerRef = nil
	}
}

func (nccT *nodeconnChain) PublishTransaction(stateIndex uint32, tx *iotago.Transaction) error {
	return nccT.nc.PublishTransaction(nccT.chainAddr, stateIndex, tx)
}

func (nccT *nodeconnChain) PullLatestOutput() {
	nccT.nc.PullLatestOutput(nccT.chainAddr)
}

func (nccT *nodeconnChain) PullTxInclusionState(txid iotago.TransactionID) {
	nccT.nc.PullTxInclusionState(nccT.chainAddr, txid)
}

func (nccT *nodeconnChain) PullOutputByID(outputID *iotago.UTXOInput) {
	nccT.nc.PullOutputByID(nccT.chainAddr, outputID)
}

func (nccT *nodeconnChain) GetMetrics() nodeconnmetrics.NodeConnectionMessagesMetrics {
	// TODO
	return nil
}

func (nccT *nodeconnChain) Close() {
	// Nothing. The ncc.nc.ctx is used for that.
	nccT.DetachFromMilestones()
	nccT.nc.DetachTxInclusionStateEvents(nccT.chainAddr, nccT.txInclusionStateHandlerRef)
	nccT.nc.UnregisterChain(nccT.chainAddr)
	nccT.log.Debugf("chain nodeconnection closed")
}
