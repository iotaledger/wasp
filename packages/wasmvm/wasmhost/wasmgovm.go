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

type WasmGoVM struct {
	WasmVMBase
	scName string
	onLoad func()
}

func NewWasmGoVM(scName string, onLoad func()) WasmVM {
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
		vm.onLoad()
		return nil
	}
	return errors.New("WasmGoVM: cannot run function: " + functionName)
}

func (vm *WasmGoVM) RunScFunction(index int32) (err error) {
	//defer func() {
	//	r := recover()
	//	if r == nil {
	//		return
	//	}
	//	switch errType := r.(type) {
	//	case error:
	//		err = errType
	//	case string:
	//		err = errors.New(errType)
	//	default:
	//		err = xerrors.Errorf("RunScFunction: %v", errType)
	//	}
	//}()
	return vm.Run(func() error {
		wasmlib.OnCall(index)
		return nil
	})
}

func (vm *WasmGoVM) UnsafeMemory() []byte {
	// no need to communicate through Wasm mem pool
	return nil
}
