// Package emulator contains the implementation of the EVMEmulator and
// subcomponents.
//
// The main components in this package are:
//
//   - [EVMEmulator], which is responsible for executing the Ethereum
//     transactions.
//
//     The [EVMEmulator] relies on the `go-ethereum` implementation of the
//     Ethereum Virtual Machine (EVM). We use a [fork] that [adds support] for
//     ISC's magic contract.
//
//   - [StateDB], which adapts go-ethereum's [vm.StateDB] interface to ISC's
//     [kv.KVStore] interface. In other words, it keeps track of the EVM state
//     (the key-value store, account balances, contract codes, etc), storing it
//     in a subpartition of the `evm` core contract's state.
//
//   - [BlockchainDB], which keeps track of the Ethereum blocks and their
//     transactions and receipts.
//
// The emulator package is mostly agnostic about ISC. It depends only on the
// [ethereum] and [kv] packages.
//
// [fork]: https://github.com/iotaledger/go-ethereum/tree/v1.12.0-wasp
// [adds support]: https://github.com/ethereum/go-ethereum/compare/v1.12.0...iotaledger:go-ethereum:v1.12.0-wasp
package emulator
