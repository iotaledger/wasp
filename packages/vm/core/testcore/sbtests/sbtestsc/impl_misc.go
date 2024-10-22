package sbtestsc

import (
	"strings"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/samber/lo"
)

// ParamCallOption
// ParamCallIntParam
// ParamHnameContract
func callOnChain(ctx isc.Sandbox) isc.CallArguments {
	ctx.Log().Debugf(FuncCallOnChain.Name)
	params := ctx.Params()
	paramIn := isc.MustArgAt[uint64](params, 0)
	hnameContract := isc.MustOptionalArgAt[isc.Hname](params, 1, ctx.Contract())
	hnameEP := isc.MustOptionalArgAt[isc.Hname](params, 2, FuncCallOnChain.Hname())

	counter := codec.StateGet[uint64](ctx.State(), VarCounter)
	codec.StateSet(ctx.State(), VarCounter, counter+1)

	ctx.Log().Infof("param IN = %d, hnameContract = %s, hnameEP = %s, counter = %d",
		paramIn, hnameContract, hnameEP, counter)

	return ctx.Call(isc.NewMessage(hnameContract, hnameEP,
		isc.NewCallArguments(codec.Encode(paramIn)),
	), nil)
}

func incCounter(ctx isc.Sandbox) {
	state := ctx.State()
	counter := codec.StateGet[uint64](state, VarCounter)
	codec.StateSet(state, VarCounter, counter+1)
}

func getCounter(ctx isc.SandboxView) uint64 {
	return codec.StateGet[uint64](ctx.StateR(), VarCounter)
}

func runRecursion(ctx isc.Sandbox) isc.CallArguments {
	params := ctx.Params()
	depth := isc.MustArgAt[uint64](params, 0)
	if depth == 0 {
		return nil
	}
	return ctx.Call(isc.NewMessage(
		ctx.Contract(),
		FuncCallOnChain.Hname(),
		isc.NewCallArguments(
			codec.Encode(depth-1),
			codec.EncodeOptionalNone(),
			codec.EncodeOptionalSome(FuncRunRecursion.Hname()),
		),
	), nil)
}

func fibonacci(n uint64) uint64 {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func getFibonacci(ctx isc.SandboxView, n uint64) uint64 {
	ctx.Log().Infof("fibonacci( %d )", n)
	return fibonacci(n)
}

func getFibonacciIndirect(ctx isc.SandboxView, n uint64) uint64 {
	ctx.Log().Infof("fibonacciIndirect( %d )", n)
	if n <= 1 {
		return n
	}

	msg := FuncGetFibonacciIndirect.Message(n - 1)
	msg.Target.Contract = ctx.Contract()
	ret1 := ctx.CallView(msg)
	n1 := lo.Must(FuncGetFibonacciIndirect.DecodeOutput(ret1))

	msg = FuncGetFibonacciIndirect.Message(n - 2)
	msg.Target.Contract = ctx.Contract()
	ret2 := ctx.CallView(msg)
	n2 := lo.Must(FuncGetFibonacciIndirect.DecodeOutput(ret2))

	return n1 + n2
}

// calls the "fib indirect" view and stores the result in the state
func calcFibonacciIndirectStoreValue(ctx isc.Sandbox, n uint64) {
	msg := FuncGetFibonacciIndirect.Message(n - 1)
	msg.Target.Contract = ctx.Contract()
	ret := ctx.CallView(msg)
	retN := lo.Must(FuncGetFibonacciIndirect.DecodeOutput(ret))
	codec.StateSet(ctx.State(), VarN, retN)
}

func viewFibResult(ctx isc.SandboxView) uint64 {
	return codec.StateGet[uint64](ctx.StateR(), VarN)
}

// ParamIntParamName
// ParamIntParamValue
func setInt(ctx isc.Sandbox, k string, v int64) {
	ctx.Log().Infof(FuncSetInt.Name)
	codec.StateSet(ctx.State(), kv.Key(k), v)
}

// ParamIntParamName
func getInt(ctx isc.SandboxView, k string) int64 {
	ctx.Log().Infof(FuncGetInt.Name)
	return codec.StateGet[int64](ctx.StateR(), kv.Key(k))
}

func infiniteLoop(ctx isc.Sandbox) {
	for {
		// do nothing, just waste gas
		ctx.State().Set("foo", []byte(strings.Repeat("dummy data", 1000)))
	}
}

func infiniteLoopView(ctx isc.SandboxView) {
	for {
		// do nothing, just waste gas
		ctx.CallView(isc.NewMessage(ctx.Contract(), FuncGetCounter.Hname(), isc.NewCallArguments()))
	}
}
