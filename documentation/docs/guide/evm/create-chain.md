---
keywords:
- create
- Chain
- EVM
- Solidity
- Deploy
- hardhat
- metamask
- JSON
- RPC
description: Create, fund and deploy a new EVM Chain using IOTA Smart Contracts.
image: /img/logo/WASP_logo_dark.png
---
# Creating an EVM Chain

EVM chains run inside IOTA Smart Contracts chains. So in order to start an EVM chain, you will first need to follow the steps to [start an IOTA Smart Contracts chain](../chains_and_nodes/setting-up-a-chain.md), or use an existing IOTA Smart Contracts chain to start the EVM chain on.

:::warning

**An IOTA Smart Contracts chain can only contain 1 EVM chain contract**. If your IOTA Smart Contracts chain already has an EVM chain contract, you should use that chain contract instead of creating a new one.

:::

## 1. Create the IOTA Smart Contracts Chain

If you don't have an IOTA Smart Contracts chain, you should create one. To do so, follow the instructions [on setting up a chain](../chains_and_nodes/setting-up-a-chain.md).

## 2. Fund Your Account on Your IOTA Smart Contracts Chain

In order to deploy the EVM chain contract, you need to have some IOTA locked on your newly created chain to fund that action. To do this, run:

```bash
wasp-cli chain deposit IOTA:10000
```

This locks 10000 IOTA into your name on your chain, which can be used on to pay for the creation of the EVM Chain contract.


## 3. Deploy the EVM Chain Contract

In order to deploy the EVM Chain contract, you need to make sure that the token that will be used on that chain is supplied right at its inception to a given address. You will have to generate a compatible address with a private key file. 

The most intuitive way to do this is by using [Metamask](https://metamask.io). In MetaMask,  you can create a wallet (it does not matter what chain it is connected to). Once a wallet is generated, you will see a wallet address under its name. You can copy this to your clipboard. This is the address that will receive the full supply of tokens on that chain.

[![MetaMask](/img/metamask.png)](/img/metamask.png)

Once you have this, you are ready to deploy the EVM chain with the following command:

```bash
wasp-cli chain evm deploy -a mychain --alloc 0x63c00c65BE86463491167eE26958a5A599BEbD2c:1000000000000000000000000
```
* The `-a` parameter indicates the name of the chain that you want to deploy your EVM chain on top of. `mychain` in this case.
* The `--alloc` parameter is the address that you copied from metamask(`0x63c00c65BE86463491167eE26958a5A599BEbD2c` in this case), followed by a `:`, and the value of Smart Contract Tokens that you want to be available to for use on EVM(`1000000` tokens in this case, followed by 18 decimal zeroes).

Once this command has been executed successfully your EVM chain is up and running, and ready to be used.

You can verify the chain has been deployed by visiting the wasp dashboard and checking the smart contracts deployed on the chain. You should be able to see an evm contract over there.

## Running the JSON-RPC Interface

In order to communicate with the EVM contract, you will need to run an additional server application that is compatible with how interactions usually takes place on other networks. This allows you to use other tools from those ecosystems like [MetaMask](https://metamask.io) and [Hardhat](https://hardhat.org/) to connect to our EVM chain. 

To run this server, run the following command: 

```bash
wasp-cli chain evm jsonrpc --chainid 1074
```

This will start the [JSON-RPC](https://www.jsonrpc.org/) server on port 8545 for you with Chain ID 1074. You can now  point MetaMask or Hardhat to that server's address on port 8545, and interact with it like any other EVM based chain.

:::caution

Re-using an existing Chain ID is not recommended, and can be a security risk. For any serious chain you will be running make sure you register a unique Chain ID on [Chainlist](https://chainlist.org/) and use that instead of the default.

:::

:::warning

The current implementation uses the wasp-cli's saves the account to pay for any fees in IOTA while using the JSON-RPC interface.  This is of course not ideal in a normal usage scenario. Upcoming releases  will address this in a more user-friendly way.
:::

## Video Tutorial

<iframe width="560" height="315" src="https://www.youtube.com/embed/JbUGX-9BTSo" title="EVM Chain Setup" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
