package hardcoded

import (
	"github.com/iotaledger/wasp/contracts"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/core"
)

func LocateHardcodedProgram(programHash hashing.HashValue) (string, bool) {
	if _, err := core.GetProcessor(programHash); err == nil {
		return core.VMType, true
	}
	if _, ok := contracts.GetExampleProcessor(programHash); ok {
		return contracts.VMType, true
	}
	return "", false
}
