// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

//go:build wasmedge
// +build wasmedge

package wasmhost

import (
	"errors"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

type WasmEdgeVM struct {
	WasmVMBase
	edge      *wasmedge.VM
	memory    *wasmedge.Memory
	module    *wasmedge.ImportObject
	store     *wasmedge.Store
	importers []*wasmedge.ImportObject
}

type HostFunction func(params []interface{}) []interface{}

const I32 = wasmedge.ValType_I32

var wasmedgerI32Params = []wasmedge.ValType{I32, I32, I32, I32, I32}

func NewWasmEdgeVM() WasmVM {
	vm := &WasmEdgeVM{}
	wasmedge.SetLogErrorLevel()

	vm.edge = wasmedge.NewVM()

	//config := wasmedge.NewConfig()
	//config.SetInterruptable(true)
	//vm.store = wasmedge.NewStore(wasmedge.NewEngineWithConfig(config))
	//vm.interrupt, _ = vm.store.InterruptHandle()
	return vm
}

func (vm *WasmEdgeVM) Instantiate() error {
	err := vm.edge.Instantiate()
	if err != nil {
		return err
	}
	vm.memory = vm.edge.GetStore().FindMemory("memory")
	if vm.memory == nil {
		return errors.New("no memory export")
	}
	return nil
}

// TODO
func (vm *WasmEdgeVM) Interrupt() {
	panic("implement me")
}

func (vm *WasmEdgeVM) LinkHost(proc *WasmProcessor) error {
	_ = vm.WasmVMBase.LinkHost(proc)

	vm.importModule(ModuleWasmLib)
	vm.importFunc(4, 1, FuncHostStateGet, vm.exportHostStateGet)
	vm.importFunc(4, 0, FuncHostStateSet, vm.exportHostStateSet)
	err := vm.edge.RegisterImport(vm.module)
	if err != nil {
		return err
	}

	// AssemblyScript Wasm versions uses this one to write panic message to console
	vm.importModule(ModuleEnv)
	vm.importFunc(4, 0, FuncAbort, vm.exportAbort)
	err = vm.edge.RegisterImport(vm.module)
	if err != nil {
		return err
	}

	// TinyGo Wasm versions uses these to write panic message to console
	vm.importModule(ModuleWasi1)
	vm.importFunc(4, 1, FuncFdWrite, vm.exportFdWrite)
	err = vm.edge.RegisterImport(vm.module)
	if err != nil {
		return err
	}
	vm.importModule(ModuleWasi2)
	vm.importFunc(4, 1, FuncFdWrite, vm.exportFdWrite)
	return vm.edge.RegisterImport(vm.module)
}

func (vm *WasmEdgeVM) LoadWasm(wasmData []byte) error {
	err := vm.edge.LoadWasmBuffer(wasmData)
	if err != nil {
		return err
	}
	err = vm.edge.Validate()
	if err != nil {
		return err
	}
	return vm.Instantiate()
}

func (vm *WasmEdgeVM) NewInstance() WasmVM {
	return NewWasmEdgeVM()
}

func (vm *WasmEdgeVM) RunFunction(functionName string, args ...interface{}) error {
	return vm.Run(func() (err error) {
		_, err = vm.edge.Execute(functionName, args...)
		return err
	})
}

func (vm *WasmEdgeVM) RunScFunction(index int32) error {
	return vm.Run(func() (err error) {
		_, err = vm.edge.Execute("on_call", index)
		return err
	})
}

func (vm *WasmEdgeVM) UnsafeMemory() []byte {
	panic("wasmedge.UnsafeMemory")
	return nil
}

func (vm *WasmEdgeVM) VMGetBytes(offset int32, size int32) []byte {
	bytes, err := vm.memory.GetData(uint(offset), uint(size))
	if err != nil {
		panic("wasmedge.VMGetBytes: " + err.Error())
	}
	return bytes
}

func (vm *WasmEdgeVM) VMGetSize() int32 {
	return int32(vm.memory.GetPageSize() << 16)
}

func (vm *WasmEdgeVM) VMSetBytes(offset int32, size int32, bytes []byte) int32 {
	if size != 0 {
		err := vm.memory.SetData(bytes, uint(offset), uint(size))
		if err != nil {
			panic("wasmedge.VMSetBytes: " + err.Error())
		}
	}
	return int32(len(bytes))
}

func (vm *WasmEdgeVM) exportAbort(args []interface{}) []interface{} {
	errMsg := args[0].(int32)
	fileName := args[1].(int32)
	line := args[2].(int32)
	col := args[3].(int32)
	vm.HostAbort(errMsg, fileName, line, col)
	return nil
}

func (vm *WasmEdgeVM) exportFdWrite(args []interface{}) []interface{} {
	fd := args[0].(int32)
	iovs := args[1].(int32)
	size := args[2].(int32)
	written := args[3].(int32)
	ret := vm.HostFdWrite(fd, iovs, size, written)
	return []interface{}{ret}
}

func (vm *WasmEdgeVM) exportHostStateGet(args []interface{}) []interface{} {
	keyRef := args[0].(int32)
	keyLen := args[1].(int32)
	valRef := args[2].(int32)
	valLen := args[3].(int32)
	ret := vm.HostStateGet(keyRef, keyLen, valRef, valLen)
	return []interface{}{ret}
}

func (vm *WasmEdgeVM) exportHostStateSet(args []interface{}) []interface{} {
	keyRef := args[0].(int32)
	keyLen := args[1].(int32)
	valRef := args[2].(int32)
	valLen := args[3].(int32)
	vm.HostStateSet(keyRef, keyLen, valRef, valLen)
	return nil
}

func (vm *WasmEdgeVM) importFunc(nrParams int, nrResults int, funcName string, function HostFunction) {
	wrapper := func(_data interface{}, _mem *wasmedge.Memory, params []interface{}) ([]interface{}, wasmedge.Result) {
		return function(params), wasmedge.Result_Success
	}
	funcType := wasmedge.NewFunctionType(wasmedgerI32Params[:nrParams], wasmedgerI32Params[:nrResults])
	funcWrapper := wasmedge.NewFunction(funcType, wrapper, nil, 0)
	vm.module.AddFunction(funcName, funcWrapper)
}

func (vm *WasmEdgeVM) importModule(name string) {
	vm.module = wasmedge.NewImportObject(name)
	vm.importers = append(vm.importers, vm.module)
}
