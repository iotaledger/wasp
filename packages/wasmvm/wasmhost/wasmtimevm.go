// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"errors"

	"github.com/bytecodealliance/wasmtime-go"
)

type WasmTimeVM struct {
	WasmVMBase
	engine     *wasmtime.Engine
	instance   *wasmtime.Instance
	linker     *wasmtime.Linker
	memory     *wasmtime.Memory
	module     *wasmtime.Module
	store      *wasmtime.Store
	lastBudget uint64
	instances  uint32
}

func NewWasmTimeVM() WasmVM {
	config := wasmtime.NewConfig()
	// config.SetInterruptable(true)
	config.SetConsumeFuel(true)
	vm := &WasmTimeVM{engine: wasmtime.NewEngineWithConfig(config)}
	vm.timeoutStarted = true // DisableWasmTimeout
	return vm
}

// GasBudget sets the gas budget for the VM.
func (vm *WasmTimeVM) GasBudget(budget uint64) {
	// save budget so we can later determine how much the VM burned
	vm.lastBudget = budget

	// new budget for VM, top up to desired budget
	err := vm.store.AddFuel(budget)
	if err != nil {
		panic("GasBudget.set: " + err.Error())
	}

	// consume 0 fuel to determine remaining budget
	remainingBudget, err := vm.store.ConsumeFuel(0)
	if err != nil {
		panic("GasBudget.determine: " + err.Error())
	}

	if remainingBudget > budget {
		// burn excess budget
		_, err = vm.store.ConsumeFuel(remainingBudget - budget)
		if err != nil {
			panic("GasBudget.burn: " + err.Error())
		}
	}
}

// GasBurned will return the gas burned since the last time GasBudget() was called
func (vm *WasmTimeVM) GasBurned() uint64 {
	// consume 0 fuel to determine remaining budget
	remainingBudget, err := vm.store.ConsumeFuel(0)
	if err != nil {
		vm.wc.proc.log.Infof("GasBurned.determine: " + err.Error())
	}

	burned := vm.lastBudget - remainingBudget
	return burned
}

func (vm *WasmTimeVM) Interrupt() {
	// interrupt, err := vm.store.InterruptHandle()
	// if err != nil {
	// 	panic(err)
	// }
	// interrupt.Interrupt()
}

func (vm *WasmTimeVM) LinkHost() (err error) {
	vm.store = wasmtime.NewStore(vm.engine)
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
	return err
}

func (vm *WasmTimeVM) NewInstance(wc *WasmContext) WasmVM {
	if vm.wc == nil {
		vm.wc = wc
	}

	// WasmTime stores instances in a store, but provides no way to release an
	// obsolete instance. They keep on being retained by the store after usage,
	// until at 10,000 instances we get an error 'max instances exceeded', i.e.
	// there is a memory leak here we need to work around.
	//
	// To combat this leak we keep track of the number of instances and when we
	// hit a magic number (256 in this case) we tell our main WasmVM to do a hard
	// relink of the Wasm module. Since this means we get a new linker with a new
	// store for the loaded Wasm module, the old store is now being abandoned.
	// Instances keep track of the store they are in through their WasmContext.
	// That means that store will stay alive as long as there still are instances
	// using it (usually because of nested calls). Once all WasmContexts that
	// reference the old store have been released, there is nothing referencing
	// the old store any more, and it will be cleaned up by WasmTime-Go.
	vm.instances++
	if (vm.instances & 0xff) == 0 {
		err := vm.LinkHost()
		if err != nil {
			panic(err)
		}
	}

	vmInstance := &WasmTimeVM{
		engine: vm.engine,
		module: vm.module,
		linker: vm.linker,
		store:  vm.store,
	}
	vmInstance.wc = wc
	vmInstance.timeoutStarted = true // DisableWasmTimeout
	err := vmInstance.newInstance()
	if err != nil {
		panic("cannot instantiate: " + err.Error())
	}
	return vmInstance
}

func (vm *WasmTimeVM) newInstance() (err error) {
	vm.GasBudget(1_000_000)
	vm.wc.GasDisable(true)
	vm.instance, err = vm.linker.Instantiate(vm.store, vm.module)
	vm.wc.GasDisable(false)
	burned := vm.GasBurned()
	_ = burned
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
		vm.store.GC()
		return err
	})
}

func (vm *WasmTimeVM) UnsafeMemory() []byte {
	return vm.memory.UnsafeData(vm.store)
}
