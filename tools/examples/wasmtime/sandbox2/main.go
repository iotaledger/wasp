package main

import (
	"fmt"
	"github.com/bytecodealliance/wasmtime-go"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
	"io/ioutil"
)

const wasmFile = "C:\\Users\\evaldas\\Documents\\proj\\Go\\src\\github.com\\iotaledger\\wasp-develop\\tools\\examples\\wasmtime\\sandbox2\\rust-wasm\\target\\wasm32-unknown-unknown\\release\\rust_wasm_call_sandbox.wasm"

func main() {
	config := wasmtime.NewConfig()
	config.SetWasmMultiValue(true)
	engine := wasmtime.NewEngineWithConfig(config)
	store := wasmtime.NewStore(engine)

	wasm, err := ioutil.ReadFile(wasmFile)
	check(err)

	module, err := wasmtime.NewModule(engine, wasm)
	check(err)

	sbox := sandbox.NewMockedSandbox()

	var memory []byte
	instance, err := wasmtime.NewInstance(store, module, getSandboxFunctions(sbox, store, &memory))
	check(err)

	mem := instance.GetExport("memory").Memory()
	mem.DataSize()
	memory = mem.UnsafeData()
	if err := callExport(instance, "entry_point1"); err != nil {
		fmt.Printf("error occured: %v\n", err)
	}
}

func callExport(instance *wasmtime.Instance, name string, args ...interface{}) error {
	var exp *wasmtime.Extern
	if exp = instance.GetExport(name); exp == nil {
		return fmt.Errorf("can't find export '%s'\n", name)
	}
	run := exp.Func()
	_, err := run.Call(args...)
	return err
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
