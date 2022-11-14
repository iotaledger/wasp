package mempool

import (
	consGR "github.com/iotaledger/wasp/packages/chain/aaa2/cons/gr"
	"github.com/iotaledger/wasp/packages/isc"
)

type Mempool interface {
	consGR.Mempool // TODO should this be unified with the Mempool interface
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
