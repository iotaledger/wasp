package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	assert2 "github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

// ParamCallOption
// ParamCallIntParam
// ParamHnameContract
func callOnChain(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf(FuncCallOnChain.Name)
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	paramIn := params.MustGetInt64(ParamIntParamValue)
	hnameContract := params.MustGetHname(ParamHnameContract, ctx.Contract())
	hnameEP := params.MustGetHname(ParamHnameEP, FuncCallOnChain.Hname())

	state := kvdecoder.New(ctx.State(), ctx.Log())
	counter := state.MustGetInt64(VarCounter, 0)
	ctx.State().Set(VarCounter, codec.EncodeInt64(counter+1))

	ctx.Log().Infof("param IN = %d, hnameContract = %s hnameEP = %s counter = %d",
		paramIn, hnameContract, hnameEP, counter)

	return ctx.Call(hnameContract, hnameEP, codec.MakeDict(map[string]interface{}{
		ParamIntParamValue: paramIn,
	}), nil)
}

func incCounter(ctx iscp.Sandbox) (dict.Dict, error) {
	state := kvdecoder.New(ctx.State(), ctx.Log())
	counter := state.MustGetInt64(VarCounter, 0)
	ctx.State().Set(VarCounter, codec.EncodeInt64(counter+1))
	return nil, nil
}

func getCounter(ctx iscp.SandboxView) (dict.Dict, error) {
	ret := dict.New()
	state := kvdecoder.New(ctx.State(), ctx.Log())
	counter := state.MustGetInt64(VarCounter, 0)
	ret.Set(VarCounter, codec.EncodeInt64(counter))
	return ret, nil
}

func runRecursion(ctx iscp.Sandbox) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	depth := params.MustGetInt64(ParamIntParamValue)
	if depth <= 0 {
		return nil, nil
	}
	return ctx.Call(ctx.Contract(), FuncCallOnChain.Hname(), codec.MakeDict(map[string]interface{}{
		ParamHnameEP:       FuncRunRecursion.Hname(),
		ParamIntParamValue: depth - 1,
	}), nil)
}

func getFibonacci(ctx iscp.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert2.NewAssert(ctx.Log())

	callInt := params.MustGetInt64(ParamIntParamValue)
	ctx.Log().Infof("fibonacci( %d )", callInt)
	ret := dict.New()
	if callInt == 0 || callInt == 1 {
		ret.Set(ParamIntParamValue, codec.EncodeInt64(callInt))
		return ret, nil
	}
	r1, err := ctx.Call(ctx.Contract(), FuncGetFibonacci.Hname(), codec.MakeDict(map[string]interface{}{
		ParamIntParamValue: callInt - 1,
	}))
	a.RequireNoError(err)
	result := kvdecoder.New(r1, ctx.Log())
	r1val := result.MustGetInt64(ParamIntParamValue)

	r2, err := ctx.Call(ctx.Contract(), FuncGetFibonacci.Hname(), codec.MakeDict(map[string]interface{}{
		ParamIntParamValue: callInt - 2,
	}))
	a.RequireNoError(err)
	result = kvdecoder.New(r2, ctx.Log())
	r2val := result.MustGetInt64(ParamIntParamValue)
	ret.Set(ParamIntParamValue, codec.EncodeInt64(r1val+r2val))
	return ret, nil
}

// ParamIntParamName
// ParamIntParamValue
func setInt(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof(FuncSetInt.Name)
	paramName, exists, err := codec.DecodeString(ctx.Params().MustGet(ParamIntParamName))
	if err != nil {
		ctx.Log().Panicf("%v", err)
	}
	if !exists {
		ctx.Log().Panicf("parameter '%s' wasn't provided", ParamIntParamName)
	}
	paramValue, exists, err := codec.DecodeInt64(ctx.Params().MustGet(ParamIntParamValue))
	if err != nil {
		ctx.Log().Panicf("%v", err)
	}
	if !exists {
		ctx.Log().Panicf("parameter '%s' wasn't provided", ParamIntParamValue)
	}
	ctx.State().Set(kv.Key(paramName), codec.EncodeInt64(paramValue))
	return nil, nil
}

// ParamIntParamName
func getInt(ctx iscp.SandboxView) (dict.Dict, error) {
	ctx.Log().Infof(FuncGetInt.Name)
	paramName, exists, err := codec.DecodeString(ctx.Params().MustGet(ParamIntParamName))
	if err != nil {
		ctx.Log().Panicf("%v", err)
	}
	if !exists {
		ctx.Log().Panicf("parameter '%s' wasn't provided", ParamIntParamName)
	}
	paramValue, exists, err := codec.DecodeInt64(ctx.State().MustGet(kv.Key(paramName)))
	if err != nil {
		ctx.Log().Panicf("%v", err)
	}
	if !exists {
		paramValue = 0
	}
	ret := dict.New()
	ret.Set(kv.Key(paramName), codec.EncodeInt64(paramValue))
	return ret, nil
}
