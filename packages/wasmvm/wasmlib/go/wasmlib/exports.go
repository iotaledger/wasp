// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

type ScExportMap struct {
	Names []string
	Funcs []ScFuncContextFunction
	Views []ScViewContextFunction
}

func FuncError(ctx ScFuncContext) {
	Panic("Invalid core func call")
}

func ViewError(ctx ScViewContext) {
	Panic("Invalid core view call")
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func ScExportsExport(exportMap *ScExportMap) {
	ExportName(-1, "WASM::GO")

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
