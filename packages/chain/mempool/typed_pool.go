// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/log"

	consGR "github.com/iotaledger/wasp/packages/chain/cons/gr"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

type RequestPool[V isc.Request] interface {
	Has(reqRef *isc.RequestRef) bool
	Get(reqRef *isc.RequestRef) V
	Add(request V)
	Remove(request V)
	// this removes requests from the pool if predicate returns false
	Cleanup(predicate func(request V, ts time.Time) bool)
	Iterate(f func(e *typedPoolEntry[V]) bool)
	StatusString() string
	WriteContent(io.Writer)
	ShouldRefreshRequests() bool
}

// TODO add gas price to on-ledger requests
// TODO this list needs to be periodically re-filled from L1 once the activity is lower
type typedPool[V isc.Request] struct {
	waitReq            WaitReq
	requests           *shrinkingmap.ShrinkingMap[isc.RequestRefKey, *typedPoolEntry[V]]
	ordered            []*typedPoolEntry[V] // TODO use a better data structure instead!!! (probably RedBlackTree)
	hasDroppedRequests bool
	maxPoolSize        int
	sizeMetric         func(int)
	timeMetric         func(time.Duration)
	log                log.Logger
}

type typedPoolEntry[V isc.Request] struct {
	req         V
	proposedFor []consGR.ConsensusID
	ts          time.Time
}

var _ RequestPool[isc.OffLedgerRequest] = &typedPool[isc.OffLedgerRequest]{}

func NewTypedPool[V isc.Request](maxOnledgerInPool int, waitReq WaitReq, sizeMetric func(int), timeMetric func(time.Duration), log log.Logger) RequestPool[V] {
	return &typedPool[V]{
		waitReq:            waitReq,
		requests:           shrinkingmap.New[isc.RequestRefKey, *typedPoolEntry[V]](),
		ordered:            []*typedPoolEntry[V]{},
		hasDroppedRequests: false,
		maxPoolSize:        maxOnledgerInPool,
		sizeMetric:         sizeMetric,
		timeMetric:         timeMetric,
		log:                log,
	}
}

func (olp *typedPool[V]) Has(reqRef *isc.RequestRef) bool {
	return olp.requests.Has(reqRef.AsKey())
}

func (olp *typedPool[V]) Get(reqRef *isc.RequestRef) V {
	entry, exists := olp.requests.Get(reqRef.AsKey())
	if !exists {
		return *new(V)
	}
	return entry.req
}

func (olp *typedPool[V]) Add(request V) {
	refKey := isc.RequestRefFromRequest(request).AsKey()
	entry := &typedPoolEntry[V]{
		req:         request,
		ts:          time.Now(),
		proposedFor: []consGR.ConsensusID{},
	}
	if !olp.requests.Set(refKey, entry) {
		return // already in pool
	}

	//
	// add the to the ordered list of requests
	{
		index, _ := slices.BinarySearchFunc(olp.ordered, entry, olp.reqSort)
		olp.ordered = append(olp.ordered, entry)
		// make room if target position is not at the end
		if index != len(olp.ordered)-1 {
			copy(olp.ordered[index+1:], olp.ordered[index:])
			olp.ordered[index] = entry
		}
	}

	// keep the pool size in check
	deleted := olp.LimitPoolSize()
	if lo.Contains(deleted, entry) {
		// this exact request was deleted from the pool, do not update metrics, or mark available
		return
	}

	//
	// update metrics and signal that the request is available
	olp.log.LogDebugf("ADD %v as key=%v", request.ID(), refKey)
	olp.sizeMetric(olp.requests.Size())
	olp.waitReq.MarkAvailable(request)
}

// LimitPoolSize drops the txs with the lowest price if the total number of requests is too big
func (olp *typedPool[V]) LimitPoolSize() []*typedPoolEntry[V] {
	if len(olp.ordered) <= olp.maxPoolSize {
		return nil // nothing to do
	}

	totalToDelete := len(olp.ordered) - olp.maxPoolSize
	reqsToDelete := make([]*typedPoolEntry[V], totalToDelete)
	j := 0
	for i := 0; i < len(olp.ordered); i++ {
		if len(olp.ordered[i].proposedFor) > 0 {
			continue // we cannot drop requests that are being used in current consensus instances
		}
		reqsToDelete[j] = olp.ordered[i]
		if j >= totalToDelete-1 {
			break
		}
	}

	for _, r := range reqsToDelete {
		olp.log.LogDebugf("LimitPoolSize dropping request: %v", r.req.ID())
		olp.Remove(r.req)
	}
	olp.hasDroppedRequests = true
	return reqsToDelete
}

func (olp *typedPool[V]) reqSort(a, b *typedPoolEntry[V]) int {
	// TODO use gas price to sort here, once on-ledger requests have a gas price field
	// use requestID as a fallback in case of matching gas price (it's random and should give roughly the same order between nodes)
	aID := a.req.ID()
	bID := b.req.ID()
	for i := range aID {
		if aID[i] == bID[i] {
			continue
		}
		if aID[i] > bID[i] {
			return 1
		}
		return -1
	}
	return 0
}

func (olp *typedPool[V]) Remove(request V) {
	refKey := isc.RequestRefFromRequest(request).AsKey()
	entry, ok := olp.requests.Get(refKey)
	if !ok {
		return
	}
	if !olp.requests.Delete(refKey) {
		return
	}

	//
	// find and delete the request from the ordered list
	{
		indexToDel := slices.IndexFunc(olp.ordered, func(e *typedPoolEntry[V]) bool {
			return refKey == isc.RequestRefFromRequest(e.req).AsKey()
		})
		olp.ordered[indexToDel] = nil // remove the pointer reference to allow GC of the entry object
		olp.ordered = slices.Delete(olp.ordered, indexToDel, indexToDel+1)
	}

	// log and update metrics
	olp.log.LogDebugf("DEL %v as key=%v", request.ID(), refKey)
	olp.sizeMetric(olp.requests.Size())
	olp.timeMetric(time.Since(entry.ts))
}

func (olp *typedPool[V]) ShouldRefreshRequests() bool {
	if !olp.hasDroppedRequests {
		return false
	}
	if olp.requests.Size() > 0 {
		return false // wait until pool is empty to refresh
	}
	// assume after this function returns true, the requests will be refreshed
	olp.hasDroppedRequests = false
	return true
}

func (olp *typedPool[V]) Cleanup(predicate func(request V, ts time.Time) bool) {
	olp.requests.ForEach(func(refKey isc.RequestRefKey, entry *typedPoolEntry[V]) bool {
		if !predicate(entry.req, entry.ts) {
			olp.Remove(entry.req)
		}
		return true
	})
	olp.sizeMetric(olp.requests.Size())
}

func (olp *typedPool[V]) Iterate(f func(e *typedPoolEntry[V]) bool) {
	orderedCopy := slices.Clone(olp.ordered)
	for _, entry := range orderedCopy {
		if !f(entry) {
			break
		}
	}

	olp.sizeMetric(olp.requests.Size())
}

func (olp *typedPool[V]) StatusString() string {
	return fmt.Sprintf("{|req|=%d}", olp.requests.Size())
}

func (olp *typedPool[V]) WriteContent(w io.Writer) {
	olp.requests.ForEach(func(_ isc.RequestRefKey, entry *typedPoolEntry[V]) bool {
		jsonData, err := isc.RequestToJSON(entry.req)
		if err != nil {
			return false // stop iteration
		}
		_, err = w.Write(codec.Encode[uint32](uint32(len(jsonData))))
		if err != nil {
			return false // stop iteration
		}
		_, err = w.Write(jsonData)
		return err == nil
	})
}
