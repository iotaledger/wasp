---
keywords:
- ISCP
- Smart Contracts
- Core Concepts
- Table of Contents
description: IOTA Smart Contract Protocol Documentation Overview
image: /img/logo/WASP_logo_dark.png
---
# ISCP

ISCP stands for IOTA Smart Contract Protocol. It is IOTA's solution for running smart contracts on top of the IOTA tangle. With ISCP we bring scalable and flexible smart contracts into the IOTA ecosystem by allowing anyone to spin up a smart contract blockchain and anchor it to the IOTA tangle. 

Allowing multiple blockchains to anchor to the tangle will solve several problems with smart contracts.

### Scaling and Fees

Due to the ordered structure and execution time of a smart contract, a single blockchain can only handle so many transactions per second before it needs to decide on which transactions it needs to postpone until other blocks are produced, as there is no processing power available for them on that chain. This eventually results in high fees on many chains, and limited functionality. 

As we allow many chains (which can communicate with other chains) to be anchored to the IOTA tangle, we can simply add additional chains once this becomes a problem. This results in lower fees over solutions that just use a single blockchain for all their smart contracts. 

### Custom Chains

Given that anyone can start a new chain, and set the rules for that chain, a lot is possible. Not only do you have full control over how the fees are handled on the chain you set up, but you also have full control over the validators and access to your chain. You can even spin up a private blockchain with no public data besides the state hash that is anchored to the main IOTA tangle. This allows parties that need private blockchains to use the security of the public IOTA network without actually exposing their blockchain to the public.

### Flexibility

You can run multiple types of virtual machines on any chain. We are starting with [Rust/WASM](https://rustwasm.github.io/docs/book/) based smart contracts, followed by [Solidity/EVM](https://docs.soliditylang.org/en/v0.8.6/) based smart contracts, but eventually all kinds of virtual machines can be supported in a ISCP chain depending on the demand. 

ISCP is more complex compared to conventional smart contracts, but this provides freedom and flexibility to allow the usage of smart contracts in a wide range of use cases.

## What is Wasp?

Wasp is the node software that we have built to let you validate smart contracts as a part of a committee while using a virtual machine of your choice. Multiple Wasp nodes connect and form a committee of validators. When they reach consensus on a virtual machine state-change, they anchor that state change to the IOTA tangle, making it immutable. 
