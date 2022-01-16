// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

//go:build wasm
// +build wasm

package wasmvmhost

import "github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"

//go:wasm-module WasmLib
//export hostGetBytes
func hostGetBytes(objID, keyID, typeID int32, value *byte, size int32) int32

//go:wasm-module WasmLib
//export hostGetKeyID
func hostGetKeyID(key *byte, size int32) int32

//go:wasm-module WasmLib
//export hostGetObjectID
func hostGetObjectID(objID, keyID, typeID int32) int32

//go:wasm-module WasmLib
//export hostSetBytes
func hostSetBytes(objID, keyID, typeID int32, value *byte, size int32)

// new Wasm VM

//go:wasm-module WasmLib
//export hostStateGet
func hostStateGet(keyRef *byte, keyLen int32, valRef *byte, valLen int32) int32

//go:wasm-module WasmLib
//export hostStateSet
func hostStateSet(keyRef *byte, keyLen int32, valRef *byte, valLen int32)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type WasmVMHost struct {
	funcs []wasmlib.ScFuncContextFunction
	views []wasmlib.ScViewContextFunction
}

// implements wasmlib.ScHost interface
var _ wasmlib.ScHost = &WasmVMHost{}

func (w *WasmVMHost) AddFunc(f wasmlib.ScFuncContextFunction) []wasmlib.ScFuncContextFunction {
	if f != nil {
		w.funcs = append(w.funcs, f)
	}
	return w.funcs
}

func (w *WasmVMHost) AddView(v wasmlib.ScViewContextFunction) []wasmlib.ScViewContextFunction {
	if v != nil {
		w.views = append(w.views, v)
	}
	return w.views
}

func (w *WasmVMHost) ConnectWasmHost() {
	wasmlib.ConnectHost(w)
}

func (w *WasmVMHost) CallFunc(objID, keyID int32, params []byte) []byte {
	args := (*byte)(nil)
	size := int32(len(params))
	if size != 0 {
		args = &params[0]
	}

	// pass params and query expected length of result
	size = hostGetBytes(objID, keyID, wasmlib.TYPE_CALL, args, size)

	// -1 means non-existent, so return default value for type
	if size <= 0 {
		return []byte(nil)
	}

	// allocate a sufficient length byte array in Wasm memory
	// and let the host copy the actual result into this Wasm byte array
	result := make([]byte, size)
	hostGetBytes(objID, keyID, wasmlib.TYPE_CALL+1, &result[0], size)
	return result
}

func (w *WasmVMHost) DelKey(objID, keyID, typeID int32) {
	// size -1 means delete
	// this removes the need for a separate hostDelete function
	hostSetBytes(objID, keyID, typeID, nil, -1)
}

func (w *WasmVMHost) Exists(objID, keyID, typeID int32) bool {
	// size -1 means only test for existence
	// returned size -1 indicates keyID not found (or error)
	// this removes the need for a separate hostExists function
	return hostGetBytes(objID, keyID, typeID, nil, -1) >= 0
}

func (w *WasmVMHost) GetBytes(objID, keyID, typeID int32) []byte {
	size := int32(wasmlib.TypeSizes[typeID])
	if size == 0 {
		// variable-sized type, first query expected length of bytes array
		// (pass zero-length buffer)
		size = hostGetBytes(objID, keyID, typeID, nil, 0)

		// -1 means non-existent, so return default value for type
		if size <= 0 {
			return []byte(nil)
		}
	}

	// allocate a sufficient length byte array in Wasm memory
	// and let the host copy the actual data bytes into this Wasm byte array
	result := make([]byte, size)
	hostGetBytes(objID, keyID, typeID, &result[0], size)
	return result
}

func (w *WasmVMHost) GetKeyIDFromBytes(bytes []byte) int32 {
	size := int32(len(bytes))
	// &bytes[0] will panic on zero length slice, so use nil instead
	// negative size indicates this was from bytes
	if size == 0 {
		return hostGetKeyID(nil, -1)
	}
	return hostGetKeyID(&bytes[0], -size-1)
}

func (w *WasmVMHost) GetKeyIDFromString(key string) int32 {
	bytes := []byte(key)
	size := int32(len(bytes))
	// &bytes[0] will panic on zero length slice, so use nil instead
	// non-negative size indicates this was from string
	if size == 0 {
		return hostGetKeyID(nil, 0)
	}
	return hostGetKeyID(&bytes[0], size)
}

func (w *WasmVMHost) GetObjectID(objID, keyID, typeID int32) int32 {
	return hostGetObjectID(objID, keyID, typeID)
}

func (w *WasmVMHost) SetBytes(objID, keyID, typeID int32, value []byte) {
	// &bytes[0] will panic on zero length slice, so use nil instead
	size := int32(len(value))
	if size == 0 {
		hostSetBytes(objID, keyID, typeID, nil, size)
		return
	}
	hostSetBytes(objID, keyID, typeID, &value[0], size)
}

func (w *WasmVMHost) Sandbox(funcNr int32, params []byte) []byte {
	// &params[0] will panic on zero length slice, so use nil instead
	par := (*byte)(nil)
	size := int32(len(params))
	if size != 0 {
		par = &params[0]
	}

	// call sandbox function, result value will be cached by host
	// always negative funcNr as keyLen indicates sandbox call
	// this removes the need for a separate hostSandbox function
	size = hostStateGet(nil, funcNr, par, size)

	// zero length, no need to retrieve cached value
	if size == 0 {
		return []byte{}
	}

	// retrieve cached result value from host
	result := make([]byte, size)
	_ = hostStateGet(nil, 0, &result[0], size)
	return result
}

func (w *WasmVMHost) StateDelete(key []byte) {
	// value size -1 means delete key
	// this removes the need for a separate hostStateDel function
	hostStateSet(&key[0], int32(len(key)), nil, -1)
}

func (w *WasmVMHost) StateExists(key []byte) bool {
	// value size -1 means only test for existence
	// returned size -1 indicates keyID not found (or error)
	// this removes the need for a separate hostStateExists function
	return hostStateGet(&key[0], int32(len(key)), nil, -1) >= 0
}

func (w *WasmVMHost) StateGet(key []byte) []byte {
	//TODO optimize when type size is known in advance?
	// or maybe pass in a larger buffer that will fit most rsults?
	//size := int32(len(value))
	//if size != 0 {
	//	// size known in advance, just get the data
	//	_ = hostStateGet(&key[0], int32(len(key)), &value[0], size)
	//	return value
	//}

	// variable sized result expected,
	// query size first by passing zero length buffer
	// value will be cached by host
	size = hostStateGet(&key[0], int32(len(key)), nil, 0)

	// -1 means non-existent
	if size < 0 {
		return []byte(nil)
	}

	// zero length, no need to retrieve cached value
	if size == 0 {
		return []byte{}
	}

	// retrieve cached value from host
	value := make([]byte, size)
	_ = hostStateGet(nil, 0, &value[0], size)
	return value
}

func (w *WasmVMHost) StateSet(key []byte, value []byte) {
	// &value[0] will panic on zero length slice, so use nil instead
	val := (*byte)(nil)
	size := int32(len(value))
	if size != 0 {
		val = &value[0]
	}
	hostStateSet(&key[0], int32(len(key)), val, size)
}
