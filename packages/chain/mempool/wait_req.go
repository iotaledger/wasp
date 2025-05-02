// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"context"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/wasp/packages/isc"
)

// WaitReq provides the synchronization between the functions asking for requests
// and the arriving requests. The former have to be notified upon reception of the latter.
type WaitReq interface {
	WaitMany(ctx context.Context, reqRefs []*isc.RequestRef, cb func(req isc.Request)) // Called per block.
	WaitAny(ctx context.Context, cb func(req isc.Request))                             // Called per block.
	MarkAvailable(req isc.Request)                                                     // Called often, per request.
}

type waitReq struct {
	reqs           *shrinkingmap.ShrinkingMap[isc.RequestRefKey, []*waitReqItem] // Wait for specific req.
	any            []*waitReqItem                                                // Wait for any req.
	cleanupCounter int                                                           // Perform cleanup when the counter goes to 0.
	cleanupEvery   int
}

type waitReqItem struct {
	ctx context.Context
	cb  func(req isc.Request)
}

func NewWaitReq(cleanupEvery int) WaitReq {
	return &waitReq{
		reqs:           shrinkingmap.New[isc.RequestRefKey, []*waitReqItem](),
		any:            []*waitReqItem{},
		cleanupCounter: cleanupEvery,
		cleanupEvery:   cleanupEvery,
	}
}

func (wr *waitReq) WaitMany(ctx context.Context, reqRefs []*isc.RequestRef, cb func(req isc.Request)) {
	for i := range reqRefs {
		reqRefKey := reqRefs[i].AsKey()
		item := &waitReqItem{ctx: ctx, cb: cb}

		requests, _ := wr.reqs.GetOrCreate(reqRefKey, func() []*waitReqItem { return make([]*waitReqItem, 0, 1) })
		wr.reqs.Set(reqRefKey, append(requests, item))
	}
	wr.maybeCleanup()
}

func (wr *waitReq) WaitAny(ctx context.Context, cb func(req isc.Request)) {
	wr.any = append(wr.any, &waitReqItem{ctx: ctx, cb: cb})
	wr.maybeCleanup()
}

func (wr *waitReq) MarkAvailable(req isc.Request) {
	if len(wr.any) > 0 {
		awaiting := wr.any // copy before resetting, so that if any of the callbacks tries to await, it doesn't get squashed
		wr.any = wr.any[:0]
		for i := range awaiting {
			if awaiting[i].ctx.Err() == nil {
				awaiting[i].cb(req)
			}
		}
	}
	reqRefKey := isc.RequestRefFromRequest(req).AsKey()
	if cbs, exists := wr.reqs.Get(reqRefKey); exists {
		for i := range cbs {
			if cbs[i].ctx.Err() == nil {
				cbs[i].cb(req)
			}
		}
		wr.reqs.Delete(reqRefKey)
	}
}

func (wr *waitReq) maybeCleanup() {
	wr.cleanupCounter++
	if wr.cleanupCounter > 0 {
		return
	}
	wr.cleanupCounter = wr.cleanupEvery
	//
	// We only care about the requests for particular requests.
	// Requests for any will be cleaned up on first request.
	wr.reqs.ForEach(func(reqRef isc.RequestRefKey, items []*waitReqItem) bool {
		newItems := []*waitReqItem{}
		for i := range items {
			if items[i].ctx.Err() == nil {
				newItems = append(newItems, items[i])
			}
		}
		if len(newItems) != 0 {
			wr.reqs.Set(reqRef, newItems)
		} else {
			wr.reqs.Delete(reqRef)
		}

		return true
	})
}
