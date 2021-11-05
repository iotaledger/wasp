package tstemplates

var mainTs = map[string]string{
	// *******************************
	"main.ts": `
// +build wasm

package main

import "github.com/iotaledger/wasp/packages/vm/wasmclient"

import "$module/ts/$package"

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
