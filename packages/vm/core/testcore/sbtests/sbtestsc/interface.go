// smart contract for testing
package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var Contract = coreutil.NewContract("testcore", "Test Core Sandbox functions")

var Processor = Contract.Processor(initialize,
	FuncChainOwnerIDView.WithHandler(testChainOwnerIDView),
	FuncChainOwnerIDFull.WithHandler(testChainOwnerIDFull),
	FuncGetMintedSupply.WithHandler(getMintedSupply),

	FuncEventLogGenericData.WithHandler(testEventLogGenericData),
	FuncEventLogEventData.WithHandler(testEventLogEventData),
	FuncEventLogDeploy.WithHandler(testEventLogDeploy),
	FuncSandboxCall.WithHandler(testSandboxCall),

	FuncPanicFullEP.WithHandler(testPanicFullEP),
	FuncPanicViewEP.WithHandler(testPanicViewEP),
	FuncCallPanicFullEP.WithHandler(testCallPanicFullEP),
	FuncCallPanicViewEPFromFull.WithHandler(testCallPanicViewEPFromFull),
	FuncCallPanicViewEPFromView.WithHandler(testCallPanicViewEPFromView),

	FuncDoNothing.WithHandler(doNothing),
	FuncSendToAddress.WithHandler(sendToAddress),

	FuncWithdrawToChain.WithHandler(withdrawToChain),
	FuncCallOnChain.WithHandler(callOnChain),
	FuncSetInt.WithHandler(setInt),
	FuncGetInt.WithHandler(getInt),
	FuncGetFibonacci.WithHandler(getFibonacci),
	FuncIncCounter.WithHandler(incCounter),
	FuncGetCounter.WithHandler(getCounter),
	FuncRunRecursion.WithHandler(runRecursion),

	FuncPassTypesFull.WithHandler(passTypesFull),
	FuncPassTypesView.WithHandler(passTypesView),
	FuncCheckContextFromFullEP.WithHandler(testCheckContextFromFullEP),
	FuncCheckContextFromViewEP.WithHandler(testCheckContextFromViewEP),

	FuncTestBlockContext1.WithHandler(testBlockContext1),
	FuncTestBlockContext2.WithHandler(testBlockContext2),
	FuncGetStringValue.WithHandler(getStringValue),

	FuncJustView.WithHandler(testJustView),

	FuncSpawn.WithHandler(spawn),
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
	ParamContractID      = "contractID"
	ParamIntParamName    = "intParamName"
	ParamIntParamValue   = "intParamValue"
	ParamHnameContract   = "hnameContract"
	ParamHnameEP         = "hnameEP"
	ParamVarName         = "varName"

	// error fragments for testing
	MsgFullPanic         = "========== panic FULL ENTRY POINT ========="
	MsgViewPanic         = "========== panic VIEW ========="
	MsgDoNothing         = "========== doing nothing"
	MsgPanicUnauthorized = "============== panic due to unauthorized call"
)
