// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"fmt"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/logger"
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
	queries        *shrinkingmap.ShrinkingMap[isc.RequestID, []*awaitReceiptReq]
	cleanupCounter int
	cleanupEvery   int
	log            *logger.Logger
}

func NewAwaitReceipt(cleanupEvery int, log *logger.Logger) AwaitReceipt {
	return &awaitReceiptImpl{
		state:          nil,
		queries:        shrinkingmap.New[isc.RequestID, []*awaitReceiptReq](),
		cleanupCounter: cleanupEvery,
		cleanupEvery:   cleanupEvery,
		log:            log,
	}
}

func (ari *awaitReceiptImpl) Await(query *awaitReceiptReq) {
	if ari.state != nil {
		receipt, err := blocklog.GetRequestReceipt(ari.state, query.requestID)
		if err != nil {
			panic(fmt.Errorf("cannot read recept from state: %w", err))
		}
		if receipt != nil {
			query.Respond(receipt)
			return
		}
	}
	ari.maybeCleanup()

	reqAwaits, _ := ari.queries.GetOrCreate(query.requestID, func() []*awaitReceiptReq { return make([]*awaitReceiptReq, 0) })
	ari.queries.Set(query.requestID, append(reqAwaits, query))
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
	ari.queries.ForEach(func(reqID isc.RequestID, reqAwaits []*awaitReceiptReq) bool {
		receipt, err := blocklog.GetRequestReceipt(state, reqID)
		if err != nil {
			panic(fmt.Errorf("cannot read recept from state: %w", err))
		}
		if receipt == nil {
			return true
		}
		for _, reqAwait := range reqAwaits {
			reqAwait.Respond(receipt)
		}

		ari.queries.Delete(reqID)
		return true
	})
}

func (ari *awaitReceiptImpl) respondByBlock(block state.Block) {
	blockReceipts, err := blocklog.RequestReceiptsFromBlock(block)
	if err != nil {
		panic(fmt.Errorf("cannot extract receipts from block: %w", err))
	}
	for _, receipt := range blockReceipts {
		requestID := receipt.Request.ID()
		if reqAwaits, exists := ari.queries.Get(requestID); exists {
			for _, reqAwait := range reqAwaits {
				reqAwait.Respond(receipt)
			}
			ari.queries.Delete(requestID)
		}
	}
}

func (ari *awaitReceiptImpl) StatusString() string {
	return fmt.Sprintf("{|queries|=%v}", ari.queries.Size())
}

func (ari *awaitReceiptImpl) maybeCleanup() {
	ari.cleanupCounter--
	if ari.cleanupCounter > 0 {
		return
	}
	ari.cleanupCounter = ari.cleanupEvery

	ari.queries.ForEach(func(reqID isc.RequestID, reqAwaits []*awaitReceiptReq) bool {
		newAwaits := []*awaitReceiptReq{}
		for _, reqAwait := range reqAwaits {
			if reqAwait.CloseIfCanceled() {
				return true
			}
			newAwaits = append(newAwaits, reqAwait)
		}
		if len(newAwaits) == 0 {
			ari.queries.Delete(reqID)
			return true
		}
		ari.queries.Set(reqID, newAwaits)
		return true
	})
}
