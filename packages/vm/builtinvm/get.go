package builtinvm

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/dummyprocessor"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const VMType = "builtinvm"

func init() {
	fmt.Printf("--------------- core contracts ------------------\n")
	fmt.Printf("    %s: '%s'  \n", root.Hname.String(), root.ContractName)
	fmt.Printf("    %s: '%s'  \n", accountsc.Hname.String(), accountsc.ContractName)
	fmt.Printf("--------------- core contracts ------------------\n")
}
func GetProcessor(programHash hashing.HashValue) (vmtypes.Processor, error) {
	switch programHash {
	case *root.ProgramHash:
		return root.GetProcessor(), nil

	case *accountsc.ProgramHash:
		return accountsc.GetProcessor(), nil

	case *dummyprocessor.ProgramHash:
		return dummyprocessor.GetProcessor(), nil
	}
	return nil, fmt.Errorf("can't find builtin processor with hash %s", programHash.String())
}
