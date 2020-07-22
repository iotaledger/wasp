package main

import (
	"fmt"
	"github.com/bytecodealliance/wasmtime-go"
)

const wat = `
      (module
        (import "" "" (func $getInt64))
        (import "" "" (func $setInt64))
        (import "" "" (func $publish))
        (import "" "" (memory 1))

        (func (export "entryPoint1")
          (call $getInt64)
          (call $setInt64)
		)
        (func (export "entryPoint2")
           (call $getInt64) 
           (call $publish)
        )
      )
`

func main() {
	engine := wasmtime.NewEngine()
	store := wasmtime.NewStore(engine)

	// Compiling modules requires WebAssembly binary input, but the wasmtime
	// package also supports converting the WebAssembly text format to the
	// binary format.
	wasm, err := wasmtime.Wat2Wasm(wat)
	check(err)

	// Once we have our binary `wasm` we can compile that into a `*Module`
	// which represents compiled JIT code.
	module, err := wasmtime.NewModule(engine, wasm)
	check(err)

	// Our `hello.wat` file imports one item1, so we create that function
	// here.
	getInt64 := wasmtime.WrapFunc(store, func() {
		fmt.Println("called getInt64!")
	})

	setInt64 := wasmtime.WrapFunc(store, func() {
		fmt.Println("called setInt64!")
	})

	publish := wasmtime.WrapFunc(store, func() {
		fmt.Println("called publish!")
	})

	ty := wasmtime.NewMemoryType(wasmtime.Limits{
		Min: 4096,
		Max: wasmtime.LimitsMaxNone,
	})
	memory := wasmtime.NewMemory(store, ty)

	// Next up we instantiate a module which is where we link in all our
	// imports. We've got one import so we pass that in here.
	instance, err := wasmtime.NewInstance(store, module, []*wasmtime.Extern{
		getInt64.AsExtern(),
		setInt64.AsExtern(),
		publish.AsExtern(),
		memory.AsExtern(),
	})
	check(err)

	// After we've instantiated we can lookup our `run` function and call
	// it.
	if err := callExport(instance, "entryPoint1"); err != nil {
		fmt.Printf("error occured: %v\n", err)
	}
	if err := callExport(instance, "entryPoint2"); err != nil {
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
