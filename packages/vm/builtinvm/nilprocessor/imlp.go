// implements nil processor
package nilprocessor

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type nilProcessorStruct struct{}

var (
	processor   = &nilProcessorStruct{}
	ProgramHash = hashing.NilHash
)

func GetProcessor() vmtypes.Processor {
	return processor
}

func (p *nilProcessorStruct) GetEntryPoint(code coretypes.EntryPointCode) (vmtypes.EntryPoint, bool) {
	return nil, false
}

func (p *nilProcessorStruct) GetDescription() string {
	return "Nil processor"
}
