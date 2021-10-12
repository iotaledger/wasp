---
keywords:
- ISCP
- Smart Contracts
- TestNet
description: A public testnet for developers to try out smart contracts 
image: /img/logo/WASP_logo_dark.png
---

# Testnet

The testnet is deployed for the community to use for testing and interacting with smart contracts. 

## Introduction

This testnet is deployed with our own GoShimmer cluster dedicated to backing smart contracts. There are multiple committee nodes that do the work for the chain as well as multiple access nodes that are exposed via the endpoints listed below. We do throttle the endpoints to prevent overloading the testnet because we are looking for functionality testing more than stress testing. 

<!--
  1. Talk about what the testnet is for
  2. List the available endpoints
  3. Have examples of deploying and interacting with a smart contract
-->

## Endpoints

The testnet can be accessed via a series of endpoints that hve been made available. 

- https://wasp.sc.iota.org
  - The Wasp dashboard to explore the node configuration and view peering/chain configurations
  - https://wasp.sc.iota.org/api/
    - The Wasp api to deploy and interact with smart contracts
    - https://wasp.sc.iota.org/api/info
      - Information about the Wasp access node
- https://goshimmer.sc.iota.org
  - The GoShimmer dashboard to explore the Tangle backing our smart contract testnet
  - https://goshimmer.sc.iota.org/api/
    - The GoShimmer api to interact with the Tangle directly
    - https://goshimmer.sc.iota.org/api/info
      - Information about GoShimmer via the API
- https://demo.sc.iota.org
  - Our FairRoulette demo application to see a live smart contract

## Smart Contract deployment and interaction

You will need to initialize `wasp-cli` in order to create a seed that will be used to generate addresses.

```
wasp-cli init
```

Now we need to tell `wasp-cli` how to reach our GoShimmer node.

```
wasp-cli set goshimmer.api goshimmer.sc.iota.org/api
```

In order to deploy a smart contract you will need some funds. The wasp-cli tool makes this easy on our testnet. Run the following command to request some funds.

```
wasp-cli request-funds
```

