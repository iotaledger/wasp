package testchain

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type MockedNodeConn struct {
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

func NewMockedNodeConnection(id string, ledger *MockedLedger) *MockedNodeConn {
	result := &MockedNodeConn{
		id:                            id,
		ledger:                        ledger,
		pullStateAllowed:              true,
		pullConfirmedOutputAllowedFun: func(*iotago.UTXOInput) bool { return true },
		receiveTxAllowedFun:           func(*iotago.Transaction) bool { return true },
	}
	result.handleTransactionFun = result.defaultHandleTransactionFun
	result.handleOutputFun = result.defaultHandleOutputFun
	result.handleUnspentAliasOutputFun = result.defaultHandleUnspentAliasOutputFun
	ledger.addNode(result)
	return result
}

func (m *MockedNodeConn) ID() string {
	return m.id
}

func (m *MockedNodeConn) PullState() {
	if m.pullStateAllowed {
		output := m.ledger.pullState()
		if output != nil {
			m.handleUnspentAliasOutputFun(output, time.Now())
		}
	}
}

func (m *MockedNodeConn) PullTransactionInclusionState(txid iotago.TransactionID) {
	// TODO
}

func (m *MockedNodeConn) PullConfirmedOutput(outputID *iotago.UTXOInput) {
	if m.pullConfirmedOutputAllowedFun(outputID) {
		output := m.ledger.pullConfirmedOutput(outputID)
		if output != nil {
			m.handleOutputFun(output, outputID)
		}
	}
}

func (m *MockedNodeConn) PostTransaction(tx *iotago.Transaction) {
	if m.receiveTxAllowedFun(tx) {
		m.ledger.receiveTx(tx)
	}
}

func (m *MockedNodeConn) SetPullStateAllowed(flag bool) {
	m.pullStateAllowed = flag
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
