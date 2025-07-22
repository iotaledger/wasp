package metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

type ChainPipeMetricsProvider struct {
	// We use Func variant of a metric here, thus we register them
	// explicitly when they are created. Therefore we need a registry here.
	reg prometheus.Registerer
}

func newChainPipeMetricsProvider() *ChainPipeMetricsProvider {
	return &ChainPipeMetricsProvider{}
}

func (p *ChainPipeMetricsProvider) register(reg prometheus.Registerer) {
	p.reg = reg
}

func (p *ChainPipeMetricsProvider) createForChain(chainID isc.ChainID) *ChainPipeMetrics {
	return &ChainPipeMetrics{
		chainID:    chainID,
		reg:        p.reg,
		lenMetrics: map[string]prometheus.Collector{},
		maxMetrics: map[string]*chainPipeMaxCollector{},
	}
}

type ChainPipeMetrics struct {
	reg        prometheus.Registerer
	chainID    isc.ChainID
	lenMetrics map[string]prometheus.Collector
	maxMetrics map[string]*chainPipeMaxCollector
	regLock    sync.RWMutex
}

type chainPipeMaxCollector struct {
	collector  prometheus.Collector
	valueFuncs map[string]func() int
}

func (m *ChainPipeMetrics) cleanup() {
	m.regLock.Lock()
	defer m.regLock.Unlock()

	if m.reg == nil {
		return
	}

	for _, collector := range m.lenMetrics {
		m.reg.Unregister(collector)
	}
	m.lenMetrics = map[string]prometheus.Collector{}

	for _, maxCollector := range m.maxMetrics {
		m.reg.Unregister(maxCollector.collector)
	}
	m.maxMetrics = map[string]*chainPipeMaxCollector{}
}

func (m *ChainPipeMetrics) makeLabels(pipeName string) prometheus.Labels {
	return prometheus.Labels{
		labelNameChain:    m.chainID.String(),
		labelNamePipeName: pipeName,
	}
}

func (m *ChainPipeMetrics) TrackPipeLen(name string, lenFunc func() int) {
	m.regLock.Lock()
	defer m.regLock.Unlock()

	if m.reg == nil {
		return
	}
	if oldCollector, ok := m.lenMetrics[name]; ok {
		m.reg.Unregister(oldCollector)
	}

	collector := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace:   "iota_wasp",
		Subsystem:   "chain_pipe",
		Name:        "len",
		Help:        "Length of a pipe",
		ConstLabels: m.makeLabels(name),
	}, func() float64 { return float64(lenFunc()) })
	m.lenMetrics[name] = collector

	if err := m.reg.Register(collector); err != nil {
		panic(fmt.Errorf("failed to register pipe %v len metric for chain %v: %w", name, m.chainID, err))
	}
}

func (m *ChainPipeMetrics) TrackPipeLenMax(name string, key string, lenFunc func() int) {
	m.regLock.Lock()
	defer m.regLock.Unlock()

	if m.reg == nil {
		return
	}

	maxCollector, found := m.maxMetrics[name]
	if !found {
		valueFuncs := map[string]func() int{}
		collector := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "iota_wasp",
			Subsystem:   "chain_pipe",
			Name:        "len",
			Help:        "Length of a pipe",
			ConstLabels: m.makeLabels(name),
		}, func() float64 {
			maxVal := 0
			for _, f := range valueFuncs {
				fVal := f()
				if maxVal < fVal {
					maxVal = fVal
				}
			}
			return float64(maxVal)
		})
		if err := m.reg.Register(collector); err != nil {
			panic(fmt.Errorf("failed to register pipe %v max len metric for chain %v: %w", name, m.chainID, err))
		}
		maxCollector = &chainPipeMaxCollector{
			collector:  collector,
			valueFuncs: valueFuncs,
		}
		m.maxMetrics[name] = maxCollector
	}
	maxCollector.valueFuncs[key] = lenFunc
}

func (m *ChainPipeMetrics) ForgetPipeLenMax(name string, key string) {
	m.regLock.Lock()
	defer m.regLock.Unlock()

	if m.reg == nil {
		return
	}
	maxCollector, found := m.maxMetrics[name]
	if !found {
		return
	}
	delete(maxCollector.valueFuncs, key)
}
