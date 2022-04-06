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

The only way a smart contract can interact with the world (access the state, call other smart contracts or send transactions) is by using the Sandbox interface.

The Sandbox provides limited and deterministic access to the state through a key/value storage abstraction.

![Sandbox](/img/sandbox.png)

Besides reading and writing to the contract state, the Sandbox interface allows smart contracts to access:

- the AgentID of the contract
- the details of the current function invocation (request or view call)
- the balances owned by the contract
- the AgentID of whoever deployed the contract
- the timestamp of the current block
- cryptographic utilities (hashing, verify signatures, obtain addresses from public keys, etc)
- events dispatch
- Entropy (deterministic randomness)
- logging (usually only used for debugging when testing)
