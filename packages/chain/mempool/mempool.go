package mempool

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

type mempool struct {
	inBuffer                map[coretypes.RequestID]coretypes.Request
	inMutex                 sync.RWMutex
	poolMutex               sync.RWMutex
	incounter               int
	outcounter              int
	stateReader             state.OptimisticStateReader
	pool                    map[coretypes.RequestID]*requestRef
	chStop                  chan struct{}
	blobCache               coretypes.BlobCache
	solidificationLoopDelay time.Duration
	log                     *logger.Logger
}

type requestRef struct {
	req          coretypes.Request
	whenReceived time.Time
}

const (
	defaultSolidificationLoopDelay = 200 * time.Millisecond
	moveToPoolLoopDelay            = 20 * time.Millisecond
)

var _ chain.Mempool = &mempool{}

func New(stateReader state.OptimisticStateReader, blobCache coretypes.BlobCache, log *logger.Logger, solidificationLoopDelay ...time.Duration) *mempool {
	ret := &mempool{
		inBuffer:    make(map[coretypes.RequestID]coretypes.Request),
		stateReader: stateReader,
		pool:        make(map[coretypes.RequestID]*requestRef),
		chStop:      make(chan struct{}),
		blobCache:   blobCache,
		log:         log.Named("m"),
	}
	if len(solidificationLoopDelay) > 0 {
		ret.solidificationLoopDelay = solidificationLoopDelay[0]
	} else {
		ret.solidificationLoopDelay = defaultSolidificationLoopDelay
	}
	go ret.moveToPoolLoop()
	go ret.solidificationLoop()
	return ret
}

func (m *mempool) addToInBuffer(req coretypes.Request) {
	// just check if it is already in the pool
	if m.HasRequest(req.ID()) {
		return
	}
	m.inMutex.Lock()
	defer m.inMutex.Unlock()
	// may be repeating but does not matter
	m.inBuffer[req.ID()] = req
}

func (m *mempool) removeFromInBuffer(req coretypes.Request) {
	m.inMutex.Lock()
	defer m.inMutex.Unlock()

	delete(m.inBuffer, req.ID())
}

// fills up the buffer with requests in the in buffer
func (m *mempool) takeInBuffer(buf []coretypes.Request) []coretypes.Request {
	buf = buf[:0]
	m.inMutex.RLock()
	defer m.inMutex.RUnlock()

	for _, req := range m.inBuffer {
		buf = append(buf, req)
	}
	return buf
}

// addToPool adds request to the pool. It may fail
// returns true if it must be removed from the input buffer
func (m *mempool) addToPool(req coretypes.Request) bool {
	if offLedgerReq, ok := req.(*request.RequestOffLedger); ok {
		if !offLedgerReq.VerifySignature() {
			// wrong signature, must be removed from in buffer
			m.log.Warnf("ReceiveRequest.VerifySignature: invalid signature")
			return true
		}
	}
	reqid := req.ID()
	m.stateReader.SetBaseline()
	alreadyProcessed, err := blocklog.IsRequestProcessed(m.stateReader.KVStoreReader(), &reqid)
	if err != nil {
		// may be invalid state. Do not remove from in buffer yet
		return false
	}
	if alreadyProcessed {
		// remove from buffer but not include into the pool
		return true
	}

	m.poolMutex.Lock()
	defer m.poolMutex.Unlock()

	if _, inPool := m.pool[reqid]; inPool {
		// already there, remove from in buffer
		return true
	}
	// put the request to the pool
	nowis := time.Now()
	m.incounter++

	m.traceIn(req)

	m.pool[reqid] = &requestRef{
		req:          req,
		whenReceived: nowis,
	}
	if _, err := req.SolidifyArgs(m.blobCache); err != nil {
		m.log.Errorf("ReceiveRequest.SolidifyArgs: %s", err)
	}
	// remove from in buffer
	return true
}

func (m *mempool) ReceiveRequests(reqs ...coretypes.Request) {
	for _, req := range reqs {
		m.addToInBuffer(req)
	}
}

func (m *mempool) RemoveRequests(reqs ...coretypes.RequestID) {
	m.poolMutex.Lock()
	defer m.poolMutex.Unlock()

	for _, rid := range reqs {
		if _, ok := m.pool[rid]; !ok {
			continue
		}
		m.outcounter++
		delete(m.pool, rid)

		m.traceOut(rid)
	}
}

const traceInOut = true

func (m *mempool) traceIn(req coretypes.Request) {
	tl := req.TimeLock()
	if traceInOut {
		if tl.IsZero() {
			m.log.Infof("IN MEMPOOL %s (+%d / -%d)", req.ID(), m.incounter, m.outcounter)
		} else {
			m.log.Infof("IN MEMPOOL %s (+%d / -%d) timelocked for %v", req.ID(), m.incounter, m.outcounter, tl.Sub(time.Now()))
		}
	} else {
		if tl.IsZero() {
			m.log.Debugf("IN MEMPOOL %s (+%d / -%d)", req.ID(), m.incounter, m.outcounter)
		} else {
			m.log.Debugf("IN MEMPOOL %s (+%d / -%d) timelocked for %v", req.ID(), m.incounter, m.outcounter, tl.Sub(time.Now()))
		}
	}
}

func (m *mempool) traceOut(reqid coretypes.RequestID) {
	if traceInOut {
		m.log.Infof("OUT MEMPOOL %s (+%d / -%d)", reqid, m.incounter, m.outcounter)
	} else {
		m.log.Debugf("OUT MEMPOOL %s (+%d / -%d)", reqid, m.incounter, m.outcounter)
	}
}

// isRequestReady for requests with paramsReady, the result is strictly deterministic
func isRequestReady(ref *requestRef, nowis time.Time) bool {
	// TODO fallback options
	if _, paramsReady := ref.req.Params(); !paramsReady {
		return false
	}
	return ref.req.TimeLock().IsZero() || ref.req.TimeLock().Before(nowis)
}

// ReadyNow returns preliminary batch for consensus.
// Note that later status of request may change due to time constraints
func (m *mempool) ReadyNow(now ...time.Time) []coretypes.Request {
	m.poolMutex.RLock()
	defer m.poolMutex.RUnlock()

	nowis := time.Now()
	if len(now) > 0 {
		nowis = now[0]
	}
	ret := make([]coretypes.Request, 0, len(m.pool))
	for _, ref := range m.pool {
		if isRequestReady(ref, nowis) {
			ret = append(ret, ref.req)
		}
	}
	return ret
}

// ReadyFromIDs if successful, function returns a deterministic list of requests for running on the VM
// - nil, false if some all requests not arrived to the mempool yet. For retry later
// - (a list of processable requests), true if the list can be deterministically calculated
// Note that (a list of processable requests) can be empty if none satisfies nowis time constraint (timelock, fallback)
// For requests which are known and solidified result is deterministic
func (m *mempool) ReadyFromIDs(nowis time.Time, reqids ...coretypes.RequestID) ([]coretypes.Request, bool) {
	ret := make([]coretypes.Request, 0, len(reqids))
	for _, reqid := range reqids {
		reqref, ok := m.pool[reqid]
		if !ok {
			// retry later
			return nil, false
		}
		if isRequestReady(reqref, nowis) {
			ret = append(ret, reqref.req)
		}
	}
	return ret, true
}

func (m *mempool) HasRequest(id coretypes.RequestID) bool {
	m.poolMutex.RLock()
	defer m.poolMutex.RUnlock()

	_, ok := m.pool[id]
	return ok
}

func (m *mempool) WaitRequestIn(reqid coretypes.RequestID, timeout ...time.Duration) bool {
	nowis := time.Now()
	deadline := nowis.Add(5 * time.Second)
	if len(timeout) > 0 {
		deadline = nowis.Add(timeout[0])
	}
	for {
		if m.HasRequest(reqid) {
			return true
		}
		time.Sleep(10 * time.Millisecond)
		if time.Now().After(deadline) {
			return false
		}
	}
}

func (m *mempool) inBufferLen() int {
	m.inMutex.RLock()
	defer m.inMutex.RUnlock()
	return len(m.inBuffer)
}

// WaitAllRequestsIn waits until in buffer becomes empty. Used in synchronous situations when the caller
// want to be sure all requests were fed into the pool
func (m *mempool) WaitAllRequestsIn(timeout ...time.Duration) bool {
	nowis := time.Now()
	deadline := nowis.Add(5 * time.Second)
	if len(timeout) > 0 {
		deadline = nowis.Add(timeout[0])
	}
	for {
		if m.inBufferLen() == 0 {
			return true
		}
		time.Sleep(10 * time.Millisecond)
		if time.Now().After(deadline) {
			return false
		}
	}
}

func (m *mempool) Stats() chain.MempoolStats {
	m.poolMutex.RLock()
	defer m.poolMutex.RUnlock()

	ret := chain.MempoolStats{
		InCounter:  m.incounter,
		OutCounter: m.outcounter,
		Total:      len(m.pool),
	}
	nowis := time.Now()
	for _, ref := range m.pool {
		if isRequestReady(ref, nowis) {
			ret.Ready++
		}
	}
	return ret
}

func (m *mempool) Close() {
	close(m.chStop)
}

func (m *mempool) moveToPoolLoop() {
	buf := make([]coretypes.Request, 0, 100)
	for {
		select {
		case <-m.chStop:
			return
		case <-time.After(moveToPoolLoopDelay):
			buf = m.takeInBuffer(buf)
			if len(buf) == 0 {
				continue
			}
			for i, req := range buf {
				if m.addToPool(req) {
					m.removeFromInBuffer(req)
				}
				buf[i] = nil // to please GC
			}
		}
	}
}

func (m *mempool) solidificationLoop() {
	for {
		select {
		case <-m.chStop:
			return
		case <-time.After(m.solidificationLoopDelay):
			m.doSolidifyRequests()
		}
	}
}

func (m *mempool) doSolidifyRequests() {
	m.poolMutex.Lock()
	defer m.poolMutex.Unlock()

	for _, ref := range m.pool {
		if ref.req != nil {
			_, _ = ref.req.SolidifyArgs(m.blobCache)
		}
	}
}
