// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

const (
	defaultTimeout      = 5 * time.Second
	FuncAbort           = "abort"
	FuncFdWrite         = "fd_write"
	FuncHostGetBytes    = "hostGetBytes"
	FuncHostGetKeyID    = "hostGetKeyID"
	FuncHostGetObjectID = "hostGetObjectID"
	FuncHostSetBytes    = "hostSetBytes"
	FuncHostStateGet    = "hostStateGet"
	FuncHostStateSet    = "hostStateSet"
	ModuleEnv           = "env"
	ModuleWasi1         = "wasi_unstable"
	ModuleWasi2         = "wasi_snapshot_preview1"
	ModuleWasmLib       = "WasmLib"
)

var (
	// DisableWasmTimeout can be used to disable the annoying timeout during debugging
	DisableWasmTimeout = false

	// HostTracing turns on debug tracing for ScHost calls
	HostTracing = false

	// HostTracingAll turns on *all* debug tracing for ScHost calls
	HostTracingAll = false

	// WasmTimeout set this to non-zero for a one-time override of the defaultTimeout
	WasmTimeout = 0 * time.Second
)

type WasmVM interface {
	Instantiate() error
	Interrupt()
	LinkHost(impl WasmVM, host *WasmHost) error
	LoadWasm(wasmData []byte) error
	NewInstance() WasmVM
	RunFunction(functionName string, args ...interface{}) error
	RunScFunction(index int32) error
	UnsafeMemory() []byte
	VMGetBytes(offset int32, size int32) []byte
	VMGetSize() int32
	VMSetBytes(offset int32, size int32, bytes []byte) int32
}

// TODO implement this in WasmContext?
type SandboxContext interface {
	CallSandbox(funcNr int32, params []byte) []byte
	StateDelete(key []byte)
	StateExists(key []byte) bool
	StateGet(key []byte) []byte
	StateSet(key []byte, value []byte)
}

type WasmVMBase struct {
	impl           WasmVM
	host           *WasmHost
	ctx            SandboxContext
	panicErr       error
	result         []byte
	resultKeyID    int32
	timeoutStarted bool
}

// catchPanicMessage is used in every host function to catch any panic.
// It will save the first panic it encounters in the WasmVMBase so that
// the caller of the Wasm function can retrieve the correct error.
// This is a workaround to WasmTime saving the *last* panic instead of
// the first, thereby reporting the wrong panic error sometimes
func (vm *WasmVMBase) catchPanicMessage() {
	panicMsg := recover()
	if panicMsg == nil {
		return
	}
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

//nolint:unparam
func (vm *WasmVMBase) getKvStore(id int32) *KvStoreHost {
	return vm.host.getKvStore(id)
}

func (vm *WasmVMBase) HostAbort(errMsg, fileName, line, col int32) {
	// crude implementation assumes texts to only use ASCII part of UTF-16

	defer vm.catchPanicMessage()

	// null-terminated UTF-16 error message
	str1 := make([]byte, 0)
	ptr := vm.impl.VMGetBytes(errMsg, 2)
	for i := errMsg; ptr[0] != 0; i += 2 {
		str1 = append(str1, ptr[0])
		ptr = vm.impl.VMGetBytes(i, 2)
	}

	// null-terminated UTF-16 file name
	str2 := make([]byte, 0)
	ptr = vm.impl.VMGetBytes(fileName, 2)
	for i := fileName; ptr[0] != 0; i += 2 {
		str2 = append(str2, ptr[0])
		ptr = vm.impl.VMGetBytes(i, 2)
	}

	panic(fmt.Sprintf("AssemblyScript panic: %s (%s %d:%d)", string(str1), string(str2), line, col))
}

func (vm *WasmVMBase) HostFdWrite(_fd, iovs, _size, written int32) int32 {
	defer vm.catchPanicMessage()

	host := vm.getKvStore(0)
	host.TraceAllf("HostFdWrite(...)")
	// very basic implementation that expects fd to be stdout and iovs to be only one element
	ptr := vm.impl.VMGetBytes(iovs, 8)
	text := int32(binary.LittleEndian.Uint32(ptr[0:4]))
	size := int32(binary.LittleEndian.Uint32(ptr[4:8]))
	msg := vm.impl.VMGetBytes(text, size)
	fmt.Print(string(msg))
	ptr = make([]byte, 4)
	binary.LittleEndian.PutUint32(ptr, uint32(size))
	vm.impl.VMSetBytes(written, size, ptr)
	return size
}

func (vm *WasmVMBase) HostGetBytes(objID, keyID, typeID, stringRef, size int32) int32 {
	defer vm.catchPanicMessage()

	host := vm.getKvStore(0)
	host.TraceAllf("HostGetBytes(o%d,k%d,t%d,r%d,s%d)", objID, keyID, typeID, stringRef, size)

	// only check for existence ?
	if size < 0 {
		if host.Exists(objID, keyID, typeID) {
			return 0
		}
		// missing key is indicated by -1
		return -1
	}

	// actual GetBytes request ?
	if (typeID & OBJTYPE_CALL) == 0 {
		bytes := host.GetBytes(objID, keyID, typeID)
		if bytes == nil {
			return -1
		}
		return vm.impl.VMSetBytes(stringRef, size, bytes)
	}

	// func call request
	switch typeID {
	case OBJTYPE_CALL:
		// func call with params, returns result length
		vm.resultKeyID = keyID
		params := vm.impl.VMGetBytes(stringRef, size)
		vm.result = host.CallFunc(objID, keyID, params)
		return int32(len(vm.result))

	case OBJTYPE_CALL + 1:
		// retrieve previous func call result
		if vm.resultKeyID == keyID {
			result := vm.result
			vm.result = nil
			vm.resultKeyID = 0
			if result == nil {
				return -1
			}
			return vm.impl.VMSetBytes(stringRef, int32(len(result)), result)
		}
	}
	panic("HostGetBytes: Invalid func call state")
}

func (vm *WasmVMBase) HostGetKeyID(keyRef, size int32) int32 {
	defer vm.catchPanicMessage()

	host := vm.getKvStore(0)
	host.TraceAllf("HostGetKeyID(r%d,s%d)", keyRef, size)
	// non-negative size means original key was a string
	if size >= 0 {
		bytes := vm.impl.VMGetBytes(keyRef, size)
		return host.GetKeyIDFromString(string(bytes))
	}

	// negative size means original key was a byte slice
	bytes := vm.impl.VMGetBytes(keyRef, -size-1)
	return host.GetKeyIDFromBytes(bytes)
}

func (vm *WasmVMBase) HostGetObjectID(objID, keyID, typeID int32) int32 {
	defer vm.catchPanicMessage()

	host := vm.getKvStore(0)
	host.TraceAllf("HostGetObjectID(o%d,k%d,t%d)", objID, keyID, typeID)
	return host.GetObjectID(objID, keyID, typeID)
}

func (vm *WasmVMBase) HostSetBytes(objID, keyID, typeID, stringRef, size int32) {
	defer vm.catchPanicMessage()

	host := vm.getKvStore(0)
	host.TraceAllf("HostSetBytes(o%d,k%d,t%d,r%d,s%d)", objID, keyID, typeID, stringRef, size)

	// delete key ?
	if size < 0 {
		host.DelKey(objID, keyID, typeID)
		return
	}

	bytes := vm.impl.VMGetBytes(stringRef, size)
	host.SetBytes(objID, keyID, typeID, bytes)
}

func (vm *WasmVMBase) HostStateGet(keyRef, keyLen, valRef, valLen int32) int32 {
	defer vm.catchPanicMessage()

	host := vm.getKvStore(0)
	host.TraceAllf("HostStateGet(k(%d,%d),v(%d,s%d))", keyRef, keyLen, valRef, valLen)

	// only check for existence ?
	if valLen < 0 {
		key := vm.impl.VMGetBytes(keyRef, keyLen)
		if vm.ctx.StateExists(key) {
			return 0
		}
		// missing key is indicated by -1
		return -1
	}

	//  get value for key request, or get cached result request (keyLen == 0)
	if keyLen >= 0 {
		if keyLen > 0 {
			// retrieve value associated with key
			key := vm.impl.VMGetBytes(keyRef, keyLen)
			vm.result = vm.ctx.StateGet(key)
		}
		if vm.result == nil {
			return -1
		}
		return vm.impl.VMSetBytes(valRef, valLen, vm.result)
	}

	// sandbox func call request, keyLen is func nr
	params := vm.impl.VMGetBytes(valRef, valLen)
	vm.result = vm.ctx.CallSandbox(keyLen, params)
	return int32(len(vm.result))
}

func (vm *WasmVMBase) HostStateSet(keyRef, keyLen, valRef, valLen int32) {
	defer vm.catchPanicMessage()

	host := vm.getKvStore(0)
	host.TraceAllf("HostStateSet(k(%d,%d),v(%d,s%d))", keyRef, keyLen, valRef, valLen)

	key := vm.impl.VMGetBytes(keyRef, keyLen)

	// delete key ?
	if valLen < 0 {
		vm.ctx.StateDelete(key)
		return
	}

	value := vm.impl.VMGetBytes(valRef, valLen)
	vm.ctx.StateSet(key, value)
}

func (vm *WasmVMBase) Instantiate() error {
	return errors.New("cannot be cloned")
}

func (vm *WasmVMBase) LinkHost(impl WasmVM, host *WasmHost) error {
	// trick vm into thinking it doesn't have to start the timeout timer
	// useful when debugging to prevent timing out on breakpoints
	vm.timeoutStarted = DisableWasmTimeout

	vm.impl = impl
	vm.host = host
	host.vm = impl
	return nil
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
			vm.impl.Interrupt()
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
	return err
}

func (vm *WasmVMBase) VMGetBytes(offset, size int32) []byte {
	ptr := vm.impl.UnsafeMemory()
	bytes := make([]byte, size)
	copy(bytes, ptr[offset:offset+size])
	return bytes
}

func (vm *WasmVMBase) VMGetSize() int32 {
	ptr := vm.impl.UnsafeMemory()
	return int32(len(ptr))
}

func (vm *WasmVMBase) VMSetBytes(offset, size int32, bytes []byte) int32 {
	if size != 0 {
		ptr := vm.impl.UnsafeMemory()
		copy(ptr[offset:offset+size], bytes)
	}
	return int32(len(bytes))
}
