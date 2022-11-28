// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
)

type typedPool[V isc.Request] struct {
	waitReq  WaitReq
	requests map[isc.RequestRefKey]V
}

var _ RequestPool[isc.OffLedgerRequest] = &typedPool[isc.OffLedgerRequest]{}

func NewTypedPool[V isc.Request](waitReq WaitReq) RequestPool[V] {
	return &typedPool[V]{
		waitReq:  waitReq,
		requests: map[isc.RequestRefKey]V{},
	}
}

func (olp *typedPool[V]) Has(reqRef *isc.RequestRef) bool {
	_, have := olp.requests[reqRef.AsKey()]
	return have
}

func (olp *typedPool[V]) Get(reqRef *isc.RequestRef) V {
	if req, ok := olp.requests[reqRef.AsKey()]; ok {
		return req
	}
	return *new(V) //nolint:gocritic
}

func (olp *typedPool[V]) Add(request V) {
	refKey := isc.RequestRefFromRequest(request).AsKey()
	olp.requests[refKey] = request
	olp.waitReq.Have(request)
}

func (olp *typedPool[V]) Remove(request V) {
	refKey := isc.RequestRefFromRequest(request).AsKey()
	delete(olp.requests, refKey)
}

func (olp *typedPool[V]) Filter(predicate func(request V) bool) {
	for refKey := range olp.requests {
		if !predicate(olp.requests[refKey]) {
			delete(olp.requests, refKey)
		}
	}
}

func (olp *typedPool[V]) StatusString() string {
	return fmt.Sprintf("{|req|=%v}", len(olp.requests))
}
