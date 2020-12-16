package test_chainlog

import (
	"strconv"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	log "github.com/iotaledger/wasp/packages/vm/builtinvm/chainlog"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("testchainlog.initialize.begin")
	ctx.Eventf("testchainlog.initialize.success hname = %s", Interface.Hname().String())
	return nil, nil
}

func example_TestStore(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("testchainlog.exampleFunction")
	state := ctx.State()
	params := ctx.Params()
	inc, ok, err := codec.DecodeInt64(params.MustGet(VarCounter))
	if err != nil {
		return nil, err
	}
	if !ok {
		inc = 1
	}
	val, _, _ := codec.DecodeInt64(state.MustGet(VarCounter))
	ctx.Eventf("incCounter: increasing counter value %d by %d", val, inc)
	state.Set(VarCounter, codec.EncodeInt64(val+inc))

	hname := log.Interface.Hname()
	initParams := dict.New()

	initParams.Set(log.ParamLog, []byte("some test text"))
	initParams.Set(log.ParamType, codec.EncodeInt64(log.TR_GENERIC_DATA))

	_, err = ctx.Call(hname, coretypes.Hn(log.FuncStoreLog), initParams, nil)

	if err != nil {
		return nil, err
	}

	return nil, nil

}

func example_TestGetLasts3(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("testchainlog.exampleFunction")
	hname := log.Interface.Hname()

	initParams := dict.New()
	initParams.Set(log.ParamLog, []byte("PostRequest Number ONE"))
	initParams.Set(log.ParamType, codec.EncodeInt64(log.TR_TOKEN_TRANSFER))

	_, err := ctx.Call(hname, coretypes.Hn(log.FuncStoreLog), initParams, nil)
	if err != nil {
		return nil, err
	}

	initParams = dict.New()
	initParams.Set(log.ParamLog, []byte("PostRequest Number TWO"))
	initParams.Set(log.ParamType, codec.EncodeInt64(log.TR_TOKEN_TRANSFER))

	_, err = ctx.Call(hname, coretypes.Hn(log.FuncStoreLog), initParams, nil)
	if err != nil {
		return nil, err
	}

	initParams = dict.New()
	initParams.Set(log.ParamLog, []byte("PostRequest Number THREE"))
	initParams.Set(log.ParamType, codec.EncodeInt64(log.TR_TOKEN_TRANSFER))

	_, err = ctx.Call(hname, coretypes.Hn(log.FuncStoreLog), initParams, nil)
	if err != nil {
		return nil, err
	}

	return nil, nil

}

func example_TestGeneric(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("testchainlog.exampleFunction")
	params := ctx.Params()
	inc, ok, err := codec.DecodeInt64(params.MustGet(VarCounter))
	if err != nil {
		return nil, err
	}
	if !ok {
		inc = 1
	}

	typeRecord, ok, err := codec.DecodeInt64(params.MustGet(TypeRecord))
	if err != nil {
		return nil, err
	}
	if !ok {
		inc = 1
	}
	hname := log.Interface.Hname()

	initParams := dict.New()
	initParams.Set(log.ParamLog, []byte("PostRequest Number "+strconv.Itoa(int(inc))))
	initParams.Set(log.ParamType, codec.EncodeInt64(typeRecord))

	_, err = ctx.Call(hname, coretypes.Hn(log.FuncStoreLog), initParams, nil)
	if err != nil {
		return nil, err
	}

	return nil, nil

}
