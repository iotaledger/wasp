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
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	paramIn := params.MustGetUint64(ParamN)
	hnameContract := params.MustGetHname(ParamHnameContract, ctx.Contract())
	hnameEP := params.MustGetHname(ParamHnameEP, FuncCallOnChain.Hname())

	state := kvdecoder.New(ctx.State(), ctx.Log())
	counter := state.MustGetUint64(VarCounter, 0)
	ctx.State().Set(VarCounter, codec.EncodeUint64(counter+1))

	ctx.Log().Infof("param IN = %d, hnameContract = %s, hnameEP = %s, counter = %d",
		paramIn, hnameContract, hnameEP, counter)

	return ctx.Call(hnameContract, hnameEP, codec.MakeDict(map[string]interface{}{
		ParamN: paramIn,
	}), nil)
}

func incCounter(ctx isc.Sandbox) dict.Dict {
	state := kvdecoder.New(ctx.State(), ctx.Log())
	counter := state.MustGetUint64(VarCounter, 0)
	ctx.State().Set(VarCounter, codec.EncodeUint64(counter+1))
	return nil
}

func getCounter(ctx isc.SandboxView) dict.Dict {
	ret := dict.New()
	state := kvdecoder.New(ctx.State(), ctx.Log())
	counter := state.MustGetUint64(VarCounter, 0)
	ret.Set(VarCounter, codec.EncodeUint64(counter))
	return ret
}

func runRecursion(ctx isc.Sandbox) dict.Dict {
	params := kvdecoder.New(ctx.Params(), ctx.Log())
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
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	n := params.MustGetUint64(ParamN)
	ctx.Log().Infof("fibonacci( %d )", n)
	result := fibonacci(n)
	ret := dict.New()
	ret.Set(ParamN, codec.EncodeUint64(result))
	return ret
}

func getFibonacciIndirect(ctx isc.SandboxView) dict.Dict {
	params := kvdecoder.New(ctx.Params(), ctx.Log())

	n := params.MustGetUint64(ParamN)
	ctx.Log().Infof("fibonacciIndirect( %d )", n)
	ret := dict.New()
	if n <= 1 {
		ret.Set(ParamN, codec.EncodeUint64(n))
		return ret
	}

	ret1 := ctx.CallView(ctx.Contract(), FuncGetFibonacciIndirect.Hname(), codec.MakeDict(map[string]interface{}{
		ParamN: n - 1,
	}))
	result := kvdecoder.New(ret1, ctx.Log())
	n1 := result.MustGetUint64(ParamN)

	ret2 := ctx.CallView(ctx.Contract(), FuncGetFibonacciIndirect.Hname(), codec.MakeDict(map[string]interface{}{
		ParamN: n - 2,
	}))
	result = kvdecoder.New(ret2, ctx.Log())
	n2 := result.MustGetUint64(ParamN)

	ret.Set(ParamN, codec.EncodeUint64(n1+n2))
	return ret
}

// ParamIntParamName
// ParamIntParamValue
func setInt(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Infof(FuncSetInt.Name)
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	paramName := params.MustGetString(ParamIntParamName)
	paramValue := params.MustGetInt64(ParamIntParamValue)
	ctx.State().Set(kv.Key(paramName), codec.EncodeInt64(paramValue))
	return nil
}

// ParamIntParamName
func getInt(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Infof(FuncGetInt.Name)
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	paramName := params.MustGetString(ParamIntParamName)
	state := kvdecoder.New(ctx.State(), ctx.Log())
	paramValue := state.MustGetInt64(kv.Key(paramName), 0)
	ret := dict.New()
	ret.Set(kv.Key(paramName), codec.EncodeInt64(paramValue))
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
