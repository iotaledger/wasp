package smGPA

import (
	"github.com/iotaledger/wasp/packages/state"
)

type blockFetchersImpl struct {
	fetchers map[state.BlockHash]blockFetcher
}

var _ blockFetchers = &blockFetchersImpl{}

func newBlockFetchers() blockFetchers {
	return &blockFetchersImpl{fetchers: make(map[state.BlockHash]blockFetcher)}
}

func (bfiT *blockFetchersImpl) getSize() int {
	return len(bfiT.fetchers)
}

func (bfiT *blockFetchersImpl) addFetcher(fetcher blockFetcher) {
	bfiT.fetchers[fetcher.getCommitment().BlockHash()] = fetcher
}

func (bfiT *blockFetchersImpl) takeFetcher(commitment *state.L1Commitment) blockFetcher {
	fetcher, ok := bfiT.fetchers[commitment.BlockHash()]
	if !ok {
		return nil
	}
	delete(bfiT.fetchers, commitment.BlockHash())
	return fetcher
}

func (bfiT *blockFetchersImpl) addCallback(commitment *state.L1Commitment, callback blockRequestCallback) bool {
	fetcher, ok := bfiT.fetchers[commitment.BlockHash()]
	if !ok {
		return false
	}
	fetcher.addCallback(callback)
	return true
}

func (bfiT *blockFetchersImpl) addRelatedFetcher(commitment *state.L1Commitment, relatedFetcher blockFetcher) bool {
	fetcher, ok := bfiT.fetchers[commitment.BlockHash()]
	if !ok {
		return false
	}
	fetcher.addRelatedFetcher(relatedFetcher)
	return true
}

func (bfiT *blockFetchersImpl) getCommitments() []*state.L1Commitment {
	result := make([]*state.L1Commitment, bfiT.getSize())
	i := 0
	for _, fetcher := range bfiT.fetchers {
		result[i] = fetcher.getCommitment()
		i++
	}
	return result
}

func (bfiT *blockFetchersImpl) cleanCallbacks() {
	for _, fetcher := range bfiT.fetchers {
		fetcher.cleanCallbacks()
	}
}
