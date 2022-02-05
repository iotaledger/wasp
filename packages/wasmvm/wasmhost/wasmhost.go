// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

type WasmStore interface {
	GetContext(id int32) *WasmContext
}

type WasmHost struct {
	codeToFunc  map[uint32]string
	funcToCode  map[string]uint32
	funcToIndex map[string]int32
	store       WasmStore
	vm          WasmVM
}

func (host *WasmHost) Init() {
	host.codeToFunc = make(map[uint32]string)
	host.funcToCode = make(map[string]uint32)
	host.funcToIndex = make(map[string]int32)
}

func (host *WasmHost) FunctionFromCode(code uint32) string {
	return host.codeToFunc[code]
}

func (host *WasmHost) SetExport(index int32, functionName string) {
	funcIndex, ok := host.funcToIndex[functionName]
	if ok {
		// TODO remove this check?
		if funcIndex != index {
			panic("SetExport: duplicate function name")
		}
		return
	}

	hashedName := uint32(iscp.Hn(functionName))
	_, ok = host.codeToFunc[hashedName]
	if ok {
		panic("SetExport: duplicate hashed name")
	}

	host.codeToFunc[hashedName] = functionName
	host.funcToCode[functionName] = hashedName
	host.funcToIndex[functionName] = index
}
