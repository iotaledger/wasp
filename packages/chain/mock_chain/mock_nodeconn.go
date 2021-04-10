package mock_chain

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

type mockedNodeConn struct {
	onPullBacklog                   func(addr ledgerstate.Address)
	onPullConfirmedTransaction      func(addr ledgerstate.Address, txid ledgerstate.TransactionID)
	onPullTransactionInclusionState func(addr ledgerstate.Address, txid ledgerstate.TransactionID)
	onPullConfirmedOutput           func(addr ledgerstate.Address, outputID ledgerstate.OutputID)
	onPostTransaction               func(tx *ledgerstate.Transaction, fromSc ledgerstate.Address, fromLeader uint16)
}

func NewMockedNodeConnection() *mockedNodeConn {
	return &mockedNodeConn{}
}

func (m *mockedNodeConn) PullBacklog(addr ledgerstate.Address) {
	m.onPullBacklog(addr)
}

func (m *mockedNodeConn) PullConfirmedTransaction(addr ledgerstate.Address, txid ledgerstate.TransactionID) {
	m.onPullConfirmedTransaction(addr, txid)
}

func (m *mockedNodeConn) PullTransactionInclusionState(addr ledgerstate.Address, txid ledgerstate.TransactionID) {
	m.onPullTransactionInclusionState(addr, txid)
}

func (m *mockedNodeConn) PullConfirmedOutput(addr ledgerstate.Address, outputID ledgerstate.OutputID) {
	m.onPullConfirmedOutput(addr, outputID)
}

func (m *mockedNodeConn) PostTransaction(tx *ledgerstate.Transaction, fromSc ledgerstate.Address, fromLeader uint16) {
	m.onPostTransaction(tx, fromSc, fromLeader)
}

func (m *mockedNodeConn) OnPullBacklog(f func(addr ledgerstate.Address)) {
	m.onPullBacklog = f
}

func (m *mockedNodeConn) OnPullConfirmedTransaction(f func(addr ledgerstate.Address, txid ledgerstate.TransactionID)) {
	m.onPullConfirmedTransaction = f
}

func (m *mockedNodeConn) OnPullTransactionInclusionState(f func(addr ledgerstate.Address, txid ledgerstate.TransactionID)) {
	m.onPullTransactionInclusionState = f
}

func (m *mockedNodeConn) OnPullConfirmedOutput(f func(addr ledgerstate.Address, outputID ledgerstate.OutputID)) {
	m.onPullConfirmedOutput = f
}

func (m *mockedNodeConn) OnPostTransaction(f func(tx *ledgerstate.Transaction, fromSc ledgerstate.Address, fromLeader uint16)) {
	m.onPostTransaction = f
}
