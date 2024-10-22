// smart contract for testing
package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
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
	FuncEventLogGenericData = coreutil.NewEP1(Contract, "testEventLogGenericData",
		coreutil.Field[uint64](),
	)
	FuncEventLogEventData = coreutil.NewEP0(Contract, "testEventLogEventData")
	FuncEventLogDeploy    = coreutil.NewEP0(Contract, "testEventLogDeploy")

	// Function sandbox test
	FuncChainOwnerIDView = coreutil.NewViewEP01(Contract, "testChainOwnerIDView",
		coreutil.Field[isc.AgentID](),
	)
	FuncChainOwnerIDFull = coreutil.NewEP01(Contract, "testChainOwnerIDFull",
		coreutil.Field[isc.AgentID](),
	)

	FuncSandboxCall = coreutil.NewViewEP01(Contract, "testSandboxCall",
		coreutil.Field[*isc.ChainInfo](),
	)
	FuncCheckContextFromFullEP = coreutil.NewEP4(Contract, "checkContextFromFullEP",
		coreutil.Field[isc.ChainID](),
		coreutil.Field[isc.AgentID](),
		coreutil.Field[isc.AgentID](),
		coreutil.Field[isc.AgentID](),
	)
	FuncCheckContextFromViewEP = coreutil.NewViewEP3(Contract, "checkContextFromViewEP",
		coreutil.Field[isc.ChainID](),
		coreutil.Field[isc.AgentID](),
		coreutil.Field[isc.AgentID](),
	)

	FuncTestCustomError         = coreutil.NewEP0(Contract, "testCustomError")
	FuncPanicFullEP             = coreutil.NewEP0(Contract, "testPanicFullEP")
	FuncPanicViewEP             = coreutil.NewViewEP0(Contract, "testPanicViewEP")
	FuncCallPanicFullEP         = coreutil.NewEP0(Contract, "testCallPanicFullEP")
	FuncCallPanicViewEPFromFull = coreutil.NewEP0(Contract, "testCallPanicViewEPFromFull")
	FuncCallPanicViewEPFromView = coreutil.NewViewEP0(Contract, "testCallPanicViewEPFromView")

	FuncWithdrawFromChain = coreutil.NewEP4(Contract, "withdrawFromChain",
		coreutil.Field[isc.ChainID](),
		coreutil.Field[uint64](),
		coreutil.FieldOptional[uint64](),
		coreutil.FieldOptional[uint64](),
	)

	FuncDoNothing = coreutil.NewEP0(Contract, "doNothing")
	// FuncSendToAddress = coreutil.NewEP(Contract,"sendToAddress")
	FuncJustView = coreutil.NewViewEP0(Contract, "justView")

	FuncCallOnChain = Contract.Func("callOnChain")
	FuncSetInt      = coreutil.NewEP2(Contract, "setInt",
		coreutil.Field[string](),
		coreutil.Field[int](),
	)
	FuncGetInt = coreutil.NewEP11(Contract, "getInt",
		coreutil.Field[string](),
		coreutil.Field[int](),
	)
	FuncGetFibonacci = coreutil.NewViewEP11(Contract, "fibonacci",
		coreutil.Field[uint64](),
		coreutil.Field[uint64](),
	)
	FuncGetFibonacciIndirect = coreutil.NewViewEP11(Contract, "fibonacciIndirect",
		coreutil.Field[uint64](),
		coreutil.Field[uint64](),
	)
	FuncCalcFibonacciIndirectStoreValue = coreutil.NewEP1(Contract, "calcFibonacciIndirectStoreValue",
		coreutil.Field[uint64](),
	)
	FuncViewCalcFibonacciResult = coreutil.NewViewEP01(Contract, "getFibCalcResult",
		coreutil.Field[uint64](),
	)
	FuncGetCounter = coreutil.NewViewEP01(Contract, "getCounter",
		coreutil.Field[int](),
	)
	FuncIncCounter   = coreutil.NewEP0(Contract, "incCounter")
	FuncRunRecursion = coreutil.NewEP11(Contract, "runRecursion",
		coreutil.Field[uint64](),
		coreutil.Field[uint64](),
	)

	FuncPassTypesFull = Contract.Func("passTypesFull")
	FuncPassTypesView = Contract.ViewFunc("passTypesView")

	FuncSpawn = coreutil.NewEP1(Contract, "spawn",
		coreutil.Field[hashing.HashValue](),
	)

	FuncSplitFunds             = coreutil.NewEP0(Contract, "splitFunds")
	FuncSplitFundsNativeTokens = coreutil.NewEP0(Contract, "splitFundsNativeTokens")
	FuncPingAllowanceBack      = coreutil.NewEP0(Contract, "pingAllowanceBack")
	FuncSendLargeRequest       = coreutil.NewEP1(Contract, "sendLargeRequest",
		coreutil.Field[int32](),
	)
	FuncEstimateMinStorageDeposit = coreutil.NewEP0(Contract, "estimateMinStorageDeposit")
	FuncInfiniteLoop              = coreutil.NewEP0(Contract, "infiniteLoop")
	FuncInfiniteLoopView          = coreutil.NewViewEP0(Contract, "infiniteLoopView")
	FuncSendNFTsBack              = coreutil.NewEP0(Contract, "sendNFTsBack")
	FuncClaimAllowance            = coreutil.NewEP0(Contract, "claimAllowance")
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
