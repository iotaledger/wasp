// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var mainGo = map[string]string{
	// *******************************
	"../main.go": `
//go:build wasm
// +build wasm

package main

import "github.com/iotaledger/wasp/packages/wasmvm/wasmvmhost"

import "$module/go/$package"

func main() {
}

func init() {
	wasmvmhost.ConnectWasmHost()
}

//export on_call
func onCall(index int32) {
	$package.OnLoad(index)
}

//export on_load
func onLoad() {
	$package.OnLoad(-1)
}
`,
}
