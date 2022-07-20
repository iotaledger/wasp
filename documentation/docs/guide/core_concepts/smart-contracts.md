---
description: Smart contracts are applications you can trust that run on a distributed network with multiple validators all executing and validating the same code.
image: /img/Banner/banner_wasp_core_concepts_smart_contracts.png
keywords:
- smart contracts
- blockchain
- parallel
- scaling
- explanation
---
# Smart Contracts

![Wasp Node Smart Contracts](/img/Banner/banner_wasp_core_concepts_smart_contracts.png)

Smart contracts are software applications that run on a distributed network with multiple validators executing and validating the same code. This prevents tampering with the execution of the program and ensures that the application behaves as expected. 

## Applications You Can Trust

Smart contracts can be trusted because it is guaranteed that the code that is being executed will never change.
The rules of the smart contract define what the contract can and can not do, making it a decentralized and a predictable decision maker.

Smart contracts are used for all kinds of purposes. A recurring reason to use a smart contract would be to automate a certain action without needing a centralized entity to enforce it. For example, a smart contract could exchange a certain amount of IOTA tokens for a certain amount of land ownership. That smart contract would accept both the IOTA tokens and the land ownership and predictably exchange them between both parties. It excludes the risk of one of the parties not delivering on their promise: with a smart contract, code is law.

## Scalable Smart Contracts

Anyone could deploy a smart contract to a public smart contract chain.
Once it is deployed, nobody will be able to change or delete it.
Smart contracts can communicate with one another, and they can also expose public functions that can be manually invoked to trigger their execution or check their state.

Smart contracts do not run on just a single computer.
Instead, each validator of the blockchain has to execute it, compare the results with others, and synchronize the state of the network.
These messages are carried over the Internet and introduce delays that cannot be solved with quicker software or faster computers.
In single-chain platforms, this issue only gets worse if one tries to add more validators to the network.
With enough requests, any traditional blockchain network will get congested, and its execution fees will ramp up.

As IOTA Smart Contracts run many independent chains, it spreads out the load and creates a network of a much larger scale. At the same time, it provides advanced means of communication between its chains and preserves the ability to create complex, composed smart contracts.
