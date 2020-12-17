// smart contract for testing
package test_sandbox

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/examples"
)

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
		contract.Func(FuncChainLogGenericData, testChainLogGenericData),
		contract.Func(FuncChainLogEventData, testChainLogEventData),
		contract.Func(FuncChainLogEventDataFormatted, testChainLogEventDataFormatted),
	})
	examples.AddProcessor(Interface.ProgramHash, Interface)
}

const (
	// function names
	FuncChainLogGenericData        = "testChainLogGenericData"
	FuncChainLogEventData          = "testChainLogEventData"
	FuncChainLogEventDataFormatted = "testChainLogEventDataFormatted"
	//Variables
	VarCounter = "counter"
)
