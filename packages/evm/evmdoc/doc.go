// Package evmdoc contains internal documentation about EVM support in
// ISC.
//
// # EVM support
//
// The main components of the EVM subsystem are:
//
//   - The EVM emulator, in package [emulator]
//   - The evm core contract interface, in package [evm]
//   - The evm core contract implementation, in package [evmimpl]
//   - The Solidity interface of the ISC magic contract, in
//     package [iscmagic]
//   - The JSONRPC service, in package [jsonrpc])
//
// Tests are grouped in the following packages:
//
//   - [github.com/iotaledger/wasp/packages/vm/core/evm/evmtest]: solo tests
//     for the EVM core contract
//   - [github.com/iotaledger/wasp/packages/evm/jsonrpc/jsonrpctest]: solo
//     tests for the JSONRPC service
//   - tools/cluster/tests/evm_jsonrpc_test.go: cluster tests
//   - [github.com/iotaledger/wasp/packages/evm/evmtest]: common Solidity
//     code used in tests
//
// # Handling Ethereum Transactions
//
// Let's follow the path of an Ethereum transaction.
//
//   - The sender connects their Metamask client to the JSON-RPC.
//
//     Each Wasp node provides a JSONRPC service (see [jsonrpc.EthService]),
//     available via HTTP and websocket transports. When setting up a chain
//     in Metamask, the user must configure the JSONRPC endpoint (like
//     `<node-api>/chain/<isc-chainid>/evm/jsonrpc`).
//
//   - The sender sends the signed EVM transaction via Metamask.
//
//     Metamask signs the EVM transaction with the sender's Ethereum private
//     key, and then calls the [eth_sendRawTransaction] JSONRPC
//     endpoint.
//
//   - The method [jsonrpc.EthService.SendRawTransaction] is invoked.
//
//   - After decoding the transaction and performing several validations, the
//     transaction is sent to the backend for processing, by calling
//     [jsonrpc.ChainBackend.EVMSendTransaction].
//
//     The main implementation of the ChainBackend interface is
//     [jsonrpc.WaspEVMBackend]. So,
//     [jsonrpc.WaspEVMBackend.EVMSendTransaction] is called.
//
//   - The Ethereum transaction is wrapped into an ISC off-ledger request by
//     calling [isc.NewEVMOffLedgerTxRequest], and sent to the mempool for later
//     processing by the ISC chain, by calling
//     [chain.ChainRequests.ReceiveOffLedgerRequest].
//
//   - Some time later the ISC request is picked up by the consensus, and
//     consequently processed by the ISC VM. The `evmOffLedgerTxRequest` acts as
//     a regular ISC off-ledger request that calls the evm core contract's
//     [evm.FuncSendTransaction] entry point with a single parameter:: the
//     serialized Ethereum transaction. (See methods CallTarget and Params of
//     type `isc.evmOffLedgerTxRequest`.)
//
//   - The [evm.FuncSendTransaction] entry point is handled by the evm core
//     contract's function `applyTransaction` in package [evmimpl].
//
//   - The evm core contract calls [emulator.EVMEmulator.SendTransaction],
//     which in turn calls [emulator.EVMEmulator.applyMessage], which calls
//     [core.ApplyMessage]. This actually executes the EVM code.
//
// # Gas Estimation
//
// Metamask usually calls [eth_estimateGas] before [eth_sendRawTransaction].
// This is processed differently:
//
//   - The method [jsonrpc.EthService.EstimateGas] is invoked instead, with the
//     unsigned call parameters instead of a signed transaction.
//
//   - [jsonrpc.ChainBackend.EVMEstimateGas] is called
//     ([jsonrpc.WaspEVMBackend.EVMEstimateGas] in the production environment),
//     which, in turn, calls [chainutil.EVMEstimateGas].
//
//   - [chainutil.EVMEstimateGas] performs a binary search, executing the call
//     with different gas limit values.
//
//     Each call is performed by wrapping it into a "fake" request, by calling
//     [isc.NewEVMOffLedgerCallRequest], and executing a VM run as if it was
//     run by the consensus. Any state changes are discarded afterwards.
//
//   - The fake off-ledger request calls the [evm.FuncCallContract] entry point of
//     the evm core contract.
//
//   - The entry point handler function, `callContract` calls
//     [emulator.EVMEmulator.CallContract], which in turn calls
//     [emulator.EVMEmulator.applyMessage], just like when processing a regular
//     transaction.
//
// # View Calls
//
// When Metamask calls [eth_call] to perform a view call, the execution path is
// similar to the gas estimation case:
//
//   - The method [jsonrpc.EthService.Call] is invoked.
//
//   - [jsonrpc.ChainBackend.EVMCall] is called
//     ([jsonrpc.WaspEVMBackend.EVMCall] in the production environment),
//     which, in turn, calls [chainutil.EVMCall].
//
//   - The call is wrapped in a fake off-ledger request via
//     [isc.NewEVMOffLedgerCallRequest], which is processed the same way as in
//     the gas estimation case.
//
// [eth_sendRawTransaction]: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_sendrawtransaction
// [eth_estimateGas]: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_estimategas
// [eth_call]: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_call
package evmdoc

import (
	"github.com/ethereum/go-ethereum/core"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

// dummy variables to keep the imports
var (
	_ = &evm.Contract
	_ *emulator.StateDB
	_ = &evmimpl.Processor
	_ = &iscmagic.Address
	_ *jsonrpc.EthService
	_ = isc.NewEVMOffLedgerTxRequest
	_ = chainutil.EVMEstimateGas
	_ chain.ChainRequests
	_ core.BlockChain
)
