// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

type ScExportMap struct {
	Names []string
	Funcs []ScFuncContextFunction
	Views []ScViewContextFunction
}

// general entrypoint for the host to call any SC function
// the host will pass the index of one of the entry points
// that was provided by on_load during SC initialization
func (m *ScExportMap) Dispatch(index int32) *ScExportMap {
	if index < 0 {
		if index < -1 {
			return m
		}
		// special dispatch for exporting entry points to host
		m.Export(ExportName)
		return nil
	}

	if (index & 0x8000) == 0 {
		// mutable full function, invoke with a WasmLib func call context
		m.Funcs[index](ScFuncContext{})
		return nil
	}
	// immutable view function, invoke with a WasmLib view call context
	m.Views[index&0x7fff](ScViewContext{})
	return nil
}

// constructs the symbol export context for the on_load function
func (m *ScExportMap) Export(exportName func(index int32, name string)) {
	exportName(-1, "WASM::GO")

	for i := range m.Funcs {
		exportName(int32(i), m.Names[i])
	}

	offset := len(m.Funcs)
	for i := range m.Views {
		exportName(int32(i)|0x8000, m.Names[offset+i])
	}
}

func FuncError(ctx ScFuncContext) {
	Panic("Invalid core func call")
}

func ViewError(ctx ScViewContext) {
	Panic("Invalid core view call")
}
