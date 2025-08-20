package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

type ChainStateMetricsProvider struct {
	blockCommitTimes            *prometheus.HistogramVec
	blockCommitNewTrieNodes     *prometheus.CounterVec
	blockCommitNewTrieValues    *prometheus.CounterVec
	blockPruneTimes             *prometheus.HistogramVec
	blockPruneDeletedTrieNodes  *prometheus.CounterVec
	blockPruneDeletedTrieValues *prometheus.CounterVec
}

func newChainStateMetricsProvider() *ChainStateMetricsProvider {
	return &ChainStateMetricsProvider{
		blockCommitTimes: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "state",
			Name:      "state_block_commit_times",
			Help:      "Time elapsed (s) committing blocks",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		blockCommitNewTrieNodes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "state",
			Name:      "state_block_commit_new_trie_nodes",
			Help:      "Newly created trie nodes",
		}, []string{labelNameChain}),
		blockCommitNewTrieValues: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "state",
			Name:      "state_block_commit_new_trie_values",
			Help:      "Newly created trie values",
		}, []string{labelNameChain}),
		blockPruneTimes: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "state",
			Name:      "state_block_prune_times",
			Help:      "Time elapsed (s) pruning blocks",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		blockPruneDeletedTrieNodes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "state",
			Name:      "state_block_prune_deleted_trie_nodes",
			Help:      "Deleted trie nodes",
		}, []string{labelNameChain}),
		blockPruneDeletedTrieValues: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "state",
			Name:      "state_block_prune_deleted_trie_values",
			Help:      "Deleted trie values",
		}, []string{labelNameChain}),
	}
}

func (p *ChainStateMetricsProvider) register(reg prometheus.Registerer) {
	reg.MustRegister(
		p.blockCommitTimes,
		p.blockCommitNewTrieNodes,
		p.blockCommitNewTrieValues,
		p.blockPruneTimes,
		p.blockPruneDeletedTrieNodes,
		p.blockPruneDeletedTrieValues,
	)
}

func (p *ChainStateMetricsProvider) createForChain(chainID isc.ChainID) *ChainStateMetrics {
	return &ChainStateMetrics{
		collector: p,
		chainID:   chainID,
	}
}

type ChainStateMetrics struct {
	collector *ChainStateMetricsProvider
	chainID   isc.ChainID
}

func (m *ChainStateMetrics) BlockCommitted(elapsed time.Duration, refcountsEnabled bool, createdNodes, createdValues uint) {
	labels := getChainLabels(m.chainID)
	m.collector.blockCommitTimes.With(labels).Observe(elapsed.Seconds())
	if refcountsEnabled {
		m.collector.blockCommitNewTrieNodes.With(labels).Add(float64(createdNodes))
		m.collector.blockCommitNewTrieValues.With(labels).Add(float64(createdValues))
	}
}

func (m *ChainStateMetrics) BlockPruned(elapsed time.Duration, deletedNodes, deletedValues uint) {
	labels := getChainLabels(m.chainID)
	m.collector.blockPruneTimes.With(labels).Observe(elapsed.Seconds())
	m.collector.blockPruneDeletedTrieNodes.With(labels).Add(float64(deletedNodes))
	m.collector.blockPruneDeletedTrieValues.With(labels).Add(float64(deletedValues))
}
