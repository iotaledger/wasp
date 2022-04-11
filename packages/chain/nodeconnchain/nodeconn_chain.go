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
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

// nodeconnChain is responsible for maintaining the information related to a single chain.
type nodeconnChain struct {
	nc      chain.NodeConnection
	chainID *iscp.ChainID
	log     *logger.Logger

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
	metrics                    nodeconnmetrics.NodeConnectionMessagesMetrics
	mutex                      sync.Mutex // NOTE: mutexes might also be separated for aliasOutput, onLedgerRequest and txInclusionState; however, it is not going to be used heavily, so the common one is used.
}

type txInclusionStateMsg struct {
	txID  iotago.TransactionID
	state string
}

var _ chain.ChainNodeConnection = &nodeconnChain{}

func NewChainNodeConnection(chainID *iscp.ChainID, nc chain.NodeConnection, log *logger.Logger) (chain.ChainNodeConnection, error) {
	var err error
	result := nodeconnChain{
		nc:                     nc,
		chainID:                chainID,
		log:                    log.Named("ncc-" + chainID.String()[2:8]),
		aliasOutputCh:          make(chan *iscp.AliasOutputWithID),
		aliasOutputStopCh:      make(chan bool),
		onLedgerRequestCh:      make(chan *iscp.OnLedgerRequestData),
		onLedgerRequestStopCh:  make(chan bool),
		txInclusionStateCh:     make(chan *txInclusionStateMsg),
		txInclusionStateStopCh: make(chan bool),
		metrics:                nc.GetMetrics().NewMessagesMetrics(chainID),
	}
	result.nc.RegisterChain(result.chainID, result.stateOutputHandler, result.outputHandler)
	result.txInclusionStateHandlerRef, err = result.nc.AttachTxInclusionStateEvents(result.chainID, result.txInclusionStateHandler)
	if err != nil {
		result.log.Errorf("cannot create chain nodeconnection: %v", err)
		return nil, err
	}
	result.log.Debugf("chain nodeconnection created")
	return &result, nil
}

func (nccT *nodeconnChain) L1Params() *parameters.L1 {
	return nccT.nc.L1Params()
}

func (nccT *nodeconnChain) stateOutputHandler(outputID iotago.OutputID, output iotago.Output) {
	nccT.metrics.GetInStateOutput().CountLastMessage(struct {
		OutputID iotago.OutputID
		Output   iotago.Output
	}{
		OutputID: outputID,
		Output:   output,
	})
	outputIDUTXO := outputID.UTXOInput()
	outputIDstring := iscp.OID(outputIDUTXO)
	nccT.log.Debugf("handling state output ID %v", outputIDstring)
	aliasOutput, ok := output.(*iotago.AliasOutput)
	if !ok {
		nccT.log.Panicf("unexpected output ID %v type %T received as state update to chain ID %s; alias output expected",
			outputIDstring, output, nccT.chainID)
	}
	if aliasOutput.AliasID.Empty() && aliasOutput.StateIndex != 0 {
		nccT.log.Panicf("unexpected output ID %v index %v with empty alias ID received as state update to chain ID %s; alias ID may be empty for initial alias output only",
			outputIDstring, aliasOutput.StateIndex, nccT.chainID)
	}
	if !util.AliasIDFromAliasOutput(aliasOutput, outputID).ToAddress().Equal(nccT.chainID.AsAddress()) {
		nccT.log.Panicf("unexpected output ID %v address %s index %v received as state update to chain ID %s, address %s",
			outputIDstring, aliasOutput.AliasID.ToAddress(), aliasOutput.StateIndex, nccT.chainID, nccT.chainID.AsAddress())
	}
	nccT.log.Debugf("handling state output ID %v: writing alias output to channel", outputIDstring)
	nccT.aliasOutputCh <- iscp.NewAliasOutputWithID(aliasOutput, outputIDUTXO)
	nccT.log.Debugf("handling state output ID %v: alias output handled", outputIDstring)
}

func (nccT *nodeconnChain) outputHandler(outputID iotago.OutputID, output iotago.Output) {
	nccT.metrics.GetInOutput().CountLastMessage(struct {
		OutputID iotago.OutputID
		Output   iotago.Output
	}{
		OutputID: outputID,
		Output:   output,
	})
	outputIDUTXO := outputID.UTXOInput()
	outputIDstring := iscp.OID(outputIDUTXO)
	nccT.log.Debugf("handling output ID %v", outputIDstring)
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
				nccT.metrics.GetInAliasOutput().CountLastMessage(aliasOutput)
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
				nccT.metrics.GetInOnLedgerRequest().CountLastMessage(onLedgerRequest)
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
				nccT.metrics.GetInTxInclusionState().CountLastMessage(msg)
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
	nccT.detachFromMilestones()
	nccT.milestonesHandlerRef = nccT.nc.AttachMilestones(handler)
}

func (nccT *nodeconnChain) DetachFromMilestones() {
	nccT.mutex.Lock()
	defer nccT.mutex.Unlock()
	nccT.detachFromMilestones()
}

func (nccT *nodeconnChain) detachFromMilestones() {
	if nccT.milestonesHandlerRef != nil {
		nccT.nc.DetachMilestones(nccT.milestonesHandlerRef)
		nccT.milestonesHandlerRef = nil
	}
}

func (nccT *nodeconnChain) PublishTransaction(stateIndex uint32, tx *iotago.Transaction) error {
	nccT.metrics.GetOutPublishTransaction().CountLastMessage(struct {
		StateIndex  uint32
		Transaction *iotago.Transaction
	}{
		StateIndex:  stateIndex,
		Transaction: tx,
	})
	return nccT.nc.PublishTransaction(nccT.chainID, stateIndex, tx)
}

func (nccT *nodeconnChain) PullLatestOutput() {
	nccT.metrics.GetOutPullLatestOutput().CountLastMessage(nil)
	nccT.nc.PullLatestOutput(nccT.chainID)
}

func (nccT *nodeconnChain) PullTxInclusionState(txID iotago.TransactionID) {
	nccT.metrics.GetOutPullTxInclusionState().CountLastMessage(txID)
	nccT.nc.PullTxInclusionState(nccT.chainID, txID)
}

func (nccT *nodeconnChain) PullOutputByID(outputID *iotago.UTXOInput) {
	nccT.metrics.GetOutPullOutputByID().CountLastMessage(outputID)
	nccT.nc.PullOutputByID(nccT.chainID, outputID)
}

func (nccT *nodeconnChain) GetMetrics() nodeconnmetrics.NodeConnectionMessagesMetrics {
	return nccT.metrics
}

func (nccT *nodeconnChain) Close() {
	nccT.DetachFromMilestones()
	_ = nccT.nc.DetachTxInclusionStateEvents(nccT.chainID, nccT.txInclusionStateHandlerRef)
	nccT.nc.UnregisterChain(nccT.chainID)
	nccT.log.Debugf("chain nodeconnection closed")
}
