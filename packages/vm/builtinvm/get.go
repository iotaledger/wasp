package builtinvm

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accounts"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/chainlog"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	VMType             = "builtinvm"
	printCoreContracts = true
)

func init() {
	if printCoreContracts {
		printReservedHnames()
	}
}

// for debugging
func printReservedHnames() {
	fmt.Printf("--------------- reserved hnames ------------------\n")
	fmt.Printf("    %10s: '%s'\n", root.Interface.Hname().String(), root.Interface.Name)
	fmt.Printf("    %10s: '%s'\n", accounts.Interface.Hname().String(), accounts.Interface.Name)
	fmt.Printf("    %10s: '%s'\n", blob.Interface.Hname().String(), blob.Interface.Name)
	fmt.Printf("    %10s: '%s'\n", chainlog.Interface.Hname().String(), chainlog.Interface.Name)
	fmt.Printf("    %10s: '%s'\n", coretypes.EntryPointInit.String(), coretypes.FuncInit)
	fmt.Printf("--------------- reserved hnames ------------------\n")
}

func GetProcessor(programHash hashing.HashValue) (vmtypes.Processor, error) {
	switch programHash {
	case root.Interface.ProgramHash:
		return root.Interface, nil

	case accounts.Interface.ProgramHash:
		return accounts.Interface, nil

	case blob.Interface.ProgramHash:
		return blob.Interface, nil

	case chainlog.Interface.ProgramHash:
		return chainlog.Interface, nil
	}
	return nil, fmt.Errorf("can't find builtin processor with hash %s", programHash.String())
}
