# Welcome to the Wasp

[Wasp](https://github.com/iotaledger/wasp) is a node software developed by the
[IOTA Foundation](http://iota.org) to run the _IOTA Smart Contract Protocol_
(_ISCP_ in short) on top of the _IOTA Tangle_. Please find here a [high level
introduction](https://blog.iota.org/an-introduction-to-iota-smart-contracts-16ea6f247936)
into ISCP.

A _smart contract_ is a distributed software agent that stores its state in the
[UTXO ledger](misc/utxo.md), and evolves with each _request_ sent to
the contract. Since the UTXO ledger is immutable, by extension the smart
contract state is also immutable.

A _committee_ of an arbitrary number of Wasp nodes runs a _chain_ of smart
contracts. The main purpose of the _committee_ is to ensure consistent
transition from the previous state to the next, according to the attached
program. This ensures that the operation of smart contracts is distributed,
fault-tolerant and leaderless.

The articles below explain how to run a Wasp node on the Goshimmer network, as
well as concepts and architecture of ISCP and Wasp.

_Disclaimer: Wasp node and articles is a work in progress, and most likely will
always be. The software presented in this repository is not ready for use in
commercial settings or whenever processing of critical data is involved._

## Prerequisites

- Go 1.16
- Access to a [GoShimmer](https://github.com/iotaledger/goshimmer) node. (GoShimmer is a developing prototype, so some things are prone to break, for a smoother development experience it is recommend to use the GoShimmer code at [this commit](https://github.com/iotaledger/goshimmer/commit/25c827e8326a))

Note: The Wasp node requires the Goshimmer node to have the
[TXStream](https://github.com/iotaledger/goshimmer/tree/master/plugins/txstream)
plugin enabled. Being an experimental plugin, it is currently disabled by default and can
be enabled via configuration.

- [RocksDB 6.15.5](https://github.com/facebook/rocksdb/blob/master/INSTALL.md) (due to an open [issue](https://github.com/tecbot/gorocksdb/issues/203#issuecomment-801067439) more recent versions of rocksdb might not work currently)

### Microsoft Windows Installation Errors

If the go install command is telling you it cannot find gcc you will need to
install [MinGW-w64](https://sourceforge.net/projects/mingw-w64/). When you do
make sure to select *x86_64* architecture instead of the preselected *i686*
architecture. After installation make sure to add this folder to your PATH variable:

```
C:\Program Files\mingw-w64\x86_64-8.1.0-posix-seh-rt_v6-rev0\mingw64\bin
```

## Compile

- Build the `wasp` binary (Wasp node): `go build -tags rocksdb`
- Build the `wasp-cli` binary (CLI client): `go build -tags rocksdb ./tools/wasp-cli`

Alternatively, build and install everything with `go install -tags rocksdb ./...`

On Windows you will need to use `go install -tags rocksdb -buildmode=exe ./...` instead

## Test

- Run all tests (including integration tests which may take several minutes): `go test -tags rocksdb -timeout 20m ./...`
- Run only unit tests: `go test -tags rocksdb -short ./...`

Note: integration tests require the `wasp` and `wasp-cli` commands
in the system path (i.e. you need to run `go install ./...` before running
tests).

## Run

- [How to run a Wasp node](misc/runwasp.md)
- [Using `wasp-cli` to deploy a chain and a contract](misc/deploy.md)

## Learn

- [Exploring IOTA Smart Contracts](tutorial/readme.md)
- [UTXO ledger and digital assets](misc/utxo.md)
- [Core types](misc/coretypes.md)
- [On-chain accounts](misc/accounts.md)
- [Wasp Publisher](misc/publisher.md)

## Tools

- [`wasp-cli`](https://github.com/iotaledger/wasp/tree/master/tools/wasp-cli): A CLI client for the Wasp node.
- [`wasp-cluster`](https://github.com/iotaledger/wasp/tree/master/tools/cluster/wasp-cluster): allows to easily run
  a network of Wasp nodes, for testing.

