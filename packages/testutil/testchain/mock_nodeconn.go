package testchain

import (
	"time"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type MockedNodeConn struct {
	log                           *logger.Logger
	ledger                        *MockedLedger
	id                            string
	pullStateAllowed              bool
	pullConfirmedOutputAllowedFun func(outputID *iotago.UTXOInput) bool
	receiveTxAllowedFun           func(tx *iotago.Transaction) bool
	handleTimeDataFun             chain.NodeConnectionHandleTimeDataFun
	handleTransactionFun          chain.NodeConnectionHandleTransactionFun
	handleOutputFun               chain.NodeConnectionHandleOutputFun
	handleUnspentAliasOutputFun   chain.NodeConnectionHandleUnspentAliasOutputFun
	stopChannel                   chan bool
}

var _ chain.ChainNodeConnection = &MockedNodeConn{}

func NewMockedNodeConnection(id string, ledger *MockedLedger, log *logger.Logger) *MockedNodeConn {
	result := &MockedNodeConn{
		log:         log.Named("nc"),
		id:          id,
		ledger:      ledger,
		stopChannel: make(chan bool),
	}
	result.handleTimeDataFun = result.defaultHandleTimeDataFun
	result.handleTransactionFun = result.defaultHandleTransactionFun
	result.handleOutputFun = result.defaultHandleOutputFun
	result.handleUnspentAliasOutputFun = result.defaultHandleUnspentAliasOutputFun
	result.SetPullStateAllowed(true)
	result.SetPullConfirmedOutputAllowed(true)
	result.SetReceiveTxAllowed(true)
	ledger.addNode(result)
	result.log.Debugf("Nodeconn created")
	go result.pushTimeDataLoop()
	return result
}

func (m *MockedNodeConn) ID() string {
	return m.id
}

func (m *MockedNodeConn) PullState() {
	m.log.Debugf("Pull state")
	if m.pullStateAllowed {
		m.log.Debugf("Pull state allowed")
		output := m.ledger.PullState()
		if output != nil {
			m.log.Debugf("Pull state successful")
			go m.handleUnspentAliasOutputFun(output, time.Now())
		}
	}
}

func (m *MockedNodeConn) PullTransactionInclusionState(txid iotago.TransactionID) {
	panic("TODO implement")
}

func (m *MockedNodeConn) PullConfirmedOutput(outputID *iotago.UTXOInput) {
	m.log.Debugf("Pull confirmed output %v", iscp.OID(outputID))
	if m.pullConfirmedOutputAllowedFun(outputID) {
		m.log.Debugf("Pull confirmed output %v allowed", iscp.OID(outputID))
		output := m.ledger.PullConfirmedOutput(outputID)
		if output != nil {
			m.log.Debugf("Pull confirmed output %v successful", iscp.OID(outputID))
			go m.handleOutputFun(output, outputID)
		}
	}
}

func (m *MockedNodeConn) PostTransaction(tx *iotago.Transaction) {
	m.log.Debugf("Post transaction")
	if m.receiveTxAllowedFun(tx) {
		m.log.Debugf("Post transaction allowed")
		m.ledger.receiveTx(tx)
	}
}

func (m *MockedNodeConn) SetPullStateAllowed(flag bool) {
	m.pullStateAllowed = flag
}

func (m *MockedNodeConn) SetPullConfirmedOutputAllowed(flag bool) {
	m.SetPullConfirmedOutputAllowedFun(func(*iotago.UTXOInput) bool { return flag })
}

func (m *MockedNodeConn) SetPullConfirmedOutputAllowedFun(fun func(*iotago.UTXOInput) bool) {
	m.pullConfirmedOutputAllowedFun = fun
}

func (m *MockedNodeConn) SetReceiveTxAllowed(flag bool) {
	m.SetReceiveTxAllowedFun(func(*iotago.Transaction) bool { return flag })
}

func (m *MockedNodeConn) SetReceiveTxAllowedFun(fun func(tx *iotago.Transaction) bool) {
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
}

func (m *MockedNodeConn) DetachFromInclusionStateReceived() { /* TODO */ }

func (m *MockedNodeConn) defaultHandleOutputFun(iotago.Output, *iotago.UTXOInput) {}

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

func (m *MockedNodeConn) pushTimeDataLoop() {
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
}
