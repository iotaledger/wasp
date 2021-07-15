// mempool implements a buffer of requests sent to the ISCP chain, essentially a backlog of requests
// It contains both on-ledger and off-ledger requests. The mempool consists of 2 parts: the in-buffer and the pool
// All incoming requests are stored into the in-buffer first. Then they are asynchronously validated
// and moved to the pool itself.
package mempool

import (
	"bytes"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/iscp/rotate"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

type Mempool struct {
	inBuffer                map[iscp.RequestID]iscp.Request
	inMutex                 sync.RWMutex
	poolMutex               sync.RWMutex
	inBufCounter            int
	outBufCounter           int
	inPoolCounter           int
	outPoolCounter          int
	stateReader             state.OptimisticStateReader
	pool                    map[iscp.RequestID]*requestRef
	chStop                  chan struct{}
	blobCache               registry.BlobCache
	solidificationLoopDelay time.Duration
	log                     *logger.Logger
}

type requestRef struct {
	req          iscp.Request
	whenReceived time.Time
}

const (
	defaultSolidificationLoopDelay = 200 * time.Millisecond
	moveToPoolLoopDelay            = 20 * time.Millisecond
)

var _ chain.Mempool = &Mempool{}

func New(stateReader state.OptimisticStateReader, blobCache registry.BlobCache, log *logger.Logger, solidificationLoopDelay ...time.Duration) *Mempool {
	ret := &Mempool{
		inBuffer:    make(map[iscp.RequestID]iscp.Request),
		stateReader: stateReader,
		pool:        make(map[iscp.RequestID]*requestRef),
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

func (m *Mempool) addToInBuffer(req iscp.Request) bool {
	// just check if it is already in the pool
	if m.HasRequest(req.ID()) {
		return false
	}
	m.inMutex.Lock()
	defer m.inMutex.Unlock()
	// may be repeating but does not matter
	m.inBuffer[req.ID()] = req
	m.inBufCounter++
	return true
}

func (m *Mempool) removeFromInBuffer(req iscp.Request) {
	m.inMutex.Lock()
	defer m.inMutex.Unlock()
	if _, ok := m.inBuffer[req.ID()]; ok {
		delete(m.inBuffer, req.ID())
		m.outBufCounter++
	}
}

// fills up the buffer with requests from the in-buffer
func (m *Mempool) takeInBuffer(buf []iscp.Request) []iscp.Request {
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
func (m *Mempool) addToPool(req iscp.Request) bool {
	if offLedgerReq, ok := req.(*request.RequestOffLedger); ok {
		if !offLedgerReq.VerifySignature() {
			// wrong signature, must be removed from in buffer
			m.log.Warnf("ReceiveRequest.VerifySignature: invalid signature")
			return true
		}
	}
	reqid := req.ID()

	// checking in the state if request is processed. Reading may fail
	m.stateReader.SetBaseline()
	alreadyProcessed, err := blocklog.IsRequestProcessed(m.stateReader.KVStoreReader(), &reqid)
	if err != nil {
		// may be invalidated state. Do not remove from in-buffer yet
		return false
	}
	if alreadyProcessed {
		// remove from the in-buffer but not include into the pool
		return true
	}
	m.poolMutex.Lock()
	defer m.poolMutex.Unlock()

	if _, inPool := m.pool[reqid]; inPool {
		// already there, remove from the in-buffer
		return true
	}

	// put the request to the pool
	nowis := time.Now()
	m.inPoolCounter++

	m.traceIn(req)

	m.pool[reqid] = &requestRef{
		req:          req,
		whenReceived: nowis,
	}
	if _, err := request.SolidifyArgs(req, m.blobCache); err != nil {
		m.log.Errorf("ReceiveRequest.SolidifyArgs: %s", err)
	}
	// return true to remove from the in-buffer
	return true
}

// ReceiveRequests places requests into the inBuffer. InBuffer is unordered and non-deterministic
func (m *Mempool) ReceiveRequests(reqs ...iscp.Request) {
	for _, req := range reqs {
		m.addToInBuffer(req)
	}
}

// ReceiveRequest used to receive off-ledger request
func (m *Mempool) ReceiveRequest(req iscp.Request) bool {
	// could be worth it to check if the request was already processed in the blocklog.
	// Not adding this check now to avoid overhead, but should be looked into in case re-gossiping happens a lot
	m.inMutex.RLock()
	if _, exists := m.inBuffer[req.ID()]; exists {
		m.inMutex.RUnlock()
		return false
	}
	m.inMutex.RUnlock()
	return m.addToInBuffer(req)
}

// RemoveRequests removes requests from the pool
func (m *Mempool) RemoveRequests(reqs ...iscp.RequestID) {
	m.poolMutex.Lock()
	defer m.poolMutex.Unlock()

	for _, rid := range reqs {
		if _, ok := m.pool[rid]; !ok {
			continue
		}
		m.outPoolCounter++
		delete(m.pool, rid)
		m.traceOut(rid)
	}
}

const traceInOut = false

func (m *Mempool) traceIn(req iscp.Request) {
	rotateStr := ""
	if rotate.IsRotateStateControllerRequest(req) {
		rotateStr = "(rotate) "
	}
	tl := req.TimeLock()
	if traceInOut {
		if tl.IsZero() {
			m.log.Infof("IN MEMPOOL %s%s (+%d / -%d)", rotateStr, req.ID(), m.inPoolCounter, m.outPoolCounter)
		} else {
			m.log.Infof("IN MEMPOOL %s%s (+%d / -%d) timelocked for %v", rotateStr, req.ID(), m.inPoolCounter, m.outPoolCounter, time.Until(tl))
		}
	} else {
		if tl.IsZero() {
			m.log.Debugf("IN MEMPOOL %s%s (+%d / -%d)", rotateStr, req.ID(), m.inPoolCounter, m.outPoolCounter)
		} else {
			m.log.Debugf("IN MEMPOOL %s%s (+%d / -%d) timelocked for %v", rotateStr, req.ID(), m.inPoolCounter, m.outPoolCounter, time.Until(tl))
		}
	}
}

func (m *Mempool) traceOut(reqid iscp.RequestID) {
	if traceInOut {
		m.log.Infof("OUT MEMPOOL %s (+%d / -%d)", reqid, m.inPoolCounter, m.outPoolCounter)
	} else {
		m.log.Debugf("OUT MEMPOOL %s (+%d / -%d)", reqid, m.inPoolCounter, m.outPoolCounter)
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

// ReadyNow returns preliminary batch of requests for consensus.
// Note that later status of request may change due to the time change and time constraints
// If there's at least one committee rotation request in the mempool, the ReadyNow returns
// batch with only one request, the oldest committee rotation request
func (m *Mempool) ReadyNow(now ...time.Time) []iscp.Request {
	m.poolMutex.RLock()
	defer m.poolMutex.RUnlock()

	nowis := time.Now()
	if len(now) > 0 {
		nowis = now[0]
	}
	var oldestRotate iscp.Request
	var oldestRotateTime time.Time

	ret := make([]iscp.Request, 0, len(m.pool))
	for _, ref := range m.pool {
		if !isRequestReady(ref, nowis) {
			continue
		}
		ret = append(ret, ref.req)
		if !rotate.IsRotateStateControllerRequest(ref.req) {
			continue
		}
		// selecting oldest rotate request
		if oldestRotate == nil {
			oldestRotate = ref.req
			oldestRotateTime = ref.whenReceived
		} else {
			switch {
			case ref.whenReceived.Before(oldestRotateTime):
				oldestRotate = ref.req
				oldestRotateTime = ref.whenReceived
			case ref.whenReceived.Equal(oldestRotateTime):
				// for full determinism we take inti account not only time but also the request id
				if bytes.Compare(ref.req.ID().Bytes(), oldestRotate.ID().Bytes()) < 0 {
					oldestRotate = ref.req
					oldestRotateTime = ref.whenReceived
				}
			}
		}
	}
	if oldestRotate != nil {
		return []iscp.Request{oldestRotate}
	}
	return ret
}

// ReadyFromIDs if successful, function returns a deterministic list of requests for running on the VM
// - (a list of missing requests), false if some requests not arrived to the mempool yet. For retry later
// - (a list of processable requests), true if the list can be deterministically calculated
// Note that (a list of processable requests) can be empty if none satisfies nowis time constraint (timelock, fallback)
// For requests which are known and solidified, the result is deterministic
func (m *Mempool) ReadyFromIDs(nowis time.Time, reqIDs ...iscp.RequestID) ([]iscp.Request, []int, bool) {
	requests := make([]iscp.Request, 0, len(reqIDs))
	missingRequestIndexes := []int{}
	m.poolMutex.RLock()
	defer m.poolMutex.RUnlock()
	for i, reqID := range reqIDs {
		reqref, ok := m.pool[reqID]
		if !ok {
			missingRequestIndexes = append(missingRequestIndexes, i)
		} else if isRequestReady(reqref, nowis) {
			requests = append(requests, reqref.req)
		}
	}
	return requests, missingRequestIndexes, len(missingRequestIndexes) == 0
}

// HasRequest checks if the request is in the pool
func (m *Mempool) HasRequest(id iscp.RequestID) bool {
	m.poolMutex.RLock()
	defer m.poolMutex.RUnlock()

	_, ok := m.pool[id]
	return ok
}

func (m *Mempool) GetRequest(id iscp.RequestID) iscp.Request {
	m.poolMutex.RLock()
	defer m.poolMutex.RUnlock()

	if reqRef, ok := m.pool[id]; ok {
		return reqRef.req
	}
	return nil
}

const waitRequestInPoolTimeoutDefault = 2 * time.Second

// WaitRequestInPool waits until the request appears in the pool but no longer than timeout
func (m *Mempool) WaitRequestInPool(reqid iscp.RequestID, timeout ...time.Duration) bool {
	nowis := time.Now()
	deadline := nowis.Add(waitRequestInPoolTimeoutDefault)
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

func (m *Mempool) inBufferLen() int {
	m.inMutex.RLock()
	defer m.inMutex.RUnlock()
	return len(m.inBuffer)
}

const waitInBufferEmptyTimeoutDefault = 5 * time.Second

// WaitAllRequestsIn waits until in buffer becomes empty. Used in synchronous situations when the caller
// want to be sure all requests were fed into the pool. May create nondeterminism when used from goroutines
func (m *Mempool) WaitInBufferEmpty(timeout ...time.Duration) bool {
	nowis := time.Now()
	deadline := nowis.Add(waitInBufferEmptyTimeoutDefault)
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

// Stats collects mempool stats
func (m *Mempool) Info() chain.MempoolInfo {
	m.poolMutex.RLock()
	defer m.poolMutex.RUnlock()

	ret := chain.MempoolInfo{
		InPoolCounter:  m.inPoolCounter,
		OutPoolCounter: m.outPoolCounter,
		InBufCounter:   m.inBufCounter,
		OutBufCounter:  m.outBufCounter,
		TotalPool:      len(m.pool),
	}
	nowis := time.Now()
	for _, ref := range m.pool {
		if isRequestReady(ref, nowis) {
			ret.ReadyCounter++
		}
	}
	return ret
}

func (m *Mempool) Close() {
	close(m.chStop)
}

// the loop validates and moves request from inBuffer to the pool
func (m *Mempool) moveToPoolLoop() {
	buf := make([]iscp.Request, 0, 100) //nolint:gomnd
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

// the loop solidifies requests
func (m *Mempool) solidificationLoop() {
	for {
		select {
		case <-m.chStop:
			return
		case <-time.After(m.solidificationLoopDelay):
			m.doSolidifyRequests()
		}
	}
}

func (m *Mempool) doSolidifyRequests() {
	m.poolMutex.Lock()
	defer m.poolMutex.Unlock()

	for _, ref := range m.pool {
		if ref.req != nil {
			_, _ = request.SolidifyArgs(ref.req, m.blobCache)
		}
	}
}
