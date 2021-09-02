# Solo

## What is Solo?

[_Solo_](https://github.com/iotaledger/wasp/tree/master/packages/solo) is a testing framework that allows developers to validate real smart contracts and entire inter-chain protocols before deploying them on the distributed network.

## Installation

_Solo_ tests are written in Go. Go (version 1.16) needs to be installed on your machine.

_Solo_ is part of the _Wasp_ codebase repository, if you clone the entire repo, you'll have access to it:

```shell
git clone https://github.com/iotaledger/wasp.git
```

Alternatively, the Solo package can be installed separately using the following command:

```shell
go get github.com/iotaledger/wasp/packages/solo
```

In Windows:

```shell
go get -buildmode=exe github.com/iotaledger/wasp/packages/solo
```

To run Rust/Wasm smart contracts you will also need `Rust` and `wasm-pack`installed.
You can use any development environment for Rust and Go.
The GoLand environment with the Rust plugin is a good combination.

You can find example implementations of smart contracts (including source code
and tests) in the Wasp repository, in the
[contracts/rust folder](https://github.com/iotaledger/wasp/tree/master/contracts/rust).

:::tip
You can find the documentation for all the functionality available in solo in [go-docs](https://pkg.go.dev/github.com/iotaledger/wasp/packages/solo).
:::

In the following pages some usage examples will be presented. The example code can be found [here](https://github.com/iotaledger/wasp/tree/develop/documentation/tutorial-examples).
