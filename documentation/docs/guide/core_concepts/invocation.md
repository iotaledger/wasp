---
description: Smart contracts can call each other if they are deployed on the same chain. If you need to invoke a smart contract from outside of the chain, you need to make a smart contract request.
image: /img/logo/WASP_logo_dark.png
keywords:
- smart contracts
- requests
- on-ledger
- off-ledger
- calls
- invocation
- explanation
---

# Smart Contracts Invocation

Just like any other computer program, a smart contract will lie dormant until someone or something would instruct it to activate. In case of smart contracts, the basic way to activate them is to call one of their [entry points](./smart-contract-anatomy.md#entry-points). It is the same as calling a program's function, and it will take a set of instructions of the smart contract and execute it over the current chain's state. View entry points can only read the state, and full entry points can both read and write to it.

Calls just execute the code, on their own they have no security checks. There is no harm in calling view entry points, but full entry points cannot be directly exposed to the outside world. Instead you need to wrap a call into a request, cryptographically sign it, and submit it to the [consensus](./consensus.md) procedure to let the chain's committee evaluate it and engrave the outcome of its execution into a new state update.

When you make a request, the committee will either take it into work and execute the wrapped call fully, or it will reject the request with all its potential changes, never modifying the state only half-way through. This means that every single request is an *atomic operation*.

Smart contract calls are deterministic and synchronous, meaning that they always produce the same result and that all instructions are executed one immediately after another. By extension, if a smart contract calls another smart contract, the resulting set of instructions is also deterministic and synchronous, meaning that for a request it makes no difference if a smart contract's entry point contains the whole set of instructions or if it is composed of multiple smart contracts of the chain. Being able to combine smart contracts in this way is called *synchronous composability*.

## Requests

A request contains a call to a smart contract and a signature of the sender (who is also the owner of the assets and funds that are going to be processed within the request). Unlike calls, requests are not executed immediately. Instead, they have to wait until the chain's validator nodes include them into a request batch. This means that requests unlike the calls have a delay and are executed in an unpredictable order, but they are subject to the protection mechanisms that allow ISC to function.

Requests are not only for humans — smart contracts can create requests too. You could request an execution of a smart contract that creates a request to, for example, a third-party decentralized exchange which would convert the user's funds from one currency to another and send them back through another request. This is called *asynchronous composability*.

### On-Ledger

An on-ledger request is a Layer 1 transaction that validator nodes retrieve from the Tangle. The Tangle acts as an arbiter between users and chains and guarantees that the transaction is valid, making it the only way to transfer assets to a chain or between the chains, albeit it is the slowest way to invoke a smart contract.

### Off-Ledger

If you have all necessary assets on a chain already, you can send a request directly to that chain's validator nodes. This way you don't have to wait for the Tangle to process the message, which makes the overall confirmation time much, much shorter. Due to the shorter delay, you should prefer sending off-ledger requests over on-ledger requests, unless you have to move assets between chains or Layer 1 accounts.