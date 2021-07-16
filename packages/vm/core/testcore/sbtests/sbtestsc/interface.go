// smart contract for testing
package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var Interface = coreutil.NewContractInterface("testcore", "Test Core Sandbox functions")

var Processor = Interface.Processor(initialize,
	FuncChainOwnerIDView.ViewHandler(testChainOwnerIDView),
	FuncChainOwnerIDFull.Handler(testChainOwnerIDFull),
	FuncGetMintedSupply.Handler(getMintedSupply),

	FuncEventLogGenericData.Handler(testEventLogGenericData),
	FuncEventLogEventData.Handler(testEventLogEventData),
	FuncEventLogDeploy.Handler(testEventLogDeploy),
	FuncSandboxCall.ViewHandler(testSandboxCall),

	FuncPanicFullEP.Handler(testPanicFullEP),
	FuncPanicViewEP.ViewHandler(testPanicViewEP),
	FuncCallPanicFullEP.Handler(testCallPanicFullEP),
	FuncCallPanicViewEPFromFull.Handler(testCallPanicViewEPFromFull),
	FuncCallPanicViewEPFromView.ViewHandler(testCallPanicViewEPFromView),

	FuncDoNothing.Handler(doNothing),
	FuncSendToAddress.Handler(sendToAddress),

	FuncWithdrawToChain.Handler(withdrawToChain),
	FuncCallOnChain.Handler(callOnChain),
	FuncSetInt.Handler(setInt),
	FuncGetInt.ViewHandler(getInt),
	FuncGetFibonacci.ViewHandler(getFibonacci),
	FuncIncCounter.Handler(incCounter),
	FuncGetCounter.ViewHandler(getCounter),
	FuncRunRecursion.Handler(runRecursion),

	FuncPassTypesFull.Handler(passTypesFull),
	FuncPassTypesView.ViewHandler(passTypesView),
	FuncCheckContextFromFullEP.Handler(testCheckContextFromFullEP),
	FuncCheckContextFromViewEP.ViewHandler(testCheckContextFromViewEP),

	FuncTestBlockContext1.Handler(testBlockContext1),
	FuncTestBlockContext2.Handler(testBlockContext2),
	FuncGetStringValue.ViewHandler(getStringValue),

	FuncJustView.ViewHandler(testJustView),

	FuncSpawn.Handler(spawn),
)

var (
	// function eventlog test
	FuncEventLogGenericData = coreutil.Func("testEventLogGenericData")
	FuncEventLogEventData   = coreutil.Func("testEventLogEventData")
	FuncEventLogDeploy      = coreutil.Func("testEventLogDeploy")

	// Function sandbox test
	FuncChainOwnerIDView = coreutil.ViewFunc("testChainOwnerIDView")
	FuncChainOwnerIDFull = coreutil.Func("testChainOwnerIDFull")

	FuncSandboxCall            = coreutil.ViewFunc("testSandboxCall")
	FuncCheckContextFromFullEP = coreutil.Func("checkContextFromFullEP")
	FuncCheckContextFromViewEP = coreutil.ViewFunc("checkContextFromViewEP")
	FuncGetMintedSupply        = coreutil.Func("getMintedSupply")

	FuncPanicFullEP             = coreutil.Func("testPanicFullEP")
	FuncPanicViewEP             = coreutil.ViewFunc("testPanicViewEP")
	FuncCallPanicFullEP         = coreutil.Func("testCallPanicFullEP")
	FuncCallPanicViewEPFromFull = coreutil.Func("testCallPanicViewEPFromFull")
	FuncCallPanicViewEPFromView = coreutil.ViewFunc("testCallPanicViewEPFromView")

	FuncTestBlockContext1 = coreutil.Func("testBlockContext1")
	FuncTestBlockContext2 = coreutil.Func("testBlockContext2")
	FuncGetStringValue    = coreutil.ViewFunc("getStringValue")

	FuncWithdrawToChain = coreutil.Func("withdrawToChain")

	FuncDoNothing     = coreutil.Func("doNothing")
	FuncSendToAddress = coreutil.Func("sendToAddress")
	FuncJustView      = coreutil.ViewFunc("justView")

	FuncCallOnChain  = coreutil.Func("callOnChain")
	FuncSetInt       = coreutil.Func("setInt")
	FuncGetInt       = coreutil.ViewFunc("getInt")
	FuncGetFibonacci = coreutil.ViewFunc("fibonacci")
	FuncGetCounter   = coreutil.ViewFunc("getCounter")
	FuncIncCounter   = coreutil.Func("incCounter")
	FuncRunRecursion = coreutil.Func("runRecursion")

	FuncPassTypesFull = coreutil.Func("passTypesFull")
	FuncPassTypesView = coreutil.ViewFunc("passTypesView")

	FuncSpawn = coreutil.Func("spawn")
)

const (
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
