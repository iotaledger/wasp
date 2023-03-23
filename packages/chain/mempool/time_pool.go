// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"time"

	"golang.org/x/exp/slices"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/wasp/packages/isc"
)

// Maintains a pool of requests that have to be postponed until specified timestamp.
type TimePool interface {
	AddRequest(timestamp time.Time, request isc.Request)
	TakeTill(timestamp time.Time) []isc.Request
	Has(reqID *isc.RequestRef) bool
	Filter(predicate func(request isc.Request, ts time.Time) bool)
}

// Here we implement TimePool. We maintain the request in a list ordered by a timestamp.
// The list is organized in slots. Each slot contains a list of requests that fit to the
// slot boundaries.
type timePoolImpl struct {
	requests *shrinkingmap.ShrinkingMap[isc.RequestRefKey, isc.Request] // All the requests in this pool.
	slots    *timeSlot                                                  // Structure to fetch them fast by their time.
}

type timeSlot struct {
	from time.Time
	till time.Time
	reqs *shrinkingmap.ShrinkingMap[time.Time, []isc.Request]
	next *timeSlot
}

const slotPrecision = time.Minute

var _ TimePool = &timePoolImpl{}

func NewTimePool() TimePool {
	return &timePoolImpl{
		requests: shrinkingmap.New[isc.RequestRefKey, isc.Request](),
		slots:    nil,
	}
}

func (tpi *timePoolImpl) AddRequest(timestamp time.Time, request isc.Request) {
	reqRefKey := isc.RequestRefFromRequest(request).AsKey()

	if tpi.requests.Has(reqRefKey) {
		return
	}

	tpi.requests.Set(reqRefKey, request)
	reqFrom, reqTill := tpi.timestampSlotBounds(timestamp)
	prevNext := &tpi.slots
	for slot := tpi.slots; ; {
		if slot == nil || slot.from.After(reqFrom) { // Add new slot (append or insert).
			newRequests := shrinkingmap.New[time.Time, []isc.Request]()
			newRequests.Set(timestamp, []isc.Request{request})

			newSlot := &timeSlot{
				from: reqFrom,
				till: reqTill,
				reqs: newRequests,
				next: slot,
			}
			*prevNext = newSlot
			return
		}
		if slot.from == reqFrom { // Add to existing slot.
			requests, _ := slot.reqs.GetOrCreate(timestamp, func() []isc.Request { return make([]isc.Request, 0, 1) })
			slot.reqs.Set(timestamp, append(requests, request))
			return
		}
		prevNext = &slot.next
		slot = slot.next
	}
}

func (tpi *timePoolImpl) TakeTill(timestamp time.Time) []isc.Request {
	resp := []isc.Request{}
	for slot := tpi.slots; slot != nil; slot = slot.next {
		if slot.from.After(timestamp) {
			break
		}
		slot.reqs.ForEach(func(ts time.Time, tsReqs []isc.Request) bool {
			if ts == timestamp || ts.Before(timestamp) {
				resp = append(resp, tsReqs...)
				for _, req := range tsReqs {
					reqRefKey := isc.RequestRefFromRequest(req).AsKey()
					tpi.requests.Delete(reqRefKey)
				}
				slot.reqs.Delete(ts)
			}
			return true
		})
		if slot.reqs.Size() != 0 {
			break
		}

		tpi.slots = slot.next
	}
	return resp
}

func (tpi *timePoolImpl) Has(reqRef *isc.RequestRef) bool {
	return tpi.requests.Has(reqRef.AsKey())
}

func (tpi *timePoolImpl) Filter(predicate func(request isc.Request, ts time.Time) bool) {
	prevNext := &tpi.slots
	for slot := tpi.slots; slot != nil; slot = slot.next {
		slot.reqs.ForEach(func(ts time.Time, tsReqs []isc.Request) bool {
			requests := tsReqs
			for i, req := range requests {
				if !predicate(req, ts) {
					requests = slices.Delete(requests, i, i+1)
				}
			}

			if len(requests) != 0 {
				slot.reqs.Set(ts, slices.Clip(requests))
			} else {
				slot.reqs.Delete(ts)
			}

			return true
		})

		if slot.reqs.Size() == 0 {
			// Drop the current slot, if it is empty, keep the prevNext the same.
			*prevNext = slot.next
		} else {
			prevNext = &slot.next
		}
	}
}

func (tpi *timePoolImpl) timestampSlotBounds(timestamp time.Time) (time.Time, time.Time) {
	from := timestamp.Truncate(slotPrecision)
	if timestamp == from {
		return from, from
	}
	return from, from.Add(slotPrecision)
}
