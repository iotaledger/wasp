---
description: How to configure and use EVM support in IOTA Smart Contracts.
image: /img/logo/WASP_logo_dark.png
keywords:
- configure
- using
- EVM
- Ethereum
- Solidity
- deploy
- hardhat
- metamask
- JSON
- RPC
- how to
---
# How to use EVM in IOTA Smart Contracts

## 1. Deploy an IOTA Smart Contracts chain

When [deploying an IOTA Smart Contracts chain](../chains_and_nodes/setting-up-a-chain.md), EVM support is automatically added with the default configuration. The `wasp-cli chain deploy` command accepts some EVM-specific options, listed below:

* `--evm-chainid <n>`: EVM chain ID (default: 1074).

  :::caution

  Re-using an existing Chain ID is not recommended, and can be a security risk. For any serious usage make sure you register a unique Chain ID on [Chainlist](https://chainlist.org/) and use that instead of the default. **It is not possible to change the EVM chain ID after deployment.**

  :::

* `--evm-block-keep-amount <n>`: Amount of blocks to keep in storage (default: keep all blocks).

* `--evm-gas-limit <n>`: Block gas limit (default: 15000000).

* `--evm-gas-ratio <a>:<b>`: ISC gas : EVM gas ratio (default 1:1). The gas ratio can be changed after deployment by calling the `setGasRatio` function of the `evm` core contract.

You can verify that the EVM support is enabled by visiting the Wasp dashboard and checking the "EVM" section in your ISC chain page. You should see the EVM chain ID and the JSON-RPC endpoint.

## 2. Fund an Ethereum account on your IOTA Smart Contracts chain

In order to send EVM transactions, you need to have an Ethereum address that owns tokens on the ISC chain (L2). These tokens will be used to cover gas fees.

The most intuitive way to do this is by using [Metamask](https://metamask.io). In MetaMask,  you can create a wallet (it does not matter what chain it is connected to). Once a wallet is generated, you will see a wallet address under its name. You can copy this to your clipboard. This is the address that will receive the full supply of tokens on that chain.

Assuming that you also have an IOTA account with some L1 funds, you can deposit some of those funds into the Ethereum address' L2 account:

```shell
wasp-cli chain deposit 0xa1b2c3d4... iota:1000000
```

## 3. Connect to the JSON-RPC service

You can point any Ethereum tool like MetaMask or Hardhat to the JSON-RPC endpoint that is displayed on the ISC chain dashboard page (Wasp dashboard / Chains / your chain). Once connected, you should be able to use your tool as if it was connected to any other EVM based chain.
