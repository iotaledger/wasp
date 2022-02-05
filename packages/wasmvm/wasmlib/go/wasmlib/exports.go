// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

type ScExportMap struct {
	Names []string
	Funcs []ScFuncContextFunction
	Views []ScViewContextFunction
}

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

func ScExportsExport(exportMap *ScExportMap) {
	ExportWasmTag()

	for i := range exportMap.Funcs {
		ExportName(int32(i), exportMap.Names[i])
	}

	offset := len(exportMap.Funcs)
	for i := range exportMap.Views {
		ExportName(int32(i)|0x8000, exportMap.Names[offset+i])
	}
}

func ScExportsCall(index int32, exportMap *ScExportMap) {
	if (index & 0x8000) == 0 {
		// mutable full function, invoke with a func context
		exportMap.Funcs[index](ScFuncContext{})
		return
	}
	// immutable view function, invoke with a view context
	exportMap.Views[index&0x7fff](ScViewContext{})
}

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
