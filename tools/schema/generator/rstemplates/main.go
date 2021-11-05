package rstemplates

var mainRs = map[string]string{
	// *******************************
	"main.rs": `
// +build wasm

package main

import "github.com/iotaledger/wasp/packages/vm/wasmclient"

import "$module/rs/$package"

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
