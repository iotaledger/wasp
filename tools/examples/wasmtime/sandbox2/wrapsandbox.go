package main

import (
	"github.com/bytecodealliance/wasmtime-go"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func getSandboxFunctions(sb vmtypes.Sandbox, store *wasmtime.Store, memory *[]byte) []*wasmtime.Extern {
	return []*wasmtime.Extern{
		// Publish
		wasmtime.WrapFunc(store, func(ptr int32, len int32) {
			str := string((*memory)[ptr : ptr+len])
			sb.Publish(str)
		}).AsExtern(),

		// GetInt64
		//wasmtime.WrapFunc(store, func (ptr int32, len int32) (int64, int32){
		//	name := string((*memory)[ptr:ptr+len])
		//	val, ok := sb.AccessState().GetInt64(kv.Key(name))
		//	var oki int32
		//	if ok{
		//		oki = 0
		//	} else {
		//		oki = 1
		//	}
		//	return val, oki
		//}).AsExtern(),

		// SetInt64
		wasmtime.WrapFunc(store, func(ptr int32, len int32, val int64) {
			name := string((*memory)[ptr : ptr+len])
			sb.AccessState().SetInt64(kv.Key(name), val)
		}).AsExtern(),
	}
}
