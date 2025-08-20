// Package evmdoc contains internal documentation about EVM support in
// ISC.
//
// Note: this article is best viewed as rendered by `godoc`
// or `gopls` "Browse package documentation".
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
//   - [github.com/iotaledger/wasp/v2/packages/vm/core/evm/evmtest]: solo tests
//     for the EVM core contract
//   - [github.com/iotaledger/wasp/v2/packages/evm/jsonrpc/jsonrpctest]: solo
//     tests for the JSONRPC service
//   - tools/cluster/tests/evm_jsonrpc_test.go: cluster tests
//   - [github.com/iotaledger/wasp/v2/packages/evm/evmtest]: common Solidity
//     code used in tests
//
// # go-ethereum
//
// The EVM support relies heavily on a [fork] of [go-ethereum], the official
// implementation of the Ethereum execution layer.
//
// The changes that the fork introduces are minimal, and the main purpose is
// to allow executing custom code on the magic contract at address 0x1074.
//
// The fork must be kept up to date with the upstream changes. In order to
// do that, and to keep the process as simple as possible, we always aim to
// have a single commit on top of the latest release tag.
//
// For example, if the latest official release tag is [v1.15.5], then we add a
// [single commit] and call that [v1.15.5-wasp].
//
// So when a new release is published on go-ethereum, we:
//
//  1. Checkout the latest release tag (e.g. v1.15.6)
//  2. Cherry-pick the commit with our changes (be careful to use the latest
//     one)
//  3. Add the new tag with the -wasp suffix (e.g. v1.15.6-wasp)
//
// # The magic contract
//
// The EVM magic contract lives in address 0x1074.
//
// The interface of the contract is defined in several .sol files, that are
// compiled with the Solidity compiler and embedded in [iscmagic.SandboxABI],
// [iscmagic.UtilABI], etc.
//
// Any calls to the magic contract address are handled in
// [evmimpl.magicContract.Run]. Given the name of the called EVM method,
// the corresponding Go function is called. For example, if
// ISCSandbox::getEntropy is called, this is translated as a call to
// [evmimpl.magicContractHandler.GetEntropy].
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
//     [services.WaspEVMBackend]. So,
//     [services.WaspEVMBackend.EVMSendTransaction] is called.
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
//     ([services.WaspEVMBackend.EVMEstimateGas] in the production environment),
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
//     ([services.WaspEVMBackend.EVMCall] in the production environment),
//     which, in turn, calls [chainutil.EVMCall].
//
//   - The call is wrapped in a fake off-ledger request via
//     [isc.NewEVMOffLedgerCallRequest], which is processed the same way as in
//     the gas estimation case.
//
// # Tracing
//
// When receiving a call to [debug_traceBlockByNumber] or similar:
//
//   - The corresponding method in [jsonrpc.EthService] is invoked, e.g.
//     [jsonrpc.EthService.TraceBlockByNumber].
//
//   - [jsonrpc.ChainBackend.EVMTrace] is called
//     ([services.WaspEVMBackend.EVMTrace] in the production environment),
//     which, in turn, calls [chainutil.EVMTrace].
//
//   - [chainutil.EVMTrace] will execute a full VM run on the given block
//     number, re-executing all transactions in the block (even if we are only
//     interested in the trace of a single transaction). For this purpose,
//     [vmimpl.Run] is called.
//
//   - The [vm.VMTask] passed to [vmimpl.Run] contains a non-null
//     [tracers.Tracer]. This comes from a call to [jsonrpc.newTracer].
//     Given the tracer type specified in the request, one of
//     [jsonrpc.newCallTracer] or [jsonrpc.newPrestateTracer] is called.
//     Each one of these contains the specific logic for the actual tracing.
//
//   - After the VM run is done, [tracers.Tracer.GetResult] is called. This
//     function returns the trace results in json format.
//
// [fork]: https://github.com/iotaledger/go-ethereum
// [go-ethereum]: https://github.com/ethereum/go-ethereum
// [single commit]: https://github.com/iotaledger/go-ethereum/commit/cd897d7b31192a9042d59b7c60e7c172de79da14
// [v1.15.5-wasp]: https://github.com/iotaledger/go-ethereum/tree/v1.15.5-wasp
// [v1.15.5]: https://github.com/ethereum/go-ethereum/tree/v1.15.5
// [eth_sendRawTransaction]: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_sendrawtransaction
// [eth_estimateGas]: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_estimategas
// [eth_call]: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_call
// [debug_traceBlockByNumber]: https://geth.ethereum.org/docs/interacting-with-geth/rpc/ns-debug#debugtraceblockbynumber
package evmdoc

import (
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/wasp/v2/packages/chain"
	"github.com/iotaledger/wasp/v2/packages/chainutil"
	"github.com/iotaledger/wasp/v2/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/v2/packages/vm/vmimpl"
	"github.com/iotaledger/wasp/v2/packages/webapi/services"
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
	_ services.WaspEVMBackend
	_ *tracers.Tracer
	_ *vm.VMTask
	_ = vmimpl.Run
)
