---
keywords:
- Smart Contracts
- EVM
- Solidity
- limitations
- fees
- sand-boxed
description: EVM based smart contract limitations. The current implementation is fully sand-boxed and not aware of IOTA or IOTA Smart Contracts. You start an EVM chain with a new supply of EVM specific tokens assigned to a single address.
image: /img/logo/WASP_logo_dark.png
---
# EVM Limitations within IOTA Smart Contracts

The current experimental EVM support for IOTA Smart Contracts allows developers to get a preview of EVM based smart contract solutions on top of IOTA. There are some limitations you should be aware of at the time, which we will be addressing in the months to come:

- **The current implementation is fully sand-boxed and not aware of IOTA or IOTA Smart Contracts**. It currently can not communicate with non-EVM smart contracts, nor can it interact with assets outside the EVM sandbox yet.
- **You start an EVM chain with a new supply of EVM specific tokens assigned to a single address** (the main token on the chain which is used for gas as well, comparable to ETH on the Ethereum network). These new tokens are in no way connected to IOTA, or any other token, but are specific for that chain for now.
- **Because EVM runs inside an IOTA Smart Contracts smart contract, any fees that need to be paid for that IOTA Smart Contracts smart contract have to be taken into account** while invoking a function on that contract. To support this right now the JSON-RPC gateway uses the wallet account connected to it. 
- **You need to manually deposit some IOTA to the chain** you are using to be able to invoke these functions. We are planning to resolve this at a later phase in a more user-friendly way.

Overall these are temporary solutions, the next release of the IOTA Smart Contracts will see a lot of these improved or resolved.
