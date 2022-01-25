// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

//export on_call
func OnCall(index int32) {
	if (index & 0x8000) == 0 {
		AddFunc(nil)[index](ScFuncContext{})
		return
	}

	AddView(nil)[index&0x7fff](ScViewContext{})
}

func FuncError(ctx ScFuncContext) {
	Panic("Invalid core func call")
}

func ViewError(ctx ScViewContext) {
	Panic("Invalid core view call")
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScExports struct{}

func NewScExports() ScExports {
	ExportWasmTag()
	return ScExports{}
}

func (ctx ScExports) AddFunc(name string, f ScFuncContextFunction) {
	index := int32(len(AddFunc(f))) - 1
	ExportName(index, name)
}

func (ctx ScExports) AddView(name string, v ScViewContextFunction) {
	index := int32(len(AddView(v))) - 1
	ExportName(index|0x8000, name)
}
