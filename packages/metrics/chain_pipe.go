package metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type IChainPipeMetrics interface {
	TrackPipeLen(name string, lenFunc func() int)
	TrackPipeLenMax(name string, key string, lenFunc func() int)
	ForgetPipeLenMax(name string, key string)
}

type emptyChainPipeMetrics struct{}

func NewEmptyChainPipeMetrics() IChainPipeMetrics                                            { return &emptyChainPipeMetrics{} }
func (m *emptyChainPipeMetrics) TrackPipeLen(name string, lenFunc func() int)                {}
func (m *emptyChainPipeMetrics) TrackPipeLenMax(name string, key string, lenFunc func() int) {}
func (m *emptyChainPipeMetrics) ForgetPipeLenMax(name string, key string)                    {}

type chainPipeMetrics struct {
	chainID    isc.ChainID
	provider   *ChainMetricsProvider
	lenMetrics map[string]prometheus.Collector
	maxMetrics map[string]*chainPipeMaxCollector
	regLock    *sync.RWMutex
}

type chainPipeMaxCollector struct {
	collector  prometheus.Collector
	valueFuncs map[string]func() int
}

func newChainPipeMetric(provider *ChainMetricsProvider, chainID isc.ChainID) *chainPipeMetrics {
	return &chainPipeMetrics{
		chainID:    chainID,
		provider:   provider,
		lenMetrics: map[string]prometheus.Collector{},
		maxMetrics: map[string]*chainPipeMaxCollector{},
		regLock:    &sync.RWMutex{},
	}
}

func (m *chainPipeMetrics) cleanup() {
	m.regLock.Lock()
	defer m.regLock.Unlock()

	reg := m.provider.pipeLenRegistry
	if reg == nil {
		return
	}

	for _, collector := range m.lenMetrics {
		reg.Unregister(collector)
	}
	m.lenMetrics = map[string]prometheus.Collector{}

	for _, maxCollector := range m.maxMetrics {
		reg.Unregister(maxCollector.collector)
	}
	m.maxMetrics = map[string]*chainPipeMaxCollector{}
}

func (m *chainPipeMetrics) makeLabels(pipeName string) prometheus.Labels {
	return prometheus.Labels{
		labelNameChain:    m.chainID.String(),
		labelNamePipeName: pipeName,
	}
}

func (m *chainPipeMetrics) TrackPipeLen(name string, lenFunc func() int) {
	m.regLock.Lock()
	defer m.regLock.Unlock()

	reg := m.provider.pipeLenRegistry
	if reg == nil {
		return
	}
	if oldCollector, ok := m.lenMetrics[name]; ok {
		reg.Unregister(oldCollector)
	}

	collector := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace:   "iota_wasp",
		Subsystem:   "chain_pipe",
		Name:        "len",
		Help:        "Length of a pipe",
		ConstLabels: m.makeLabels(name),
	}, func() float64 { return float64(lenFunc()) })
	m.lenMetrics[name] = collector

	if err := reg.Register(collector); err != nil {
		panic(fmt.Errorf("failed to register pipe %v len metric for chain %v: %w", name, m.chainID, err))
	}
}

func (m *chainPipeMetrics) TrackPipeLenMax(name string, key string, lenFunc func() int) {
	m.regLock.Lock()
	defer m.regLock.Unlock()

	reg := m.provider.pipeLenRegistry
	if reg == nil {
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
			max := 0
			for _, f := range valueFuncs {
				fVal := f()
				if max < fVal {
					max = fVal
				}
			}
			return float64(max)
		})
		if err := reg.Register(collector); err != nil {
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

func (m *chainPipeMetrics) ForgetPipeLenMax(name string, key string) {
	m.regLock.Lock()
	defer m.regLock.Unlock()

	reg := m.provider.pipeLenRegistry
	if reg == nil {
		return
	}
	maxCollector, found := m.maxMetrics[name]
	if !found {
		return
	}
	delete(maxCollector.valueFuncs, key)
}
