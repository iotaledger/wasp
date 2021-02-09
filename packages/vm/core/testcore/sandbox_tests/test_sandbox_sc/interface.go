// smart contract for testing
package test_sandbox_sc

import (
	"github.com/iotaledger/wasp/contracts"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
)

const (
	Name        = "test_sandbox"
	description = "Test Sandbox functions"
)

var (
	Interface = &contract.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []contract.ContractFunctionInterface{
		contract.ViewFunc(FuncChainOwnerIDView, testChainOwnerIDView),
		contract.Func(FuncChainOwnerIDFull, testChainOwnerIDFull),
		contract.ViewFunc(FuncContractIDView, testContractIDView),
		contract.Func(FuncContractIDFull, testContractIDFull),

		contract.Func(FuncEventLogGenericData, testEventLogGenericData),
		contract.Func(FuncEventLogEventData, testEventLogEventData),
		contract.Func(FuncEventLogDeploy, testEventLogDeploy),
		contract.ViewFunc(FuncSandboxCall, testSandboxCall),

		contract.Func(FuncPanicFullEP, testPanicFullEP),
		contract.ViewFunc(FuncPanicViewEP, testPanicViewEP),
		contract.Func(FuncCallPanicFullEP, testCallPanicFullEP),
		contract.Func(FuncCallPanicViewEPFromFull, testCallPanicViewEPFromFull),
		contract.ViewFunc(FuncCallPanicViewEPFromView, testCallPanicViewEPFromView),

		contract.Func(FuncDoNothing, doNothing),
		contract.Func(FuncSendToAddress, sendToAddress),

		contract.Func(FuncWithdrawToChain, withdrawToChain),
		contract.Func(FuncCallOnChain, callOnChain),
		contract.Func(FuncSetInt, setInt),
		contract.ViewFunc(FuncGetInt, getInt),
		contract.ViewFunc(FuncGetFibonacci, getFibonacci),
		contract.ViewFunc(FuncGetCounter, getCounter),
		contract.Func(FuncRunRecursion, runRecursion),

		contract.Func(FuncPassTypesFull, passTypesFull),
		contract.ViewFunc(FuncPassTypesView, passTypesView),
		contract.Func(FuncCheckContextFromFullEP, testCheckContextFromFullEP),
		contract.ViewFunc(FuncCheckContextFromViewEP, testCheckContextFromViewEP),

		contract.ViewFunc(FuncJustView, testJustView),
	})
	contracts.AddProcessor(Interface)
}

const (
	// function eventlog test
	FuncEventLogGenericData = "testEventLogGenericData"
	FuncEventLogEventData   = "testEventLogEventData"
	FuncEventLogDeploy      = "testEventLogDeploy"

	//Function sandbox test
	FuncChainOwnerIDView = "testChainOwnerIDView"
	FuncChainOwnerIDFull = "testChainOwnerIDFull"
	FuncContractIDView   = "testContractIDView"
	FuncContractIDFull   = "testContractIDFull"

	FuncSandboxCall            = "testSandboxCall"
	FuncCheckContextFromFullEP = "checkContextFromFullEP"
	FuncCheckContextFromViewEP = "checkContextFromViewEP"

	FuncPanicFullEP             = "testPanicFullEP"
	FuncPanicViewEP             = "testPanicViewEP"
	FuncCallPanicFullEP         = "testCallPanicFullEP"
	FuncCallPanicViewEPFromFull = "testCallPanicViewEPFromFull"
	FuncCallPanicViewEPFromView = "testCallPanicViewEPFromView"

	FuncWithdrawToChain = "withdrawToChain"

	FuncDoNothing     = "doNothing"
	FuncSendToAddress = "sendToAddress"
	FuncJustView      = "justView"

	FuncCallOnChain  = "callOnChain"
	FuncSetInt       = "setInt"
	FuncGetInt       = "getInt"
	FuncGetFibonacci = "fibonacci"
	FuncGetCounter   = "getCounter"
	FuncRunRecursion = "runRecursion"

	FuncPassTypesFull = "passTypesFull"
	FuncPassTypesView = "passTypesView"

	//Variables
	VarCounter              = "counter"
	VarChainOwner           = "chainOwner"
	VarContractID           = "contractID"
	VarSandboxCall          = "sandboxCall"
	VarContractNameDeployed = "exampleDeployTR"

	// parameters
	ParamFail            = "initFailParam"
	ParamAddress         = "address"
	ParamChainID         = "chainid"
	ParamChainOwnerID    = "chainOwnerID"
	ParamCaller          = "caller"
	ParamContractID      = "contractID"
	ParamAgentID         = "agentID"
	ParamContractCreator = "contractCreator"
	ParamIntParamName    = "intParamName"
	ParamIntParamValue   = "intParamValue"
	ParamHnameContract   = "hnameContract"
	ParamHnameEP         = "hnameEP"

	// error fragments for testing
	MsgFullPanic         = "========== panic FULL ENTRY POINT ========="
	MsgViewPanic         = "========== panic VIEW ========="
	MsgDoNothing         = "========== doing nothing"
	MsgPanicUnauthorized = "============== panic due to unauthorized call"
)
