// smart contract for testing
package test_sandbox

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/examples"
)

const (
	Name        = "test_sandbox"
	description = "Test Sandbox functions"
)

var (
	Interface = &contract.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: *hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []contract.ContractFunctionInterface{
		contract.Func(FuncChainLogGenericData, testChainLogGenericData),
		contract.Func(FuncChainLogEventData, testChainLogEventData),
		contract.Func(FuncChainOwnerID, testChainOwnerID),
		contract.Func(FuncChainlogDeploy, testChainlogDeploy),
		contract.Func(FuncPanicFullEntryPoint, testPanicFullEntryPoint),
		contract.ViewFunc(FuncPanicViewEntryPoint, testPanicViewEntryPoint),
		contract.ViewFunc(FuncChainID, testChainID),
		contract.ViewFunc(FuncSandboxCall, testSandboxCall),
	})
	examples.AddProcessor(Interface)
}

const (
	// function chainlog test
	FuncChainLogGenericData = "testChainLogGenericData"
	FuncChainLogEventData   = "testChainLogEventData"
	FuncChainlogDeploy      = "testChainlogDeploy"

	//Function sandbox test
	FuncChainOwnerID        = "testChainOwnerID"
	FuncChainID             = "testChainID"
	FuncSandboxCall         = "testSandboxCall"
	FuncPanicFullEntryPoint = "testPanicFullEntryPoint"
	FuncPanicViewEntryPoint = "testPanicViewEntryPoint"

	//Variables
	VarCounter              = "counter"
	VarChainOwner           = "chainOwner"
	VarChainID              = "chainID"
	VarSandboxCall          = "sandboxCall"
	VarContractNameDeployed = "exampleDeployTR"

	// error fragments for testing
	ErrorFullPanic = "========== panic FULL ENTRY POINT ========="
	ErrorViewPanic = "========== panic VIEW ========="
)
