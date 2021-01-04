package hardcoded

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/examples"
)

func LocateHardcodedProgram(programHash hashing.HashValue) (string, bool) {
	if _, err := core.GetProcessor(programHash); err == nil {
		return core.VMType, true
	}
	if _, ok := examples.GetExampleProcessor(programHash); ok {
		return examples.VMType, true
	}
	return "", false
}
