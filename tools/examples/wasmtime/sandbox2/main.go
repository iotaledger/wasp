package main

import (
	"fmt"
	"github.com/bytecodealliance/wasmtime-go"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
	"io/ioutil"
	"reflect"
	"unsafe"
)

const wasmFile = "C:\\Users\\evaldas\\Documents\\proj\\Go\\src\\github.com\\iotaledger\\wasp-develop\\tools\\examples\\wasmtime\\sandbox2\\rust-wasm\\target\\wasm32-unknown-unknown\\release\\rust_wasm_call_sandbox.wasm"

func main() {
	engine := wasmtime.NewEngine()
	store := wasmtime.NewStore(engine)

	// Compiling modules requires WebAssembly binary input, but the wasmtime
	// package also supports converting the WebAssembly text format to the
	// binary format.
	//wasm, err := wasmtime.Wat2Wasm(wat)

	wasm, err := ioutil.ReadFile(wasmFile)
	check(err)

	module, err := wasmtime.NewModule(engine, wasm)
	check(err)

	// TODO mocked empty sandbox
	var nilAddr address.Address
	sbox := sandbox.NewSandbox(&vm.VMContext{
		VirtualState: state.NewVirtualState(mapdb.NewMapDB(), &nilAddr),
		StateUpdate:  state.NewStateUpdate(nil),
	})
	publish := wasmtime.WrapFunc(store, func(ptr int32, len int32) {
		sh := &reflect.SliceHeader{
			Data: uintptr(ptr),
			Len:  int(len),
			Cap:  int(len),
		}

		str := string(*(*[]byte)(unsafe.Pointer(sh)))
		sbox.Publish(str)
		fmt.Println("called publish!")
	})

	// Next up we instantiate a module which is where we link in all our
	// imports. We've got one import so we pass that in here.
	instance, err := wasmtime.NewInstance(store, module, []*wasmtime.Extern{
		publish.AsExtern(),
	})
	check(err)

	// After we've instantiated we can lookup our `run` function and call
	// it.
	if err := callExport(instance, "entry_point1"); err != nil {
		fmt.Printf("error occured: %v\n", err)
	}
}

func callExport(instance *wasmtime.Instance, name string) error {
	var exp *wasmtime.Extern
	if exp = instance.GetExport(name); exp == nil {
		return fmt.Errorf("can't find export '%s'\n", name)
	}
	run := exp.Func()
	_, err := run.Call()
	return err
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
