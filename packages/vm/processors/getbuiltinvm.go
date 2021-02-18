package processors

import (
	"github.com/iotaledger/wasp/contracts/examples_core"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/core"
)

// GetBuiltinProcessorType locates hardcoded processor: core contract or example
func GetBuiltinProcessorType(programHash hashing.HashValue) (string, bool) {
	if _, err := core.GetProcessor(programHash); err == nil {
		return core.VMType, true
	}
	if _, ok := examples_core.GetProcessor(programHash); ok {
		return examples_core.VMType, true
	}
	return "", false
}
