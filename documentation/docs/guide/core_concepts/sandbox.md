---
description: Smart Contracts can only interact with the world by using the Sandbox interface which provides limited and deterministic access to the state through a key/value storage abstraction.
image: /img/sandbox.png
keywords:
- smart contracts
- sandbox
- interface
- storage abstraction
- explanation
---

# Sandbox Interface

A smart contract's access to the world has to be limited. Imagine a smart contract that would directly tap into a weather forecast website: as the weather changes, the result of the contract's execution changes. This smart contract is not deterministic anymore, meaning that you cannot reproduce the result to verify it.

The access to the chain's state has to be curated, too. The owner of the chain and developers of individual smart contracts are not necessary the same entity, and a single malicious contract could ruin the whole chain if not limited to its own space. Instead of working on the state as a whole, each smart contract can only modify a part of it.

The only way for smart contracts to access the data is to use the Sandbox interface. It provides them only with deterministic data and exposes the state as a structure of key/value pairs, only the ones the smart contract has access to.

![Sandbox](/img/sandbox.png)

Besides reading and writing to the state, the Sandbox interface allows smart contracts to access:

- The [ID](./accounts/how-accounts-work#agent-id) of the contract.
- The details of the current request or view call.
- The balances owned by the contract.
- The ID of whoever had deployed the contract.
- The timestamp of the current block.
- Cryptographic utilities like hashing, signature verification, and so on.
- The [events](../schema/events.mdx) dispatch.
- Entropy, which emulates randomness in an unpredictable yet deterministic way.
- Logging, which is usually only used for debugging in a test environment.
