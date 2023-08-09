// Package jsonrpc contains the implementation of the JSONRPC service which is
// called by Ethereum clients (e.g. Metamask).
//
// The most relevant types in this package are:
//   - [EthService], [NetService], [DebugService], etc. which contain the
//     implementations for the `eth_*`, `net_*`, `debug_*`, etc., JSONRPC endpoints.
//     Each endpoint corresponds to a public receiver with the same name. For
//     example, `eth_getTransactionCount` corresponds to
//     [EthService.GetTransactionCount].
//   - [EVMChain], which provides common functionality to interact with the EVM
//     state.
//   - [ChainBackend], which provides access to the underlying ISC chain, and
//     has two implementations:
//   - [WaspEVMBackend] for the production environment.
//   - [solo.jsonRPCSoloBackend] for Solo tests.
package jsonrpc
