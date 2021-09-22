// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

//export on_call
func OnCall(index int32) {
	ctx := ScFuncContext{}
	ctx.Require(GetObjectID(OBJ_ID_ROOT, KeyState, TYPE_MAP) == OBJ_ID_STATE, "object id mismatch")
	ctx.Require(GetObjectID(OBJ_ID_ROOT, KeyParams, TYPE_MAP) == OBJ_ID_PARAMS, "object id mismatch")
	ctx.Require(GetObjectID(OBJ_ID_ROOT, KeyResults, TYPE_MAP) == OBJ_ID_RESULTS, "object id mismatch")

	if (index & 0x8000) == 0 {
		AddFunc(nil)[index](ctx)
		return
	}

	AddView(nil)[index&0x7fff](ScViewContext{})
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScExports struct {
	exports ScMutableStringArray
}

func NewScExports() ScExports {
	exports := Root.GetStringArray(KeyExports)
	// tell host what our highest predefined key is
	// this helps detect missing or extra keys
	exports.GetString(int32(KeyZzzzzzz)).SetValue("Go:KeyZzzzzzz")
	return ScExports{exports: exports}
}

func (ctx ScExports) AddFunc(name string, f func(ctx ScFuncContext)) {
	index := int32(len(AddFunc(f))) - 1
	ctx.exports.GetString(index).SetValue(name)
}

func (ctx ScExports) AddView(name string, v func(ctx ScViewContext)) {
	index := int32(len(AddView(v))) - 1
	ctx.exports.GetString(index | 0x8000).SetValue(name)
}
