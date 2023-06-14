// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var cargoToml = map[string]string{
	// *******************************
	"../Cargo.toml": `
[package]
$#emit package$cargoMain
license = "$license"
version = "$version"
authors = ["$author"]
edition = "2021"
$#if repository userRepository

[lib]
crate-type = ["cdylib", "rlib"]

$#emit dependencies$cargoMain

[dev-dependencies]
wasm-bindgen-test = "0.3.34"
`,
	// *******************************
	"userRepository": `
repository = "$repository"
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
$env_wasmlib$+wasmlib = { git = "https://github.com/iotaledger/wasp", branch = "develop" }
`,
	// *******************************
	"dependenciesImpl": `
[dependencies]
$package = { path = "../$package" }
$env_wasmlib$+wasmlib = { git = "https://github.com/iotaledger/wasp", branch = "develop" }
`,
	// *******************************
	"dependenciesWasm": `
[features]
default = ["console_error_panic_hook"]

[dependencies]
$package$+impl = { path = "../$package$+impl" }
$env_wasmvmhost$+wasmvmhost = { git = "https://github.com/iotaledger/wasp", branch = "develop" }
console_error_panic_hook = { version = "0.1.7", optional = true }
wee_alloc = { version = "0.4.5", optional = true }
`,
}
