package chain

// TODO: This package contains a temporary solution. It will not work with rollbacks or reorgs.

import (
	"context"

	"github.com/iotaledger/hive.go/core/generics/lo"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

type WaitConfirmed interface {
	AwaitUntilConfirmed(ctx context.Context, receipt *blocklog.RequestReceipt, responseCh chan<- *blocklog.RequestReceipt)
	StateIndexConfirmed(stateIndex uint32)
}

type waitConfirmedImpl struct {
	lastIndex uint32
	entries   map[uint32][]*waitConfirmedEntry
}

type waitConfirmedEntry struct {
	ctx        context.Context
	receipt    *blocklog.RequestReceipt
	responseCh chan<- *blocklog.RequestReceipt
}

func NewWaitConfirmed() WaitConfirmed {
	return &waitConfirmedImpl{
		lastIndex: 0,
		entries:   map[uint32][]*waitConfirmedEntry{},
	}
}

func (wci *waitConfirmedImpl) AwaitUntilConfirmed(ctx context.Context, receipt *blocklog.RequestReceipt, responseCh chan<- *blocklog.RequestReceipt) {
	if receipt.BlockIndex <= wci.lastIndex {
		responseCh <- receipt
		return
	}
	entry := &waitConfirmedEntry{
		ctx:        ctx,
		receipt:    receipt,
		responseCh: responseCh,
	}
	if entries, ok := wci.entries[receipt.BlockIndex]; ok {
		wci.entries[receipt.BlockIndex] = append(entries, entry)
		return
	}
	wci.entries[receipt.BlockIndex] = []*waitConfirmedEntry{entry}
}

func (wci *waitConfirmedImpl) StateIndexConfirmed(stateIndex uint32) {
	if stateIndex <= wci.lastIndex {
		return
	}
	wci.lastIndex = stateIndex
	for si, siEntries := range wci.entries {
		filteredSiEntries := lo.Filter(siEntries, func(e *waitConfirmedEntry) bool {
			if e.ctx.Err() != nil {
				close(e.responseCh)
				return false
			}
			if si <= stateIndex {
				e.responseCh <- e.receipt
				close(e.responseCh)
				return false
			}
			return true
		})
		if si <= stateIndex {
			delete(wci.entries, si)
		} else {
			wci.entries[si] = filteredSiEntries
		}
	}
}
