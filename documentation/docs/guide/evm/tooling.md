---
keywords:
- Smart Contracts
- Chain
- EVM
- Solidity
- Tooling
- wasp-cli
- hardhat
- metamask
- JSON-RPC
description: Existing EVM tooling is compatible and can be used directly with an IOTA Smart Contracts chain running EVM. You can configure hardhat, metamask, remix, Ether.js and Web3.js among others.
image: /img/logo/WASP_logo_dark.png
---
# EVM Tooling

EVM on IOTA Smart Contracts has been integrated in a way that the existing EVM tooling is compatible, and can be used directly with an IOTA Smart Contracts chain running EVM as long as a couple of things are taken into account.

## Tooling Considerations

 1. Please make sure you use the correct JSON-RPC endpoint URL in your tooling for your chain. If you run locally this will simply be `localhost:8545`.
 2. Please make sure you use the right `Chain ID` as configured while starting the JSON-RPC service. If you did not explicitly define this while starting the service, the default Chain ID will be `1074`. 
 3. Fees are being handled on the IOTA Smart Contracts chain level, not EVM level. Because of this, you can simply use a gas price of 0 on EVM level at this point in time.

:::caution

Re-using an existing Chain ID is not recommended and can be a security risk. For any serious chain you will be running make sure you register a unique Chain ID on [Chainlist](https://chainlist.org/) and use that instead of the default.

:::

## Wasp-cli

The Wasp CLI has some very basic functionalities to manage an EVM chain. Given the compatibility with existing tooling, only the basics are covered to get started with IOTA Smart Contracts and EVM. You can currently either run a JSON-RPC server, or deploy the EVM Chain itself on an IOTA Smart Contracts chain. To see the available options and configuration parameters simply run:

```bash
wasp-cli chain evm
```

## MetaMask

[MetaMask](https://metamask.io/) is a popular EVM compatible wallet which runs in a browser extension that allows you to let your wallet interact with web applications utilizing an EVM chain (dApps). 

To use your EVM chain with MetaMask, simply open up MetaMask and click on the network drop-down list at the very top. At the bottom of this list you will see the option `Custom RPC`, click on this. For a local setup use the values as shown in the image below:

[![MetaMask Network](/img/metamask_network.png)](/img/metamask_network.png)

Make sure that your `RPC Url` and `Chain ID` are set correctly and match the settings you've chosen running your JSON-RPC endpoint. The `Network Name` can be whatever you see fit.

If you wish to use additional EVM chains with Metamask, you can simply add more Custom RPC networks, as long as they have a unique `Chain ID` and `RPC Url`. Once this is done, you can start using MetaMask to manage your EVM wallet or issue/sign transactions with any dApp running on that network. 

### Remix 

If you also want to use the [Remix IDE](https://remix.ethereum.org/) to deploy any regular Solidity Smart Contract, you should set the environment as **Injected Web3**, which should then connect with your MetaMask wallet.

## Hardhat

[Hardhat](https://hardhat.org/) is a commandline toolbox that allows you to deploy, test, verify, and interact with Solidity smart contracts on an EVM chain. EVM chains running on IOTA Smart Contracts are compatible with Hardhat; simply make sure you add the correct network parameters to your `hardhat.config.js`, for example:

```javascript
networks: {
    local: {
        url: 'http://localhost:8545',
        chainId: 1074,
        accounts: [privkey],
        timeout: 60000
    }
}
```

:::caution

Currently, there is no validation service available for EVM/Solidity smart contracts on IOTA Smart Contracts, which is often offered through block explorer API's.

:::


## Ethers.js/Web3.js

As long as you input the right configuration parameters for the JSON-RPC endpoint to talk to, [Ethers.js](https://docs.ethers.io/) and [Web3.js](https://web3js.readthedocs.io/) are also compatible with EVM chains on IOTA Smart Contracts. Alternatively you can let both interact through MetaMask instead so that it uses the network as configured in MetaMask. For more information on this, read their documentation.

## Other Tooling

Most other tooling available will be compatible as well as long as you enter the correct `Chain ID` and `RPC Url`. 
