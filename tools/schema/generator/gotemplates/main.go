package gotemplates

var mainGo = map[string]string{
	// *******************************
	"../main.go": `
// +build wasm

package main

import "github.com/iotaledger/wasp/packages/wasmvm/wasmvmhost"

import "$module/go/$package"

func main() {
}

//export on_load
func onLoad() {
	h := &wasmvmhost.WasmVMHost{}
	h.ConnectWasmHost()
	$package.OnLoad()
}
`,
}
