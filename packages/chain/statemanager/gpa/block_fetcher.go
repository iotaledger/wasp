package gpa

import (
	"time"

	"github.com/iotaledger/wasp/v2/packages/state"
)

type blockFetcherImpl struct {
	start      time.Time
	stateIndex uint32
	commitment *state.L1Commitment
	callbacks  []blockRequestCallback
	related    []blockFetcher
}

var _ blockFetcher = &blockFetcherImpl{}

func newBlockFetcher(stateIndex uint32, commitment *state.L1Commitment) blockFetcher {
	return &blockFetcherImpl{
		start:      time.Now(),
		stateIndex: stateIndex,
		commitment: commitment,
		callbacks:  make([]blockRequestCallback, 0),
		related:    make([]blockFetcher, 0),
	}
}

func newBlockFetcherWithCallback(stateIndex uint32, commitment *state.L1Commitment, callback blockRequestCallback) blockFetcher {
	result := newBlockFetcher(stateIndex, commitment)
	result.addCallback(callback)
	return result
}

func newBlockFetcherWithRelatedFetcher(commitment *state.L1Commitment, fetcher blockFetcher) blockFetcher {
	newStateIndex := fetcher.getStateIndex() - 1
	result := newBlockFetcher(newStateIndex, commitment)
	result.addRelatedFetcher(fetcher)
	return result
}

func (bfiT *blockFetcherImpl) getStateIndex() uint32 {
	return bfiT.stateIndex
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

func (bfiT *blockFetcherImpl) notifyFetched() []blockFetcher {
	for _, callback := range bfiT.callbacks {
		if callback.isValid() {
			callback.requestCompleted()
		}
	}
	return bfiT.related
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
