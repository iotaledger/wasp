package builtinvm

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/chainlog"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/dummyprocessor"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	VMType             = "builtinvm"
	printCoreContracts = false
)

func init() {
	if printCoreContracts {
		PrintCoreContracts()
	}
}

func PrintCoreContracts() {
	fmt.Printf("--------------- core contracts ------------------\n")
	fmt.Printf("    %s: '%s'  \n", root.Interface.Hname().String(), root.Interface.Name)
	fmt.Printf("    %s: '%s'  \n", accountsc.Interface.Hname().String(), accountsc.Interface.Name)
	fmt.Printf("    %s: '%s'  \n", blob.Interface.Hname().String(), blob.Interface.Name)
	fmt.Printf("    %s: '%s'  \n", chainlog.Interface.Hname().String(), chainlog.Interface.Name)
	fmt.Printf("--------------- core contracts ------------------\n")
}

func GetProcessor(programHash hashing.HashValue) (vmtypes.Processor, error) {
	switch programHash {
	case root.Interface.ProgramHash:
		return root.GetProcessor(), nil

	case accountsc.Interface.ProgramHash:
		return accountsc.GetProcessor(), nil

	case blob.Interface.ProgramHash:
		return blob.GetProcessor(), nil

	case chainlog.Interface.ProgramHash:
		return chainlog.GetProcessor(), nil

	case *dummyprocessor.ProgramHash:
		return dummyprocessor.GetProcessor(), nil
	}
	return nil, fmt.Errorf("can't find builtin processor with hash %s", programHash.String())
}
