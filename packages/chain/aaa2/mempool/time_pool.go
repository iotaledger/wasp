package mempool

import (
	"time"

	"github.com/iotaledger/wasp/packages/isc"
)

// Maintains a pool of requests that have to be postponed until specified timestamp.
type TimePool interface {
	AddRequest(timestamp time.Time, request isc.Request)
	TakeTill(timestamp time.Time) []isc.Request
}

// Here we implement TimePool. We maintain the request in a list ordered by a timestamp.
// The list is organized in slots. Each slot contains a list of requests that fit to the
// slot boundaries.
type timePoolImpl struct {
	slots *timeSlot
}

type timeSlot struct {
	from time.Time
	till time.Time
	reqs map[time.Time][]isc.Request
	next *timeSlot
}

const slotPrecision = time.Minute

var _ TimePool = &timePoolImpl{}

func NewTimePool() TimePool {
	return &timePoolImpl{
		slots: nil,
	}
}

func (tpi *timePoolImpl) AddRequest(timestamp time.Time, request isc.Request) {
	reqFrom, reqTill := tpi.timestampSlotBounds(timestamp)
	prevNext := &tpi.slots
	for slot := tpi.slots; ; {
		if slot == nil || slot.from.After(reqFrom) { // Add new slot (append or insert).
			newSlot := &timeSlot{
				from: reqFrom,
				till: reqTill,
				reqs: map[time.Time][]isc.Request{timestamp: {request}},
				next: slot,
			}
			*prevNext = newSlot
			return
		}
		if slot.from == reqFrom { // Add to existing slot.
			if _, ok := slot.reqs[timestamp]; !ok {
				slot.reqs[timestamp] = make([]isc.Request, 0, 1)
			}
			slot.reqs[timestamp] = append(slot.reqs[timestamp], request)
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
		for ts, tsReqs := range slot.reqs {
			if ts == timestamp || ts.Before(timestamp) {
				resp = append(resp, tsReqs...)
				delete(slot.reqs, ts)
			}
		}
		if len(slot.reqs) == 0 {
			tpi.slots = slot.next
		} else {
			break
		}
	}
	return resp
}

func (tpi *timePoolImpl) timestampSlotBounds(timestamp time.Time) (time.Time, time.Time) {
	from := timestamp.Truncate(slotPrecision)
	if timestamp == from {
		return from, from
	}
	return from, from.Add(slotPrecision)
}
