// Package chaindb provides database functionality and operations for blockchain data storage.
package chaindb

const (
	PrefixBlockByTrieRoot         = 0
	PrefixTrie                    = 1
	PrefixLatestTrieRoot          = 2
	PrefixLargestPrunedBlockIndex = 3
	PrefixHealthTracker           = 255
)
