// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"errors"
	"strings"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
)

type WasmGoVM struct {
	WasmVMBase
	scName string
	onLoad func()
}

var _ WasmVM = &WasmGoVM{}

func NewWasmGoVM(scName string, onLoad func()) *WasmGoVM {
	return &WasmGoVM{scName: scName, onLoad: onLoad}
}

func Connect(h wasmlib.ScHost) wasmlib.ScHost {
	return wasmlib.ConnectHost(h)
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
	vm.onLoad()
	return nil
}

func (vm *WasmGoVM) RunFunction(functionName string, args ...interface{}) error {
	// already ran on_load in LoadWasm, other functions are not supported
	if functionName != "on_load" {
		return errors.New("WasmGoVM: cannot run function: " + functionName)
	}
	return nil
}

func (vm *WasmGoVM) RunScFunction(index int32) error {
	return vm.Run(func() error {
		wasmlib.OnCall(index)
		return nil
	})
}

func (vm *WasmGoVM) UnsafeMemory() []byte {
	// no need to communicate through Wasm mem pool
	return nil
}
