// Package sbtestsc defines a smart contract for testing
package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract("testcore")

var Processor = Contract.Processor(nil,
	FuncChainAdminView.WithHandler(testChainAdminView),
	FuncChainAdminFull.WithHandler(testChainAdminFull),

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
	// FuncSendToAddress.WithHandler(sendToAddress),

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

	FuncCheckContextFromFullEP.WithHandler(testCheckContextFromFullEP),
	FuncCheckContextFromViewEP.WithHandler(testCheckContextFromViewEP),

	FuncJustView.WithHandler(testJustView),

	FuncSplitFunds.WithHandler(testSplitFunds),
	FuncSplitFundsNativeTokens.WithHandler(testSplitFundsNativeTokens),
	FuncPingAllowanceBack.WithHandler(pingAllowanceBack),
	FuncInfiniteLoop.WithHandler(infiniteLoop),
	FuncInfiniteLoopView.WithHandler(infiniteLoopView),
	FuncSendObjectsBack.WithHandler(sendObjectsBack),
	FuncClaimAllowance.WithHandler(claimAllowance),
	FuncStackOverflow.WithHandler(stackOverflow),
)

var (
	// function eventlog test
	FuncEventLogGenericData = coreutil.NewEP1(Contract, "testEventLogGenericData",
		coreutil.FieldOptional[uint64](""),
	)
	FuncEventLogEventData = coreutil.NewEP0(Contract, "testEventLogEventData")
	FuncEventLogDeploy    = coreutil.NewEP0(Contract, "testEventLogDeploy")

	// Function sandbox test
	FuncChainAdminView = coreutil.NewViewEP01(Contract, "testChainAdminView",
		coreutil.Field[isc.AgentID](""),
	)
	FuncChainAdminFull = coreutil.NewEP01(Contract, "testChainAdminFull",
		coreutil.Field[isc.AgentID](""),
	)

	FuncSandboxCall            = Contract.ViewFunc("testSandboxCall")
	FuncCheckContextFromFullEP = coreutil.NewEP3(Contract, "checkContextFromFullEP",
		coreutil.Field[isc.AgentID](""),
		coreutil.Field[isc.AgentID](""),
		coreutil.Field[isc.AgentID](""),
	)
	FuncCheckContextFromViewEP = coreutil.NewViewEP2(Contract, "checkContextFromViewEP",
		coreutil.Field[isc.AgentID](""),
		coreutil.Field[isc.AgentID](""),
	)

	FuncTestCustomError         = coreutil.NewEP0(Contract, "testCustomError")
	FuncPanicFullEP             = coreutil.NewEP0(Contract, "testPanicFullEP")
	FuncPanicViewEP             = coreutil.NewViewEP0(Contract, "testPanicViewEP")
	FuncCallPanicFullEP         = Contract.Func("testCallPanicFullEP")
	FuncCallPanicViewEPFromFull = Contract.Func("testCallPanicViewEPFromFull")
	FuncCallPanicViewEPFromView = Contract.ViewFunc("testCallPanicViewEPFromView")

	FuncDoNothing = coreutil.NewEP0(Contract, "doNothing")
	// FuncSendToAddress = coreutil.NewEP(Contract,"sendToAddress")
	FuncJustView = coreutil.NewViewEP0(Contract, "justView")

	FuncCallOnChain = Contract.Func("callOnChain")
	FuncSetInt      = coreutil.NewEP2(Contract, "setInt",
		coreutil.Field[string](""),
		coreutil.Field[int64](""),
	)
	FuncGetInt = coreutil.NewViewEP11(Contract, "getInt",
		coreutil.Field[string](""),
		coreutil.Field[int64](""),
	)
	FuncGetFibonacci = coreutil.NewViewEP11(Contract, "fibonacci",
		coreutil.Field[uint64](""),
		coreutil.Field[uint64](""),
	)
	FuncGetFibonacciIndirect = coreutil.NewViewEP11(Contract, "fibonacciIndirect",
		coreutil.Field[uint64](""),
		coreutil.Field[uint64](""),
	)
	FuncCalcFibonacciIndirectStoreValue = coreutil.NewEP1(Contract, "calcFibonacciIndirectStoreValue",
		coreutil.Field[uint64](""),
	)
	FuncViewCalcFibonacciResult = coreutil.NewViewEP01(Contract, "getFibCalcResult",
		coreutil.Field[uint64](""),
	)
	FuncGetCounter = coreutil.NewViewEP01(Contract, "getCounter",
		coreutil.Field[uint64](""),
	)
	FuncIncCounter   = coreutil.NewEP0(Contract, "incCounter")
	FuncRunRecursion = Contract.Func("runRecursion")

	FuncSplitFunds             = coreutil.NewEP0(Contract, "splitFunds")
	FuncSplitFundsNativeTokens = coreutil.NewEP0(Contract, "splitFundsNativeTokens")
	FuncPingAllowanceBack      = coreutil.NewEP0(Contract, "pingAllowanceBack")

	FuncInfiniteLoop     = coreutil.NewEP0(Contract, "infiniteLoop")
	FuncInfiniteLoopView = coreutil.NewViewEP0(Contract, "infiniteLoopView")
	FuncSendObjectsBack  = coreutil.NewEP0(Contract, "sendObjectsBack")
	FuncClaimAllowance   = coreutil.NewEP0(Contract, "claimAllowance")
	FuncStackOverflow    = coreutil.NewEP0(Contract, "stackOverflow")
)

const (
	// State variables
	VarCounter              = "counter"
	VarContractNameDeployed = "exampleDeployTR"
	VarN                    = "n"

	// error fragments for testing
	MsgDoNothing = "========== doing nothing =========="
	MsgFullPanic = "========== panic FULL ENTRY POINT =========="
	MsgViewPanic = "========== panic VIEW =========="
)
