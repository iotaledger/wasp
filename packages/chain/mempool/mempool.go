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
	mutex                   sync.RWMutex
	incounter               int
	outcounter              int
	stateReader             state.StateReader
	requestRefs             map[coretypes.RequestID]*requestRef
	chStop                  chan struct{}
	blobCache               coretypes.BlobCache
	solidificationLoopDelay time.Duration
	log                     *logger.Logger
}

type requestRef struct {
	req          coretypes.Request
	whenReceived time.Time
}

const defaultSolidificationLoopDelay = 200 * time.Millisecond

var _ chain.Mempool = &mempool{}

func New(stateReader state.StateReader, blobCache coretypes.BlobCache, log *logger.Logger, solidificationLoopDelay ...time.Duration) *mempool {
	ret := &mempool{
		stateReader: stateReader,
		requestRefs: make(map[coretypes.RequestID]*requestRef),
		chStop:      make(chan struct{}),
		blobCache:   blobCache,
		log:         log.Named("m"),
	}
	if len(solidificationLoopDelay) > 0 {
		ret.solidificationLoopDelay = solidificationLoopDelay[0]
	} else {
		ret.solidificationLoopDelay = defaultSolidificationLoopDelay
	}
	go ret.solidificationLoop()
	return ret
}

func (m *mempool) ReceiveRequests(reqs ...coretypes.Request) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, req := range reqs {
		// only allow off-ledger requests with valid signature
		if offLedgerReq, ok := req.(*request.RequestOffLedger); ok {
			if !offLedgerReq.VerifySignature() {
				m.log.Errorf("ReceiveRequest.VerifySignature: invalid signature")
				continue
			}
		}
		id := req.ID()
		if blocklog.IsRequestProcessed(m.stateReader, &id) {
			return
		}
		if _, ok := m.requestRefs[req.ID()]; ok {
			continue
		}

		// attempt solidification for those requests that do not require blobs
		// instead of having to wait for the solidification goroutine to kick in
		// also weeds out requests with solidification errors
		_, err := req.SolidifyArgs(m.blobCache)
		if err != nil {
			m.log.Errorf("ReceiveRequest.SolidifyArgs: %s", err)
			continue
		}
		nowis := time.Now()
		m.incounter++
		tl := req.TimeLock()
		if tl.IsZero() {
			m.log.Debugf("IN MEMPOOL %s (+%d / -%d)", req.ID(), m.incounter, m.outcounter)
		} else {
			m.log.Debugf("IN MEMPOOL %s (+%d / -%d) timelocked for %v", req.ID(), m.incounter, m.outcounter, tl.Sub(time.Now()))
		}
		m.requestRefs[req.ID()] = &requestRef{
			req:          req,
			whenReceived: nowis,
		}
	}
}

func (m *mempool) RemoveRequests(reqs ...coretypes.RequestID) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, rid := range reqs {
		if _, ok := m.requestRefs[rid]; !ok {
			continue
		}
		m.outcounter++
		delete(m.requestRefs, rid)
		m.log.Debugf("OUT MEMPOOL %s (+%d / -%d)", rid, m.incounter, m.outcounter)
	}
}

// isRequestReady for requests with paramsReady, the result is strictly deterministic
func (m *mempool) isRequestReady(ref *requestRef, nowis time.Time) bool {
	// TODO fallback options
	if _, paramsReady := ref.req.Params(); !paramsReady {
		return false
	}
	return ref.req.TimeLock().IsZero() || ref.req.TimeLock().Before(nowis)
}

// ReadyNow returns preliminary batch for consensus.
// Note that later status of request may change due to time constraints
func (m *mempool) ReadyNow() []coretypes.Request {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	ret := make([]coretypes.Request, 0, len(m.requestRefs))
	nowis := time.Now()
	for _, ref := range m.requestRefs {
		if m.isRequestReady(ref, nowis) {
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
		reqref, ok := m.requestRefs[reqid]
		if !ok {
			// retry later
			return nil, false
		}
		if m.isRequestReady(reqref, nowis) {
			ret = append(ret, reqref.req)
		}
	}
	return ret, true
}

func (m *mempool) HasRequest(id coretypes.RequestID) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	rec, ok := m.requestRefs[id]
	return ok && rec.req != nil
}

func (m *mempool) Stats() chain.MempoolStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	ret := chain.MempoolStats{
		InCounter:  m.incounter,
		OutCounter: m.outcounter,
		Total:      len(m.requestRefs),
	}
	nowis := time.Now()
	for _, ref := range m.requestRefs {
		if m.isRequestReady(ref, nowis) {
			ret.Ready++
		}
	}
	return ret
}

func (m *mempool) Close() {
	close(m.chStop)
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
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, ref := range m.requestRefs {
		if ref.req != nil {
			_, _ = ref.req.SolidifyArgs(m.blobCache)
		}
	}
}
