// smart contract for testing
package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract("testcore")

var Processor = Contract.Processor(initialize,
	FuncChainOwnerIDView.WithHandler(testChainOwnerIDView),
	FuncChainOwnerIDFull.WithHandler(testChainOwnerIDFull),

	FuncEventLogGenericData.WithHandler(testEventLogGenericData),
	FuncEventLogEventData.WithHandler(testEventLogEventData),
	FuncEventLogDeploy.WithHandler(testEventLogDeploy),
	FuncSandboxCall.WithHandler(testSandboxCall),

	FuncTestCustomError.WithHandler(testCustomError),
	FuncPanicFullEP.WithHandler(testPanicFullEP),
	FuncPanicViewEP.WithHandler(testPanicViewEP),
	FuncCallPanicFullEP.WithHandler(testCallPanicFullEP),
	FuncCallPanicViewEPFromFull.WithHandler(testCallPanicViewEPFromFull),
	FuncCallPanicViewEPFromView.WithHandler(testCallPanicViewEPFromView),

	FuncDoNothing.WithHandler(doNothing),
	// FuncSendToAddress.WithHandler(sendToAddress),

	FuncWithdrawFromChain.WithHandler(withdrawFromChain),
	FuncCallOnChain.WithHandler(callOnChain),
	FuncSetInt.WithHandler(setInt),
	FuncGetInt.WithHandler(getInt),
	FuncGetFibonacci.WithHandler(getFibonacci),
	FuncGetFibonacciIndirect.WithHandler(getFibonacciIndirect),
	FuncCalcFibonacciIndirectStoreValue.WithHandler(calcFibonacciIndirectStoreValue),
	FuncViewCalcFibonacciResult.WithHandler(viewFibResult),
	FuncIncCounter.WithHandler(incCounter),
	FuncGetCounter.WithHandler(getCounter),
	FuncRunRecursion.WithHandler(runRecursion),

	FuncPassTypesFull.WithHandler(passTypesFull),
	FuncPassTypesView.WithHandler(passTypesView),
	FuncCheckContextFromFullEP.WithHandler(testCheckContextFromFullEP),
	FuncCheckContextFromViewEP.WithHandler(testCheckContextFromViewEP),

	FuncJustView.WithHandler(testJustView),

	FuncSpawn.WithHandler(spawn),

	FuncSplitFunds.WithHandler(testSplitFunds),
	FuncSplitFundsNativeTokens.WithHandler(testSplitFundsNativeTokens),
	FuncPingAllowanceBack.WithHandler(pingAllowanceBack),
	FuncSendLargeRequest.WithHandler(sendLargeRequest),
	FuncEstimateMinStorageDeposit.WithHandler(testEstimateMinimumStorageDeposit),
	FuncInfiniteLoop.WithHandler(infiniteLoop),
	FuncInfiniteLoopView.WithHandler(infiniteLoopView),
	FuncSendNFTsBack.WithHandler(sendNFTsBack),
	FuncClaimAllowance.WithHandler(claimAllowance),
)

var (
	// function eventlog test
	FuncEventLogGenericData = Contract.Func("testEventLogGenericData")
	FuncEventLogEventData   = Contract.Func("testEventLogEventData")
	FuncEventLogDeploy      = Contract.Func("testEventLogDeploy")

	// Function sandbox test
	FuncChainOwnerIDView = Contract.ViewFunc("testChainOwnerIDView")
	FuncChainOwnerIDFull = Contract.Func("testChainOwnerIDFull")

	FuncSandboxCall            = Contract.ViewFunc("testSandboxCall")
	FuncCheckContextFromFullEP = Contract.Func("checkContextFromFullEP")
	FuncCheckContextFromViewEP = Contract.ViewFunc("checkContextFromViewEP")

	FuncTestCustomError         = Contract.Func("testCustomError")
	FuncPanicFullEP             = Contract.Func("testPanicFullEP")
	FuncPanicViewEP             = Contract.ViewFunc("testPanicViewEP")
	FuncCallPanicFullEP         = Contract.Func("testCallPanicFullEP")
	FuncCallPanicViewEPFromFull = Contract.Func("testCallPanicViewEPFromFull")
	FuncCallPanicViewEPFromView = Contract.ViewFunc("testCallPanicViewEPFromView")

	FuncWithdrawFromChain = Contract.Func("withdrawFromChain")

	FuncDoNothing = Contract.Func("doNothing")
	// FuncSendToAddress = Contract.Func("sendToAddress")
	FuncJustView = Contract.ViewFunc("justView")

	FuncCallOnChain                     = Contract.Func("callOnChain")
	FuncSetInt                          = Contract.Func("setInt")
	FuncGetInt                          = Contract.ViewFunc("getInt")
	FuncGetFibonacci                    = Contract.ViewFunc("fibonacci")
	FuncGetFibonacciIndirect            = Contract.ViewFunc("fibonacciIndirect")
	FuncCalcFibonacciIndirectStoreValue = Contract.Func("calcFibonacciIndirectStoreValue")
	FuncViewCalcFibonacciResult         = Contract.ViewFunc("getFibCalcResult")
	FuncGetCounter                      = Contract.ViewFunc("getCounter")
	FuncIncCounter                      = Contract.Func("incCounter")
	FuncRunRecursion                    = Contract.Func("runRecursion")

	FuncPassTypesFull = Contract.Func("passTypesFull")
	FuncPassTypesView = Contract.ViewFunc("passTypesView")

	FuncSpawn = Contract.Func("spawn")

	FuncSplitFunds                = Contract.Func("splitFunds")
	FuncSplitFundsNativeTokens    = Contract.Func("splitFundsNativeTokens")
	FuncPingAllowanceBack         = Contract.Func("pingAllowanceBack")
	FuncSendLargeRequest          = Contract.Func("sendLargeRequest")
	FuncEstimateMinStorageDeposit = Contract.Func("estimateMinStorageDeposit")
	FuncInfiniteLoop              = Contract.Func("infiniteLoop")
	FuncInfiniteLoopView          = Contract.ViewFunc("infiniteLoopView")
	FuncSendNFTsBack              = Contract.Func("sendNFTsBack")
	FuncClaimAllowance            = Contract.Func("claimAllowance")
)

const (
	// State variables
	VarCounter              = "counter"
	VarContractNameDeployed = "exampleDeployTR"

	// parameters
	ParamAddress                          = "address"
	ParamAgentID                          = "agentID"
	ParamCaller                           = "caller"
	ParamChainID                          = "chainID"
	ParamChainOwnerID                     = "chainOwnerID"
	ParamGasReserve                       = "gasReserve"
	ParamGasReserveTransferAccountToChain = "gasReserveTransferAccountToChain"
	ParamContractID                       = "contractID"
	ParamFail                             = "initFailParam"
	ParamHnameContract                    = "hnameContract"
	ParamHnameEP                          = "hnameEP"
	ParamIntParamName                     = "intParamName"
	ParamIntParamValue                    = "intParamValue"
	ParamBaseTokens                       = "baseTokens"
	ParamN                                = "n"
	ParamProgHash                         = "progHash"
	ParamSize                             = "size"

	// error fragments for testing
	MsgDoNothing = "========== doing nothing =========="
	MsgFullPanic = "========== panic FULL ENTRY POINT =========="
	MsgViewPanic = "========== panic VIEW =========="
)
