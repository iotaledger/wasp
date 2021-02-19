package processors

import (
	"github.com/iotaledger/wasp/contracts/native"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/core"
)

// GetBuiltinProcessorType locates hardcoded processor: core contract or example
func GetBuiltinProcessorType(programHash hashing.HashValue) (string, bool) {
	if _, err := core.GetProcessor(programHash); err == nil {
		return core.VMType, true
	}
	if _, ok := native.GetProcessor(programHash); ok {
		return native.VMType, true
	}
	return "", false
}
