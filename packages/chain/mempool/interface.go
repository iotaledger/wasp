package mempool

import (
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
)

type Mempool interface {
	ReceiveRequests(reqs ...iscp.Request)
	ReceiveRequest(req iscp.Request) bool
	RemoveRequests(reqs ...iscp.RequestID)
	ReadyNow(currentTime time.Time) []iscp.Request
	ReadyFromIDs(currentTime time.Time, reqIDs ...iscp.RequestID) ([]iscp.Request, []int, bool)
	HasRequest(id iscp.RequestID) bool
	GetRequest(id iscp.RequestID) iscp.Request
	Info(currentTime time.Time) MempoolInfo
	WaitRequestInPool(reqid iscp.RequestID, timeout ...time.Duration) bool // for testing
	WaitInBufferEmpty(timeout ...time.Duration) bool                       // for testing
	WaitPoolEmpty(timeout ...time.Duration) bool                           // for testing
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
