// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

type (
	ScFuncContextFunction func(ScFuncContext)
	ScViewContextFunction func(ScViewContext)

	ScHost interface {
		ExportName(index int32, name string)
		Sandbox(funcNr int32, params []byte) []byte
		StateDelete(key []byte)
		StateExists(key []byte) bool
		StateGet(key []byte) []byte
		StateSet(key, value []byte)
	}
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\
var host ScHost

func ConnectHost(h ScHost) ScHost {
	oldHost := host
	host = h
	return oldHost
}

func ExportName(index int32, name string) {
	host.ExportName(index, name)
}

func Sandbox(funcNr int32, params []byte) []byte {
	return host.Sandbox(funcNr, params)
}

func StateDelete(key []byte) {
	host.StateDelete(key)
}

func StateExists(key []byte) bool {
	return host.StateExists(key)
}

func StateGet(key []byte) []byte {
	return host.StateGet(key)
}

func StateSet(key, value []byte) {
	host.StateSet(key, value)
}
