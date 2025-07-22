// Package processors implements the vm processor
package processors

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type VMConstructor func(binaryCode []byte) (isc.VMProcessor, error)

type Config struct {
	// coreContracts is the collection of core contracts
	coreContracts map[isc.Hname]isc.VMProcessor
}

func NewConfig() *Config {
	return &Config{
		coreContracts: make(map[isc.Hname]isc.VMProcessor),
	}
}

func (p *Config) WithCoreContracts(coreContracts map[isc.Hname]isc.VMProcessor) *Config {
	p.coreContracts = coreContracts
	return p
}

func (p *Config) GetCoreProcessor(hn isc.Hname) (isc.VMProcessor, bool) {
	proc, ok := p.coreContracts[hn]
	return proc, ok
}
