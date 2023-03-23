// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/wasp/packages/isc"
)

type typedPool[V isc.Request] struct {
	waitReq  WaitReq
	requests *shrinkingmap.ShrinkingMap[isc.RequestRefKey, *typedPoolEntry[V]]
}

type typedPoolEntry[V isc.Request] struct {
	req V
	ts  time.Time
}

var _ RequestPool[isc.OffLedgerRequest] = &typedPool[isc.OffLedgerRequest]{}

func NewTypedPool[V isc.Request](waitReq WaitReq) RequestPool[V] {
	return &typedPool[V]{
		waitReq:  waitReq,
		requests: shrinkingmap.New[isc.RequestRefKey, *typedPoolEntry[V]](),
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
	olp.requests.Set(refKey, &typedPoolEntry[V]{req: request, ts: time.Now()})
	olp.waitReq.Have(request)
}

func (olp *typedPool[V]) Remove(request V) {
	refKey := isc.RequestRefFromRequest(request).AsKey()
	olp.requests.Delete(refKey)
}

func (olp *typedPool[V]) Filter(predicate func(request V, ts time.Time) bool) {
	olp.requests.ForEach(func(refKey isc.RequestRefKey, entry *typedPoolEntry[V]) bool {
		if !predicate(entry.req, entry.ts) {
			olp.requests.Delete(refKey)
		}
		return true
	})
}

func (olp *typedPool[V]) StatusString() string {
	return fmt.Sprintf("{|req|=%d}", olp.requests.Size())
}
