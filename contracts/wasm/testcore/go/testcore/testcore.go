// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

//nolint:dupl
package testcore

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coregovernance"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

const (
	ContractNameDeployed = "exampleDeployTR"
	MsgCoreOnlyPanic     = "========== core only ========="
	MsgCounterNumber     = "[GenericData] Counter Number: "
	MsgDoNothing         = "========== doing nothing"
	MsgFailOnPurpose     = "failing on purpose"
	MsgFullPanic         = "========== panic FULL ENTRY POINT ========="
	MsgJustView          = "calling empty view entry point"
	MsgTestingEvent      = "[Event] - Testing Event..."
	MsgViewPanic         = "========== panic VIEW ========="
)

func funcCallOnChain(ctx wasmlib.ScFuncContext, f *CallOnChainContext) {
	n := f.Params.N().Value()

	hnameContract := ctx.Contract()
	if f.Params.HnameContract().Exists() {
		hnameContract = f.Params.HnameContract().Value()
	}

	hnameEP := HFuncCallOnChain
	if f.Params.HnameEP().Exists() {
		hnameEP = f.Params.HnameEP().Value()
	}

	counter := f.State.Counter()

	ctx.Log("param IN = " + f.Params.N().String() +
		", hnameContract = " + hnameContract.String() +
		", hnameEP = " + hnameEP.String() +
		", counter = " + counter.String())

	counter.SetValue(counter.Value() + 1)

	params := wasmlib.NewScDict()
	key := []byte(ParamN)
	params.Set(key, wasmtypes.Uint64ToBytes(n))
	ret := ctx.Call(hnameContract, hnameEP, params, nil)
	retVal := wasmtypes.Uint64FromBytes(ret.Get(key))
	f.Results.N().SetValue(retVal)
}

func funcCheckContextFromFullEP(ctx wasmlib.ScFuncContext, f *CheckContextFromFullEPContext) {
	ctx.Require(f.Params.AgentID().Value() == ctx.AccountID(), "fail: agentID")
	ctx.Require(f.Params.Caller().Value() == ctx.Caller(), "fail: caller")
	ctx.Require(f.Params.ChainID().Value() == ctx.CurrentChainID(), "fail: chainID")
	ctx.Require(f.Params.ChainOwnerID().Value() == ctx.ChainOwnerID(), "fail: chainOwnerID")
}

func funcClaimAllowance(ctx wasmlib.ScFuncContext, _ *ClaimAllowanceContext) {
	allowance := ctx.Allowance()
	transfer := wasmlib.NewScTransferFromBalances(allowance)
	ctx.TransferAllowed(ctx.AccountID(), transfer, false)
}

func funcDoNothing(ctx wasmlib.ScFuncContext, _ *DoNothingContext) {
	ctx.Log(MsgDoNothing)
}

func funcEstimateMinStorageDeposit(ctx wasmlib.ScFuncContext, _ *EstimateMinStorageDepositContext) {
	provided := ctx.Allowance().BaseTokens()
	dummy := ScFuncs.EstimateMinStorageDeposit(ctx)
	required := ctx.EstimateStorageDeposit(dummy.Func)
	ctx.Require(provided >= required, "not enough funds")
}

func funcIncCounter(_ wasmlib.ScFuncContext, f *IncCounterContext) {
	counter := f.State.Counter()
	counter.SetValue(counter.Value() + 1)
}

func funcInfiniteLoop(_ wasmlib.ScFuncContext, _ *InfiniteLoopContext) {
	for {
		// do nothing, just waste gas
	}
}

func funcInit(ctx wasmlib.ScFuncContext, f *InitContext) {
	if f.Params.Fail().Exists() {
		ctx.Panic(MsgFailOnPurpose)
	}
}

func funcPassTypesFull(ctx wasmlib.ScFuncContext, f *PassTypesFullContext) {
	hash := ctx.Utility().HashBlake2b([]byte(ParamHash))
	ctx.Require(f.Params.Hash().Value() == hash, "wrong hash")
	ctx.Require(f.Params.Hname().Value() == ctx.Utility().Hname(ParamHname), "wrong hname")
	ctx.Require(f.Params.HnameZero().Value() == 0, "wrong hname-0")
	ctx.Require(f.Params.Int64().Value() == 42, "wrong int64")
	ctx.Require(f.Params.Int64Zero().Value() == 0, "wrong int64-0")
	ctx.Require(f.Params.String().Value() == ParamString, "wrong string")
	ctx.Require(f.Params.StringZero().Value() == "", "wrong string-0")
	// TODO more?
}

func funcPingAllowanceBack(ctx wasmlib.ScFuncContext, _ *PingAllowanceBackContext) {
	caller := ctx.Caller()
	ctx.Require(caller.IsAddress(), "pingAllowanceBack: caller expected to be a L1 address")
	transfer := wasmlib.NewScTransferFromBalances(ctx.Allowance())
	ctx.TransferAllowed(ctx.AccountID(), transfer, false)
	ctx.Send(caller.Address(), transfer)
}

func funcRunRecursion(ctx wasmlib.ScFuncContext, f *RunRecursionContext) {
	depth := f.Params.N().Value()
	if depth <= 0 {
		return
	}

	callOnChain := ScFuncs.CallOnChain(ctx)
	callOnChain.Params.N().SetValue(depth - 1)
	callOnChain.Params.HnameEP().SetValue(HFuncRunRecursion)
	callOnChain.Func.Call()
	retVal := callOnChain.Results.N().Value()
	f.Results.N().SetValue(retVal)
}

func funcSendLargeRequest(_ wasmlib.ScFuncContext, _ *SendLargeRequestContext) {
}

func funcSendNFTsBack(ctx wasmlib.ScFuncContext, _ *SendNFTsBackContext) {
	address := ctx.Caller().Address()
	allowance := ctx.Allowance()
	transfer := wasmlib.NewScTransferFromBalances(allowance)
	ctx.TransferAllowed(ctx.AccountID(), transfer, false)
	for nftID := range allowance.NftIDs() {
		transfer = wasmlib.NewScTransferNFT(&nftID)
		ctx.Send(address, transfer)
	}
}

func funcSendToAddress(_ wasmlib.ScFuncContext, _ *SendToAddressContext) {
	// transfer := wasmlib.NewScTransferFromBalances(ctx.Balances())
	// ctx.Send(f.Params.Address().Value(), transfer)
}

func funcSetInt(_ wasmlib.ScFuncContext, f *SetIntContext) {
	f.State.Ints().GetInt64(f.Params.Name().Value()).SetValue(f.Params.IntValue().Value())
}

func funcSpawn(ctx wasmlib.ScFuncContext, f *SpawnContext) {
	programHash := f.Params.ProgHash().Value()
	spawnName := ScName + "_spawned"
	spawnDescr := "spawned contract description"
	ctx.DeployContract(programHash, spawnName, spawnDescr, nil)

	spawnHname := wasmtypes.NewScHname(spawnName)
	for i := 0; i < 5; i++ {
		ctx.Call(spawnHname, HFuncIncCounter, nil, nil)
	}
}

func funcSplitFunds(ctx wasmlib.ScFuncContext, _ *SplitFundsContext) {
	tokens := ctx.Allowance().BaseTokens()
	address := ctx.Caller().Address()
	tokensToTransfer := uint64(1_000_000)
	transfer := wasmlib.NewScTransferBaseTokens(tokensToTransfer)
	for ; tokens >= tokensToTransfer; tokens -= tokensToTransfer {
		ctx.TransferAllowed(ctx.AccountID(), transfer, false)
		ctx.Send(address, transfer)
	}
}

func funcSplitFundsNativeTokens(ctx wasmlib.ScFuncContext, _ *SplitFundsNativeTokensContext) {
	tokens := ctx.Allowance().BaseTokens()
	address := ctx.Caller().Address()
	transfer := wasmlib.NewScTransferBaseTokens(tokens)
	ctx.TransferAllowed(ctx.AccountID(), transfer, false)
	for _, token := range ctx.Allowance().TokenIDs() {
		one := wasmtypes.NewScBigInt(1)
		transfer = wasmlib.NewScTransferTokens(token, one)
		tokens := ctx.Allowance().Balance(token)
		for ; tokens.Cmp(one) >= 0; tokens = tokens.Sub(one) {
			ctx.TransferAllowed(ctx.AccountID(), transfer, false)
			ctx.Send(address, transfer)
		}
	}
}

func funcTestBlockContext1(ctx wasmlib.ScFuncContext, _ *TestBlockContext1Context) {
	ctx.Panic(MsgCoreOnlyPanic)
}

func funcTestBlockContext2(ctx wasmlib.ScFuncContext, _ *TestBlockContext2Context) {
	ctx.Panic(MsgCoreOnlyPanic)
}

func funcTestCallPanicFullEP(ctx wasmlib.ScFuncContext, _ *TestCallPanicFullEPContext) {
	ctx.Log("will be calling entry point '" + FuncTestPanicFullEP + "' from full EP")
	ScFuncs.TestPanicFullEP(ctx).Func.Call()
}

func funcTestCallPanicViewEPFromFull(ctx wasmlib.ScFuncContext, _ *TestCallPanicViewEPFromFullContext) {
	ctx.Log("will be calling entry point '" + ViewTestPanicViewEP + "' from full EP")
	ScFuncs.TestPanicViewEP(ctx).Func.Call()
}

func funcTestChainOwnerIDFull(ctx wasmlib.ScFuncContext, f *TestChainOwnerIDFullContext) {
	f.Results.ChainOwnerID().SetValue(ctx.ChainOwnerID())
}

func funcTestEventLogDeploy(ctx wasmlib.ScFuncContext, _ *TestEventLogDeployContext) {
	// deploy the same contract with another name
	programHash := ctx.Utility().HashBlake2b([]byte(ScName))
	ctx.DeployContract(programHash, ContractNameDeployed, "test contract deploy log", nil)
}

func funcTestEventLogEventData(ctx wasmlib.ScFuncContext, _ *TestEventLogEventDataContext) {
	ctx.Event(MsgTestingEvent)
}

func funcTestEventLogGenericData(ctx wasmlib.ScFuncContext, f *TestEventLogGenericDataContext) {
	event := MsgCounterNumber + f.Params.Counter().String()
	ctx.Event(event)
}

func funcTestPanicFullEP(ctx wasmlib.ScFuncContext, _ *TestPanicFullEPContext) {
	ctx.Panic(MsgFullPanic)
}

func funcWithdrawFromChain(ctx wasmlib.ScFuncContext, f *WithdrawFromChainContext) {
	targetChain := f.Params.ChainID().Value()
	withdrawal := f.Params.BaseTokensWithdrawal().Value()
	// gasBudget := f.Params.GasBudget().Value()

	// TODO more
	availableTokens := ctx.Allowance().BaseTokens()
	// requiredStorageDepositDeposit := ctx.EstimateRequiredStorageDepositDeposit(request)
	if availableTokens < 1000 {
		ctx.Panic("not enough base tokens sent to cover StorageDeposit deposit")
	}
	transfer := wasmlib.NewScTransferFromBalances(ctx.Allowance())
	ctx.TransferAllowed(ctx.AccountID(), transfer, false)

	//request := isc.RequestParameters{
	//	TargetAddress:  targetChain.AsAddress(),
	//	FungibleTokens: isc.NewFungibleBaseTokens(availableTokens),
	//	Metadata: &isc.SendMetadata{
	//		TargetContract: accounts.Contract.Hname(),
	//		EntryPoint:     accounts.FuncWithdraw.Hname(),
	//		GasBudget:      gasBudget,
	//		Allowance:      isc.NewAllowanceBaseTokens(withdrawal),
	//	},
	//}

	withdraw := coreaccounts.ScFuncs.Withdraw(ctx)
	withdraw.Func.TransferBaseTokens(withdrawal).PostToChain(targetChain)
}

func viewCheckContextFromViewEP(ctx wasmlib.ScViewContext, f *CheckContextFromViewEPContext) {
	ctx.Require(f.Params.AgentID().Value() == ctx.AccountID(), "fail: agentID")
	ctx.Require(f.Params.ChainID().Value() == ctx.CurrentChainID(), "fail: chainID")
	ctx.Require(f.Params.ChainOwnerID().Value() == ctx.ChainOwnerID(), "fail: chainOwnerID")
}

func fibonacci(n uint64) uint64 {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func viewFibonacci(_ wasmlib.ScViewContext, f *FibonacciContext) {
	n := f.Params.N().Value()
	result := fibonacci(n)
	f.Results.N().SetValue(result)
}

func viewFibonacciIndirect(ctx wasmlib.ScViewContext, f *FibonacciIndirectContext) {
	n := f.Params.N().Value()
	if n == 0 || n == 1 {
		f.Results.N().SetValue(n)
		return
	}

	fib := ScFuncs.FibonacciIndirect(ctx)
	fib.Params.N().SetValue(n - 1)
	fib.Func.Call()
	n1 := fib.Results.N().Value()

	fib.Params.N().SetValue(n - 2)
	fib.Func.Call()
	n2 := fib.Results.N().Value()

	f.Results.N().SetValue(n1 + n2)
}

func viewGetCounter(_ wasmlib.ScViewContext, f *GetCounterContext) {
	f.Results.Counter().SetValue(f.State.Counter().Value())
}

func viewGetInt(ctx wasmlib.ScViewContext, f *GetIntContext) {
	name := f.Params.Name().Value()
	value := f.State.Ints().GetInt64(name)
	ctx.Require(value.Exists(), "param '"+name+"' not found")
	f.Results.Values().GetInt64(name).SetValue(value.Value())
}

func viewGetStringValue(ctx wasmlib.ScViewContext, _ *GetStringValueContext) {
	ctx.Panic(MsgCoreOnlyPanic)
	// varName := f.Params.VarName().Value()
	// value := f.State.Strings().GetString(varName).Value()
	// f.Results.Vars().GetString(varName).SetValue(value)
}

func viewInfiniteLoopView(_ wasmlib.ScViewContext, _ *InfiniteLoopViewContext) {
	for {
		// do nothing, just waste gas
	}
}

func viewJustView(ctx wasmlib.ScViewContext, _ *JustViewContext) {
	ctx.Log(MsgJustView)
}

func viewPassTypesView(ctx wasmlib.ScViewContext, f *PassTypesViewContext) {
	hash := ctx.Utility().HashBlake2b([]byte(ParamHash))
	ctx.Require(f.Params.Hash().Value() == hash, "wrong hash")
	ctx.Require(f.Params.Hname().Value() == ctx.Utility().Hname(ParamHname), "wrong hname")
	ctx.Require(f.Params.HnameZero().Value() == 0, "wrong hname-0")
	ctx.Require(f.Params.Int64().Value() == 42, "wrong int64")
	ctx.Require(f.Params.Int64Zero().Value() == 0, "wrong int64-0")
	ctx.Require(f.Params.String().Value() == ParamString, "wrong string")
	ctx.Require(f.Params.StringZero().Value() == "", "wrong string-0")
	// TODO more?
}

func viewTestCallPanicViewEPFromView(ctx wasmlib.ScViewContext, _ *TestCallPanicViewEPFromViewContext) {
	ctx.Log("will be calling entry point '" + ViewTestPanicViewEP + "' from view EP")
	ScFuncs.TestPanicViewEP(ctx).Func.Call()
}

func viewTestChainOwnerIDView(ctx wasmlib.ScViewContext, f *TestChainOwnerIDViewContext) {
	f.Results.ChainOwnerID().SetValue(ctx.ChainOwnerID())
}

func viewTestPanicViewEP(ctx wasmlib.ScViewContext, _ *TestPanicViewEPContext) {
	ctx.Panic(MsgViewPanic)
}

func viewTestSandboxCall(ctx wasmlib.ScViewContext, f *TestSandboxCallContext) {
	getChainInfo := coregovernance.ScFuncs.GetChainInfo(ctx)
	getChainInfo.Func.Call()
	f.Results.SandboxCall().SetValue(getChainInfo.Results.Description().Value())
}
