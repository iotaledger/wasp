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
# Testing Smart Contracts with Solo

## What is Solo?

[_Solo_](https://github.com/iotaledger/wasp/tree/develop/packages/solo) is a testing framework that allows developers to validate real smart contracts and entire inter-chain protocols before deploying them on the distributed network.

## Installation

_Solo_ tests are written in Go. Go (version 1.18) needs to be installed on your machine.

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

:::tip
You can browse the Solo Go API reference (updated to the `master` branch) in [go-docs](https://pkg.go.dev/github.com/iotaledger/wasp/packages/solo).
:::

Of course, along with Solo you will need a smart contract to test.
You can find example implementations of Rust/Wasm smart contracts (including source code and tests) in the Wasp repository, in the [contracts/wasm folder](https://github.com/iotaledger/wasp/tree/develop/contracts/wasm).
For information about how to create Wasm smart contracts, refer to the [Wasm VM chapter](../wasm_vm/intro.mdx).

In the following sections we will present some Solo usage examples. The example code can be found in the [Wasp repository](https://github.com/iotaledger/wasp/tree/develop/documentation/tutorial-examples).
