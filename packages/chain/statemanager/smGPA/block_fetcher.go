package smGPA

import (
	"time"

	"github.com/iotaledger/wasp/packages/state"
)

type blockFetcherImpl struct {
	start      time.Time
	commitment *state.L1Commitment
	callbacks  []blockRequestCallback
	related    []blockFetcher
}

var _ blockFetcher = &blockFetcherImpl{}

func newBlockFetcher(commitment *state.L1Commitment) blockFetcher {
	return &blockFetcherImpl{
		start:      time.Now(),
		commitment: commitment,
		callbacks:  make([]blockRequestCallback, 0),
		related:    make([]blockFetcher, 0),
	}
}

func newBlockFetcherWithCallback(commitment *state.L1Commitment, callback blockRequestCallback) blockFetcher {
	result := newBlockFetcher(commitment)
	result.addCallback(callback)
	return result
}

func newBlockFetcherWithRelatedFetcher(commitment *state.L1Commitment, fetcher blockFetcher) blockFetcher {
	result := newBlockFetcher(commitment)
	result.addRelatedFetcher(fetcher)
	return result
}

func (bfiT *blockFetcherImpl) getCommitment() *state.L1Commitment {
	return bfiT.commitment
}

func (bfiT *blockFetcherImpl) getCallbacksCount() int {
	return len(bfiT.callbacks)
}

func (bfiT *blockFetcherImpl) addCallback(callback blockRequestCallback) {
	bfiT.callbacks = append(bfiT.callbacks, callback)
}

func (bfiT *blockFetcherImpl) addRelatedFetcher(fetcher blockFetcher) {
	bfiT.related = append(bfiT.related, fetcher)
}

func (bfiT *blockFetcherImpl) notifyFetched(notifyFun func(blockFetcher) bool) {
	if notifyFun(bfiT) {
		for _, callback := range bfiT.callbacks {
			if callback.isValid() {
				callback.requestCompleted()
			}
		}
		for _, fetcher := range bfiT.related {
			fetcher.notifyFetched(notifyFun)
		}
	}
}

func (bfiT *blockFetcherImpl) cleanCallbacks() {
	outI := 0
	for _, callback := range bfiT.callbacks {
		if callback.isValid() { // Callback is valid - keeping it
			bfiT.callbacks[outI] = callback
			outI++
		}
	}
	for i := outI; i < len(bfiT.callbacks); i++ {
		bfiT.callbacks[i] = nil // Not needed callbacks at the end - freeing memory
	}
	bfiT.callbacks = bfiT.callbacks[:outI]
}
