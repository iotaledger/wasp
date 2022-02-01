---
description: Solo is a testing framework that allows developers to validate real smart contracts and entire inter-chain protocols
image: /img/logo/WASP_logo_dark.png
keywords:
- testing framework
- golang
- rust
- inter-chain protocols
- validate smart contracts
- install
---
# Solo

## What is Solo?

[_Solo_](https://github.com/iotaledger/wasp/tree/master/packages/solo) is a testing framework that allows developers to validate real smart contracts and entire inter-chain protocols before deploying them on the distributed network.

## Installation

_Solo_ tests are written in Go. Go (version 1.16) needs to be installed on your machine.

_Solo_ is part of the [_Wasp_ codebase repository](https://github.com/iotaledger/wasp.git). You can access the Solo framework by cloning the repository with the following command:

```shell
git clone https://github.com/iotaledger/wasp.git
```

Alternatively, you can install the Solo package separately using the following command:

In Linux/macOS:

```shell
go get github.com/iotaledger/wasp/packages/solo
```

In Windows:

```shell
go get -buildmode=exe github.com/iotaledger/wasp/packages/solo
```

To run Rust/Wasm smart contracts you will also need [Rust](https://www.rust-lang.org/tools/install) and [wasm-pack](https://rustwasm.github.io/wasm-pack/installer/) installed.
You can use any development environment for Rust and Go.
The GoLang environment with the Rust plugin is a good combination.

You can find example implementations of smart contracts (including source code
and tests) in the Wasp repository, in the
[contracts/wasm folder](https://github.com/iotaledger/wasp/tree/master/contracts/wasm).

:::tip
You can find the documentation for all the functionalities available in solo in [go-docs](https://pkg.go.dev/github.com/iotaledger/wasp/packages/solo).
:::

In the following pages some usage examples will be presented. The example code can be found in the [Wasp repository](https://github.com/iotaledger/wasp/tree/develop/documentation/tutorial-examples).
