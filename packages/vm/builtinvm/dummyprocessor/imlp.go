// implements nil processor
package dummyprocessor

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type dummyProcessorStruct struct{}

var (
	processor   = &dummyProcessorStruct{}
	ProgramHash = hashing.AllFHash
)

func GetProcessor() vmtypes.Processor {
	return processor
}

func (p *dummyProcessorStruct) GetEntryPoint(code coretypes.EntryPointCode) (vmtypes.EntryPoint, bool) {
	return nil, false
}

func (p *dummyProcessorStruct) GetDescription() string {
	return "Nil processor"
}
