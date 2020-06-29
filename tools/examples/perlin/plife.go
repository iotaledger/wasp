package main

import (
	"errors"
	"fmt"
	"github.com/perlin-network/life/exec"
	"io/ioutil"
	"os"
)

type Resolver struct{}

func (r *Resolver) ResolveFunc(module, field string) exec.FunctionImport {
	switch module {
	case "env":
		switch field {
		case "_log_from_wasm":
			return func(vm *exec.VirtualMachine) int64 {
				ptr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				msgLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				msg := vm.Memory[ptr : ptr+msgLen]
				fmt.Printf("[app] %s\n", string(msg))
				return 0
			}

		default:
			panic(fmt.Errorf("unknown import resolved: %s", field))
		}
	default:
		panic(fmt.Errorf("unknown module: %s", module))
	}
}

func (r *Resolver) ResolveGlobal(module, field string) int64 {
	panic("we're not resolving global variables for now")
}

const fname = "rust_wasm_example1.wasm"

func main() {
	fmt.Println("---------------------------------------")
	fmt.Printf("An example of compiling Rust code to WebAssembly and then\nrunning it on Perlin Wasm VM" +
		"Wasm code calls back functions on Go host.\nThat is what is needed for the Wasp VM PoC\n\n")
	input, err := ioutil.ReadFile(fname)
	checkErr(err)

	vmach, err := exec.NewVirtualMachine(input, exec.VMConfig{}, new(Resolver), nil)
	entryID, ok := vmach.GetFunctionExport("app_main") // can be changed to your own exported function
	checkErr(err)
	if !ok {
		checkErr(errors.New("entry not found"))
	}
	ret, err := vmach.Run(entryID)
	if err != nil {
		vmach.PrintStackTrace()
		panic(err)
	}
	if ret == 0 {
		fmt.Printf("Success!!\n")
	} else {
		fmt.Printf("return value = %d\n", ret)
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
}
