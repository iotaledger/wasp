package hardcoded

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/builtinvm"
	"github.com/iotaledger/wasp/packages/vm/examples"
)

func LocateHardcodedProgram(programHash hashing.HashValue) (string, bool) {
	if _, err := builtinvm.GetProcessor(programHash); err == nil {
		return builtinvm.VMType, true
	}
	if _, ok := examples.GetExampleProcessor(programHash.String()); ok {
		return examples.VMType, true
	}
	return "", false
}
