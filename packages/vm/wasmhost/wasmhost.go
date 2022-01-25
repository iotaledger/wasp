// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"errors"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
)

type WasmStore interface {
	GetContext(id int32) *WasmContext
}

type WasmHost struct {
	codeToFunc  map[uint32]string
	funcToCode  map[string]uint32
	funcToIndex map[string]int32
	funcs       []wasmlib.ScFuncContextFunction
	views       []wasmlib.ScViewContextFunction
	store       WasmStore
	vm          WasmVM
}

func (host *WasmHost) AddFunc(f wasmlib.ScFuncContextFunction) []wasmlib.ScFuncContextFunction {
	if f != nil {
		host.funcs = append(host.funcs, f)
	}
	return host.funcs
}

func (host *WasmHost) AddView(v wasmlib.ScViewContextFunction) []wasmlib.ScViewContextFunction {
	if v != nil {
		host.views = append(host.views, v)
	}
	return host.views
}

func (host *WasmHost) InitVM(vm WasmVM, store WasmStore) error {
	host.store = store
	return vm.LinkHost(vm, host)
}

func (host *WasmHost) Init() {
	host.codeToFunc = make(map[uint32]string)
	host.funcToCode = make(map[string]uint32)
	host.funcToIndex = make(map[string]int32)
}

func (host *WasmHost) FunctionFromCode(code uint32) string {
	return host.codeToFunc[code]
}

func (host *WasmHost) Instantiate() error {
	return host.vm.Instantiate()
}

func (host *WasmHost) IsView(function string) bool {
	return (host.funcToIndex[function] & 0x8000) != 0
}

func (host *WasmHost) LoadWasm(wasmData []byte) error {
	return host.vm.LoadWasm(wasmData)
}

func (host *WasmHost) NewInstance() WasmVM {
	return host.vm.NewInstance()
}

func (host *WasmHost) RunFunction(functionName string, args ...interface{}) (err error) {
	return host.vm.RunFunction(functionName, args...)
}

func (host *WasmHost) RunScFunction(functionName string) (err error) {
	index, ok := host.funcToIndex[functionName]
	if !ok {
		return errors.New("unknown SC function name: " + functionName)
	}
	return host.vm.RunScFunction(index)
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
