## Package Solo

Package `solo` is a development tool for writing unit tests for IOTA Smart Contracts (ISCP).

### A development tool
The package is intended for developers of smart contracts as well as contributors to the development
of the ISCP and the [Wasp node](https://github.com/iotaledger/wasp) itself.

Normally, the smart contract is developed and tested in the `solo` environment before trying it out on the network of Wasp nodes.
Running and testing the smart contract on 'solo' does not require to run the Wasp
nodes nor committee of nodes: just ordinary 'go test' environment.

See here the GoDoc documentation of the `solo` package:
 [![Go Reference](https://pkg.go.dev/badge/iotaledger/wasp/packages/solo.svg)](https://pkg.go.dev/github.com/iotaledger/wasp/packages/solo)

### Native environment
`solo` shares the same code of Virtual Machine with the Wasp node. This guarantees that smart contract programs
can later be deployed on chains which are run by the network of Wasp nodes without any modifications.

The `solo` environment uses in-memory UTXO ledger called _UTXODB_ to validate and store transactions. The _UTXODB_
mocks Goshimmer UTXO ledger. It uses same value transaction structure, colored tokens, signature
schemes as well as transaction and signature validation as in Value Tangle of [Goshimmer (Pollen release)](https://github.com/iotaledger/goshimmer/tree/wasp).
The only difference with the Value Tangle is that _UTXODB_ provides full synchronicity of ledger updates.

The _virtual state_ of the chain (key/value database) in `solo` is an in-memory database. It provides exactly the same
interface of access to it as the database of the Wasp node.

### Writing smart contracts

The smart contracts are usually written in Rust using Rust libraries provided
in the [wasplib repository](https://github.com/iotaledger/wasplib).
Rust code is compiled into the WebAssembly (Wasm) binary.
The Wasm binary is uploaded by `solo` onto the chain and then loaded into the VM
and executed.

Another option to write and run ISCP smart contracts is to use the native Go environment
of the Wasp node and `Sandbox` interface, provided by the Wasp for the VM: the "hardcoded" mode. 
The latter approach is not normally used to develop apps.
However, is is how the 4 builtin smart contracts which constitutes the core of the ISCP chains, are written.
The approach to write "hardcoded" smart contracts may also be a useful for
the development and debugging of the smart contract logic in IDEs such as GoLand, before writing it as
a Rust/Wasm smart contract.

