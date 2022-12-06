// Here we implement a trivial mempool, just for solo tests.
// We don't use the mempool from the chain/consensus because
// it does a lot of functions not needed in this context.
// The interface of this mempool has a little in common with
// the real mempool implementation.

package solo

import (
	"time"

	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/isc"
)

type Mempool interface {
	ReceiveRequests(reqs ...isc.Request)
	RequestBatchProposal() []isc.Request
	RemoveRequests(reqs ...isc.RequestID)
	Info() MempoolInfo
}

type MempoolInfo struct {
	TotalPool      int
	InPoolCounter  int
	OutPoolCounter int
}

type mempoolImpl struct {
	requests map[isc.RequestID]isc.Request
	info     MempoolInfo
}

func newMempool() Mempool {
	return &mempoolImpl{
		requests: map[isc.RequestID]isc.Request{},
		info:     MempoolInfo{},
	}
}

func (mi *mempoolImpl) ReceiveRequests(reqs ...isc.Request) {
	for _, req := range reqs {
		if _, ok := mi.requests[req.ID()]; !ok {
			mi.info.TotalPool++
			mi.info.InPoolCounter++
		}
		mi.requests[req.ID()] = req
	}
}

func (mi *mempoolImpl) RequestBatchProposal() []isc.Request {
	now := time.Now()
	batch := []isc.Request{}
	for rid, request := range mi.requests {
		switch request := request.(type) {
		case isc.OnLedgerRequest:
			reqUnlockCondSet := request.Output().UnlockConditionSet()
			timeLock := reqUnlockCondSet.Timelock()
			expiration := reqUnlockCondSet.Expiration()
			if expiration != nil && timeLock.UnixTime >= expiration.UnixTime {
				// can never be processed, just reject
				delete(mi.requests, rid)
				continue
			}
			if timeLock == nil || timeLock.UnixTime <= uint32(now.Unix()) {
				batch = append(batch, request)
				continue
			}
		case isc.OffLedgerRequest:
			batch = append(batch, request)
		default:
			panic(xerrors.Errorf("unexpected request type %T: %+v", request, request))
		}
	}
	return batch
}

func (mi *mempoolImpl) RemoveRequests(reqIDs ...isc.RequestID) {
	for _, rid := range reqIDs {
		if _, ok := mi.requests[rid]; ok {
			mi.info.OutPoolCounter++
			mi.info.TotalPool--
		}
		delete(mi.requests, rid)
	}
}

func (mi *mempoolImpl) Info() MempoolInfo {
	return mi.info
}
