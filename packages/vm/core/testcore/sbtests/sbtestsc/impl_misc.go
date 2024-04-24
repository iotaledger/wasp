package sbtestsc

import (
	"strings"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

// ParamCallOption
// ParamCallIntParam
// ParamHnameContract
func callOnChain(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Debugf(FuncCallOnChain.Name)
	params := ctx.Params()
	paramIn := params.MustGetUint64(ParamN)
	hnameContract := params.MustGetHname(ParamHnameContract, ctx.Contract())
	hnameEP := params.MustGetHname(ParamHnameEP, FuncCallOnChain.Hname())

	state := ctx.State()
	decoder := kvdecoder.New(state, ctx.Log())
	counter := decoder.MustGetUint64(VarCounter, 0)
	state.Set(VarCounter, codec.Uint64.Encode(counter+1))

	ctx.Log().Infof("param IN = %d, hnameContract = %s, hnameEP = %s, counter = %d",
		paramIn, hnameContract, hnameEP, counter)

	return ctx.Call(hnameContract, hnameEP, codec.MakeDict(map[string]interface{}{
		ParamN: paramIn,
	}), nil)
}

func incCounter(ctx isc.Sandbox) dict.Dict {
	state := ctx.State()
	decoder := kvdecoder.New(state, ctx.Log())
	counter := decoder.MustGetUint64(VarCounter, 0)
	state.Set(VarCounter, codec.Uint64.Encode(counter+1))
	return nil
}

func getCounter(ctx isc.SandboxView) dict.Dict {
	ret := dict.New()
	decoder := kvdecoder.New(ctx.StateR(), ctx.Log())
	counter := decoder.MustGetUint64(VarCounter, 0)
	ret.Set(VarCounter, codec.Uint64.Encode(counter))
	return ret
}

func runRecursion(ctx isc.Sandbox) dict.Dict {
	params := ctx.Params()
	depth := params.MustGetUint64(ParamN)
	if depth == 0 {
		return nil
	}
	return ctx.Call(ctx.Contract(), FuncCallOnChain.Hname(), codec.MakeDict(map[string]interface{}{
		ParamHnameEP: FuncRunRecursion.Hname(),
		ParamN:       depth - 1,
	}), nil)
}

func fibonacci(n uint64) uint64 {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func getFibonacci(ctx isc.SandboxView) dict.Dict {
	params := ctx.Params()
	n := params.MustGetUint64(ParamN)
	ctx.Log().Infof("fibonacci( %d )", n)
	result := fibonacci(n)
	ret := dict.New()
	ret.Set(ParamN, codec.Uint64.Encode(result))
	return ret
}

func getFibonacciIndirect(ctx isc.SandboxView) dict.Dict {
	params := ctx.Params()

	n := params.MustGetUint64(ParamN)
	ctx.Log().Infof("fibonacciIndirect( %d )", n)
	ret := dict.New()
	if n <= 1 {
		ret.Set(ParamN, codec.Uint64.Encode(n))
		return ret
	}

	hContract := ctx.Contract()
	hFibonacci := FuncGetFibonacciIndirect.Hname()
	ret1 := ctx.CallView(hContract, hFibonacci, codec.MakeDict(map[string]interface{}{
		ParamN: n - 1,
	}))
	decoder := kvdecoder.New(ret1, ctx.Log())
	n1 := decoder.MustGetUint64(ParamN)

	ret2 := ctx.CallView(hContract, hFibonacci, codec.MakeDict(map[string]interface{}{
		ParamN: n - 2,
	}))
	decoder = kvdecoder.New(ret2, ctx.Log())
	n2 := decoder.MustGetUint64(ParamN)

	ret.Set(ParamN, codec.Uint64.Encode(n1+n2))
	return ret
}

// calls the "fib indirect" view and stores the result in the state
func calcFibonacciIndirectStoreValue(ctx isc.Sandbox) dict.Dict {
	ret := ctx.CallView(ctx.Contract(), FuncGetFibonacciIndirect.Hname(), dict.Dict{
		ParamN: ctx.Params().Get(ParamN),
	})
	ctx.State().Set(ParamN, ret.Get(ParamN))
	return nil
}

func viewFibResult(ctx isc.SandboxView) dict.Dict {
	return dict.Dict{
		ParamN: ctx.StateR().Get(ParamN),
	}
}

// ParamIntParamName
// ParamIntParamValue
func setInt(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Infof(FuncSetInt.Name)
	params := ctx.Params()
	paramName := params.MustGetString(ParamIntParamName)
	paramValue := params.MustGetInt64(ParamIntParamValue)
	ctx.State().Set(kv.Key(paramName), codec.Int64.Encode(paramValue))
	return nil
}

// ParamIntParamName
func getInt(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Infof(FuncGetInt.Name)
	params := ctx.Params()
	paramName := params.MustGetString(ParamIntParamName)
	decoder := kvdecoder.New(ctx.StateR(), ctx.Log())
	paramValue := decoder.MustGetInt64(kv.Key(paramName), 0)
	ret := dict.New()
	ret.Set(kv.Key(paramName), codec.Int64.Encode(paramValue))
	return ret
}

func infiniteLoop(ctx isc.Sandbox) dict.Dict {
	for {
		// do nothing, just waste gas
		ctx.State().Set("foo", []byte(strings.Repeat("dummy data", 1000)))
	}
}

func infiniteLoopView(ctx isc.SandboxView) dict.Dict {
	for {
		// do nothing, just waste gas
		ctx.CallView(ctx.Contract(), FuncGetCounter.Hname(), dict.Dict{})
	}
}
