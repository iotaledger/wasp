package smGPA

import (
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/state"
)

type blockFetchersImpl struct {
	fetchers *shrinkingmap.ShrinkingMap[smGPAUtils.BlockKey, blockFetcher]
}

var _ blockFetchers = &blockFetchersImpl{}

func newBlockFetchers() blockFetchers {
	return &blockFetchersImpl{fetchers: shrinkingmap.New[smGPAUtils.BlockKey, blockFetcher]()}
}

func (bfiT *blockFetchersImpl) getSize() int {
	return bfiT.fetchers.Size()
}

func (bfiT *blockFetchersImpl) addFetcher(fetcher blockFetcher) {
	bfiT.fetchers.Set(smGPAUtils.NewBlockKey(fetcher.getCommitment()), fetcher)
}

func (bfiT *blockFetchersImpl) takeFetcher(commitment *state.L1Commitment) blockFetcher {
	blockKey := smGPAUtils.NewBlockKey(commitment)
	fetcher, exists := bfiT.fetchers.Get(blockKey)
	if !exists {
		return nil
	}
	bfiT.fetchers.Delete(blockKey)
	return fetcher
}

func (bfiT *blockFetchersImpl) addCallback(commitment *state.L1Commitment, callback blockRequestCallback) bool {
	fetcher, exists := bfiT.fetchers.Get(smGPAUtils.NewBlockKey(commitment))
	if !exists {
		return false
	}
	fetcher.addCallback(callback)
	return true
}

func (bfiT *blockFetchersImpl) addRelatedFetcher(commitment *state.L1Commitment, relatedFetcher blockFetcher) bool {
	fetcher, exists := bfiT.fetchers.Get(smGPAUtils.NewBlockKey(commitment))
	if !exists {
		return false
	}
	fetcher.addRelatedFetcher(relatedFetcher)
	return true
}

func (bfiT *blockFetchersImpl) getCommitments() []*state.L1Commitment {
	result := make([]*state.L1Commitment, bfiT.getSize())
	i := 0
	bfiT.fetchers.ForEach(func(_ smGPAUtils.BlockKey, fetcher blockFetcher) bool {
		result[i] = fetcher.getCommitment()
		i++
		return true
	})
	return result
}

func (bfiT *blockFetchersImpl) cleanCallbacks() {
	bfiT.fetchers.ForEach(func(_ smGPAUtils.BlockKey, fetcher blockFetcher) bool {
		fetcher.cleanCallbacks()
		return true
	})
}
