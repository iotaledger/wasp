// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"context"

	"github.com/iotaledger/wasp/packages/isc"
)

// This object provides the synchronization between the functions asking for requests
// and the arriving requests. The former have to be notified upon reception of the latter.
type WaitReq interface {
	WaitMany(ctx context.Context, reqRefs []*isc.RequestRef, cb func(req isc.Request)) // Called per block.
	WaitAny(ctx context.Context, cb func(req isc.Request))                             // Called per block.
	Have(req isc.Request)                                                              // Called often, per request.
}

type waitReq struct {
	reqs           map[isc.RequestRefKey][]*waitReqItem // Wait for specific req.
	any            []*waitReqItem                       // Wait for any req.
	cleanupCounter int                                  // Perform cleanup when the counter goes to 0.
	cleanupEvery   int
}

type waitReqItem struct {
	ctx context.Context
	cb  func(req isc.Request)
}

func NewWaitReq(cleanupEvery int) WaitReq {
	return &waitReq{
		reqs:           map[isc.RequestRefKey][]*waitReqItem{},
		any:            []*waitReqItem{},
		cleanupCounter: cleanupEvery,
		cleanupEvery:   cleanupEvery,
	}
}

func (wr *waitReq) WaitMany(ctx context.Context, reqRefs []*isc.RequestRef, cb func(req isc.Request)) {
	for i := range reqRefs {
		reqRefKey := reqRefs[i].AsKey()
		item := &waitReqItem{ctx: ctx, cb: cb}
		if _, ok := wr.reqs[reqRefKey]; !ok {
			wr.reqs[reqRefKey] = []*waitReqItem{item}
		} else {
			wr.reqs[reqRefKey] = append(wr.reqs[reqRefKey], item)
		}
	}
	wr.maybeCleanup()
}

func (wr *waitReq) WaitAny(ctx context.Context, cb func(req isc.Request)) {
	wr.any = append(wr.any, &waitReqItem{ctx: ctx, cb: cb})
	wr.maybeCleanup()
}

func (wr *waitReq) Have(req isc.Request) {
	if len(wr.any) > 0 {
		for i := range wr.any {
			if wr.any[i].ctx.Err() == nil {
				wr.any[i].cb(req)
			}
		}
		wr.any = wr.any[:0]
	}
	reqRefKey := isc.RequestRefFromRequest(req).AsKey()
	if cbs, ok := wr.reqs[reqRefKey]; ok {
		for i := range cbs {
			if cbs[i].ctx.Err() == nil {
				cbs[i].cb(req)
			}
		}
		delete(wr.reqs, reqRefKey)
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
	for reqRef, items := range wr.reqs {
		newItems := []*waitReqItem{}
		for i := range items {
			if items[i].ctx.Err() == nil {
				newItems = append(newItems, items[i])
			}
		}
		if len(newItems) == 0 {
			delete(wr.reqs, reqRef)
		} else {
			wr.reqs[reqRef] = newItems
		}
	}
}
