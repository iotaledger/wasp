// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var cargoToml = map[string]string{
	// *******************************
	"../Cargo.toml": `
# Copyright 2020 IOTA Stiftung
# SPDX-License-Identifier: Apache-2.0

[package]
$#emit package$cargoMain
license = "Apache-2.0"
version = "0.1.0"
authors = ["$author"]
edition = "2021"
repository = "https://github.com/iotaledger/wasp"

[lib]
crate-type = ["cdylib", "rlib"]

$#emit dependencies$cargoMain

[dev-dependencies]
wasm-bindgen-test = "0.3.34"
`,
	// *******************************
	"packageLib": `
name = "$package"
description = "Interface library for: $scDesc"
`,
	// *******************************
	"packageImpl": `
name = "$package$+impl"
description = "Implementation library for: $scDesc"
`,
	// *******************************
	"packageWasm": `
name = "$package$+wasm"
description = "Wasm VM host stub for: $scDesc"
`,
	// *******************************
	"dependenciesLib": `
[dependencies]
wasmlib = { git = "https://github.com/iotaledger/wasp", branch = "wasmclient" }
`,
	// *******************************
	"dependenciesImpl": `
[dependencies]
$package = { path = "../$package" }
$#if packagetestwasmlib dependencyErc721
wasmlib = { git = "https://github.com/iotaledger/wasp", branch = "wasmclient" }
`,
	// *******************************
	"dependencyErc721": `
erc721 = { path = "../../../erc721/rs/erc721" }
`,
	// *******************************
	"dependenciesWasm": `
[features]
default = ["console_error_panic_hook"]

[dependencies]
$package$+impl = { path = "../$package$+impl" }
wasmvmhost = { git = "https://github.com/iotaledger/wasp", branch = "wasmclient" }
console_error_panic_hook = { version = "0.1.7", optional = true }
wee_alloc = { version = "0.4.5", optional = true }
`,
}
