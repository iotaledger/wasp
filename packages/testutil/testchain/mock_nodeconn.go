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
	handleTransactionFun          chain.NodeConnectionHandleTransactionFun
	handleOutputFun               chain.NodeConnectionHandleOutputFun
	handleUnspentAliasOutputFun   chain.NodeConnectionHandleUnspentAliasOutputFun
}

var _ chain.ChainNodeConnection = &MockedNodeConn{}

func NewMockedNodeConnection(id string, ledger *MockedLedger, log *logger.Logger) *MockedNodeConn {
	result := &MockedNodeConn{
		log:    log.Named("nc"),
		id:     id,
		ledger: ledger,
	}
	result.handleTransactionFun = result.defaultHandleTransactionFun
	result.handleOutputFun = result.defaultHandleOutputFun
	result.handleUnspentAliasOutputFun = result.defaultHandleUnspentAliasOutputFun
	result.SetPullStateAllowed(true)
	result.SetPullConfirmedOutputAllowed(true)
	result.SetReceiveTxAllowed(true)
	ledger.addNode(result)
	result.log.Debugf("Nodeconn created")
	return result
}

func (m *MockedNodeConn) ID() string {
	return m.id
}

func (m *MockedNodeConn) PullState() {
	m.log.Debugf("Pull state")
	if m.pullStateAllowed {
		m.log.Debugf("Pull state allowed")
		output := m.ledger.pullState()
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
		output := m.ledger.pullConfirmedOutput(outputID)
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

func (m *MockedNodeConn) Close() {}

func (m *MockedNodeConn) GetMetrics() nodeconnmetrics.NodeConnectionMessagesMetrics {
	return nodeconnmetrics.NewEmptyNodeConnectionMessagesMetrics()
}
