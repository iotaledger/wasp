---
description: The current release of IOTA Smart Contracts also has experimental support for EVM/Solidity, providing limited compatibility with existing smart contracts and tooling from other EVM based chains like Ethereum.
image: /img/logo/WASP_logo_dark.png
keywords:
- EVM
- Solidity
- smart contracts
- Ethereum
- explanation
---
# EVM/Solidity Based Smart Contracts

The current release of IOTA Smart Contracts has experimental support for
EVM/Solidity smart contracts as well as Wasm based smart contracts, providing
limited compatibility with existing smart contracts and tooling from other EVM
based chains like Ethereum. This allows us to offer the existing ecosystem
around EVM/Solidity a familiar alternative.

### What is EVM/Solidity

[EVM](https://ethereum.org/en/developers/docs/evm/) stands for "Ethereum Virtual Machine", which currently is the tried and proven virtual machine running most smart contract implementations. [Solidity](https://soliditylang.org/) is the programming language of choice with EVM, and has been created for this specific purpose.

The main benefit of using EVM/Solidity right now is the sheer amount of resources available from it from years of development and experimentation on Ethereum. There are many articles, tutorials, examples and tools available for EVM/Solidity, and the IOTA Smart Contracts implementation is fully compatible with them. If you have experience developing on other EVM based chains you will feel right at home, and any existing contracts you've written will probably need no (or very minimal) changes to function on IOTA Smart Contracts as well.

### How IOTA Smart Contracts Work With EVM

Every deployed IOTA Smart Contracts chain automatically includes a core contract called `evm`. This core contract is responsible for running EVM code and storing the EVM state. The Wasp node also provides a standard JSON-RPC service, which allows you to interact with the EVM layer using existing tooling like [MetaMask](https://metamask.io/), [Remix](https://remix.ethereum.org/) or [Hardhat](https://hardhat.org/). Deploying EVM contracts is as easy as pointing your tools to the JSON-RPC endpoint.

