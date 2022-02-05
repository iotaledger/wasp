// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

type WasmFuncTable struct {
	codeToFunc  map[uint32]string
	funcToCode  map[string]uint32
	funcToIndex map[string]int32
}

func NewWasmFuncTable() *WasmFuncTable {
	return &WasmFuncTable{
		codeToFunc:  make(map[uint32]string),
		funcToCode:  make(map[string]uint32),
		funcToIndex: make(map[string]int32),
	}
}

func (t *WasmFuncTable) FunctionFromCode(code uint32) string {
	return t.codeToFunc[code]
}

func (t *WasmFuncTable) SetExport(index int32, functionName string) {
	funcIndex, ok := t.funcToIndex[functionName]
	if ok {
		// paranoia check
		if funcIndex != index {
			panic("SetExport: duplicate function name")
		}
		return
	}

	hashedName := uint32(iscp.Hn(functionName))
	_, ok = t.codeToFunc[hashedName]
	if ok {
		panic("SetExport: duplicate hashed name")
	}

	t.codeToFunc[hashedName] = functionName
	t.funcToCode[functionName] = hashedName
	t.funcToIndex[functionName] = index
}
