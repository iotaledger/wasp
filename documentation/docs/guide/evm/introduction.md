---
description: The current release of IOTA Smart Contracts also has experimental support for EVM/Solidity,this means that existing smart contracts and tooling from other EVM based chains like Ethereum are fully compatible with EVM chains running on IOTA Smart Contracts. 
image: /img/logo/WASP_logo_dark.png
keywords:
- EVM
- EVM chain
- Solidity
- smart contracts
- Ethereum
- explanation
---
# EVM/Solidity Based Smart Contracts

The current release of IOTA Smart Contracts has experimental support for EVM/Solidity smart contracts as well as Wasm based smart contracts. This means that existing smart contracts and tooling from other EVM based chains like Ethereum are fully compatible with EVM chains running on IOTA Smart Contracts. This allows us to offer the existing ecosystem around EVM/Solidity a familiar alternative.

:::caution

This experimental implementation currently does not have the ability yet to interact with Layer 1 IOTA tokens. We will bring support for this in a later release.

:::

### What is EVM/Solidity

[EVM](https://ethereum.org/en/developers/docs/evm/) stands for "Ethereum Virtual Machine", which currently is the tried and proven virtual machine running most smart contract implementations. [Solidity](https://soliditylang.org/) is the programming language of choice with EVM, and has been created for this specific purpose.

The main benefit of using EVM/Solidity right now is the sheer amount of resources available from it from years of development and experimentation on Ethereum. There are many articles, tutorials, examples and tools available for EVM/Solidity, and the IOTA Smart Contracts implementation is fully compatible with them. If you have experience developing on other EVM based chains you will feel right at home, and any existing contracts you've written will probably need no (or very minimal) changes to function on IOTA Smart Contracts as well.

### How IOTA Smart Contracts Work With EVM

With IOTA Smart Contracts, an EVM based chain runs inside an IOTA Smart Contracts chain as an IOTA Smart Contracts smart contract. Because of this, it is possible to run both Wasm based smart contracts and an EVM chain in a single IOTA Smart Contracts chain. We offer an EVM compatible JSON-RPC server as part of the `wasp-cli`, which allows you to connect to these EVM Chains using existing tooling like [MetaMask](https://metamask.io/), [Remix](https://remix.ethereum.org/) or [Hardhat](https://hardhat.org/). Deploying to a new EVM chain is as easy as pointing your tools to the address of your JSON-RPC gateway.

