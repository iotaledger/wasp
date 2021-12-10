---
keywords:
- Smart Contracts
- TestNet
description: A public testnet for developers to try out smart contracts 
image: /img/logo/WASP_logo_dark.png
---

# Testnet

The testnet is deployed for the community to use for testing and interacting with smart contracts. 

:::caution
To make the barrier to entry for trying out the testnet as low as possible we decided to set all possible fees to zero. Since we expect that this decision will pollute the chain quite quickly we’ll perform regular, unscheduled resets of this network if needed.
:::

## Introduction

This testnet is deployed with our own GoShimmer cluster dedicated to backing smart contracts. There are multiple committee nodes that do the work for the chain as well as multiple access nodes that are exposed via the endpoints listed below. We do throttle the endpoints to prevent overloading the testnet because we are looking for functionality testing more than stress testing. 

<!--
  1. Talk about what the testnet is for
  2. List the available endpoints
  3. Have examples of deploying and interacting with a smart contract
-->

## Endpoints

The testnet can be accessed via a series of endpoints that have been made available. 

- https://wasp.sc.iota.org . **Username**: wasp **Password**: wasp
  - The Wasp dashboard to explore the node configuration and view peering/chain configurations
  - https://api.wasp.sc.iota.org
    - The Wasp api to deploy and interact with smart contracts
    - https://api.wasp.sc.iota.org/info
      - Information about the Wasp access node
    - https://api.wasp.sc.iota.org/doc
      - API reference for the Wasp API
  - https://evm.wasp.sc.iota.org
    - The URL to interact with the Ethereum Virtual Machine on our testnet
  - https://explorer.wasp.sc.iota.org
    - The URL to interact with the Ethereum Block Explorer on our testnet
- https://goshimmer.sc.iota.org
  - The GoShimmer dashboard to explore the Tangle backing our smart contract testnet
  - goshimmer.sc.iota.org:5000
    - The TXStream TCP port to use for the `nodeconn` address with Wasp nodes.
  - https://api.goshimmer.sc.iota.org/
    - The GoShimmer api to interact with the Tangle directly
    - https://api.goshimmer.sc.iota.org/info
      - Information about GoShimmer via the API
- https://demo.sc.iota.org
  - Our FairRoulette demo application to see a live smart contract in action
- https://metrics.wasp.sc.iota.org
  - System metrics

## Configuring _wasp-cli_

You will need to initialize `wasp-cli` in order to create a seed that will be used to generate addresses.

```
wasp-cli init
```

Now we need to tell `wasp-cli` how to reach our GoShimmer node.

```
wasp-cli set goshimmer.api https://api.goshimmer.sc.iota.org
```

In order to deploy a smart contract you will need some funds. The wasp-cli tool makes this easy on our testnet. Run the following command to request some funds.

```
wasp-cli request-funds
```

We need to let _wasp-cli_ know how to reach _Wasp_ by configuring the API address.

```
wasp-cli set wasp.0.api https://api.wasp.sc.iota.org
```

Now you need to set the chain ID in _wasp-cli_ so that the correct chain can be found. Yo can find the ChainID by navigating to the (chains)[https://wasp.sc.iota.org/chains] page of the wasp dashboard. Click on the ChainID of the chain you will be able to copy the ChainID from the next page. It will be formatted like `jaSDxeZNtum7kLuRg8oWQ6nXKgYjb3XVq7yiwnvtUG3C`. 

Use the ChainID to tell _wasp-cli_ which chain you want to interact with. 

```
wasp-cli set chains.testchain jaSDxeZNtum7kLuRg8oWQ6nXKgYjb3XVq7yiwnvtUG3C
wasp-cli set chain testchain
```

On the __testchain__ we have deployed a FairRoulette game that you can use to make sure your configuration is correct.

```
wasp-cli --verbose chain post-request fairroulete placeBet string number int 2
```

For simplicity, here is the full set of commands to configure _wasp-cli_.

```
wasp-cli init
wasp-cli set goshimmer.api https://api.goshimmer.sc.iota.org
wasp-cli request-funds
wasp-cli set wasp.0.api https://api.wasp.sc.iota.org
wasp-cli set chains.testchain jaSDxeZNtum7kLuRg8oWQ6nXKgYjb3XVq7yiwnvtUG3C
wasp-cli set chain testchain
```

## Interact with EVM

We have deployed an experimental EVM chain that you can interact with. To begin, add a custom network to Metamask with the following configuration:

| Key | Value |
| --- | ----- |
| RPC URL | https://evm.wasp.sc.iota.org |
| Chain ID | 1074 |
| Block Explorer URL | https://explorer.wasp.sc.iota.org |

It should look similar to this image. 

![MetaMask](/img/metamask_testnet.png)

:::note
The other values (network name and currency symbol) can be whatever value you like. 
:::
