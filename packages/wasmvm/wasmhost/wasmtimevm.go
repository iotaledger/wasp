// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"errors"

	"github.com/bytecodealliance/wasmtime-go"
)

type WasmTimeVM struct {
	WasmVMBase
	engine       *wasmtime.Engine
	instance     *wasmtime.Instance
	interrupt    *wasmtime.InterruptHandle
	linker       *wasmtime.Linker
	memory       *wasmtime.Memory
	module       *wasmtime.Module
	store        *wasmtime.Store
	remainingGas uint64
}

func NewWasmTimeVM() WasmVM {
	vm := &WasmTimeVM{}
	config := wasmtime.NewConfig()
	config.SetInterruptable(true)
	config.SetConsumeFuel(true)
	vm.engine = wasmtime.NewEngineWithConfig(config)
	return vm
}

// GasBudget sets gas budget. Each time GasBudget is called, we will restart counting the total burned gas
func (vm *WasmTimeVM) GasBudget(budget uint64) error {
	remaining, err := vm.store.ConsumeFuel(0)
	if err != nil {
		return err
	}

	vm.remainingGas = remaining

	if budget > remaining {
		return vm.store.AddFuel(budget - remaining)
	}

	_, err = vm.store.ConsumeFuel(remaining - budget)
	if err != nil {
		return err
	}
	return nil
}

// GasBurned will return the gas has been burned by wasmtime since the last time WasmTimeVM.GasBudget was called
func (vm *WasmTimeVM) GasBurned() uint64 {
	gasBurned, ok := vm.store.FuelConsumed()
	if !ok {
		return 0
	}

	// gasBurned is the gas has been burned since the vm.store instance has been created
	// so we need to substrate the remaining gas before calling vm.GasBudget
	return gasBurned - vm.remainingGas
}

func (vm *WasmTimeVM) Instantiate() (err error) {
	vm.instance, err = vm.linker.Instantiate(vm.store, vm.module)
	if err != nil {
		return err
	}
	memory := vm.instance.GetExport(vm.store, "memory")
	if memory == nil {
		return errors.New("no memory export")
	}
	vm.memory = memory.Memory()
	if vm.memory == nil {
		return errors.New("not a memory type")
	}
	return nil
}

func (vm *WasmTimeVM) Interrupt() {
	vm.interrupt.Interrupt()
}

func (vm *WasmTimeVM) LinkHost(proc *WasmProcessor) (err error) {
	_ = vm.WasmVMBase.LinkHost(proc)

	vm.store = wasmtime.NewStore(vm.engine)
	vm.interrupt, err = vm.store.InterruptHandle()
	if err != nil {
		return err
	}

	vm.linker = wasmtime.NewLinker(vm.engine)

	// new Wasm VM interface
	err = vm.linker.DefineFunc(vm.store, ModuleWasmLib, FuncHostStateGet, vm.HostStateGet)
	if err != nil {
		return err
	}
	err = vm.linker.DefineFunc(vm.store, ModuleWasmLib, FuncHostStateSet, vm.HostStateSet)
	if err != nil {
		return err
	}

	// AssemblyScript Wasm versions uses this one to write panic message to console
	err = vm.linker.DefineFunc(vm.store, ModuleEnv, FuncAbort, vm.HostAbort)
	if err != nil {
		return err
	}

	// TinyGo Wasm versions uses this one to write panic message to console
	err = vm.linker.DefineFunc(vm.store, ModuleWasi1, FuncFdWrite, vm.HostFdWrite)
	if err != nil {
		return err
	}
	return vm.linker.DefineFunc(vm.store, ModuleWasi2, FuncFdWrite, vm.HostFdWrite)
}

func (vm *WasmTimeVM) LoadWasm(wasmData []byte) (err error) {
	vm.module, err = wasmtime.NewModule(vm.engine, wasmData)
	if err != nil {
		return err
	}
	return vm.Instantiate()
}

func (vm *WasmTimeVM) NewInstance() WasmVM {
	return &WasmTimeVM{engine: vm.engine, module: vm.module}
}

func (vm *WasmTimeVM) RunFunction(functionName string, args ...interface{}) error {
	export := vm.instance.GetExport(vm.store, functionName)
	if export == nil {
		return errors.New("unknown export function: '" + functionName + "'")
	}
	return vm.Run(func() (err error) {
		_, err = export.Func().Call(vm.store, args...)
		return err
	})
}

func (vm *WasmTimeVM) RunScFunction(index int32) error {
	export := vm.instance.GetExport(vm.store, "on_call")
	if export == nil {
		return errors.New("unknown export function: 'on_call'")
	}

	return vm.Run(func() (err error) {
		_, err = export.Func().Call(vm.store, index)
		return err
	})
}

func (vm *WasmTimeVM) UnsafeMemory() []byte {
	return vm.memory.UnsafeData(vm.store)
}
