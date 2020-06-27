package loader

import (
	flag "github.com/spf13/pflag"
)

const (
	CfgWasmBinaryDir = "loader.wasm"
)

func init() {
	flag.String(CfgWasmBinaryDir, "wasm", "path where Wasm binaries are located (using file:// schema")
}
