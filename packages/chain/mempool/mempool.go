package mempool

import (
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/state"
	"sort"
	"sync"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type mempool struct {
	mutex      sync.RWMutex
	chainState kvstore.KVStore
	requests   map[coretypes.RequestID]*request
	chStop     chan bool
	blobCache  coretypes.BlobCache
	log        *logger.Logger
}

type request struct {
	req             coretypes.Request
	timestamp       time.Time
	whenMsgReceived time.Time
	seen            map[uint16]bool
}

const constSolidificationLoopDelay = 200 * time.Millisecond

var _ chain.Mempool = &mempool{}

func New(chainState kvstore.KVStore, blobCache coretypes.BlobCache, log *logger.Logger) chain.Mempool {
	ret := &mempool{
		chainState: chainState,
		requests:   make(map[coretypes.RequestID]*request),
		chStop:     make(chan bool),
		blobCache:  blobCache,
		log:        log.Named("m"),
	}
	go ret.solidificationLoop()
	return ret
}

func (m *mempool) ReceiveRequest(req coretypes.Request) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	isCompleted, err := state.IsRequestCompleted(m.chainState, req.ID())
	if err != nil {
		m.log.Errorf("ReceiveRequest.IsRequestCompleted: %s", err)
		return
	}
	if isCompleted {
		return
	}
	if _, ok := m.requests[req.ID()]; ok {
		return
	}
	if onLedgerRequest, ok := req.(*sctransaction.RequestOnLedger); ok {
		tl := onLedgerRequest.TimeLock()
		if tl.IsZero() {
			m.log.Infof("IN MEMPOOL %s", req.ID())
		} else {
			m.log.Infof("IN MEMPOOL %s timelocked for %v", req.ID(), tl.Sub(time.Now()))
		}
	}
	m.requests[req.ID()] = &request{
		req:             req,
		whenMsgReceived: time.Now(),
		seen:            make(map[uint16]bool),
	}
}

func (m *mempool) MarkSeenByCommitteePeer(reqid *coretypes.RequestID, peerIndex uint16) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.requests[*reqid]; !ok {
		m.requests[*reqid] = &request{
			seen: make(map[uint16]bool),
		}
	}
	m.requests[*reqid].seen[peerIndex] = true
}

func (m *mempool) ClearSeenMarks() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, rec := range m.requests {
		rec.seen = make(map[uint16]bool)
	}
}

func (m *mempool) RemoveRequests(reqs ...coretypes.RequestID) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, rid := range reqs {
		delete(m.requests, rid)
		m.log.Infof("OUT MEMPOOL %s", rid)
	}
}

const timeAheadTolerance = 1000 * time.Nanosecond

func isRequestReady(req *request, seenThreshold uint16, nowis time.Time) bool {
	if req.req == nil {
		return false
	}
	if len(req.seen) < int(seenThreshold) {
		return false
	}
	if _, paramsReady := req.req.Params(); !paramsReady {
		return false
	}
	if r, ok := req.req.(*sctransaction.RequestOnLedger); ok {
		timeBaseline := nowis.Add(timeAheadTolerance)
		if !r.TimeLock().IsZero() && r.TimeLock().After(timeBaseline) {
			return false
		}
	}
	return true
}

func (m *mempool) GetReadyList(seenThreshold uint16) []coretypes.Request {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	ret := make([]coretypes.Request, 0, len(m.requests))
	nowis := time.Now()
	for _, req := range m.requests {
		if isRequestReady(req, seenThreshold, nowis) {
			ret = append(ret, req.req)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Order() < ret[j].Order()
	})
	return ret
}

func (m *mempool) GetReadyListFull(seenThreshold uint16) []*chain.ReadyListRecord {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	ret := make([]*chain.ReadyListRecord, 0, len(m.requests))
	nowis := time.Now()
	for _, req := range m.requests {
		if isRequestReady(req, seenThreshold, nowis) {
			rec := &chain.ReadyListRecord{
				Request: req.req,
				Seen:    make(map[uint16]bool),
			}
			for p := range req.seen {
				rec.Seen[p] = true
			}
			ret = append(ret, rec)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Request.Order() < ret[j].Request.Order()
	})
	return ret
}

func (m *mempool) TakeAllReady(nowis time.Time, reqids ...coretypes.RequestID) ([]coretypes.Request, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	ret := make([]coretypes.Request, len(reqids))
	for i := range reqids {
		req, ok := m.requests[reqids[i]]
		if !ok {
			return nil, false
		}
		if !isRequestReady(req, 0, nowis) {
			return nil, false
		}
		ret[i] = req.req
	}
	return ret, true
}

func (m *mempool) HasRequest(id coretypes.RequestID) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	rec, ok := m.requests[id]
	return ok && rec.req != nil
}

// Stats return total number, number with messages, number solid
func (m *mempool) Stats() (int, int, int) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	total := len(m.requests)
	withMsg, solid := 0, 0
	for _, req := range m.requests {
		if req.req == nil {
			continue
		}
		withMsg++
		if onTangleRequest, ok := req.req.(*sctransaction.RequestOnLedger); ok {
			if isSolid, _ := onTangleRequest.SolidifyArgs(m.blobCache); isSolid {
				solid++
			}
		}
	}
	return total, withMsg, solid
}

func (m *mempool) Close() {
	close(m.chStop)
}

func (m *mempool) solidificationLoop() {
	for {
		select {
		case <-m.chStop:
			return
		default:
			m.doSolidifyRequests()
		}
		time.Sleep(constSolidificationLoopDelay)
	}
}

func (m *mempool) doSolidifyRequests() {
	m.mutex.Lock()
	for _, req := range m.requests {
		if req.req == nil {
			continue
		}
		if onTangleRequest, ok := req.req.(*sctransaction.RequestOnLedger); ok {
			_, _ = onTangleRequest.SolidifyArgs(m.blobCache)
		}
	}
	m.mutex.Unlock()
}
