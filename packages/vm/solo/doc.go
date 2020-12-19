// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package 'solo' is a development tool for writing unit tests for IOTA Smart Contracts (ISCP).
//
// The package is intended for developers of smart contracts and for contributors to the development
// of the ISCP and the Wasp node itself.
//
// Running and testing the smart contract on 'solo' does not require to run the Wasp
// nodes nor committee of nodes: just ordinary 'go test' environment.
// However, 'solo' uses same native code of the Wasp Virtual Machine and therefore smart contract programs
// can later be deployed on chains run by the network of Wasp nodes without any modifications.
//
// The smart contract code usually is written in Rust using Rust libraries provided
// in the 'wasplib' repository at https://github.com/iotaledger/wasplib.
// It then is compiled into the WebAssembly (Wasm) binary.
// The Wasm binary is loaded by 'solo' into the VM environment of the Wasp and executed.
//
// Another option to write and run ISCP smart contracts is to use the native Go environment
// of the Wasp node: the "hardcoded" mode. The latter approach is not normally used to develop apps,
// however is used for the core contract of ISCP chains. This approach may also be very useful for
// development and debugging of the smart contract smart contract logic in IDE such as GoLand.
package solo
