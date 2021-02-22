// smart contract for testing
package sbtestsc

import (
	"github.com/iotaledger/wasp/contracts/native"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
)

const (
	Name        = "test_sandbox"
	description = "Test Sandbox functions"
)

var (
	Interface = &coreutil.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		coreutil.ViewFunc(FuncChainOwnerIDView, testChainOwnerIDView),
		coreutil.Func(FuncChainOwnerIDFull, testChainOwnerIDFull),
		coreutil.ViewFunc(FuncContractIDView, testContractIDView),
		coreutil.Func(FuncContractIDFull, testContractIDFull),
		coreutil.Func(FuncGetMintedSupply, getMintedSupply),

		coreutil.Func(FuncEventLogGenericData, testEventLogGenericData),
		coreutil.Func(FuncEventLogEventData, testEventLogEventData),
		coreutil.Func(FuncEventLogDeploy, testEventLogDeploy),
		coreutil.ViewFunc(FuncSandboxCall, testSandboxCall),

		coreutil.Func(FuncPanicFullEP, testPanicFullEP),
		coreutil.ViewFunc(FuncPanicViewEP, testPanicViewEP),
		coreutil.Func(FuncCallPanicFullEP, testCallPanicFullEP),
		coreutil.Func(FuncCallPanicViewEPFromFull, testCallPanicViewEPFromFull),
		coreutil.ViewFunc(FuncCallPanicViewEPFromView, testCallPanicViewEPFromView),

		coreutil.Func(FuncDoNothing, doNothing),
		coreutil.Func(FuncSendToAddress, sendToAddress),

		coreutil.Func(FuncWithdrawToChain, withdrawToChain),
		coreutil.Func(FuncCallOnChain, callOnChain),
		coreutil.Func(FuncSetInt, setInt),
		coreutil.ViewFunc(FuncGetInt, getInt),
		coreutil.ViewFunc(FuncGetFibonacci, getFibonacci),
		coreutil.Func(FuncIncCounter, incCounter),
		coreutil.ViewFunc(FuncGetCounter, getCounter),
		coreutil.Func(FuncRunRecursion, runRecursion),

		coreutil.Func(FuncPassTypesFull, passTypesFull),
		coreutil.ViewFunc(FuncPassTypesView, passTypesView),
		coreutil.Func(FuncCheckContextFromFullEP, testCheckContextFromFullEP),
		coreutil.ViewFunc(FuncCheckContextFromViewEP, testCheckContextFromViewEP),

		coreutil.ViewFunc(FuncJustView, testJustView),
	})
	native.AddProcessor(Interface)
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
	FuncGetMintedSupply        = "getMintedSupply"

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
	FuncIncCounter   = "incCounter"
	FuncRunRecursion = "runRecursion"

	FuncPassTypesFull = "passTypesFull"
	FuncPassTypesView = "passTypesView"

	//Variables
	VarCounter              = "counter"
	VarContractID           = "contractID"
	VarSandboxCall          = "sandboxCall"
	VarContractNameDeployed = "exampleDeployTR"
	VarMintedSupply         = "mintedSupply"

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
