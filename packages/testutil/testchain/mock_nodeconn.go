package testchain

import (
	"fmt"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/parameters"
)

type MockedNodeConn struct {
	log                            *logger.Logger
	ledgers                        *MockedLedgers
	id                             string
	publishTransactionAllowedFun   func(chainID *iscp.ChainID, stateIndex uint32, tx *iotago.Transaction) bool
	pullLatestOutputAllowed        bool
	pullTxInclusionStateAllowedFun func(chainID *iscp.ChainID, txID iotago.TransactionID) bool
	pullOutputByIDAllowedFun       func(chainID *iscp.ChainID, outputID *iotago.UTXOInput) bool
	stopChannel                    chan bool
}

var _ chain.NodeConnection = &MockedNodeConn{}

func NewMockedNodeConnection(id string, ledgers *MockedLedgers, log *logger.Logger) *MockedNodeConn {
	result := &MockedNodeConn{
		log:         log.Named("mnc"),
		id:          id,
		ledgers:     ledgers,
		stopChannel: make(chan bool),
	}
	result.SetPublishTransactionAllowed(true)
	result.SetPullLatestOutputAllowed(true)
	result.SetPullTxInclusionStateAllowed(true)
	result.SetPullOutputByIDAllowed(true)
	result.log.Debugf("Nodeconn created")
	return result
}

func (mncT *MockedNodeConn) ID() string {
	return mncT.id
}

func (mncT *MockedNodeConn) RegisterChain(chainID *iscp.ChainID, stateOutputHandler, outputHandler func(iotago.OutputID, iotago.Output)) {
	mncT.ledgers.GetLedger(chainID).Register(mncT.id, stateOutputHandler, outputHandler)
}

func (mncT *MockedNodeConn) UnregisterChain(chainID *iscp.ChainID) {
	mncT.ledgers.GetLedger(chainID).Unregister(mncT.id)
}

func (mncT *MockedNodeConn) PublishTransaction(chainID *iscp.ChainID, stateIndex uint32, tx *iotago.Transaction) error {
	if mncT.publishTransactionAllowedFun(chainID, stateIndex, tx) {
		return mncT.ledgers.GetLedger(chainID).PublishTransaction(stateIndex, tx)
	}
	return fmt.Errorf("Publishing transaction for address %s of index %v is not allowed", chainID, stateIndex)
}

func (mncT *MockedNodeConn) PullLatestOutput(chainID *iscp.ChainID) {
	if mncT.pullLatestOutputAllowed {
		mncT.ledgers.GetLedger(chainID).PullLatestOutput(mncT.id)
	} else {
		mncT.log.Errorf("Pull latest output for address %s is not allowed", chainID)
	}
}

func (mncT *MockedNodeConn) PullTxInclusionState(chainID *iscp.ChainID, txid iotago.TransactionID) {
	if mncT.pullTxInclusionStateAllowedFun(chainID, txid) {
		mncT.ledgers.GetLedger(chainID).PullTxInclusionState(mncT.id, txid)
	} else {
		mncT.log.Errorf("Pull transaction inclusion state for address %s txID %v is not allowed", chainID, iscp.TxID(&txid))
	}
}

func (mncT *MockedNodeConn) PullStateOutputByID(chainID *iscp.ChainID, id *iotago.UTXOInput) {
	if mncT.pullOutputByIDAllowedFun(chainID, id) {
		mncT.ledgers.GetLedger(chainID).PullStateOutputByID(mncT.id, id)
	} else {
		mncT.log.Errorf("Pull output by ID for address %s ID %v is not allowed", chainID, iscp.OID(id))
	}
}

func (mncT *MockedNodeConn) AttachTxInclusionStateEvents(chainID *iscp.ChainID, handler chain.NodeConnectionInclusionStateHandlerFun) (*events.Closure, error) {
	return mncT.ledgers.GetLedger(chainID).AttachTxInclusionStateEvents(mncT.id, handler)
}

func (mncT *MockedNodeConn) DetachTxInclusionStateEvents(chainID *iscp.ChainID, closure *events.Closure) error {
	return mncT.ledgers.GetLedger(chainID).DetachTxInclusionStateEvents(mncT.id, closure)
}

func (mncT *MockedNodeConn) AttachMilestones(handler chain.NodeConnectionMilestonesHandlerFun) *events.Closure {
	return mncT.ledgers.AttachMilestones(handler)
}

func (mncT *MockedNodeConn) DetachMilestones(attachID *events.Closure) {
	mncT.ledgers.DetachMilestones(attachID)
}

func (mncT *MockedNodeConn) GetMetrics() nodeconnmetrics.NodeConnectionMetrics {
	return nodeconnmetrics.NewEmptyNodeConnectionMetrics()
}

func (mncT *MockedNodeConn) Close() {
	// TODO
}

func (mncT *MockedNodeConn) SetPublishTransactionAllowed(flag bool) {
	mncT.SetPublishTransactionAllowedFun(func(*iscp.ChainID, uint32, *iotago.Transaction) bool { return flag })
}

func (mncT *MockedNodeConn) SetPublishTransactionAllowedFun(fun func(chainID *iscp.ChainID, stateIndex uint32, tx *iotago.Transaction) bool) {
	mncT.publishTransactionAllowedFun = fun
}

func (mncT *MockedNodeConn) SetPullLatestOutputAllowed(flag bool) {
	mncT.pullLatestOutputAllowed = flag
}

func (mncT *MockedNodeConn) SetPullTxInclusionStateAllowed(flag bool) {
	mncT.SetPullTxInclusionStateAllowedFun(func(*iscp.ChainID, iotago.TransactionID) bool { return flag })
}

func (mncT *MockedNodeConn) SetPullTxInclusionStateAllowedFun(fun func(chainID *iscp.ChainID, txID iotago.TransactionID) bool) {
	mncT.pullTxInclusionStateAllowedFun = fun
}

func (mncT *MockedNodeConn) SetPullOutputByIDAllowed(flag bool) {
	mncT.SetPullOutputByIDAllowedFun(func(*iscp.ChainID, *iotago.UTXOInput) bool { return flag })
}

func (mncT *MockedNodeConn) SetPullOutputByIDAllowedFun(fun func(chainID *iscp.ChainID, outputID *iotago.UTXOInput) bool) {
	mncT.pullOutputByIDAllowedFun = fun
}

func (mncT *MockedNodeConn) L1Params() *parameters.L1 {
	return parameters.L1ForTesting()
}

/*func (m *MockedNodeConn) PullLatestOutput() {
	m.log.Debugf("Pull latest state output")
	if m.pullLatestStateOutputAllowed {
		m.log.Debugf("Pull latest state output allowed")
		output := m.ledger.PullState()
		if output != nil {
			m.log.Debugf("Pull latest state output successful")
			go m.handleUnspentAliasOutputFun(output, time.Now())
		}
	}
}

func (m *MockedNodeConn) PullTxInclusionState(txid iotago.TransactionID) {
	panic("TODO implement")
}

func (m *MockedNodeConn) PullOutputByID(outputID *iotago.UTXOInput) {
	m.log.Debugf("Pull output by id %v", iscp.OID(outputID))
	if m.pullOutputByIDAllowedFun(outputID) {
		m.log.Debugf("Pull output by id %v allowed", iscp.OID(outputID))
		output := m.ledger.PullConfirmedOutput(outputID)
		if output != nil {
			m.log.Debugf("Pull confirmed output %v successful", iscp.OID(outputID))
			go m.handleOutputFun(output, outputID)
		}
	}
}

func (m *MockedNodeConn) PublishTransaction(stateIndex uint32, tx *iotago.Transaction) error {
	m.log.Debugf("Publishing transaction for state %v", stateIndex)
	if m.receiveTxAllowedFun(stateIndex, tx) {
		m.log.Debugf("Publishing transaction for state %v allowed", stateIndex)
		m.ledger.receiveTx(tx)
		return nil
	}
	return fmt.Errorf("Publishing transaction for state %v not allowed", stateIndex)
}

func (m *MockedNodeConn) SetPullLatestStateOutputAllowed(flag bool) {
	m.pullLatestStateOutputAllowed = flag
}

func (m *MockedNodeConn) SetPullConfirmedOutputAllowed(flag bool) {
	m.SetPullConfirmedOutputAllowedFun(func(*iotago.UTXOInput) bool { return flag })
}

func (m *MockedNodeConn) SetPullOutputByIDAllowedFun(fun func(*iotago.UTXOInput) bool) {
	m.pullOutputByIDAllowedFun = fun
}

func (m *MockedNodeConn) SetReceiveTxAllowed(flag bool) {
	m.SetReceiveTxAllowedFun(func(uint32, *iotago.Transaction) bool { return flag })
}

func (m *MockedNodeConn) SetReceiveTxAllowedFun(fun func(stateIndex uint32, tx *iotago.Transaction) bool) {
	m.receiveTxAllowedFun = fun
}

func (m *MockedNodeConn) defaultHandleTimeDataFun(*iscp.TimeData) {}

func (m *MockedNodeConn) AttachToTimeData(fun chain.NodeConnectionHandleTimeDataFun) {
	m.handleTimeDataFun = fun
}

func (m *MockedNodeConn) DetachFromTimeData() {
	m.handleTimeDataFun = m.defaultHandleTimeDataFun
}

func (m *MockedNodeConn) defaultHandleTransactionFun(*iotago.Transaction) {}

func (m *MockedNodeConn) AttachToTransactionReceived(fun chain.NodeConnectionHandleTransactionFun) {
	m.handleTransactionFun = fun
}

func (m *MockedNodeConn) DetachFromTransactionReceived() {
	m.handleTransactionFun = m.defaultHandleTransactionFun
}*/

// func (m *MockedNodeConn) DetachFromInclusionStateReceived() { /* TODO */ }

/*func (m *MockedNodeConn) defaultHandleOutputFun(iotago.Output, *iotago.UTXOInput) {}

func (m *MockedNodeConn) AttachToOutputReceived(fun chain.NodeConnectionHandleOutputFun) {
	m.handleOutputFun = fun
}

func (m *MockedNodeConn) DetachFromOutputReceived() {
	m.handleOutputFun = m.defaultHandleOutputFun
}

func (m *MockedNodeConn) defaultHandleUnspentAliasOutputFun(*iscp.AliasOutputWithID, time.Time) {}

func (m *MockedNodeConn) AttachToUnspentAliasOutputReceived(fun chain.NodeConnectionHandleUnspentAliasOutputFun) {
	m.handleUnspentAliasOutputFun = fun
}

func (m *MockedNodeConn) DetachFromUnspentAliasOutputReceived() {
	m.handleUnspentAliasOutputFun = m.defaultHandleUnspentAliasOutputFun
}

func (m *MockedNodeConn) Close() {
	close(m.stopChannel)
}

func (m *MockedNodeConn) GetMetrics() nodeconnmetrics.NodeConnectionMessagesMetrics {
	return nodeconnmetrics.NewEmptyNodeConnectionMessagesMetrics()
}

func (m *MockedNodeConn) pushMilestonesLoop() {
	milestone := uint32(0)
	for {
		select {
		case <-time.After(100 * time.Millisecond):
			m.handleTimeDataFun(&iscp.TimeData{
				MilestoneIndex: milestone,
				Time:           time.Now(),
			})
			milestone++
		case <-m.stopChannel:
			return
		}
	}
}*/

/*AttachToAliasOutput(NodeConnectionAliasOutputHandlerFun)
DetachFromAliasOutput()
AttachToOnLedgerRequest(NodeConnectionOnLedgerRequestHandlerFun)
DetachFromOnLedgerRequest()
AttachToTxInclusionState(NodeConnectionInclusionStateHandlerFun)
DetachFromTxInclusionState()
AttachToMilestones(NodeConnectionMilestonesHandlerFun)
DetachFromMilestones()
Close()

+PublishTransaction(stateIndex uint32, tx *iotago.Transaction) error
+PullLatestOutput()
+PullTxInclusionState(txid iotago.TransactionID)
PullOutputByID(*iotago.UTXOInput)
*/
