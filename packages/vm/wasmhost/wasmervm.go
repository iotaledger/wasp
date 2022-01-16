// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// +build wasmer

package wasmhost

import (
	"github.com/wasmerio/wasmer-go/wasmer"
)

type WasmerVM struct {
	WasmVMBase
	instance *wasmer.Instance
	linker   *wasmer.ImportObject
	memory   *wasmer.Memory
	module   *wasmer.Module
	store    *wasmer.Store
}

var i32 = []wasmer.ValueKind{wasmer.I32, wasmer.I32, wasmer.I32, wasmer.I32, wasmer.I32}

func NewWasmerVM() WasmVM {
	vm := &WasmerVM{}
	vm.store = wasmer.NewStore(wasmer.NewEngine())
	return vm
}

func (vm *WasmerVM) NewInstance() WasmVM {
	return &WasmerVM{store: vm.store}
}

//TODO
func (vm *WasmerVM) Interrupt() {
	panic("implement me")
}

func (vm *WasmerVM) LinkHost(impl WasmVM, host *WasmHost) error {
	vm.linker = wasmer.NewImportObject()
	_ = vm.WasmVMBase.LinkHost(impl, host)

	funcs := map[string]wasmer.IntoExtern{
		FuncHostGetBytes:    vm.importFunc(5, 1, vm.exportHostGetBytes),
		FuncHostGetKeyID:    vm.importFunc(2, 1, vm.exportHostGetKeyID),
		FuncHostGetObjectID: vm.importFunc(3, 1, vm.exportHostGetObjectID),
		FuncHostSetBytes:    vm.importFunc(5, 0, vm.exportHostSetBytes),

		// new Wasm VM interface
		FuncHostStateGet: vm.importFunc(4, 1, vm.exportHostStateGet),
		FuncHostStateSet: vm.importFunc(4, 0, vm.exportHostStateSet),
	}
	vm.linker.Register(ModuleWasmLib, funcs)

	funcs = map[string]wasmer.IntoExtern{
		FuncAbort: vm.importFunc(4, 0, vm.exportAbort),
	}
	vm.linker.Register(ModuleEnv, funcs)

	// TinyGo Wasm implementation uses this one to write panic message to console
	funcs = map[string]wasmer.IntoExtern{
		FuncFdWrite: vm.importFunc(4, 1, vm.exportFdWrite),
	}
	vm.linker.Register(ModuleWasi1, funcs)
	vm.linker.Register(ModuleWasi2, funcs)
	return nil
}

func (vm *WasmerVM) importFunc(nrParams, nrResults int, function func([]wasmer.Value) ([]wasmer.Value, error)) *wasmer.Extern {
	params := wasmer.NewValueTypes(i32[:nrParams]...)
	results := wasmer.NewValueTypes(i32[:nrResults]...)
	funcType := wasmer.NewFunctionType(params, results)
	return wasmer.NewFunction(vm.store, funcType, function).IntoExtern()
}

func (vm *WasmerVM) LoadWasm(wasmData []byte) error {
	var err error
	vm.module, err = wasmer.NewModule(vm.store, wasmData)
	if err != nil {
		return err
	}
	vm.instance, err = wasmer.NewInstance(vm.module, vm.linker)
	if err != nil {
		return err
	}
	vm.memory, err = vm.instance.Exports.GetMemory("memory")
	return err
}

func (vm *WasmerVM) RunFunction(functionName string, args ...interface{}) error {
	export, err := vm.instance.Exports.GetFunction(functionName)
	if err != nil {
		return err
	}
	return vm.Run(func() error {
		_, err = export(args...)
		return err
	})
}

func (vm *WasmerVM) RunScFunction(index int32) error {
	export, err := vm.instance.Exports.GetFunction("on_call")
	if err != nil {
		return err
	}
	frame := vm.PreCall()
	err = vm.Run(func() error {
		_, err = export(index)
		return err
	})
	vm.PostCall(frame)
	return err
}

func (vm *WasmerVM) UnsafeMemory() []byte {
	return vm.memory.Data()
}

func (vm *WasmerVM) exportAbort(args []wasmer.Value) ([]wasmer.Value, error) {
	errMsg := args[0].I32()
	fileName := args[1].I32()
	line := args[2].I32()
	col := args[3].I32()
	vm.EnvAbort(errMsg, fileName, line, col)
	return nil, nil
}

func (vm *WasmerVM) exportFdWrite(args []wasmer.Value) ([]wasmer.Value, error) {
	fd := args[0].I32()
	iovs := args[1].I32()
	size := args[2].I32()
	written := args[3].I32()
	ret := vm.HostFdWrite(fd, iovs, size, written)
	return []wasmer.Value{wasmer.NewI32(ret)}, nil
}

func (vm *WasmerVM) exportHostGetBytes(args []wasmer.Value) ([]wasmer.Value, error) {
	objID := args[0].I32()
	keyID := args[1].I32()
	typeID := args[2].I32()
	stringRef := args[3].I32()
	size := args[4].I32()
	ret := vm.HostGetBytes(objID, keyID, typeID, stringRef, size)
	return []wasmer.Value{wasmer.NewI32(ret)}, nil
}

func (vm *WasmerVM) exportHostGetKeyID(args []wasmer.Value) ([]wasmer.Value, error) {
	keyRef := args[0].I32()
	size := args[1].I32()
	ret := vm.HostGetKeyID(keyRef, size)
	return []wasmer.Value{wasmer.NewI32(ret)}, nil
}

func (vm *WasmerVM) exportHostGetObjectID(args []wasmer.Value) ([]wasmer.Value, error) {
	objID := args[0].I32()
	keyID := args[1].I32()
	typeID := args[2].I32()
	ret := vm.HostGetObjectID(objID, keyID, typeID)
	return []wasmer.Value{wasmer.NewI32(ret)}, nil
}

func (vm *WasmerVM) exportHostSetBytes(args []wasmer.Value) ([]wasmer.Value, error) {
	objID := args[0].I32()
	keyID := args[1].I32()
	typeID := args[2].I32()
	stringRef := args[3].I32()
	size := args[4].I32()
	vm.HostSetBytes(objID, keyID, typeID, stringRef, size)
	return nil, nil
}

func (vm *WasmerVM) exportHostStateGet(args []wasmer.Value) ([]wasmer.Value, error) {
	keyRef := args[0].I32()
	keyLen := args[1].I32()
	valRef := args[2].I32()
	valLen := args[3].I32()
	ret := vm.HostStateGet(keyRef, keyLen, valRef, valLen)
	return []wasmer.Value{wasmer.NewI32(ret)}, nil
}

func (vm *WasmerVM) exportHostStateSet(args []wasmer.Value) ([]wasmer.Value, error) {
	keyRef := args[0].I32()
	keyLen := args[1].I32()
	valRef := args[2].I32()
	valLen := args[3].I32()
	vm.HostStateSet(keyRef, keyLen, valRef, valLen)
	return nil, nil
}
