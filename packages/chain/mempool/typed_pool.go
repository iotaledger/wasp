// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/isc"
)

type typedPool[V isc.Request] struct {
	waitReq  WaitReq
	requests map[isc.RequestRefKey]*typedPoolEntry[V]
}

type typedPoolEntry[V isc.Request] struct {
	req V
	ts  time.Time
}

var _ RequestPool[isc.OffLedgerRequest] = &typedPool[isc.OffLedgerRequest]{}

func NewTypedPool[V isc.Request](waitReq WaitReq) RequestPool[V] {
	return &typedPool[V]{
		waitReq:  waitReq,
		requests: map[isc.RequestRefKey]*typedPoolEntry[V]{},
	}
}

func (olp *typedPool[V]) Has(reqRef *isc.RequestRef) bool {
	_, have := olp.requests[reqRef.AsKey()]
	return have
}

func (olp *typedPool[V]) Get(reqRef *isc.RequestRef) V {
	if entry, ok := olp.requests[reqRef.AsKey()]; ok {
		return entry.req
	}
	return *new(V) //nolint:gocritic
}

func (olp *typedPool[V]) Add(request V) {
	refKey := isc.RequestRefFromRequest(request).AsKey()
	olp.requests[refKey] = &typedPoolEntry[V]{req: request, ts: time.Now()}
	olp.waitReq.Have(request)
}

func (olp *typedPool[V]) Remove(request V) {
	refKey := isc.RequestRefFromRequest(request).AsKey()
	delete(olp.requests, refKey)
}

func (olp *typedPool[V]) Filter(predicate func(request V, ts time.Time) bool) {
	for refKey, entry := range olp.requests {
		if !predicate(entry.req, entry.ts) {
			delete(olp.requests, refKey)
		}
	}
}

func (olp *typedPool[V]) StatusString() string {
	return fmt.Sprintf("{|req|=%v}", len(olp.requests))
}
