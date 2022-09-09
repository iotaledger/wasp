---
description: A public testnet for developers to try out smart contracts
image: /img/logo/WASP_logo_dark.png
keywords:
- Smart Contracts
- TestNet

---

# Testnet

The testnet is deployed for the community to use for testing and interacting with smart contracts.

:::caution unscheduled network resets
To make the barrier to entry for trying out the testnet as low as possible, we decided to set all possible fees to zero.
Since we expect this decision will pollute the chain quickly, we’ll perform regular, unscheduled network resets if
needed.
:::

## Introduction

This testnet is deployed with our GoShimmer cluster that backs smart contracts. Multiple committee nodes do the work for
the chain and multiple access nodes are exposed via the endpoints listed below. We do throttle the endpoints to prevent
overloading the testnet because we are looking for functionality testing more than stress testing.

## Endpoints

You can access the testnet via the following endpoints:

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
- https://wasp.sc.iota.org/metrics
    - System metrics

## Configuring `wasp-cli`

### Initialize `wasp-cli`

You will need to initialize `wasp-cli` to create a seed that will be used to generate addresses.

```shell
wasp-cli init
```

### Set the GoShimmer API URL

Next, you should tell `wasp-cli` how to reach our GoShimmer node.

```shell
wasp-cli set goshimmer.api https://api.goshimmer.sc.iota.org
```

### Request Funds

To deploy a smart contract, you will need some funds. The `wasp-cli` tool makes this easy on our testnet. Run the
following command to request some funds.

```shell
wasp-cli request-funds
```

### Configure the Wasp API URL

Next, you need to let `wasp-cli` know how to reach _Wasp_ by configuring the API address.

```shell
wasp-cli set wasp.0.api https://api.wasp.sc.iota.org
```

### Set the Chain ID

You will need to set the chain ID in `wasp-cli` to find the correct chain. You can find the ChainID by navigating to
the [chains](https://wasp.sc.iota.org/chains) page of the wasp dashboard. Click on the ChainID of the chain. You will be
able to copy the ChainID from the next page. It will be formatted like `jaSDxeZNtum7kLuRg8oWQ6nXKgYjb3XVq7yiwnvtUG3C`.

Use the ChainID to tell `wasp-cli` which chain you want to interact with:

```shell
wasp-cli set chains.testchain jaSDxeZNtum7kLuRg8oWQ6nXKgYjb3XVq7yiwnvtUG3C
wasp-cli set chain testchain
```

### Test Your Chain

We have deployed a FairRoulette game on the __testchain__. You can use to ensure your configuration is correct.

```shell
wasp-cli --verbose chain post-request fairroulete placeBet string number int 2
```

### Putting It All Together

For simplicity, here is the full set of commands to configure `wasp-cli`.

```shell
wasp-cli init
wasp-cli set goshimmer.api https://api.goshimmer.sc.iota.org
wasp-cli request-funds
wasp-cli set wasp.0.api https://api.wasp.sc.iota.org
wasp-cli set chains.testchain jaSDxeZNtum7kLuRg8oWQ6nXKgYjb3XVq7yiwnvtUG3C
wasp-cli set chain testchain
```

## Interact with EVM

We have deployed an experimental EVM chain that you can interact with. To begin, add a custom network to Metamask with
the following configuration:

| Key                | Value                             |
|--------------------|-----------------------------------|
| RPC URL            | https://evm.wasp.sc.iota.org      |
| Chain ID           | 1074                              |
| Block Explorer URL | https://explorer.wasp.sc.iota.org |

It should look similar to this image.

![MetaMask](/img/metamask_testnet.png)

:::note

The other values (network name and currency symbol) can be whatever value you like.

:::



