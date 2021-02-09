![Wasp logo](WASP_logo_dark.png)

# Welcome to the Wasp repository!

[Wasp](https://github.com/iotaledger/wasp) is a node software developed by the
[IOTA Foundation](http://iota.org) to run the _IOTA Smart Contract Protocol_
(_ISCP_ in short) on top of the _IOTA Tangle_.  Please find here a [high level
introduction](https://blog.iota.org/an-introduction-to-iota-smart-contracts-16ea6f247936)
into ISCP.

A _smart contract_ is a distributed software agent that stores its state in the
[UTXO ledger](articles/intro/utxo.md), and evolves with each _request_ sent to
the contrct. Since the UTXO ledger is immutable, by extension the smart
contract state is also immutable.

A _committee_ of an arbitrary number of Wasp nodes runs a _chain_ of smart
contracts.  The main purpose of the _committee_ is to ensure consistent
transition from the previous state to the next, according to the attached
program.  This ensures that the operation of smart contracts is distributed,
fault-tolerant and leaderless.

The articles below explain how to run a Wasp node on the Pollen network, as
well as concepts and architecture of ISCP and Wasp.

_Disclaimer: Wasp node and articles is a work in progress, and most likely will
always be.  The software presented in this repository is not ready for use in
commercial settings or whenever processing of critical data is involved._

## Prerequisites

- Go 1.15

## Compile

- Build the `wasp` binary (Wasp node): `go build`
- Build the `wasp-cli` binary (CLI client): `go build ./tools/wasp-cli`

## Test

- Run all tests (including `tools/cluster` tests which may take several minutes): `go test -timeout 20m ./...`
- Run only unit tests: `go test -short ./...`

## Run

To run a Wasp node you need at least one Goshimmer node with the
[WaspConn](https://github.com/iotaledger/goshimmer/tree/master+wasp/dapps/waspconn)
plugin. This version of Goshimmer is located in the
[`master+wasp` branch of the Goshimmer repository](https://github.com/iotaledger/goshimmer/tree/master+wasp).

The only difference between standard Goshimmer (the `develop` branch) and the
`master+wasp` branch is the `WaspConn` plugin, which accepts connections from Wasp
nodes.

- [How to run a Wasp node on Pollen](articles/docs/runwasp.md)

## Develop

Below are some articles describing the architecture:

- [Core types](articles/docs/coretypes.md)
- [On-chain accounts](articles/docs/accounts.md)
- [Wasp Publisher](articles/docs/publisher.md)

## Tools

- [`wasp-cli`](tools/wasp-cli/README.md): A CLI client for the Wasp node.
- [`wasp-cluster`](tools/cluster/wasp-cluster/README.md): allows to easily run
  a cluster of Wasp nodes, for testing.

## PoC smart contracts

- [Main concepts with _DonateWithFeedback_](articles/intro/dwf.md)
- [Deployment of the smart contract](articles/intro/deploy.md)
- [Handling tagged tokens with _TokenRegistry_ and _FairAuction_ smart contracts](articles/intro/tr-fa.md)
- [Short intoduction to UTXO ledger and digital assets](articles/intro/utxo.md)

