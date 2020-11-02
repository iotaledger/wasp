package builtinvm

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/dummyprocessor"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const VMType = "builtinvm"

func GetProcessor(programHash hashing.HashValue) (vmtypes.Processor, error) {
	switch programHash {
	case *root.ProgramHash:
		return root.GetProcessor(), nil

	case *dummyprocessor.ProgramHash:
		return dummyprocessor.GetProcessor(), nil
	}
	return nil, fmt.Errorf("can't find builtin processor with hash %s", programHash.String())
}
