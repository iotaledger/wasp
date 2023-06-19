package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type IStateMetrics interface {
	BlockCommitted(elapsed time.Duration, createdNodes, createdValues uint)
	BlockPruned(elapsed time.Duration, deletedNodes, deletedValues uint)
}

var (
	_ IStateMetrics = &emptyStateMetrics{}
	_ IStateMetrics = &stateMetrics{}
)

type emptyStateMetrics struct{}

func NewEmptyStateMetrics() IStateMetrics {
	return &emptyStateMetrics{}
}

func (*emptyStateMetrics) BlockCommitted(elapsed time.Duration, createdNodes, createdValues uint) {
}

func (*emptyStateMetrics) BlockPruned(elapsed time.Duration, deletedNodes, deletedValues uint) {
}

type stateMetrics struct {
	provider *ChainMetricsProvider
	chainID  isc.ChainID
}

func newStateMetrics(provider *ChainMetricsProvider, chainID isc.ChainID) *stateMetrics {
	return &stateMetrics{
		provider: provider,
		chainID:  chainID,
	}
}

func (m *stateMetrics) BlockCommitted(elapsed time.Duration, createdNodes, createdValues uint) {
	labels := getChainLabels(m.chainID)
	m.provider.stateBlockCommitTimes.With(labels).Observe(elapsed.Seconds())
	m.provider.stateBlockCommitNewTrieNodes.With(labels).Add(float64(createdNodes))
	m.provider.stateBlockCommitNewTrieValues.With(labels).Add(float64(createdValues))
}

func (m *stateMetrics) BlockPruned(elapsed time.Duration, deletedNodes, deletedValues uint) {
	labels := getChainLabels(m.chainID)
	m.provider.stateBlockPruneTimes.With(labels).Observe(elapsed.Seconds())
	m.provider.stateBlockPruneDeletedTrieNodes.With(labels).Add(float64(deletedNodes))
	m.provider.stateBlockPruneDeletedTrieValues.With(labels).Add(float64(deletedValues))
}

func initStateMetrics(m *ChainMetricsProvider) {
	m.stateBlockCommitTimes = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "iota_wasp",
		Subsystem: "state",
		Name:      "state_block_commit_times",
		Help:      "Time elapsed (s) committing blocks",
		Buckets:   execTimeBuckets,
	}, []string{labelNameChain})
	m.stateBlockCommitNewTrieNodes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "iota_wasp",
		Subsystem: "state",
		Name:      "state_block_commit_new_trie_nodes",
		Help:      "Newly created trie nodes",
	}, []string{labelNameChain})
	m.stateBlockCommitNewTrieValues = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "iota_wasp",
		Subsystem: "state",
		Name:      "state_block_commit_new_trie_values",
		Help:      "Newly created trie values",
	}, []string{labelNameChain})
	m.stateBlockPruneTimes = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "iota_wasp",
		Subsystem: "state",
		Name:      "state_block_prune_times",
		Help:      "Time elapsed (s) pruning blocks",
		Buckets:   execTimeBuckets,
	}, []string{labelNameChain})
	m.stateBlockPruneDeletedTrieNodes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "iota_wasp",
		Subsystem: "state",
		Name:      "state_block_prune_deleted_trie_nodes",
		Help:      "Deleted trie nodes",
	}, []string{labelNameChain})
	m.stateBlockPruneDeletedTrieValues = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "iota_wasp",
		Subsystem: "state",
		Name:      "state_block_prune_deleted_trie_values",
		Help:      "Deleted trie values",
	}, []string{labelNameChain})
}
