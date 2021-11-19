// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

const defaultTimeout = 5 * time.Second

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
	PoolSize() int
	RunFunction(functionName string, args ...interface{}) error
	RunScFunction(index int32) error
	SaveMemory()
	UnsafeMemory() []byte
	VMGetBytes(offset int32, size int32) []byte
	VMGetSize() int32
	VMSetBytes(offset int32, size int32, bytes []byte) int32
}

type WasmVMBase struct {
	impl           WasmVM
	host           *WasmHost
	memoryCopy     []byte
	memoryDirty    bool
	memoryNonZero  int32
	result         []byte
	resultKeyID    int32
	timeoutStarted bool
}

func (vm *WasmVMBase) EnvAbort(errMsg, fileName, line, col int32) {
	// crude implementation assumes texts to only use ASCII part of UTF-16

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

//nolint:unparam
func (vm *WasmVMBase) getKvStore(id int32) *KvStoreHost {
	return vm.host.getKvStore(id)
}

func (vm *WasmVMBase) HostFdWrite(_fd, iovs, _size, written int32) int32 {
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
	host := vm.getKvStore(0)
	host.TraceAllf("HostGetObjectID(o%d,k%d,t%d)", objID, keyID, typeID)
	return host.GetObjectID(objID, keyID, typeID)
}

func (vm *WasmVMBase) HostSetBytes(objID, keyID, typeID, stringRef, size int32) {
	host := vm.getKvStore(0)
	host.TraceAllf("HostSetBytes(o%d,k%d,t%d,r%d,s%d)", objID, keyID, typeID, stringRef, size)
	bytes := vm.impl.VMGetBytes(stringRef, size)
	host.SetBytes(objID, keyID, typeID, bytes)
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

func (vm *WasmVMBase) PoolSize() int {
	return 0
}

func (vm *WasmVMBase) PreCall() []byte {
	if vm.PoolSize() != 0 {
		return nil
	}

	bytes := vm.impl.VMGetSize()
	frame := vm.impl.VMGetBytes(0, bytes)
	if vm.memoryDirty {
		// clear memory and restore initialized data range
		vm.impl.VMSetBytes(0, bytes, make([]byte, bytes))
		size := int32(len(vm.memoryCopy))
		vm.impl.VMSetBytes(vm.memoryNonZero, size, vm.memoryCopy)
	}
	vm.memoryDirty = true
	return frame
}

func (vm *WasmVMBase) PostCall(frame []byte) {
	if vm.PoolSize() != 0 {
		return
	}

	vm.impl.VMSetBytes(0, int32(len(frame)), frame)
}

func (vm *WasmVMBase) Run(runner func() error) (err error) {
	if vm.timeoutStarted {
		// no need to wrap nested calls in timeout code
		return runner()
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
	return err
}

func (vm *WasmVMBase) SaveMemory() {
	if vm.PoolSize() != 0 {
		return
	}

	// find initialized data range in memory
	bytes := vm.impl.VMGetSize()
	ptr := vm.impl.VMGetBytes(0, bytes)
	if ptr == nil {
		// this vm implementation does not communicate via mem pool
		return
	}
	firstNonZero := -1
	lastNonZero := 0
	for i, b := range ptr {
		if b != 0 {
			if firstNonZero < 0 {
				firstNonZero = i
			}
			lastNonZero = i
		}
	}

	// save copy of initialized data range
	vm.memoryNonZero = bytes
	if firstNonZero >= 0 {
		vm.memoryNonZero = int32(firstNonZero)
		size := lastNonZero + 1 - firstNonZero
		vm.memoryCopy = make([]byte, size)
		copy(vm.memoryCopy, ptr[vm.memoryNonZero:])
	}
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
