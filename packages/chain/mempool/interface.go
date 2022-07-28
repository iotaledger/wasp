package mempool

import (
	"time"

	"github.com/iotaledger/wasp/packages/isc"
)

type Mempool interface {
	ReceiveRequests(reqs ...isc.Request)
	ReceiveRequest(req isc.Request) bool
	RemoveRequests(reqs ...isc.RequestID)
	ReadyNow(currentTime time.Time) []isc.Request
	ReadyFromIDs(currentTime time.Time, reqIDs ...isc.RequestID) ([]isc.Request, []int, bool)
	HasRequest(id isc.RequestID) bool
	GetRequest(id isc.RequestID) isc.Request
	Info(currentTime time.Time) MempoolInfo
	WaitRequestInPool(reqid isc.RequestID, timeout ...time.Duration) bool // for testing
	WaitInBufferEmpty(timeout ...time.Duration) bool                      // for testing
	WaitPoolEmpty(timeout ...time.Duration) bool                          // for testing
	Close()
}

type MempoolInfo struct {
	TotalPool      int
	ReadyCounter   int
	InBufCounter   int
	OutBufCounter  int
	InPoolCounter  int
	OutPoolCounter int
}
