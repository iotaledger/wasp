---
keywords:
- ISCP
- Smart Contracts
- EVM
- Solidity
description: EVM based smart contracts
image: /img/logo/WASP_logo_dark.png
---
# EVM/Solidity based smart contracts

Next to the WASM based smart contracts ISCP has to offer, the current release of ISCP also has experimental support for EVM/Solidity smart contracts as well. This means that existing smart contracts and tooling from other EVM based chains like Ethereum are fully compatible with EVM chains running on ISCP. This allows us to offer the existing ecosystem around EVM/Solidity a familiar alternative.

:::caution

This experimental implementation currently does not have the ability yet to interact with Layer 1 IOTA tokens. We'll bring support for this in a later release.

:::

### What is EVM / Solidity

[EVM](https://ethereum.org/en/developers/docs/evm/) stands for "Ethereum Virtual Machine" which currently is the tried and proven virtual machine running most smart contract implementations. [Solidity](https://soliditylang.org/) is the programming language of choice with EVM and has been created for this specific purpose.
The main benefit of using EVM / Solidity right now is the sheer amount of resources available for it from years of development and experimentation on Ethereum. There are many articles, tutorials, examples and tools available for EVM / Solidity and the ISCP implementation is fully compatible with it. If you have experience developing on other EVM based chains you will feel right at home and any existing contracts you've written probably need no or very minimal changes to function on ISCP as well.

### How ISCP works with EVM

With ISCP, an EVM based chain runs inside a ISCP chain as a ISCP smart contract. It is possible to run both WASM based smart contracts and a EVM chain in a single ISCP chain because of this. We offer a EVM compatible JSON/RPC server as part of the `wasp-cli` which allows you to connect to these EVM Chains using existing tooling like [MetaMask](https://metamask.io/), [Remix](https://remix.ethereum.org/) or [Hardhat](https://hardhat.org/). Deploying to a new EVM chain is as easy as pointing your tools to the address of your JSON / RPC gateway.

