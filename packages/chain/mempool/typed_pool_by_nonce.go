// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/logger"
	consGR "github.com/iotaledger/wasp/packages/chain/cons/cons_gr"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// keeps a map of requests ordered by nonce for each account
type TypedPoolByNonce[V isc.OffLedgerRequest] struct {
	waitReq WaitReq
	refLUT  *shrinkingmap.ShrinkingMap[isc.RequestRefKey, *OrderedPoolEntry[V]]
	// reqsByAcountOrdered keeps an ordered map of reqsByAcountOrdered for each account by nonce
	reqsByAcountOrdered *shrinkingmap.ShrinkingMap[string, []*OrderedPoolEntry[V]] // string is isc.AgentID.String()
	sizeMetric          func(int)
	timeMetric          func(time.Duration)
	log                 *logger.Logger
}

func NewTypedPoolByNonce[V isc.OffLedgerRequest](waitReq WaitReq, sizeMetric func(int), timeMetric func(time.Duration), log *logger.Logger) *TypedPoolByNonce[V] {
	return &TypedPoolByNonce[V]{
		waitReq:             waitReq,
		reqsByAcountOrdered: shrinkingmap.New[string, []*OrderedPoolEntry[V]](),
		refLUT:              shrinkingmap.New[isc.RequestRefKey, *OrderedPoolEntry[V]](),
		sizeMetric:          sizeMetric,
		timeMetric:          timeMetric,
		log:                 log,
	}
}

type OrderedPoolEntry[V isc.OffLedgerRequest] struct {
	req         V
	old         bool
	ts          time.Time
	proposedFor []consGR.ConsensusID
}

func (p *TypedPoolByNonce[V]) Has(reqRef *isc.RequestRef) bool {
	return p.refLUT.Has(reqRef.AsKey())
}

func (p *TypedPoolByNonce[V]) Get(reqRef *isc.RequestRef) V {
	entry, exists := p.refLUT.Get(reqRef.AsKey())
	if !exists {
		return *new(V)
	}
	return entry.req
}

func (p *TypedPoolByNonce[V]) Add(request V) {
	ref := isc.RequestRefFromRequest(request)
	entry := &OrderedPoolEntry[V]{req: request, ts: time.Now()}
	account := request.SenderAccount().String()

	if !p.refLUT.Set(ref.AsKey(), entry) {
		p.log.Debugf("NOT ADDED, already exists. reqID: %v as key=%v, senderAccount: ", request.ID(), ref, account)
		return // not added already exists
	}

	defer func() {
		p.log.Debugf("ADD %v as key=%v, senderAccount: %s", request.ID(), ref, account)
		p.sizeMetric(p.refLUT.Size())
		p.waitReq.MarkAvailable(request)
	}()

	reqsForAcount, exists := p.reqsByAcountOrdered.Get(account)
	if !exists {
		// no other requests for this account
		p.reqsByAcountOrdered.Set(account, []*OrderedPoolEntry[V]{entry})
		return
	}

	// add to the account requests, keep the slice ordered

	// find the index where the new entry should be added
	index, exists := slices.BinarySearchFunc(reqsForAcount, entry,
		func(a, b *OrderedPoolEntry[V]) int {
			aNonce := a.req.Nonce()
			bNonce := b.req.Nonce()
			if aNonce == bNonce {
				return 0
			}
			if aNonce > bNonce {
				return 1
			}
			return -1
		},
	)
	if exists {
		// same nonce, mark the existing request with overlapping nonce as "old", place the new one
		// NOTE: do not delete the request here, as it might already be part of an on-going consensus round
		reqsForAcount[index].old = true
	}

	reqsForAcount = append(reqsForAcount, entry) // add to the end of the list (thus extending the array)

	// make room if target position is not at the end
	if index != len(reqsForAcount)+1 {
		copy(reqsForAcount[index+1:], reqsForAcount[index:])
		reqsForAcount[index] = entry
	}
	p.reqsByAcountOrdered.Set(account, reqsForAcount)
}

func (p *TypedPoolByNonce[V]) Remove(request V) {
	refKey := isc.RequestRefFromRequest(request).AsKey()
	entry, exists := p.refLUT.Get(refKey)
	if !exists {
		return // does not exist
	}
	defer func() {
		p.sizeMetric(p.refLUT.Size())
		p.timeMetric(time.Since(entry.ts))
	}()
	if p.refLUT.Delete(refKey) {
		p.log.Debugf("DEL %v as key=%v", request.ID(), refKey)
	}
	account := entry.req.SenderAccount().String()
	reqsByAccount, exists := p.reqsByAcountOrdered.Get(account)
	if !exists {
		p.log.Error("inconsistency trying to DEL %v as key=%v, no request list for account %s", request.ID(), refKey, account)
		return
	}
	// find the request in the accounts map
	indexToDel := slices.IndexFunc(reqsByAccount, func(e *OrderedPoolEntry[V]) bool {
		return refKey == isc.RequestRefFromRequest(e.req).AsKey()
	})
	if indexToDel == -1 {
		p.log.Error("inconsistency trying to DEL %v as key=%v, request not found in list for account %s", request.ID(), refKey, account)
		return
	}
	if len(reqsByAccount) == 1 { // just remove the entire array for the account
		p.reqsByAcountOrdered.Delete(account)
		return
	}
	reqsByAccount[indexToDel] = nil // remove the pointer reference to allow GC of the entry object
	reqsByAccount = slices.Delete(reqsByAccount, indexToDel, indexToDel+1)
	p.reqsByAcountOrdered.Set(account, reqsByAccount)
}

func (p *TypedPoolByNonce[V]) Iterate(f func(account string, requests []*OrderedPoolEntry[V])) {
	p.reqsByAcountOrdered.ForEach(func(acc string, entries []*OrderedPoolEntry[V]) bool {
		f(acc, slices.Clone(entries))
		return true
	})
}

func (p *TypedPoolByNonce[V]) Filter(predicate func(request V, ts time.Time) bool) {
	p.refLUT.ForEach(func(refKey isc.RequestRefKey, entry *OrderedPoolEntry[V]) bool {
		if !predicate(entry.req, entry.ts) {
			p.Remove(entry.req)
		}
		return true
	})
	p.sizeMetric(p.refLUT.Size())
}

func (p *TypedPoolByNonce[V]) StatusString() string {
	return fmt.Sprintf("{|req|=%d}", p.refLUT.Size())
}

func (p *TypedPoolByNonce[V]) WriteContent(w io.Writer) {
	p.reqsByAcountOrdered.ForEach(func(_ string, list []*OrderedPoolEntry[V]) bool {
		for _, entry := range list {
			jsonData, err := isc.RequestToJSON(entry.req)
			if err != nil {
				return false // stop iteration
			}
			_, err = w.Write(codec.EncodeUint32(uint32(len(jsonData))))
			if err != nil {
				return false // stop iteration
			}
			_, err = w.Write(jsonData)
			if err != nil {
				return false // stop iteration
			}
		}
		return true
	})
}
