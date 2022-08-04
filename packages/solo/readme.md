## Package Solo

Package `solo` is a development tool for writing unit tests for IOTA Smart Contracts (ISC).

The package is intended for developers of smart contracts as well as contributors to the development
of the ISC and the [Wasp node](https://github.com/iotaledger/wasp) itself.

Normally, the smart contract is developed and tested in the `solo` environment before trying 
it out on the network of Wasp nodes. Running and testing the smart contract on 'solo' 
does not require to run the Wasp nodes nor committee of nodes: just ordinary 'go test' environment. 
Same time, the smart contracts in `solo` is run in native environment, including transactions, tokens, signatures and
virtual state access. This allows deployment of smart contracts on Wasp network without any changes.

See here the GoDoc documentation of the `solo` package:
 [![Go Reference](https://pkg.go.dev/badge/iotaledger/wasp/packages/solo.svg)](https://pkg.go.dev/github.com/iotaledger/wasp/packages/solo)
