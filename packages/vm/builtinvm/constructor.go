package builtinvm

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/nilprocessor"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func Constructor(binaryCode []byte) (vmtypes.Processor, error) {
	programHash, err := hashing.HashValueFromBytes(binaryCode)
	if err != nil {
		return nil, err
	}
	switch programHash.String() {
	case root.ProgramHash.String():
		return root.GetProcessor(), nil

	case nilprocessor.ProgramHash.String():
		return nilprocessor.GetProcessor(), nil

	default:
		ret, ok := examples.GetExampleProcessor(programHash.String())
		if !ok {
			return nil, fmt.Errorf("can't load example processor with hash %s", programHash.String())
		}
		return ret, nil

	}
}
