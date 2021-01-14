package test_sandbox

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// ParamCallOption
// ParamCallDepth
// ParamHname
func callOnChain(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof(FuncCallOnChain)
	callOption, exists, err := codec.DecodeString(ctx.Params().MustGet(ParamCallOption))
	if err != nil {
		ctx.Log().Panicf("%v", err)
	}
	if !exists {
		callOption = ""
	}
	callDepth, exists, err := codec.DecodeInt64(ctx.Params().MustGet(ParamCallDepth))
	if err != nil {
		ctx.Log().Panicf("%v", err)
	}
	if !exists {
		ctx.Log().Panicf("parameter '%s' wasn't provided", ParamCallDepth)
	}
	hname, exists, err := codec.DecodeHname(ctx.Params().MustGet(ParamHname))
	if err != nil {
		ctx.Log().Panicf("%v", err)
	}
	if !exists {
		ctx.Log().Panicf("parameter '%s' wasn't provided", ParamHname)
	}
	ctx.Log().Infof("call depth = %d, option = %s, hname = %s", callDepth, callOption, hname)
	if callDepth <= 0 {
		return nil, nil
	}
	callDepth--

	return ctx.Call(hname, coretypes.Hn(FuncCallOnChain), codec.MakeDict(map[string]interface{}{
		ParamCallOption: []byte(callOption),
		ParamCallDepth:  callDepth,
		ParamHname:      hname,
	}), nil)
}

// ParamIntParamName
// ParamIntParamValue
func setInt(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof(FuncSetInt)
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
func getInt(ctx vmtypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Infof(FuncGetInt)
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
