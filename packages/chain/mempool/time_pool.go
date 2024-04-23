// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"slices"
	"time"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/isc"
)

// Maintains a pool of requests that have to be postponed until specified timestamp.
type TimePool interface {
	AddRequest(timestamp time.Time, request isc.OnLedgerRequest)
	TakeTill(timestamp time.Time) []isc.OnLedgerRequest
	Has(reqID *isc.RequestRef) bool
	Cleanup(predicate func(request isc.OnLedgerRequest, ts time.Time) bool)
	ShouldRefreshRequests() bool
}

// Here we implement TimePool. We maintain the request in a list ordered by a timestamp.
// The list is organized in slots. Each slot contains a list of requests that fit to the
// slot boundaries.
type timePoolImpl struct {
	requests           *shrinkingmap.ShrinkingMap[isc.RequestRefKey, isc.OnLedgerRequest] // All the requests in this pool.
	slots              *timeSlot                                                          // Structure to fetch them fast by their time.
	hasDroppedRequests bool
	maxPoolSize        int
	sizeMetric         func(int)
	log                *logger.Logger
}

type timeSlot struct {
	from time.Time
	till time.Time
	reqs *shrinkingmap.ShrinkingMap[time.Time, []isc.OnLedgerRequest]
	next *timeSlot
}

const slotPrecision = time.Minute

var _ TimePool = &timePoolImpl{}

func NewTimePool(maxTimedInPool int, sizeMetric func(int), log *logger.Logger) TimePool {
	return &timePoolImpl{
		requests:           shrinkingmap.New[isc.RequestRefKey, isc.OnLedgerRequest](),
		slots:              nil,
		hasDroppedRequests: false,
		maxPoolSize:        maxTimedInPool,
		sizeMetric:         sizeMetric,
		log:                log,
	}
}

func (tpi *timePoolImpl) AddRequest(timestamp time.Time, request isc.OnLedgerRequest) {
	reqRefKey := isc.RequestRefFromRequest(request).AsKey()

	if tpi.requests.Has(reqRefKey) {
		return
	}

	if !tpi.requests.Set(reqRefKey, request) {
		return
	}

	reqFrom, reqTill := tpi.timestampSlotBounds(timestamp)
	prevNext := &tpi.slots
	for slot := tpi.slots; ; {
		if slot == nil || slot.from.After(reqFrom) { // Add new slot (append or insert).
			newRequests := shrinkingmap.New[time.Time, []isc.OnLedgerRequest]()
			newRequests.Set(timestamp, []isc.OnLedgerRequest{request})

			newSlot := &timeSlot{
				from: reqFrom,
				till: reqTill,
				reqs: newRequests,
				next: slot,
			}
			*prevNext = newSlot
			break
		}
		if slot.from == reqFrom { // Add to existing slot.
			requests, _ := slot.reqs.GetOrCreate(timestamp, func() []isc.OnLedgerRequest { return make([]isc.OnLedgerRequest, 0, 1) })
			slot.reqs.Set(timestamp, append(requests, request))
			break
		}
		prevNext = &slot.next
		slot = slot.next
	}

	//
	// keep the size of this pool limited
	if tpi.requests.Size() > tpi.maxPoolSize {
		// remove the slot most far out in the future
		var prev *timeSlot
		lastSlot := tpi.slots
		for {
			if lastSlot.next == nil {
				break
			}
			prev = lastSlot
			lastSlot = lastSlot.next
		}

		// remove the link to the lastSlot
		if prev == nil {
			tpi.slots = nil
		} else {
			prev.next = nil
		}

		// delete the requests included in the last slot
		reqsToDelete := lastSlot.reqs.Values()
		for _, reqs := range reqsToDelete {
			for _, req := range reqs {
				rKey := isc.RequestRefFromRequest(req).AsKey()
				tpi.requests.Delete(rKey)
			}
		}
		tpi.hasDroppedRequests = true
	}

	// log and update metrics
	tpi.log.Debugf("ADD %v as key=%v", request.ID(), reqRefKey)
	tpi.sizeMetric(tpi.requests.Size())
}

func (tpi *timePoolImpl) ShouldRefreshRequests() bool {
	if !tpi.hasDroppedRequests {
		return false
	}
	if tpi.requests.Size() > 0 {
		return false // wait until pool is empty to refresh
	}
	// assume after this function returns true, the requests will be refreshed
	tpi.hasDroppedRequests = false
	return true
}

func (tpi *timePoolImpl) TakeTill(timestamp time.Time) []isc.OnLedgerRequest {
	resp := []isc.OnLedgerRequest{}
	for slot := tpi.slots; slot != nil; slot = slot.next {
		if slot.from.After(timestamp) {
			break
		}
		slot.reqs.ForEach(func(ts time.Time, tsReqs []isc.OnLedgerRequest) bool {
			if ts == timestamp || ts.Before(timestamp) {
				resp = append(resp, tsReqs...)
				for _, req := range tsReqs {
					reqRefKey := isc.RequestRefFromRequest(req).AsKey()
					if tpi.requests.Delete(reqRefKey) {
						tpi.log.Debugf("DEL %v as key=%v", req.ID(), reqRefKey)
					}
				}
				tpi.sizeMetric(tpi.requests.Size())
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

func (tpi *timePoolImpl) Cleanup(predicate func(request isc.OnLedgerRequest, ts time.Time) bool) {
	prevNext := &tpi.slots
	for slot := tpi.slots; slot != nil; slot = slot.next {
		slot.reqs.ForEach(func(ts time.Time, tsReqs []isc.OnLedgerRequest) bool {
			requests := tsReqs
			for i, req := range requests {
				if !predicate(req, ts) {
					requests = slices.Delete(requests, i, i+1)
					reqRefKey := isc.RequestRefFromRequest(req).AsKey()
					if tpi.requests.Delete(reqRefKey) {
						tpi.log.Debugf("DEL %v as key=%v", req.ID(), reqRefKey)
					}
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
	tpi.sizeMetric(tpi.requests.Size())
}

func (tpi *timePoolImpl) timestampSlotBounds(timestamp time.Time) (time.Time, time.Time) {
	from := timestamp.Truncate(slotPrecision)
	if timestamp == from {
		return from, from
	}
	return from, from.Add(slotPrecision)
}
