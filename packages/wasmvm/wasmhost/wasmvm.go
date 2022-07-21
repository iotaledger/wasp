// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
)

const (
	defaultTimeout   = 5 * time.Second
	FuncAbort        = "abort"
	FuncFdWrite      = "fd_write"
	FuncHostStateGet = "hostStateGet"
	FuncHostStateSet = "hostStateSet"
	ModuleEnv        = "env"
	ModuleWasi1      = "wasi_unstable"
	ModuleWasi2      = "wasi_snapshot_preview1"
	ModuleWasmLib    = "WasmLib"
)

var (
	// DisableWasmTimeout can be used to disable the annoying timeout during debugging
	DisableWasmTimeout = true

	// HostTracing turns on debug tracing for ScHost calls
	HostTracing = false

	// WasmTimeout set this to non-zero for a one-time override of the defaultTimeout
	WasmTimeout = 0 * time.Second
)

type WasmVM interface {
	GasBudget(budget uint64)
	GasBurned() uint64
	Interrupt()
	LinkHost() error
	LoadWasm(wasmData []byte) error
	NewInstance(wc *WasmContext) WasmVM
	RunFunction(functionName string, args ...interface{}) error
	RunScFunction(index int32) error
	UnsafeMemory() []byte
	VMGetBytes(offset int32, size int32) []byte
	VMGetSize() int32
	VMSetBytes(offset int32, size int32, bytes []byte) int32
}

type WasmVMBase struct {
	cachedResult   []byte
	panicErr       error
	timeoutStarted bool
	wc             *WasmContext
}

func (vm *WasmVMBase) GasBudget(budget uint64) {
	// ignore gas budget
}

func (vm *WasmVMBase) GasBurned() uint64 {
	// burn nothing
	return 0
}

func (vm *WasmVMBase) getContext() *WasmContext {
	return vm.wc.proc.GetCurrentContext()
}

func (vm *WasmVMBase) HostAbort(errMsg, fileName, line, col int32) {
	wc := vm.getContext()
	vm.reportGasBurned(wc)
	defer vm.wrapUp(wc)

	// crude implementation assumes texts to only use ASCII part of UTF-16
	impl := wc.vm

	// null-terminated UTF-16 error message
	str1 := make([]byte, 0)
	ptr := impl.VMGetBytes(errMsg, 2)
	for i := errMsg; ptr[0] != 0; i += 2 {
		str1 = append(str1, ptr[0])
		ptr = impl.VMGetBytes(i, 2)
	}

	// null-terminated UTF-16 file name
	str2 := make([]byte, 0)
	ptr = impl.VMGetBytes(fileName, 2)
	for i := fileName; ptr[0] != 0; i += 2 {
		str2 = append(str2, ptr[0])
		ptr = impl.VMGetBytes(i, 2)
	}

	panic(fmt.Sprintf("AssemblyScript panic: %s (%s %d:%d)", string(str1), string(str2), line, col))
}

func (vm *WasmVMBase) HostFdWrite(_fd, iovs, _size, written int32) int32 {
	wc := vm.getContext()
	vm.reportGasBurned(wc)
	defer vm.wrapUp(wc)

	wc.log().Debugf("HostFdWrite(...)")
	impl := wc.vm

	// very basic implementation that expects fd to be stdout and iovs to be only one element
	ptr := impl.VMGetBytes(iovs, 8)
	text := int32(binary.LittleEndian.Uint32(ptr[0:4]))
	size := int32(binary.LittleEndian.Uint32(ptr[4:8]))
	// msg := vm.impl.VMGetBytes(text, size)
	// fmt.Print(string(msg))
	ptr = make([]byte, 4)
	binary.LittleEndian.PutUint32(ptr, uint32(size))
	impl.VMSetBytes(written, size, ptr)

	// strip off "panic: " prefix and call sandbox panic function
	vm.HostStateGet(0, wasmlib.FnPanic, text+7, size)
	return size
}

func (vm *WasmVMBase) HostStateGet(keyRef, keyLen, valRef, valLen int32) int32 {
	wc := vm.getContext()
	vm.reportGasBurned(wc)
	defer vm.wrapUp(wc)

	impl := wc.vm

	// only check for existence ?
	if valLen < 0 {
		key := impl.VMGetBytes(keyRef, keyLen)
		if wc.StateExists(key) {
			return 0
		}
		// missing key is indicated by -1
		return -1
	}

	//  get value for key request, or get cached result request (keyLen == 0)
	if keyLen >= 0 {
		if keyLen > 0 {
			// retrieve value associated with key
			key := impl.VMGetBytes(keyRef, keyLen)
			vm.cachedResult = wc.StateGet(key)
		}
		if vm.cachedResult == nil {
			return -1
		}
		return impl.VMSetBytes(valRef, valLen, vm.cachedResult)
	}

	// sandbox func call request, keyLen is func nr
	params := impl.VMGetBytes(valRef, valLen)
	vm.cachedResult = wc.Sandbox(keyLen, params)
	return int32(len(vm.cachedResult))
}

func (vm *WasmVMBase) HostStateSet(keyRef, keyLen, valRef, valLen int32) {
	wc := vm.getContext()
	vm.reportGasBurned(wc)
	defer vm.wrapUp(wc)

	impl := wc.vm

	// export name?
	if keyRef == 0 {
		name := string(impl.VMGetBytes(valRef, valLen))
		if keyLen < 0 {
			// ExportWasmTag, log the wasm tag name
			wc.log().Infof(name)

			// // potentially adjust gas fudge factor multiplier here
			// if strings.Contains(name, "WASM::TYPESCRIPT") {
			// 	wc.proc.gasFactorX = 1
			// }
			return
		}
		wc.ExportName(keyLen, name)
		return
	}

	key := impl.VMGetBytes(keyRef, keyLen)

	// delete key ?
	if valLen < 0 {
		wc.StateDelete(key)
		return
	}

	// set key
	value := impl.VMGetBytes(valRef, valLen)
	wc.StateSet(key, value)
}

func (vm *WasmVMBase) Instantiate(proc *WasmProcessor) error {
	return errors.New("cannot be cloned")
}

func (vm *WasmVMBase) LinkHost() error {
	// trick vm into thinking it doesn't have to start the timeout timer
	// useful when debugging to prevent timing out on breakpoints
	return nil
}

// reportGasBurned updates the sandbox gas budget with the amount burned by the VM
func (vm *WasmVMBase) reportGasBurned(wc *WasmContext) {
	if !wc.gasDisabled {
		wc.GasBurned(wc.vm.GasBurned() / wc.proc.gasFactor())
	}
}

func (vm *WasmVMBase) Run(runner func() error) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		// could be the wrong panic message due to a WasmTime bug, so we always
		// rethrow our intercepted first panic instead of WasmTime's last panic
		if vm.panicErr != nil {
			panic(vm.panicErr)
		}
		panic(r)
	}()

	if vm.timeoutStarted {
		// no need to wrap nested calls in timeout code
		err = runner()
		if vm.panicErr != nil {
			err = vm.panicErr
			vm.panicErr = nil
		}
		if err != nil && strings.Contains(err.Error(), "all fuel consumed") {
			err = errors.New("gas budget exceeded in Wasm VM")
		}
		return err
	}

	timeout := defaultTimeout
	if WasmTimeout != 0 {
		timeout = WasmTimeout
		WasmTimeout = 0
	}

	done := make(chan bool, 2)

	// start timeout handler
	go func() {
		select {
		case <-done: // runner was done before timeout
		case <-time.After(timeout):
			// timeout: interrupt Wasm
			vm.wc.vm.Interrupt()
			// wait for runner to finish
			<-done
		}
	}()

	vm.timeoutStarted = true
	err = runner()
	done <- true
	vm.timeoutStarted = false
	if vm.panicErr != nil {
		err = vm.panicErr
		vm.panicErr = nil
	}
	if err != nil && strings.Contains(err.Error(), "all fuel consumed") {
		err = errors.New("gas budget exceeded in Wasm VM")
	}
	return err
}

func (vm *WasmVMBase) VMGetBytes(offset, size int32) []byte {
	ptr := vm.wc.vm.UnsafeMemory()
	bytes := make([]byte, size)
	copy(bytes, ptr[offset:offset+size])
	return bytes
}

func (vm *WasmVMBase) VMGetSize() int32 {
	ptr := vm.wc.vm.UnsafeMemory()
	return int32(len(ptr))
}

func (vm *WasmVMBase) VMSetBytes(offset, size int32, bytes []byte) int32 {
	if size != 0 {
		ptr := vm.wc.vm.UnsafeMemory()
		copy(ptr[offset:offset+size], bytes)
	}
	return int32(len(bytes))
}

// wrapUp is used in every host function to catch any panic.
// It will save the first panic it encounters in the WasmVMBase so that
// the caller of the Wasm function can retrieve the correct error.
// This is a workaround to WasmTime saving the *last* panic instead of
// the first, thereby reporting the wrong panic error sometimes
// wrapUp will also update the Wasm code that initiated the call with
// the remaining gas budget (the Wasp node may have burned some)
func (vm *WasmVMBase) wrapUp(wc *WasmContext) {
	panicMsg := recover()
	if panicMsg == nil {
		if !wc.gasDisabled {
			// update VM gas budget to reflect what sandbox burned
			wc.vm.GasBudget(wc.GasBudget() * wc.proc.gasFactor())
		}
		return
	}

	// panic means no need to update gas budget
	if vm.panicErr == nil {
		switch msg := panicMsg.(type) {
		case error:
			vm.panicErr = msg
		default:
			vm.panicErr = fmt.Errorf("%v", msg)
		}
	}

	// rethrow and let nature run its course...
	panic(panicMsg)
}
