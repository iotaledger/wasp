// Package evmimpl contains the implementation of the `evm` core contract.
//
// The `evm` core contract is responsible for executing and storing:
//   - The [github.com/iotaledger/wasp/v2/packages/vm/core/evm/emulator.EVMEmulator].
//   - The ISC magic contract (see iscmagic.go).
//
// See:
//   - The [SetInitialState] function, which initializes EVM on a newly created
//     ISC chain.
//   - The [evm.FuncSendTransaction] entry point, which executes an Ethereum transaction.
//   - The [evm.FuncCallContract] entry point, which executes a view call.
package evmimpl
