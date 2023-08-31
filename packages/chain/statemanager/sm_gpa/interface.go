package sm_gpa

import (
	"time"

	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_snapshots"
	"github.com/iotaledger/wasp/packages/state"
)

type StateManagerOutput interface {
	addBlockCommitted(uint32, *state.L1Commitment)
	TakeBlocksCommitted() []sm_snapshots.SnapshotInfo
}

type SnapshotExistsFun func(uint32, *state.L1Commitment) bool

type blockRequestCallback interface {
	isValid() bool
	requestCompleted()
}

type blockFetcher interface {
	getStateIndex() uint32
	getCommitment() *state.L1Commitment
	getCallbacksCount() int
	commitAndNotifyFetched(func(blockFetcher) bool) // calls fun for this block, notifies waiting callbacks of this fetcher and does the same for each related fetcher recursively; fun for parent block is always called before fun for related block
	notifyFetched(func(blockFetcher) bool)          // notifies waiting callbacks of this fetcher, then calls fun and notifies waiting callbacks of all related fetchers recursively; fun for parent block is always called before fun for related block
	addCallback(blockRequestCallback)
	addRelatedFetcher(blockFetcher)
	cleanCallbacks()
}

type blockFetchers interface {
	getSize() int
	getCallbacksCount() int
	addFetcher(blockFetcher)
	takeFetcher(*state.L1Commitment) blockFetcher
	addCallback(*state.L1Commitment, blockRequestCallback) bool
	addRelatedFetcher(*state.L1Commitment, blockFetcher) bool
	getCommitments() []*state.L1Commitment
	cleanCallbacks()
}

type blockFetchersMetrics interface {
	inc()
	dec()
	duration(time.Duration)
}
