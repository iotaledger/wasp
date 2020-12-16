package test_chainlog

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// +++ we do not need special contract to test log. We just deploy chainlog contract on 'solo' tool
// and write unit tests

const (
	Name        = "test_chainlog"
	Version     = "0.1"
	fullName    = Name + "-" + Version
	description = "Test chainlog contract"
)

var (
	Interface = &contract.ContractInterface{
		Name:        fullName,
		Description: description,
		ProgramHash: *hashing.HashStrings(fullName),
	}
)

func init() {
	Interface.WithFunctions(initialize, []contract.ContractFunctionInterface{
		contract.Func(FuncTestGeneric, example_TestGeneric),
	})
	examples.AddProcessor(Interface.ProgramHash, Interface)

}

const (
	// function names
	FuncTestGeneric = "example_TestGeneric"
	//Variables
	VarCounter = "counter"
)

func GetProcessor() vmtypes.Processor {
	return Interface
}
