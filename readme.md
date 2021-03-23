![Wasp logo](WASP_logo_dark.png)

# Welcome to the Wasp repository!

[Wasp](https://github.com/iotaledger/wasp) is a node software developed by the
[IOTA Foundation](http://iota.org) to run the _IOTA Smart Contract Protocol_
(_ISCP_ in short) on top of the _IOTA Tangle_.  Please find here a [high level
introduction](https://blog.iota.org/an-introduction-to-iota-smart-contracts-16ea6f247936)
into ISCP.

A _smart contract_ is a distributed software agent that stores its state in the
[UTXO ledger](docs/intro/utxo.md), and evolves with each _request_ sent to
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
- The `goshimmer` command, compiled from [Goshimmer `develop+wasp` branch](https://github.com/iotaledger/goshimmer/tree/develop+wasp)

```
$ git clone -b develop+wasp https://github.com/iotaledger/goshimmer.git
$ cd goshimmer
$ go install
```

Note: The only difference between standard Goshimmer (`master` branch) and the
`develop+wasp` branch is the
[WaspConn](https://github.com/iotaledger/goshimmer/tree/develop+wasp/dapps/waspconn)
plugin, which accepts connections from Wasp nodes.

###Microsoft Windows installation errors:

If the go install command is telling you it cannot find gcc you will need to
install [MinGW-w64](https://sourceforge.net/projects/mingw-w64/). When you do
make sure to select *x86_64* architecture instead of the preselected *i686*
architecture. After installation make sure to add this folder to your PATH variable:
```
C:\Program Files\mingw-w64\x86_64-8.1.0-posix-seh-rt_v6-rev0\mingw64\bin
```

## Compile

- Build the `wasp` binary (Wasp node): `go build`
- Build the `wasp-cli` binary (CLI client): `go build ./tools/wasp-cli`

Alternatively, build and install everything with `go install ./...`

On Windows you will need to use `go install -buildmode=exe ./...` instead

## Test

- Run all tests (including integration tests which may take several minutes): `go test -timeout 20m ./...`
- Run only unit tests: `go test -short ./...`

Note: integration tests require the `goshimmer`, `wasp` and `wasp-cli` commands
in the system path (i.e. you need to run `go install ./...` before running
tests).

## Run

- [How to run a Wasp node on Pollen](docs/docs/runwasp.md)
- [Using `wasp-cli` to deploy a chain and a contract](docs/docs/deploy.md)

## Learn

- [Exploring IOTA Smart Contracts](docs/tutorial/readme.md)
- [UTXO ledger and digital assets](docs/intro/utxo.md)
- [Core types](docs/docs/coretypes.md)
- [On-chain accounts](docs/docs/accounts.md)
- [Wasp Publisher](docs/docs/publisher.md)

## Tools

- [`wasp-cli`](tools/wasp-cli/README.md): A CLI client for the Wasp node.
- [`wasp-cluster`](tools/cluster/wasp-cluster/README.md): allows to easily run
  a network of Wasp nodes, for testing.

## Contributing

We are accepting pull requests! Please create your branch based on `develop`.
You can also join our [Discord server](https://discord.iota.org/) and ping us
in `#smartcontracts-dev`.
