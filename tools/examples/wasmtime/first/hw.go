package main

import (
	"fmt"
	"github.com/bytecodealliance/wasmtime-go"
)

func main() {
	// Almost all operations in wasmtime require a contextual `store`
	// argument to share, so create that first
	engine := wasmtime.NewEngine()
	store := wasmtime.NewStore(engine)

	// Compiling modules requires WebAssembly binary input, but the wasmtime
	// package also supports converting the WebAssembly text format to the
	// binary format.
	wasm, err := wasmtime.Wat2Wasm(`
      (module
        (import "" "" (func $hello))
        (import "" "" (func $mumu))
        (import "" "" (func $kuku))
        (func (export "run1")
          (call $hello))
        (func (export "run2")
          (call $mumu) (call $kuku))
      )
    `)
	check(err)

	// Once we have our binary `wasm` we can compile that into a `*Module`
	// which represents compiled JIT code.
	module, err := wasmtime.NewModule(engine, wasm)
	check(err)

	// Our `hello.wat` file imports one item1, so we create that function
	// here.
	item1 := wasmtime.WrapFunc(store, func() {
		fmt.Println("Hello from Go!")
	})

	item2 := wasmtime.WrapFunc(store, func() {
		fmt.Println("Mumu from Go!")
	})

	item3 := wasmtime.WrapFunc(store, func() {
		fmt.Println("Kuku from Go!")
	})

	// Next up we instantiate a module which is where we link in all our
	// imports. We've got one import so we pass that in here.
	instance, err := wasmtime.NewInstance(store, module, []*wasmtime.Extern{item1.AsExtern(), item2.AsExtern(), item3.AsExtern()})
	check(err)

	// After we've instantiated we can lookup our `run` function and call
	// it.
	if err := callExport(instance, "run"); err != nil {
		fmt.Printf("error occured: %v\n", err)
	}
	if err := callExport(instance, "run1"); err != nil {
		fmt.Printf("error occured: %v\n", err)
	}
	if err := callExport(instance, "run2"); err != nil {
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
