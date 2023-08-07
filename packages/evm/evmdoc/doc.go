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
// # Handling of Ethereum Transactions
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
// [eth_sendRawTransaction]: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_sendrawtransaction
package evmdoc

import (
	"github.com/ethereum/go-ethereum/core"

	"github.com/iotaledger/wasp/packages/chain"
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
	_ chain.ChainRequests
	_ core.BlockChain
)
