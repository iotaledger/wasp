package gotemplates

var mainGo = map[string]string{
	// *******************************
	"../main.go": `
// +build wasm

package main

import "github.com/iotaledger/wasp/packages/vm/wasmclient"

import "$module/go/$package"

func main() {
}

//export on_load
func onLoad() {
	h := &wasmclient.WasmVMHost{}
	h.ConnectWasmHost()
	$package.OnLoad()
}
`,
}
