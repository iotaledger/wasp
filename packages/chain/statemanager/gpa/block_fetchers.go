package gpa

import (
	"time"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/wasp/packages/chain/statemanager/gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/state"
)

type blockFetcherWithTime struct {
	fetcher    blockFetcher
	createTime time.Time
}

type blockFetchersImpl struct {
	fetchers *shrinkingmap.ShrinkingMap[sm_gpa_utils.BlockKey, *blockFetcherWithTime]
	metrics  blockFetchersMetrics
}

var _ blockFetchers = &blockFetchersImpl{}

func newBlockFetchers(metrics blockFetchersMetrics) blockFetchers {
	return &blockFetchersImpl{
		fetchers: shrinkingmap.New[sm_gpa_utils.BlockKey, *blockFetcherWithTime](),
		metrics:  metrics,
	}
}

func (bfiT *blockFetchersImpl) getSize() int {
	return bfiT.fetchers.Size()
}

func (bfiT *blockFetchersImpl) getCallbacksCount() int {
	result := 0
	bfiT.fetchers.ForEach(func(_ sm_gpa_utils.BlockKey, fetcherWithTime *blockFetcherWithTime) bool {
		result += fetcherWithTime.fetcher.getCallbacksCount()
		return true
	})
	return result
}

func (bfiT *blockFetchersImpl) addFetcher(fetcher blockFetcher) {
	key := sm_gpa_utils.NewBlockKey(fetcher.getCommitment())
	_, exists := bfiT.fetchers.Get(key)
	if !exists {
		bfiT.metrics.inc()
	}
	bfiT.fetchers.Set(key, &blockFetcherWithTime{
		fetcher:    fetcher,
		createTime: time.Now(),
	})
}

func (bfiT *blockFetchersImpl) takeFetcher(commitment *state.L1Commitment) blockFetcher {
	blockKey := sm_gpa_utils.NewBlockKey(commitment)
	fetcherWithTime, exists := bfiT.fetchers.Get(blockKey)
	if !exists {
		return nil
	}
	bfiT.fetchers.Delete(blockKey)
	bfiT.metrics.dec()
	bfiT.metrics.duration(time.Since(fetcherWithTime.createTime))
	return fetcherWithTime.fetcher
}

func (bfiT *blockFetchersImpl) addCallback(commitment *state.L1Commitment, callback blockRequestCallback) bool {
	fetcherWithTime, exists := bfiT.fetchers.Get(sm_gpa_utils.NewBlockKey(commitment))
	if !exists {
		return false
	}
	fetcherWithTime.fetcher.addCallback(callback)
	return true
}

func (bfiT *blockFetchersImpl) addRelatedFetcher(commitment *state.L1Commitment, relatedFetcher blockFetcher) bool {
	fetcherWithTime, exists := bfiT.fetchers.Get(sm_gpa_utils.NewBlockKey(commitment))
	if !exists {
		return false
	}
	fetcherWithTime.fetcher.addRelatedFetcher(relatedFetcher)
	return true
}

func (bfiT *blockFetchersImpl) getCommitments() []*state.L1Commitment {
	result := make([]*state.L1Commitment, bfiT.getSize())
	i := 0
	bfiT.fetchers.ForEach(func(_ sm_gpa_utils.BlockKey, fetcherWithTime *blockFetcherWithTime) bool {
		result[i] = fetcherWithTime.fetcher.getCommitment()
		i++
		return true
	})
	return result
}

func (bfiT *blockFetchersImpl) cleanCallbacks() {
	bfiT.fetchers.ForEach(func(_ sm_gpa_utils.BlockKey, fetcherWithTime *blockFetcherWithTime) bool {
		fetcherWithTime.fetcher.cleanCallbacks()
		return true
	})
}
