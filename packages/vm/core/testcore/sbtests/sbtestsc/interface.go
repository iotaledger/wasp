// smart contract for testing
package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
)

const (
	Name        = "testcore"
	description = "Test Core Sandbox functions"
)

var Interface = &coreutil.ContractInterface{
	Name:        Name,
	Description: description,
	ProgramHash: hashing.HashStrings(Name),
}

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		coreutil.ViewFunc(FuncChainOwnerIDView, testChainOwnerIDView),
		coreutil.Func(FuncChainOwnerIDFull, testChainOwnerIDFull),
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

		coreutil.Func(FuncTestBlockContext1, testBlockContext1),
		coreutil.Func(FuncTestBlockContext2, testBlockContext2),
		coreutil.ViewFunc(FuncGetStringValue, getStringValue),

		coreutil.ViewFunc(FuncJustView, testJustView),

		coreutil.Func(FuncSpawn, spawn),
	})
}

const (
	// function eventlog test
	FuncEventLogGenericData = "testEventLogGenericData"
	FuncEventLogEventData   = "testEventLogEventData"
	FuncEventLogDeploy      = "testEventLogDeploy"

	// Function sandbox test
	FuncChainOwnerIDView = "testChainOwnerIDView"
	FuncChainOwnerIDFull = "testChainOwnerIDFull"

	FuncSandboxCall            = "testSandboxCall"
	FuncCheckContextFromFullEP = "checkContextFromFullEP"
	FuncCheckContextFromViewEP = "checkContextFromViewEP"
	FuncGetMintedSupply        = "getMintedSupply"

	FuncPanicFullEP             = "testPanicFullEP"
	FuncPanicViewEP             = "testPanicViewEP"
	FuncCallPanicFullEP         = "testCallPanicFullEP"
	FuncCallPanicViewEPFromFull = "testCallPanicViewEPFromFull"
	FuncCallPanicViewEPFromView = "testCallPanicViewEPFromView"

	FuncTestBlockContext1 = "testBlockContext1"
	FuncTestBlockContext2 = "testBlockContext2"
	FuncGetStringValue    = "getStringValue"

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

	FuncSpawn = "spawn"

	// State variables
	VarCounter              = "counter"
	VarSandboxCall          = "sandboxCall"
	VarContractNameDeployed = "exampleDeployTR"
	VarMintedSupply         = "mintedSupply"
	VarMintedColor          = "mintedColor"

	// parameters
	ParamFail            = "initFailParam"
	ParamAddress         = "address"
	ParamChainID         = "chainID"
	ParamChainOwnerID    = "chainOwnerID"
	ParamCaller          = "caller"
	ParamAgentID         = "agentID"
	ParamContractCreator = "contractCreator"
	ParamIntParamName    = "intParamName"
	ParamIntParamValue   = "intParamValue"
	ParamHnameContract   = "hnameContract"
	ParamHnameEP         = "hnameEP"
	ParamVarName         = "paramVar"

	// error fragments for testing
	MsgFullPanic         = "========== panic FULL ENTRY POINT ========="
	MsgViewPanic         = "========== panic VIEW ========="
	MsgDoNothing         = "========== doing nothing"
	MsgPanicUnauthorized = "============== panic due to unauthorized call"
)
