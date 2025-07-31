package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

// ChainMetrics holds all metrics for a single chain
type ChainMetrics struct {
	Pipe         *ChainPipeMetrics
	BlockWAL     *ChainBlockWALMetrics
	CmtLog       *ChainCmtLogMetrics
	Consensus    *ChainConsensusMetrics
	Mempool      *ChainMempoolMetrics
	Message      *ChainMessageMetrics
	StateManager *ChainStateManagerMetrics
	Snapshots    *ChainSnapshotsMetrics
	NodeConn     *ChainNodeConnMetrics
	WebAPI       *ChainWebAPIMetrics
	State        *ChainStateMetrics
}

// ChainMetricsProvider holds all metrics for all chains per chain
type ChainMetricsProvider struct {
	mu     sync.RWMutex
	chains map[isc.ChainID]*ChainMetrics

	Pipe         *ChainPipeMetricsProvider
	BlockWAL     *ChainBlockWALMetricsProvider
	CmtLog       *ChainCmtLogMetricsProvider
	Consensus    *ChainConsensusMetricsProvider
	Mempool      *ChainMempoolMetricsProvider
	Message      *ChainMessageMetricsProvider
	StateManager *ChainStateManagerMetricsProvider
	Snapshots    *ChainSnapshotsMetricsProvider
	NodeConn     *ChainNodeConnMetricsProvider
	WebAPI       *ChainWebAPIMetricsProvider
	State        *ChainStateMetricsProvider
}

func NewChainMetricsProvider() *ChainMetricsProvider {
	return &ChainMetricsProvider{
		chains: make(map[isc.ChainID]*ChainMetrics),

		Pipe:         newChainPipeMetricsProvider(),
		BlockWAL:     newChainBlockWALMetricsProvider(),
		CmtLog:       newChainCmtLogMetricsProvider(),
		Consensus:    newChainConsensusMetricsProvider(),
		Mempool:      newChainMempoolMetricsProvider(),
		Message:      newChainMessageMetricsProvider(),
		StateManager: newChainStateManagerMetricsProvider(),
		Snapshots:    newChainSnapshotsMetricsProvider(),
		NodeConn:     newChainNodeConnMetricsProvider(),
		WebAPI:       NewChainWebAPIMetricsProvider(),
		State:        newChainStateMetricsProvider(),
	}
}

func (m *ChainMetricsProvider) Register(reg prometheus.Registerer) {
	m.Pipe.register(reg)
	m.BlockWAL.register(reg)
	m.CmtLog.register(reg)
	m.Consensus.register(reg)
	m.Mempool.register(reg)
	m.Message.register(reg)
	m.StateManager.register(reg)
	m.Snapshots.register(reg)
	m.NodeConn.register(reg)
	m.WebAPI.register(reg)
	m.State.register(reg)
}

func (m *ChainMetricsProvider) GetChainMetrics(chainID isc.ChainID) *ChainMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cm, ok := m.chains[chainID]; ok {
		return cm
	}
	cm := &ChainMetrics{
		Pipe:         m.Pipe.createForChain(chainID),
		BlockWAL:     m.BlockWAL.createForChain(chainID),
		CmtLog:       m.CmtLog.createForChain(chainID),
		Consensus:    m.Consensus.createForChain(chainID),
		Mempool:      m.Mempool.createForChain(chainID),
		Message:      m.Message.createForChain(chainID),
		StateManager: m.StateManager.createForChain(chainID),
		Snapshots:    m.Snapshots.createForChain(chainID),
		NodeConn:     m.NodeConn.createForChain(chainID),
		WebAPI:       m.WebAPI.CreateForChain(chainID),
		State:        m.State.createForChain(chainID),
	}
	m.chains[chainID] = cm
	return cm
}

func (m *ChainMetricsProvider) RegisterChain(chainID isc.ChainID) {
	m.GetChainMetrics(chainID)
}

func (m *ChainMetricsProvider) UnregisterChain(chainID isc.ChainID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cm, ok := m.chains[chainID]; ok {
		cm.Pipe.cleanup()
		delete(m.chains, chainID)
	}
}

func (m *ChainMetricsProvider) RegisteredChains() []isc.ChainID {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return lo.Keys(m.chains)
}
