package mempool

import (
	"time"

	consGR "github.com/iotaledger/wasp/packages/chain/aaa2/cons/gr"
	"github.com/iotaledger/wasp/packages/isc"
)

type Mempool interface {
	consGR.Mempool // TODO should this be unified with the Mempool interface
	ReceiveRequests(reqs ...isc.Request) []bool
	HasRequest(id isc.RequestID) bool
	GetRequest(id isc.RequestID) isc.Request
	Info(currentTime time.Time) MempoolInfo
}

// for testing (only for use in solo)
type SoloMempool interface {
	Mempool
	WaitPoolEmpty(timeout ...time.Duration) bool
}

type MempoolInfo struct {
	TotalPool      int
	InPoolCounter  int
	OutPoolCounter int
}
