---
keywords:
- ISCP
- Smart Contracts
- Chain
- EVM
- Solidity
- Tooling
description: Available tooling for EVM chains
image: /img/logo/WASP_logo_dark.png
---
# EVM Tooling

EVM on ISCP has been integrated in such a way that existing EVM tooling is compatible and can be used directly with a ISCP chain running EVM as long as a couple of things are taken into account.

## Tooling considerations

 1. Please make sure you use the correct JSON/RPC endpoint URL in your tooling for your chain. If you run locally this will simply be `localhost:8545`
 2. Please make sure you are using the right `Chain ID` as configured while starting the JSON/RPC service. If you did not explicitly define this while starting the service the default Chain ID will be `1074`. 
 3. Fees are being handled on the ISCP chain level, not EVM level, because of this you can simply always use a gas price of 0 on EVM level at this point in time.

:::caution

re-using a existing Chain ID is not recommended and can be a security risk. For any serious chain you will be running make sure you register a unique Chain ID on [Chainlist](https://chainlist.org/) and use that instead of the default.

:::

## Wasp-cli

The Wasp CLI has some very basic functionality to manage a EVM chain. Given the compatibility with existing tooling only the basics are covered to get started with ISCP and EVM. You can currently either run a JSON/RPC server or deploy the EVM Chain itself on a ISCP chain. To see the available options and configuration parameters simply run

```
wasp-cli chain evm
```

## MetaMask

[MetaMask](https://metamask.io/) is a popular EVM compatible wallet running in a browser extension that allows you to let your wallet interact with web applications utilizing a EVM chain (dApps). 

To use your EVM chain with MetaMask simply open up MetaMask and click on the network drop-down list at the very top. At the bottom of this list you'll see the option `Custom RPC`, click on this. For a local setup use the values as shown in the image below:

![MetaMask Network](/img/metamask_network.png)

Make sure that your `RPC Url` and `Chain ID` are set correctly and match the settings you've chosen running your JSON/RPC endpoint.

If you wish to use additional EVM chains with Metamask you can simply add more Custom RPC networks, as long as they have a unique `Chain ID` and `RPC Url`. Once this is done you can start using MetaMask to manage your EVM wallet or issue/sign transactions with any dApp running on that network. 

## Hardhat

[Hardhat](https://hardhat.org/) is a commandline toolbox that allows you to deploy, test, verify and interact with Solidity smart contracts on an EVM chain. EVM chains running on ISCP are compatible with Hardhat; Simply make sure you add the correct network parameters to your `hardhat.config.js`, for example:

```javascript=
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

Currently there is no validation service available yet for EVM/Solidity smart contracts on ISCP which is often offered through block explorer API's.

:::


## Ethers.js / Web3.js

[Ethers.js](https://docs.ethers.io/) and [Web3.js](https://web3js.readthedocs.io/) are also compatible with EVM chains on ISCP as long as you input the right configuration parameters for the JSON/RPC endpoint to talk to. Alternatively you can let both interact through MetaMask instead so that it uses the network as configured in MetaMask. For more information on this read their documentation.

## Other tooling

Most other tooling available will be compatible as well as long as you enter the correct `Chain ID` and `RPC Url`. 
