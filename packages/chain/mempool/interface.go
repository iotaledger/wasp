package mempool

import (
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
)

type Mempool interface {
	ReceiveRequests(reqs ...iscp.RequestRaw)
	ReceiveRequest(req iscp.RequestRaw) bool
	RemoveRequests(reqs ...iscp.RequestID)
	ReadyNow(currentTime ...time.Time) []iscp.RequestRaw
	ReadyFromIDs(currentTime time.Time, reqIDs ...iscp.RequestID) ([]iscp.RequestRaw, []int, bool)
	HasRequest(id iscp.RequestID) bool
	GetRequest(id iscp.RequestID) iscp.RequestRaw
	Info() MempoolInfo
	WaitRequestInPool(reqid iscp.RequestID, timeout ...time.Duration) bool // for testing
	WaitInBufferEmpty(timeout ...time.Duration) bool                       // for testing
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
