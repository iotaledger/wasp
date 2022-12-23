// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

// AwaitReceipt implements await for request receipts.
type AwaitReceipt interface {
	Await(query *awaitReceiptReq)
	ConsiderState(state state.State, blocksAdded []state.Block) // Respond to all queries, that have receipts in the state.
	StatusString() string
}

type awaitReceiptImpl struct {
	state          state.State
	queries        map[isc.RequestID][]*awaitReceiptReq
	cleanupCounter int
	cleanupEvery   int
}

func NewAwaitReceipt(cleanupEvery int) AwaitReceipt {
	return &awaitReceiptImpl{
		state:          nil,
		queries:        map[isc.RequestID][]*awaitReceiptReq{},
		cleanupCounter: cleanupEvery,
		cleanupEvery:   cleanupEvery,
	}
}

func (ari *awaitReceiptImpl) Await(query *awaitReceiptReq) {
	if ari.state != nil {
		receipt, err := blocklog.GetRequestReceipt(ari.state, &query.requestID)
		if err != nil {
			panic(fmt.Errorf("cannot read recept from state: %w", err))
		}
		if receipt != nil {
			query.Respond(receipt)
			return
		}
	}
	ari.maybeCleanup()
	if reqAwaits, ok := ari.queries[query.requestID]; ok {
		ari.queries[query.requestID] = append(reqAwaits, query)
		return
	}
	ari.queries[query.requestID] = []*awaitReceiptReq{query}
}

func (ari *awaitReceiptImpl) ConsiderState(state state.State, addedBlocks []state.Block) {
	if ari.state == nil {
		ari.respondByState(state)
	} else {
		for _, block := range addedBlocks {
			ari.respondByBlock(block)
		}
	}
	ari.state = state
}

func (ari *awaitReceiptImpl) respondByState(state state.State) {
	for reqID, reqAwaits := range ari.queries {
		receipt, err := blocklog.GetRequestReceipt(state, &reqID)
		if err != nil {
			panic(fmt.Errorf("cannot read recept from state: %w", err))
		}
		if receipt == nil {
			continue
		}
		for _, reqAwait := range reqAwaits {
			reqAwait.Respond(receipt)
		}
		delete(ari.queries, reqID)
	}
}

func (ari *awaitReceiptImpl) respondByBlock(block state.Block) {
	blockReceipts, err := blocklog.RequestReceiptsFromBlock(block)
	if err != nil {
		panic(fmt.Errorf("cannot extract receipts from block: %w", err))
	}
	for _, receipt := range blockReceipts {
		requestID := receipt.Request.ID()
		if reqAwaits, ok := ari.queries[requestID]; ok {
			for _, reqAwait := range reqAwaits {
				reqAwait.Respond(receipt)
			}
			delete(ari.queries, requestID)
		}
	}
}

func (ari *awaitReceiptImpl) StatusString() string {
	return fmt.Sprintf("{|queries|=%v}", len(ari.queries))
}

func (ari *awaitReceiptImpl) maybeCleanup() {
	ari.cleanupCounter--
	if ari.cleanupCounter > 0 {
		return
	}
	ari.cleanupCounter = ari.cleanupEvery
	for reqID, reqAwaits := range ari.queries {
		newAwaits := []*awaitReceiptReq{}
		for _, reqAwait := range reqAwaits {
			if reqAwait.CloseIfCanceled() {
				continue
			}
			newAwaits = append(newAwaits, reqAwait)
		}
		if len(newAwaits) == 0 {
			delete(ari.queries, reqID)
			continue
		}
		ari.queries[reqID] = newAwaits
	}
}
