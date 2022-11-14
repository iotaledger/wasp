package mempool

import (
	consGR "github.com/iotaledger/wasp/packages/chain/aaa2/cons/gr"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

type Mempool interface {
	consGR.Mempool // TODO should this be unified with the Mempool interface
	// Invoked by the chain when a new off-ledger request is received from a node user.
	// Inter-node off-ledger dissemination is NOT performed via this function.
	ReceiveOnLedgerRequest(request isc.OnLedgerRequest)
	// Invoked by the chain when a set of access nodes has changed.
	// These nodes should be used to disseminate the off-ledger requests.
	AccessNodesUpdated(committeePubKeys []*cryptolib.PublicKey, accessNodePubKeys []*cryptolib.PublicKey)
	//
	ReceiveRequests(reqs ...isc.Request) []bool
	RemoveRequests(reqs ...isc.RequestID)
	HasRequest(id isc.RequestID) bool
	GetRequest(id isc.RequestID) isc.Request
	Info() MempoolInfo
}

type MempoolInfo struct {
	TotalPool      int
	InPoolCounter  int
	OutPoolCounter int
}
