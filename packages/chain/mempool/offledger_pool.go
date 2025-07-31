// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"fmt"
	"math/big"
	"slices"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/log"

	consGR "github.com/iotaledger/wasp/v2/packages/chain/cons/gr"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

// OffLedgerPool keeps a map of requests ordered by nonce for each account
type OffLedgerPool struct {
	waitReq WaitReq
	refLUT  *shrinkingmap.ShrinkingMap[isc.RequestRefKey, *OrderedPoolEntry]
	// reqsByAcountOrdered keeps an ordered map of reqsByAcountOrdered for each account by nonce
	reqsByAcountOrdered *shrinkingmap.ShrinkingMap[string, []*OrderedPoolEntry] // string is isc.AgentID.String()
	// orderedByGasPrice keeps a list ordered by the highest gas price
	orderedByGasPrice []*OrderedPoolEntry // TODO use a better data structure instead!!! (probably RedBlackTree)
	minGasPrice       *big.Int
	maxPoolSize       int
	maxPerAccount     int
	sizeMetric        func(int)
	timeMetric        func(time.Duration)
	log               log.Logger
}

func NewOffledgerPool(maxPoolSize int, maxPerAccount int, waitReq WaitReq, sizeMetric func(int), timeMetric func(time.Duration), log log.Logger) *OffLedgerPool {
	return &OffLedgerPool{
		waitReq:             waitReq,
		refLUT:              shrinkingmap.New[isc.RequestRefKey, *OrderedPoolEntry](),
		reqsByAcountOrdered: shrinkingmap.New[string, []*OrderedPoolEntry](),
		orderedByGasPrice:   []*OrderedPoolEntry{},
		minGasPrice:         big.NewInt(1),
		maxPoolSize:         maxPoolSize,
		maxPerAccount:       maxPerAccount,
		sizeMetric:          sizeMetric,
		timeMetric:          timeMetric,
		log:                 log,
	}
}

type OrderedPoolEntry struct {
	req         isc.OffLedgerRequest
	old         bool
	ts          time.Time
	proposedFor []consGR.ConsensusID
}

func (ope *OrderedPoolEntry) markProposed(consID consGR.ConsensusID) {
	ope.proposedFor = append(ope.proposedFor, consID)
}

func (ope *OrderedPoolEntry) proposedForAny(consInsts []consGR.ConsensusID) bool {
	return lo.Some(consInsts, ope.proposedFor)
}

func cmpOrderedPoolEntryByNonce(a, b *OrderedPoolEntry) int {
	aNonce := a.req.Nonce()
	bNonce := b.req.Nonce()
	if aNonce == bNonce {
		return 0
	}
	if aNonce > bNonce {
		return 1
	}
	return -1
}

func (p *OffLedgerPool) Has(reqRef *isc.RequestRef) bool {
	return p.refLUT.Has(reqRef.AsKey())
}

func (p *OffLedgerPool) Get(reqRef *isc.RequestRef) isc.OffLedgerRequest {
	entry, exists := p.refLUT.Get(reqRef.AsKey())
	if !exists {
		return isc.OffLedgerRequest(nil)
	}
	return entry.req
}

func (p *OffLedgerPool) Add(request isc.OffLedgerRequest) bool {
	ref := isc.RequestRefFromRequest(request)
	entry := &OrderedPoolEntry{req: request, ts: time.Now()}
	account := request.SenderAccount().String()

	//
	// add the request to the "request ref" Lookup Table
	if !p.refLUT.Set(ref.AsKey(), entry) {
		p.log.LogDebugf("OffLedger Request NOT ADDED, already exists. reqID: %v as key=%v, senderAccount: %v", request.ID(), ref, account)
		return true // not added already exists
	}

	//
	// add to the account requests, keep the slice ordered
	{
		reqsForAcount, exists := p.reqsByAcountOrdered.Get(account)
		if !exists {
			// no other requests for this account
			p.reqsByAcountOrdered.Set(account, []*OrderedPoolEntry{entry})
		} else {
			// find the index where the new entry should be added
			index, exists := slices.BinarySearchFunc(reqsForAcount, entry, cmpOrderedPoolEntryByNonce)

			if exists && len(reqsForAcount[index].proposedFor) == 0 {
				// Name nonce, but the existing request was not proposed yet.
				// Thus we just replace it.
				oldEntry := reqsForAcount[index]
				reqsForAcount[index] = entry
				p.orderedByGasPrice = lo.Filter(p.orderedByGasPrice, func(e *OrderedPoolEntry, _ int) bool {
					return e.req.ID() != oldEntry.req.ID()
				})
			} else {
				reqsInAccount := len(reqsForAcount)
				if reqsInAccount >= p.maxPerAccount {
					// User has too much requests pending.
					// Reject the new ones, unless they replace the existing ones (the case above).
					p.log.LogDebugf(
						"Not accepting request %v, account %v already has %v requests in mempool.",
						ref, account, reqsInAccount,
					)
					return false
				}

				if exists {
					// same nonce, mark the existing request with overlapping nonce as "old", place the new one
					// NOTE: do not delete the request here, as it might already be part of an on-going consensus round
					reqsForAcount[index].old = true
				}

				// add to the end of the list (thus extending the array)
				reqsForAcount = append(reqsForAcount, entry)

				// make room if target position is not at the end
				if index != len(reqsForAcount)-1 {
					copy(reqsForAcount[index+1:], reqsForAcount[index:])
					reqsForAcount[index] = entry
				}
			}

			p.reqsByAcountOrdered.Set(account, reqsForAcount)
		}
	}

	//
	// add the to the ordered list of requests by gas price
	{
		index, _ := slices.BinarySearchFunc(p.orderedByGasPrice, entry, p.reqSort)
		p.orderedByGasPrice = append(p.orderedByGasPrice, entry)
		// make room if target position is not at the end
		if index != len(p.orderedByGasPrice)-1 {
			copy(p.orderedByGasPrice[index+1:], p.orderedByGasPrice[index:])
			p.orderedByGasPrice[index] = entry
		}
	}

	// keep the pool size in check
	deleted := p.LimitPoolSize()
	if lo.Contains(deleted, entry) {
		// this exact request was deleted from the pool, do not update metrics, or mark available
		p.log.LogDebugf("OffLedger Request NOT ADDED, was removed already. reqID: %v as key=%v, senderAccount: %v", request.ID(), ref, account)
		return false
	}

	//
	// update metrics and signal that the request is available
	p.log.LogDebugf("ADD %v as key=%v, senderAccount: %s", request.ID(), ref, account)
	p.sizeMetric(p.refLUT.Size())
	p.waitReq.MarkAvailable(request)
	return true
}

// LimitPoolSize drops the txs with the lowest price if the total number of requests is too big
func (p *OffLedgerPool) LimitPoolSize() []*OrderedPoolEntry {
	if len(p.orderedByGasPrice) <= p.maxPoolSize {
		return nil // nothing to do
	}

	totalToDelete := len(p.orderedByGasPrice) - p.maxPoolSize
	reqsToDelete := make([]*OrderedPoolEntry, totalToDelete)
	j := 0
	for i := 0; i < len(p.orderedByGasPrice); i++ {
		if len(p.orderedByGasPrice[i].proposedFor) > 0 {
			continue // we cannot drop requests that are being used in current consensus instances
		}
		reqsToDelete[j] = p.orderedByGasPrice[i]
		if j >= totalToDelete-1 {
			break
		}
	}

	for _, r := range reqsToDelete {
		p.log.LogDebugf("LimitPoolSize dropping request: %v", r.req.ID())
		p.Remove(r.req)
	}
	return reqsToDelete
}

func (p *OffLedgerPool) GasPrice(e *OrderedPoolEntry) *big.Int {
	price := e.req.GasPrice()
	if price != nil {
		return price
	}
	// requests without a price specified are assigned the minimum gas price
	return p.minGasPrice
}

func (p *OffLedgerPool) SetMinGasPrice(newPrice *big.Int) {
	if p.minGasPrice.Cmp(newPrice) == 0 {
		// no change
		return
	}
	// update the price and re-order the transactions
	p.minGasPrice = newPrice
	slices.SortFunc(p.orderedByGasPrice, p.reqSort)
}

func (p *OffLedgerPool) reqSort(a, b *OrderedPoolEntry) int {
	cmp := p.GasPrice(a).Cmp(p.GasPrice(b))
	if cmp != 0 {
		return cmp
	}
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

func (p *OffLedgerPool) Remove(request isc.OffLedgerRequest) {
	refKey := isc.RequestRefFromRequest(request).AsKey()
	entry, exists := p.refLUT.Get(refKey)
	if !exists {
		return // does not exist
	}
	defer func() {
		p.sizeMetric(p.refLUT.Size())
		p.timeMetric(time.Since(entry.ts))
	}()

	//
	// delete from the "requests reference" LookupTable
	if p.refLUT.Delete(refKey) {
		p.log.LogDebugf("DEL %v as key=%v", request.ID(), refKey)
	}

	//
	// find the request in the accounts map and delete it
	{
		account := entry.req.SenderAccount().String()
		reqsByAccount, exists := p.reqsByAcountOrdered.Get(account)
		if !exists {
			p.log.LogErrorf("inconsistency trying to DEL %v as key=%v, no request list for account %s", request.ID(), refKey, account)
			return
		}
		indexToDel := slices.IndexFunc(reqsByAccount, func(e *OrderedPoolEntry) bool {
			return refKey == isc.RequestRefFromRequest(e.req).AsKey()
		})
		if indexToDel == -1 {
			p.log.LogErrorf("inconsistency trying to DEL %v as key=%v, request not found in list for account %s", request.ID(), refKey, account)
			return
		}
		if len(reqsByAccount) == 1 { // just remove the entire array for the account
			p.reqsByAcountOrdered.Delete(account)
		} else {
			reqsByAccount[indexToDel] = nil // remove the pointer reference to allow GC of the entry object
			reqsByAccount = slices.Delete(reqsByAccount, indexToDel, indexToDel+1)
			p.reqsByAcountOrdered.Set(account, reqsByAccount)
		}
	}

	//
	// find and delete the request from the gas price ordered list
	{
		indexToDel := slices.IndexFunc(p.orderedByGasPrice, func(e *OrderedPoolEntry) bool {
			return refKey == isc.RequestRefFromRequest(e.req).AsKey()
		})
		p.orderedByGasPrice[indexToDel] = nil // remove the pointer reference to allow GC of the entry object
		p.orderedByGasPrice = slices.Delete(p.orderedByGasPrice, indexToDel, indexToDel+1)
	}
}

func (p *OffLedgerPool) Iterate(f func(account string, requests []*OrderedPoolEntry) bool) {
	p.reqsByAcountOrdered.ForEach(func(acc string, entries []*OrderedPoolEntry) bool {
		return f(acc, slices.Clone(entries))
	})
}

func (p *OffLedgerPool) Cleanup(predicate func(request isc.OffLedgerRequest, ts time.Time) bool) {
	p.refLUT.ForEach(func(refKey isc.RequestRefKey, entry *OrderedPoolEntry) bool {
		if !predicate(entry.req, entry.ts) {
			p.Remove(entry.req)
		}
		return true
	})
	p.sizeMetric(p.refLUT.Size())
}

func (p *OffLedgerPool) StatusString() string {
	return fmt.Sprintf("{|req|=%d}", p.refLUT.Size())
}
