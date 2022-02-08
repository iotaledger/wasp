// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"errors"
	"strings"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
)

// provide implementation for Wasm-only function
func Connect(h wasmlib.ScHost) wasmlib.ScHost {
	return wasmlib.ConnectHost(h)
}

type ScOnloadFunc func(index int32)

type WasmGoVM struct {
	WasmVMBase
	scName string
	onLoad ScOnloadFunc
}

func NewWasmGoVM(scName string, onLoad ScOnloadFunc) WasmVM {
	return &WasmGoVM{scName: scName, onLoad: onLoad}
}

func (vm *WasmGoVM) NewInstance() WasmVM {
	return nil
}

func (vm *WasmGoVM) Instantiate() error {
	return nil
}

func (vm *WasmGoVM) Interrupt() {
	// disabled for now
	// panic("implement me")
}

func (vm *WasmGoVM) LoadWasm(wasmData []byte) error {
	scName := string(wasmData)
	if !strings.HasPrefix(scName, "go:") {
		return errors.New("WasmGoVM: not a Go contract: " + scName)
	}
	if scName[3:] != vm.scName {
		return errors.New("WasmGoVM: unknown contract: " + scName)
	}
	return nil
}

func (vm *WasmGoVM) RunFunction(functionName string, args ...interface{}) error {
	if functionName == "on_load" {
		// note: on_load is funneled through onload()
		vm.onLoad(-1)
		return nil
	}
	return errors.New("WasmGoVM: cannot run function: " + functionName)
}

func (vm *WasmGoVM) RunScFunction(index int32) (err error) {
	return vm.Run(func() error {
		// note: on_call is funneled through onload()
		vm.onLoad(index)
		return nil
	})
}

func (vm *WasmGoVM) UnsafeMemory() []byte {
	// no need to communicate through Wasm mem pool
	return nil
}
