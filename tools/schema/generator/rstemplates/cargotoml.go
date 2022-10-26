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
version = "0.1.0"
authors = ["Eric Hop <eric@iota.org>"]
edition = "2021"
repository = "https://github.com/iotaledger/wasp"

[lib]
crate-type = ["cdylib", "rlib"]

$#emit dependencies$cargoMain

[dev-dependencies]
wasm-bindgen-test = "0.3.13"
`,
	// *******************************
	"packageSc": `
name = "$package"
description = "$scDesc"
`,
	// *******************************
	"packageMain": `
name = "main"
description = "Wasm VM host stub for: $scDesc"
`,
	// *******************************
	"dependenciesSc": `
[dependencies]
wasmlib = { git = "https://github.com/iotaledger/wasp", branch = "wasmclient" }
`,
	// *******************************
	"dependenciesMain": `
[features]
default = ["console_error_panic_hook"]

[dependencies]
$package = { path = "../$package" }
wasmvmhost = { git = "https://github.com/iotaledger/wasp", branch = "wasmclient" }
console_error_panic_hook = { version = "0.1.6", optional = true }
wee_alloc = { version = "0.4.5", optional = true }
`,
}
