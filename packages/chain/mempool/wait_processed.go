// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mempool

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

// WaitProcessed implements await of requests to be processed queried
// by some external clients.
type WaitProcessed interface {
	Await(query *reqAwaitRequestProcessed)
	Processed(receipt *blocklog.RequestReceipt)
	StatusString() string
}

type waitProcessedImpl struct {
	queries        map[isc.RequestID][]*reqAwaitRequestProcessed
	cleanupCounter int
	cleanupEvery   int
}

func NewWaitProcessed(cleanupEvery int) WaitProcessed {
	return &waitProcessedImpl{
		queries:        map[isc.RequestID][]*reqAwaitRequestProcessed{},
		cleanupCounter: cleanupEvery,
		cleanupEvery:   cleanupEvery,
	}
}

func (wpi *waitProcessedImpl) Await(query *reqAwaitRequestProcessed) {
	wpi.maybeCleanup()
	if reqAwaits, ok := wpi.queries[query.requestID]; ok {
		wpi.queries[query.requestID] = append(reqAwaits, query)
		return
	}
	wpi.queries[query.requestID] = []*reqAwaitRequestProcessed{query}
}

func (wpi *waitProcessedImpl) Processed(receipt *blocklog.RequestReceipt) {
	wpi.maybeCleanup()
	requestID := receipt.Request.ID()
	if reqAwaits, ok := wpi.queries[requestID]; ok {
		for _, reqAwait := range reqAwaits {
			reqAwait.responseCh <- receipt
			close(reqAwait.responseCh)
		}
		delete(wpi.queries, requestID)
	}
}

func (wpi *waitProcessedImpl) StatusString() string {
	return fmt.Sprintf("{|queries|=%v}", len(wpi.queries))
}

func (wpi *waitProcessedImpl) maybeCleanup() {
	wpi.cleanupCounter--
	if wpi.cleanupCounter > 0 {
		return
	}
	wpi.cleanupCounter = wpi.cleanupEvery
	for reqID, reqAwaits := range wpi.queries {
		newAwaits := []*reqAwaitRequestProcessed{}
		for _, reqAwait := range reqAwaits {
			if reqAwait.ctx.Err() != nil {
				close(reqAwait.responseCh)
				continue
			}
			newAwaits = append(newAwaits, reqAwait)
		}
		if len(newAwaits) == 0 {
			delete(wpi.queries, reqID)
			continue
		}
		wpi.queries[reqID] = newAwaits
	}
}
