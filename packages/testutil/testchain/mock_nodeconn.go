package testchain

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

type MockedNodeConn struct {
	onPullBacklog                   func(addr *ledgerstate.AliasAddress)
	onPullState                     func(addr *ledgerstate.AliasAddress)
	onPullConfirmedTransaction      func(addr ledgerstate.Address, txid ledgerstate.TransactionID)
	onPullTransactionInclusionState func(addr ledgerstate.Address, txid ledgerstate.TransactionID)
	onPullConfirmedOutput           func(addr ledgerstate.Address, outputID ledgerstate.OutputID)
	onPostTransaction               func(tx *ledgerstate.Transaction)
}

func NewMockedNodeConnection() *MockedNodeConn {
	return &MockedNodeConn{}
}

func (m *MockedNodeConn) PullBacklog(addr *ledgerstate.AliasAddress) {
	m.onPullBacklog(addr)
}

func (n *MockedNodeConn) PullState(addr *ledgerstate.AliasAddress) {
	n.onPullState(addr)
}

func (m *MockedNodeConn) PullConfirmedTransaction(addr ledgerstate.Address, txid ledgerstate.TransactionID) {
	m.onPullConfirmedTransaction(addr, txid)
}

func (m *MockedNodeConn) PullTransactionInclusionState(addr ledgerstate.Address, txid ledgerstate.TransactionID) {
	m.onPullTransactionInclusionState(addr, txid)
}

func (m *MockedNodeConn) PullConfirmedOutput(addr ledgerstate.Address, outputID ledgerstate.OutputID) {
	m.onPullConfirmedOutput(addr, outputID)
}

func (m *MockedNodeConn) PostTransaction(tx *ledgerstate.Transaction) {
	m.onPostTransaction(tx)
}

func (m *MockedNodeConn) OnPullBacklog(f func(addr *ledgerstate.AliasAddress)) {
	m.onPullBacklog = f
}

func (m *MockedNodeConn) OnPullState(f func(addr *ledgerstate.AliasAddress)) {
	m.onPullState = f
}

func (m *MockedNodeConn) OnPullConfirmedTransaction(f func(addr ledgerstate.Address, txid ledgerstate.TransactionID)) {
	m.onPullConfirmedTransaction = f
}

func (m *MockedNodeConn) OnPullTransactionInclusionState(f func(addr ledgerstate.Address, txid ledgerstate.TransactionID)) {
	m.onPullTransactionInclusionState = f
}

func (m *MockedNodeConn) OnPullConfirmedOutput(f func(addr ledgerstate.Address, outputID ledgerstate.OutputID)) {
	m.onPullConfirmedOutput = f
}

func (m *MockedNodeConn) OnPostTransaction(f func(tx *ledgerstate.Transaction)) {
	m.onPostTransaction = f
}
