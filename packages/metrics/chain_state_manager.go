package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

type ChainStateManagerMetricsProvider struct {
	// chain state / tips
	chainActiveStateWant    *prometheus.GaugeVec
	chainActiveStateHave    *prometheus.GaugeVec
	chainConfirmedStateWant *prometheus.GaugeVec
	chainConfirmedStateHave *prometheus.GaugeVec
	chainConfirmedStateLag  ChainStateLag

	// state manager
	cacheSize                  *prometheus.GaugeVec
	blocksFetching             *prometheus.GaugeVec
	blocksPending              *prometheus.GaugeVec
	blocksCommitted            *countAndMaxMetrics
	requestsWaiting            *prometheus.GaugeVec
	cspHandlingDuration        *prometheus.HistogramVec
	cdsHandlingDuration        *prometheus.HistogramVec
	cbpHandlingDuration        *prometheus.HistogramVec
	fsdHandlingDuration        *prometheus.HistogramVec
	btcHandlingDuration        *prometheus.HistogramVec
	ttHandlingDuration         *prometheus.HistogramVec
	blockFetchDuration         *prometheus.HistogramVec
	pruningRunDuration         *prometheus.HistogramVec
	pruningSingleStateDuration *prometheus.HistogramVec
	pruningStatesInRun         *prometheus.HistogramVec
	statesPruned               *countAndMaxMetrics
}

//nolint:funlen
func newChainStateManagerMetricsProvider() *ChainStateManagerMetricsProvider {
	return &ChainStateManagerMetricsProvider{
		//
		// chain state / tips.
		//
		chainActiveStateWant: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "chain",
			Name:      "active_state_want",
			Help:      "We try to get blocks till this StateIndex for the active state.",
		}, []string{labelNameChain}),
		chainActiveStateHave: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "chain",
			Name:      "active_state_have",
			Help:      "We received blocks till this StateIndex for the active state.",
		}, []string{labelNameChain}),
		chainConfirmedStateWant: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "chain",
			Name:      "confirmed_state_want",
			Help:      "We try to get blocks till this StateIndex for the confirmed state.",
		}, []string{labelNameChain}),
		chainConfirmedStateHave: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "chain",
			Name:      "confirmed_state_have",
			Help:      "We received blocks till this StateIndex for the confirmed state.",
		}, []string{labelNameChain}),
		//
		// state manager
		//
		cacheSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "cache_size",
			Help:      "Number of blocks stored in cache",
		}, []string{labelNameChain}),
		blocksFetching: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "blocks_fetching",
			Help:      "Number of blocks the node is waiting from other nodes",
		}, []string{labelNameChain}),
		blocksPending: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "blocks_pending",
			Help:      "Number of blocks the node has fetched but hasn't committed, because the node doesn't have their ancestors",
		}, []string{labelNameChain}),
		blocksCommitted: newCountAndMaxMetrics(
			prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "iota_wasp",
				Subsystem: "state_manager",
				Name:      "blocks_committed",
				Help:      "Number of blocks the node has committed to the store",
			}, []string{labelNameChain}),
			prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "iota_wasp",
				Subsystem: "state_manager",
				Name:      "max_blocks_index_committed",
				Help:      "Largest index of block committed to the store",
			}, []string{labelNameChain}),
		),
		requestsWaiting: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "requests_waiting",
			Help:      "Number of requests from other components of the node waiting for response from the state manager. Note that StateDiff request is counted as two requests as it has to obtain two possibly different blocks.",
		}, []string{labelNameChain}),
		cspHandlingDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "consensus_state_proposal_duration",
			Help:      "The duration (s) from starting handling ConsensusStateProposal request till responding to the consensus",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		cdsHandlingDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "consensus_decided_state_duration",
			Help:      "The duration (s) from starting handling ConsensusDecidedState request till responding to the consensus",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		cbpHandlingDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "consensus_block_produced_duration",
			Help:      "The duration (s) from starting till finishing handling ConsensusBlockProduced, which includes responding to the consensus",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		fsdHandlingDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "chain_fetch_state_diff_duration",
			Help:      "The duration (s) from starting handling ChainFetchStateDiff request till responding to the chain",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		btcHandlingDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "self_blocks_to_commit_duration",
			Help:      "The duration (s) from starting handling StateManagerBlocksToCommit request till block is committed and other block(s) are marked to be committed (if needed)",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		ttHandlingDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "timer_tick_duration",
			Help:      "The duration (s) from starting till finishing handling StateManagerTimerTick request",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		blockFetchDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "block_fetch_duration",
			Help:      "The duration (s) from starting fetching block from other till it is received in this node",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		pruningRunDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "pruning_run_duration",
			Help:      "The duration (s) from starting till finishing pruning run, which may include pruning several states from store",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		pruningSingleStateDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "pruning_single_state_duration",
			Help:      "The duration (s) from starting till finishing pruning single state from store",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		pruningStatesInRun: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "state_manager",
			Name:      "pruning_states_in_run",
			Help:      "Number of states pruned in single pruning run (should be 1 in normally running nodes)",
			Buckets:   recCountBuckets,
		}, []string{labelNameChain}),
		statesPruned: newCountAndMaxMetrics(
			prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "iota_wasp",
				Subsystem: "state_manager",
				Name:      "sates_pruned",
				Help:      "Number of states pruned in total since starting the node",
			}, []string{labelNameChain}),
			prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "iota_wasp",
				Subsystem: "state_manager",
				Name:      "max_state_index_pruned",
				Help:      "Largest index of state pruned from the store",
			}, []string{labelNameChain}),
		),
		chainConfirmedStateLag: make(ChainStateLag),
	}
}

func (p *ChainStateManagerMetricsProvider) register(reg prometheus.Registerer) {
	reg.MustRegister(
		p.chainActiveStateWant,
		p.chainActiveStateHave,
		p.chainConfirmedStateWant,
		p.chainConfirmedStateHave,
	)
	reg.MustRegister(
		p.cacheSize,
		p.blocksFetching,
		p.blocksPending,
		p.requestsWaiting,
		p.cspHandlingDuration,
		p.cdsHandlingDuration,
		p.cbpHandlingDuration,
		p.fsdHandlingDuration,
		p.ttHandlingDuration,
		p.blockFetchDuration,
		p.pruningRunDuration,
		p.pruningSingleStateDuration,
		p.pruningStatesInRun,
	)
	reg.MustRegister(p.blocksCommitted.collectors()...)
	reg.MustRegister(p.statesPruned.collectors()...)
}

func (p *ChainStateManagerMetricsProvider) createForChain(chainID isc.ChainID) *ChainStateManagerMetrics {
	labels := getChainLabels(chainID)

	// init values so they appear in prometheus
	p.chainActiveStateWant.With(labels)
	p.chainActiveStateHave.With(labels)
	p.chainConfirmedStateWant.With(labels)
	p.chainConfirmedStateHave.With(labels)

	p.cacheSize.With(labels)
	p.blocksFetching.With(labels)
	p.blocksPending.With(labels)
	p.blocksCommitted.with(labels)
	p.requestsWaiting.With(labels)
	p.cspHandlingDuration.With(labels)
	p.cdsHandlingDuration.With(labels)
	p.cbpHandlingDuration.With(labels)
	p.fsdHandlingDuration.With(labels)
	p.ttHandlingDuration.With(labels)
	p.blockFetchDuration.With(labels)
	p.pruningRunDuration.With(labels)
	p.pruningSingleStateDuration.With(labels)
	p.pruningStatesInRun.With(labels)
	p.statesPruned.with(labels)

	return &ChainStateManagerMetrics{
		chainID:    chainID,
		collectors: p,
		labels:     labels,
	}
}

func (p *ChainStateManagerMetricsProvider) MaxChainConfirmedStateLag() uint32 {
	return p.chainConfirmedStateLag.MaxLag()
}

type ChainStateManagerMetrics struct {
	chainID    isc.ChainID
	labels     prometheus.Labels
	collectors *ChainStateManagerMetricsProvider
}

func (m *ChainStateManagerMetrics) SetChainActiveStateWant(stateIndex uint32) {
	m.collectors.chainActiveStateWant.With(m.labels).Set(float64(stateIndex))
}

func (m *ChainStateManagerMetrics) SetChainActiveStateHave(stateIndex uint32) {
	m.collectors.chainActiveStateHave.With(m.labels).Set(float64(stateIndex))
}

func (m *ChainStateManagerMetrics) SetChainConfirmedStateWant(stateIndex uint32) {
	m.collectors.chainConfirmedStateWant.With(m.labels).Set(float64(stateIndex))
	m.collectors.chainConfirmedStateLag.Want(m.chainID, stateIndex)
}

func (m *ChainStateManagerMetrics) SetChainConfirmedStateHave(stateIndex uint32) {
	m.collectors.chainConfirmedStateHave.With(m.labels).Set(float64(stateIndex))
	m.collectors.chainConfirmedStateLag.Have(m.chainID, stateIndex)
}

func (m *ChainStateManagerMetrics) SetCacheSize(size int) {
	m.collectors.cacheSize.With(m.labels).Set(float64(size))
}

func (m *ChainStateManagerMetrics) IncBlocksFetching() {
	m.collectors.blocksFetching.With(m.labels).Inc()
}

func (m *ChainStateManagerMetrics) DecBlocksFetching() {
	m.collectors.blocksFetching.With(m.labels).Dec()
}

func (m *ChainStateManagerMetrics) IncBlocksPending() {
	m.collectors.blocksPending.With(m.labels).Inc()
}

func (m *ChainStateManagerMetrics) DecBlocksPending() {
	m.collectors.blocksPending.With(m.labels).Dec()
}

func (m *ChainStateManagerMetrics) BlockIndexCommitted(blockIndex uint32) {
	m.collectors.blocksCommitted.countValue(m.labels, float64(blockIndex))
}

func (m *ChainStateManagerMetrics) IncRequestsWaiting() {
	m.collectors.requestsWaiting.With(m.labels).Inc()
}

func (m *ChainStateManagerMetrics) SubRequestsWaiting(count int) {
	m.collectors.requestsWaiting.With(m.labels).Sub(float64(count))
}

func (m *ChainStateManagerMetrics) SetRequestsWaiting(count int) {
	m.collectors.requestsWaiting.With(m.labels).Set(float64(count))
}

func (m *ChainStateManagerMetrics) ConsensusStateProposalHandled(duration time.Duration) {
	m.collectors.cspHandlingDuration.With(m.labels).Observe(duration.Seconds())
}

func (m *ChainStateManagerMetrics) ConsensusDecidedStateHandled(duration time.Duration) {
	m.collectors.cdsHandlingDuration.With(m.labels).Observe(duration.Seconds())
}

func (m *ChainStateManagerMetrics) ConsensusBlockProducedHandled(duration time.Duration) {
	m.collectors.cbpHandlingDuration.With(m.labels).Observe(duration.Seconds())
}

func (m *ChainStateManagerMetrics) ChainFetchStateDiffHandled(duration time.Duration) {
	m.collectors.fsdHandlingDuration.With(m.labels).Observe(duration.Seconds())
}

func (m *ChainStateManagerMetrics) StateManagerBlocksToCommitHandled(duration time.Duration) {
	m.collectors.btcHandlingDuration.With(m.labels).Observe(duration.Seconds())
}

func (m *ChainStateManagerMetrics) StateManagerTimerTickHandled(duration time.Duration) {
	m.collectors.ttHandlingDuration.With(m.labels).Observe(duration.Seconds())
}

func (m *ChainStateManagerMetrics) StateManagerBlockFetched(duration time.Duration) {
	m.collectors.blockFetchDuration.With(m.labels).Observe(duration.Seconds())
}

func (m *ChainStateManagerMetrics) StatePruned(duration time.Duration, stateIndex uint32) {
	m.collectors.pruningSingleStateDuration.With(m.labels).Observe(duration.Seconds())
	m.collectors.statesPruned.countValue(m.labels, float64(stateIndex))
}

func (m *ChainStateManagerMetrics) PruningCompleted(duration time.Duration, statesPruned int) {
	m.collectors.pruningRunDuration.With(m.labels).Observe(duration.Seconds())
	m.collectors.pruningStatesInRun.With(m.labels).Observe(float64(statesPruned))
}
