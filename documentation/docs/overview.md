---
description: IOTA Smart Contracts allow you to run smart contracts on top of the IOTA Tangle.
image: /img/logo/WASP_logo_dark.png
keywords:
- smart contracts
- core concepts
- scaling
- flexibility
- explanation
---

# IOTA Smart Contracts

The IOTA Smart Contracts is a protocol that brings scalable and flexible smart contracts into the IOTA ecosystem. It allows anyone to spin up a smart contract blockchain and anchor it to the IOTA Tangle, a design that offers a multiple advantages.

## ISC Advantages

### Scaling and Fees

Due to the ordered structure and execution time of a smart contract, a single blockchain can only handle so many transactions per second before it needs to decide on which transactions it needs to postpone until other blocks are produced, as there is no processing power available for them on that chain. This eventually results in high fees on many chains, and limited functionality. 

As ISC allows many chains to be anchored to the IOTA Tangle and lets them communicate with each other, we can add additional chains once this becomes a problem. This results in lower fees over solutions that use a single blockchain for all their smart contracts.

### Custom Chains

Given that anyone can start a new chain and set its rules, a lot is possible. Not only do you have full control over how the fees are handled on the chain you set up, but you also have full control over the validators and access to your chain. You can even spin up a private blockchain with no public data besides the state hash that is anchored to the main IOTA Tangle. This allows parties that need private blockchains to use the security of the public IOTA network without actually exposing their blockchain to the public.

### Flexibility

You can run multiple types of virtual machines on any chain. We are starting with 
[Rust/Wasm](https://rustwasm.github.io/docs/book/)-based smart contracts, followed by 
[Solidity/EVM](https://docs.soliditylang.org/en/v0.8.6/)-based smart contracts, but eventually all kinds of virtual machines can be supported in a IOTA Smart Contract chain depending on the demand. 

IOTA Smart Contracts are more complex compared to conventional smart contracts, but this provides freedom and flexibility to allow the usage of smart contracts in a wide range of use cases.

## Wasp

Wasp is the node software that we have built to let you validate smart contracts as a part of a committee while using a virtual machine of your choice. Multiple Wasp nodes connect and form a committee of validators. When they reach consensus on a virtual machine state-change, they anchor that state change to the IOTA tangle, making it immutable. 

## Feedback

We are eager to receive your feedback about your experiences with the IOTA Smart Contracts Beta. You can use [this form](https://docs.google.com/forms/d/e/1FAIpQLSd4jcmLzCPUNDIijEwGzuWerO23MS0Jmgzszgs-D6_awJUWow/viewform) to share your developer experience with us. This feedback will help us improve the product in future releases.