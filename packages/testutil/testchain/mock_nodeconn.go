package testchain

import iotago "github.com/iotaledger/iota.go/v3"

type MockedNodeConn struct {
	id                              string
	onPullBacklog                   func(addr *iotago.AliasAddress)
	onPullState                     func(addr *iotago.AliasAddress)
	onPullConfirmedTransaction      func(addr iotago.Address, txid iotago.TransactionID)
	onPullTransactionInclusionState func(addr iotago.Address, txid iotago.TransactionID)
	onPullConfirmedOutput           func(addr iotago.Address, outputID iotago.OutputID)
	onPostTransaction               func(tx *iotago.Transaction)
}

func NewMockedNodeConnection(id string) *MockedNodeConn {
	return &MockedNodeConn{id: id}
}

func (m *MockedNodeConn) ID() string {
	return m.id
}

func (m *MockedNodeConn) PullBacklog(addr *iotago.AliasAddress) {
	m.onPullBacklog(addr)
}

func (m *MockedNodeConn) PullState(addr *iotago.AliasAddress) {
	m.onPullState(addr)
}

func (m *MockedNodeConn) PullConfirmedTransaction(addr iotago.Address, txid iotago.TransactionID) {
	m.onPullConfirmedTransaction(addr, txid)
}

func (m *MockedNodeConn) PullTransactionInclusionState(addr iotago.Address, txid iotago.TransactionID) {
	m.onPullTransactionInclusionState(addr, txid)
}

func (m *MockedNodeConn) PullConfirmedOutput(addr iotago.Address, outputID iotago.OutputID) {
	m.onPullConfirmedOutput(addr, outputID)
}

func (m *MockedNodeConn) PostTransaction(tx *iotago.Transaction) {
	m.onPostTransaction(tx)
}

func (m *MockedNodeConn) OnPullBacklog(f func(addr *iotago.AliasAddress)) {
	m.onPullBacklog = f
}

func (m *MockedNodeConn) OnPullState(f func(addr *iotago.AliasAddress)) {
	m.onPullState = f
}

func (m *MockedNodeConn) OnPullConfirmedTransaction(f func(addr iotago.Address, txid iotago.TransactionID)) {
	m.onPullConfirmedTransaction = f
}

func (m *MockedNodeConn) OnPullTransactionInclusionState(f func(addr iotago.Address, txid iotago.TransactionID)) {
	m.onPullTransactionInclusionState = f
}

func (m *MockedNodeConn) OnPullConfirmedOutput(f func(addr iotago.Address, outputID iotago.OutputID)) {
	m.onPullConfirmedOutput = f
}

func (m *MockedNodeConn) OnPostTransaction(f func(tx *iotago.Transaction)) {
	m.onPostTransaction = f
}
