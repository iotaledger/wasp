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
	paramInt := f.Params.IntValue().Value()

	hnameContract := ctx.Contract()
	if f.Params.HnameContract().Exists() {
		hnameContract = f.Params.HnameContract().Value()
	}

	hnameEP := HFuncCallOnChain
	if f.Params.HnameEP().Exists() {
		hnameEP = f.Params.HnameEP().Value()
	}

	counter := f.State.Counter()

	ctx.Log("param IN = " + f.Params.IntValue().String() +
		", hnameContract = " + hnameContract.String() +
		", hnameEP = " + hnameEP.String() +
		", counter = " + counter.String())

	counter.SetValue(counter.Value() + 1)

	params := wasmlib.NewScDict()
	key := []byte(ParamIntValue)
	params.Set(key, wasmtypes.Int64ToBytes(paramInt))
	ret := ctx.Call(hnameContract, hnameEP, params, nil)
	retVal := wasmtypes.Int64FromBytes(ret.Get(key))
	f.Results.IntValue().SetValue(retVal)
}

func funcCheckContextFromFullEP(ctx wasmlib.ScFuncContext, f *CheckContextFromFullEPContext) {
	ctx.Require(f.Params.AgentID().Value() == ctx.AccountID(), "fail: agentID")
	ctx.Require(f.Params.Caller().Value() == ctx.Caller(), "fail: caller")
	ctx.Require(f.Params.ChainID().Value() == ctx.ChainID(), "fail: chainID")
	ctx.Require(f.Params.ChainOwnerID().Value() == ctx.ChainOwnerID(), "fail: chainOwnerID")
	ctx.Require(f.Params.ContractCreator().Value() == ctx.ContractCreator(), "fail: contractCreator")
}

//nolint:unparam
func funcDoNothing(ctx wasmlib.ScFuncContext, f *DoNothingContext) {
	ctx.Log(MsgDoNothing)
}

//func funcGetMintedSupply(ctx wasmlib.ScFuncContext, f *GetMintedSupplyContext) {
//	//minted := ctx.Minted()
//	//mintedColors := minted.Colors()
//	//ctx.Require(len(mintedColors) == 1, "test only supports one minted color")
//	//color := mintedColors[0]
//	//amount := minted.Balance(color)
//	//f.Results.MintedColor().SetValue(color)
//	//f.Results.MintedSupply().SetValue(amount)
//}

func funcIncCounter(ctx wasmlib.ScFuncContext, f *IncCounterContext) {
	counter := f.State.Counter()
	counter.SetValue(counter.Value() + 1)
}

//nolint:unparam
func funcInfiniteLoop(ctx wasmlib.ScFuncContext, f *InfiniteLoopContext) {
	//nolint:staticcheck
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

func funcRunRecursion(ctx wasmlib.ScFuncContext, f *RunRecursionContext) {
	depth := f.Params.IntValue().Value()
	if depth <= 0 {
		return
	}

	callOnChain := ScFuncs.CallOnChain(ctx)
	callOnChain.Params.IntValue().SetValue(depth - 1)
	callOnChain.Params.HnameEP().SetValue(HFuncRunRecursion)
	callOnChain.Func.Call()
	retVal := callOnChain.Results.IntValue().Value()
	f.Results.IntValue().SetValue(retVal)
}

func funcSendToAddress(ctx wasmlib.ScFuncContext, f *SendToAddressContext) {
	// transfer := wasmlib.NewScTransferFromBalances(ctx.Balances())
	// ctx.Send(f.Params.Address().Value(), transfer)
}

func funcSetInt(ctx wasmlib.ScFuncContext, f *SetIntContext) {
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

//nolint:unparam
func funcTestCallPanicFullEP(ctx wasmlib.ScFuncContext, f *TestCallPanicFullEPContext) {
	ctx.Log("will be calling entry point '" + FuncTestPanicFullEP + "' from full EP")
	ScFuncs.TestPanicFullEP(ctx).Func.Call()
}

//nolint:unparam
func funcTestCallPanicViewEPFromFull(ctx wasmlib.ScFuncContext, f *TestCallPanicViewEPFromFullContext) {
	ctx.Log("will be calling entry point '" + ViewTestPanicViewEP + "' from full EP")
	ScFuncs.TestPanicViewEP(ctx).Func.Call()
}

func funcTestChainOwnerIDFull(ctx wasmlib.ScFuncContext, f *TestChainOwnerIDFullContext) {
	f.Results.ChainOwnerID().SetValue(ctx.ChainOwnerID())
}

//nolint:unparam
func funcTestEventLogDeploy(ctx wasmlib.ScFuncContext, f *TestEventLogDeployContext) {
	// deploy the same contract with another name
	programHash := ctx.Utility().HashBlake2b([]byte(ScName))
	ctx.DeployContract(programHash, ContractNameDeployed, "test contract deploy log", nil)
}

//nolint:unparam
func funcTestEventLogEventData(ctx wasmlib.ScFuncContext, f *TestEventLogEventDataContext) {
	ctx.Event(MsgTestingEvent)
}

func funcTestEventLogGenericData(ctx wasmlib.ScFuncContext, f *TestEventLogGenericDataContext) {
	event := MsgCounterNumber + f.Params.Counter().String()
	ctx.Event(event)
}

//nolint:unparam
func funcTestPanicFullEP(ctx wasmlib.ScFuncContext, f *TestPanicFullEPContext) {
	ctx.Panic(MsgFullPanic)
}

//func funcWithdrawToChain(ctx wasmlib.ScFuncContext, f *WithdrawToChainContext) {
//	//coreaccounts.ScFuncs.Withdraw(ctx).Func.PostToChain(f.Params.ChainID().Value())
//}

func funcWithdrawFromChain(ctx wasmlib.ScFuncContext, f *WithdrawFromChainContext) {
	targetChain := f.Params.ChainID().Value()
	iotasToWithdrawal := f.Params.IotasWithdrawal().Value()
	// gasBudget := f.Params.GasBudget().Value()

	// TODO more
	availableIotas := ctx.Allowance().Iotas()
	// requiredDustDeposit := ctx.EstimateRequiredDustDeposit(request)
	if availableIotas < 1000 {
		ctx.Panic("no enough iotas sent to cover dust deposit")
	}
	transfer := wasmlib.NewScTransferFromBalances(ctx.Allowance())
	ctx.TransferAllowed(ctx.AccountID(), transfer, false)

	//request := iscp.RequestParameters{
	//	TargetAddress:  targetChain.AsAddress(),
	//	FungibleTokens: iscp.NewTokensIotas(availableIotas),
	//	Metadata: &iscp.SendMetadata{
	//		TargetContract: accounts.Contract.Hname(),
	//		EntryPoint:     accounts.FuncWithdraw.Hname(),
	//		GasBudget:      gasBudget,
	//		Allowance:      iscp.NewAllowanceIotas(iotasToWithdrawal),
	//	},
	//}

	withdraw := coreaccounts.ScFuncs.Withdraw(ctx)
	withdraw.Func.TransferIotas(iotasToWithdrawal).PostToChain(targetChain)
}

func viewCheckContextFromViewEP(ctx wasmlib.ScViewContext, f *CheckContextFromViewEPContext) {
	ctx.Require(f.Params.AgentID().Value() == ctx.AccountID(), "fail: agentID")
	ctx.Require(f.Params.ChainID().Value() == ctx.ChainID(), "fail: chainID")
	ctx.Require(f.Params.ChainOwnerID().Value() == ctx.ChainOwnerID(), "fail: chainOwnerID")
	ctx.Require(f.Params.ContractCreator().Value() == ctx.ContractCreator(), "fail: contractCreator")
}

func viewFibonacci(ctx wasmlib.ScViewContext, f *FibonacciContext) {
	n := f.Params.IntValue().Value()
	if n == 0 || n == 1 {
		f.Results.IntValue().SetValue(n)
		return
	}

	fib := ScFuncs.Fibonacci(ctx)
	fib.Params.IntValue().SetValue(n - 1)
	fib.Func.Call()
	n1 := fib.Results.IntValue().Value()

	fib.Params.IntValue().SetValue(n - 2)
	fib.Func.Call()
	n2 := fib.Results.IntValue().Value()

	f.Results.IntValue().SetValue(n1 + n2)
}

func viewGetCounter(ctx wasmlib.ScViewContext, f *GetCounterContext) {
	f.Results.Counter().SetValue(f.State.Counter().Value())
}

func viewGetInt(ctx wasmlib.ScViewContext, f *GetIntContext) {
	name := f.Params.Name().Value()
	value := f.State.Ints().GetInt64(name)
	ctx.Require(value.Exists(), "param '"+name+"' not found")
	f.Results.Values().GetInt64(name).SetValue(value.Value())
}

//nolint:unparam
func viewJustView(ctx wasmlib.ScViewContext, f *JustViewContext) {
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

//nolint:unparam
func viewTestCallPanicViewEPFromView(ctx wasmlib.ScViewContext, f *TestCallPanicViewEPFromViewContext) {
	ctx.Log("will be calling entry point '" + ViewTestPanicViewEP + "' from view EP")
	ScFuncs.TestPanicViewEP(ctx).Func.Call()
}

func viewTestChainOwnerIDView(ctx wasmlib.ScViewContext, f *TestChainOwnerIDViewContext) {
	f.Results.ChainOwnerID().SetValue(ctx.ChainOwnerID())
}

//nolint:unparam
func viewTestPanicViewEP(ctx wasmlib.ScViewContext, f *TestPanicViewEPContext) {
	ctx.Panic(MsgViewPanic)
}

func viewTestSandboxCall(ctx wasmlib.ScViewContext, f *TestSandboxCallContext) {
	getChainInfo := coregovernance.ScFuncs.GetChainInfo(ctx)
	getChainInfo.Func.Call()
	f.Results.SandboxCall().SetValue(getChainInfo.Results.Description().Value())
}

//nolint:unparam
func funcTestBlockContext1(ctx wasmlib.ScFuncContext, f *TestBlockContext1Context) {
	ctx.Panic(MsgCoreOnlyPanic)
}

//nolint:unparam
func funcTestBlockContext2(ctx wasmlib.ScFuncContext, f *TestBlockContext2Context) {
	ctx.Panic(MsgCoreOnlyPanic)
}

//nolint:unparam
func viewGetStringValue(ctx wasmlib.ScViewContext, f *GetStringValueContext) {
	ctx.Panic(MsgCoreOnlyPanic)
	// varName := f.Params.VarName().Value()
	// value := f.State.Strings().GetString(varName).Value()
	// f.Results.Vars().GetString(varName).SetValue(value)
}

//nolint:unparam
func viewInfiniteLoopView(ctx wasmlib.ScViewContext, f *InfiniteLoopViewContext) {
	//nolint:staticcheck
	for {
		// do nothing, just waste gas
	}
}

//nolint:unparam
func funcClaimAllowance(ctx wasmlib.ScFuncContext, f *ClaimAllowanceContext) {
	allowance := ctx.Allowance()
	transfer := wasmlib.NewScTransferFromBalances(allowance)
	ctx.TransferAllowed(ctx.AccountID(), transfer, false)
}

//nolint:unparam
func funcEstimateMinDust(ctx wasmlib.ScFuncContext, f *EstimateMinDustContext) {
	provided := ctx.Allowance().Iotas()
	dummy := ScFuncs.EstimateMinDust(ctx)
	required := ctx.EstimateDust(dummy.Func)
	ctx.Require(provided >= required, "not enough funds")
}

//nolint:unparam
func funcPingAllowanceBack(ctx wasmlib.ScFuncContext, f *PingAllowanceBackContext) {
	caller := ctx.Caller()
	ctx.Require(caller.IsAddress(), "pingAllowanceBack: caller expected to be a L1 address")
	transfer := wasmlib.NewScTransferFromBalances(ctx.Allowance())
	ctx.TransferAllowed(ctx.AccountID(), transfer, false)
	ctx.Send(caller.Address(), transfer)
}

func funcSendLargeRequest(ctx wasmlib.ScFuncContext, f *SendLargeRequestContext) {
}

//nolint:unparam
func funcSendNFTsBack(ctx wasmlib.ScFuncContext, f *SendNFTsBackContext) {
	address := ctx.Caller().Address()
	allowance := ctx.Allowance()
	transfer := wasmlib.NewScTransferFromBalances(allowance)
	ctx.TransferAllowed(ctx.AccountID(), transfer, false)
	for _, nftID := range allowance.NftIDs() {
		transfer = wasmlib.NewScTransferNFT(nftID)
		ctx.Send(address, transfer)
	}
}

//nolint:unparam
func funcSplitFunds(ctx wasmlib.ScFuncContext, f *SplitFundsContext) {
	iotas := ctx.Allowance().Iotas()
	address := ctx.Caller().Address()
	transfer := wasmlib.NewScTransferIotas(200)
	for ; iotas >= 200; iotas -= 200 {
		ctx.TransferAllowed(ctx.AccountID(), transfer, false)
		ctx.Send(address, transfer)
	}
}

//nolint:unparam
func funcSplitFundsNativeTokens(ctx wasmlib.ScFuncContext, f *SplitFundsNativeTokensContext) {
	iotas := ctx.Allowance().Iotas()
	address := ctx.Caller().Address()
	transfer := wasmlib.NewScTransferIotas(iotas)
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
