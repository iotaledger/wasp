---
keywords:
- ISCP
- Smart Contracts
- Chain
- EVM
- Solidity
- Deploy
description: How to create a new EVM Chain and get started
image: /img/logo/WASP_logo_dark.png
---
# Creating an EVM chain

EVM chains run inside ISCP chains. So in order to start a EVM chain you will first need to follow the steps to start a ISCP chain or use a existing ISCP chain to start the EVM chain on.

:::warning

An ISCP chain can only contain 1 EVM chain contract. If your ISCP chain already has a EVM chain contract that is the one that should be used instead of creating a new one.

:::

### 1. Creating the ISCP chain

If you don't have a ISCP chain you should create one first. To do so follow the instructions [on setting up a chain](../../chains_and_nodes/setting-up-a-chain/).

### 2. Fund your account on your newly created chain

In order to deploy the EVM chain contract you first need to have some IOTA locked on your newly created chain to fund that action. In order to do this simply run:

```bash=
wasp-cli chain deposit IOTA:10000
```

This locks 10000 IOTA into your name on your newly created chain which can be used on that chain to pay for the creation of the EVM Chain contract.

### 3. Deploy the EVM Chain Contract

In order to deploy the EVM Chain contract we first need to make sure that the token that will be used on that chain is supplied right at its inception to a given address. You will have to generate a compatible address first with a private key file. Probably the most intuitive way to do this is by utilizing [Metamask](https://metamask.io). In MetaMask we create a wallet (it doesn't matter what chain it is connected to). Once a wallet is generated you'll see a wallet address under it's name. You can copy this to your clipboard. This is the address that will receive the full supply of the tokens on that chain.


![MetaMask](/img/metamask.png)

Once you have this we are ready to deploy the EVM chain with the following command:

```bash=
wasp-cli chain evm deploy --alloc 0x0ad2406D4A50C199ddD08086bC89E8189e86d9d4:1000000000000000000000000
```

The `--alloc` parameter makes sure the provided address (`0x0ad2406D4A50C199ddD08086bC89E8189e86d9d4` in this case) will receive 1 million tokens (it has a additional 18 zero's for decimals).

Once this command has been executed successfully your EVM chain is up and running and ready to be used.

### Running the JSON/RPC interface

In order to communicate with the EVM contract we need to run a additional server application that is compatible with how interaction usually takes place on other networks. This allows us to use other tools from those ecosystems to connect to our EVM chain like [MetaMask](https://metamask.io) and [Hardhat](https://hardhat.org/). To run this server simply run 

```bash=
wasp-cli chain evm jsonrpc --chainid 1074
```

This will start the JSON/RPC server on port 8545 for us with Chain ID 1074. We can now simply point MetaMask or Hardhat to that servers address on port 8545 and interact with it like any other EVM based chain.

:::caution

re-using a existing Chain ID is not recommended and can be a security risk. For any serious chain you will be running make sure you register a unique Chain ID on [Chainlist](https://chainlist.org/) and use that instead of the default.

:::

:::warning

The current implementation uses the wasp-cli's saved account to pay for any fees in IOTA while using the JSON/RPC interface, this is of course not ideal in a normal usage scenario. Releases after this one will address this in a more user friendly way.
:::
