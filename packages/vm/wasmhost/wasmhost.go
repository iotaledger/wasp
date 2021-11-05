// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"errors"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
)

type WasmStore interface {
	GetKvStore(id int32) *KvStoreHost
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

func (host *WasmHost) getKvStore(id int32) *KvStoreHost {
	return host.store.GetKvStore(id)
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

func (host *WasmHost) IsView(function string) bool {
	return (host.funcToIndex[function] & 0x8000) != 0
}

func (host *WasmHost) LoadWasm(wasmData []byte) error {
	err := host.vm.LoadWasm(wasmData)
	if err != nil {
		return err
	}
	err = host.RunFunction("on_load")
	if err != nil {
		return err
	}
	host.vm.SaveMemory()
	return nil
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
	if index < 0 {
		// double check that predefined keys are in sync
		if index == KeyZzzzzzz {
			return
		}
		panic("SetExport: predefined key value mismatch")
	}

	funcIndex, ok := host.funcToIndex[functionName]
	if ok {
		if funcIndex != index {
			panic("SetExport: duplicate function name")
		}
		return
	}

	hn := iscp.Hn(functionName)
	hashedName := uint32(hn)
	_, ok = host.codeToFunc[hashedName]
	if ok {
		panic("SetExport: duplicate hashed name")
	}
	host.codeToFunc[hashedName] = functionName
	host.funcToCode[functionName] = hashedName
	host.funcToIndex[functionName] = index
}
