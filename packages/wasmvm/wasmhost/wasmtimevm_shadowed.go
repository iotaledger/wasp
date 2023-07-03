// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
//go:build no_wasmhost

package wasmhost

type WasmTimeVM struct {
	WasmVMBase
}

func NewWasmTimeVM() WasmVM {
	return nil
}

// GasBudget sets the gas budget for the VM.
func (vm *WasmTimeVM) GasBudget(budget uint64) {

}

// GasBurned will return the gas burned since the last time GasBudget() was called
func (vm *WasmTimeVM) GasBurned() uint64 {
	return 0
}

func (vm *WasmTimeVM) Interrupt() {

}

func (vm *WasmTimeVM) LinkHost() (err error) {
	return nil
}

func (vm *WasmTimeVM) LoadWasm(wasmData []byte) (err error) {
	return nil
}

func (vm *WasmTimeVM) NewInstance(wc *WasmContext) WasmVM {
	return nil
}

func (vm *WasmTimeVM) newInstance() (err error) {

	return nil
}

func (vm *WasmTimeVM) RunFunction(functionName string, args ...interface{}) error {
	return nil
}

func (vm *WasmTimeVM) RunScFunction(index int32) error {
	return nil
}

func (vm *WasmTimeVM) UnsafeMemory() []byte {
	return nil
}
