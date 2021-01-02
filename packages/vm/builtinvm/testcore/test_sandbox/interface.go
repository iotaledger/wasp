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
		contract.ViewFunc(FuncChainID, testChainID),
		contract.ViewFunc(FuncSandboxCall, testSandboxCall),

		contract.Func(FuncCheckContextFromFullEP, testCheckContextFromFullEP),
		contract.ViewFunc(FuncCheckContextFromViewEP, testCheckContextFromViewEP),

		contract.Func(FuncPanicFullEP, testPanicFullEP),
		contract.ViewFunc(FuncPanicViewEP, testPanicViewEP),
		contract.Func(FuncCallPanicFullEP, testCallPanicFullEP),
		contract.Func(FuncCallPanicViewEPFromFull, testCallPanicViewEPFromFull),
		contract.ViewFunc(FuncCallPanicViewEPFromView, testCallPanicViewEPFromView),

		contract.Func(FuncDoNothing, doNothing),
		contract.Func(FuncSendToAddress, sendToAddress),
	})
	examples.AddProcessor(Interface)
}

const (
	// function chainlog test
	FuncChainLogGenericData = "testChainLogGenericData"
	FuncChainLogEventData   = "testChainLogEventData"
	FuncChainlogDeploy      = "testChainlogDeploy"

	//Function sandbox test
	FuncChainOwnerID           = "testChainOwnerID"
	FuncChainID                = "testChainID"
	FuncSandboxCall            = "testSandboxCall"
	FuncCheckContextFromFullEP = "testCheckContextFromFullEP"
	FuncCheckContextFromViewEP = "testCheckContextFromViewEP"

	FuncPanicFullEP             = "testPanicFullEP"
	FuncPanicViewEP             = "testPanicViewEP"
	FuncCallPanicFullEP         = "testCallPanicFullEP"
	FuncCallPanicViewEPFromFull = "testCallPanicViewEPFromFull"
	FuncCallPanicViewEPFromView = "testCallPanicViewEPFromView"

	FuncDoNothing     = "doNothing"
	FuncSendToAddress = "sendToAddress"

	//Variables
	VarCounter              = "counter"
	VarChainOwner           = "chainOwner"
	VarChainID              = "chainID"
	VarSandboxCall          = "sandboxCall"
	VarContractNameDeployed = "exampleDeployTR"

	// parameters
	ParamAddress         = "address"
	ParamChainID         = "chainid"
	ParamChainOwnerID    = "chainOwnerID"
	ParamCaller          = "caller"
	ParamContractID      = "contractID"
	ParamAgentID         = "agentID"
	ParamContractCreator = "contractCreator"

	// error fragments for testing
	MsgFullPanic = "========== panic FULL ENTRY POINT ========="
	MsgViewPanic = "========== panic VIEW ========="
	MsgDoNothing = "========== doing nothing"
)
