---
description: Smart contracts can be invoked through their entry points, from outside via a request, or from inside via a call.
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

# Calling a Smart Contract

Just like any other computer program, a smart contract will lie dormant until someone or something would instruct it to activate. In the case of smart contracts, the basic way to activate them is to call one of their [entry points](./smart-contract-anatomy.md#entry-points). It is the same as calling a program's function, and it will take a set of instructions of the smart contract and execute it over the current chain's state. View entry points can only read the state, and full entry points can both read and write to it.

In order to invoke a smart contract from outside the chain, the _sender_ (some entity that needs to be identified by a private/public key pair) has to wrap the call to the entry point into a _request_.
The request has to be cryptographically signed and submitted to the [consensus](./consensus.md) procedure to let the chain's committee evaluate it and engrave the outcome of its execution into a new state update.

Upon receiving a request, the committee will either take it into work and execute the wrapped call fully, or it will reject the request with all its potential changes, never modifying the state only half-way through.
This means that every single request is an atomic operation.

After being invoked by a request, the smart contract code is allowed to invoke entry points of other smart contracts on the same chain; i.e. it can _call_ other smart contracts.

Smart contract calls are deterministic and synchronous, meaning that they always produce the same result and that all instructions are executed one immediately after another.
By extension, if a smart contract calls another smart contract, the resulting set of instructions is also deterministic and synchronous, meaning that for a request it makes no difference if a smart contract's entry point contains the whole set of instructions or if it is composed by multiple calls to different smart contracts of the chain.
Being able to combine smart contracts in this way is called *synchronous composability*.

---

## Requests

A request contains a call to a smart contract and a signature of the sender (who is also the owner of the assets and funds that are going to be processed within the request).
Unlike calls between smart contracts, requests are not executed immediately.
Instead, they have to wait until the chain's validator nodes include them into a request batch.
This means that requests have a delay and are executed in an unpredictable order.

Requests are not only sent by humans; smart contracts can create requests too.
For example, a user could send a request to a smart contract that, in turn, creates a request to a third-party decentralized exchange which would convert the user's funds from one currency to another and send them back through another request.
This is called *asynchronous composability*.

### On-Ledger

An on-ledger request is a Layer 1 transaction that validator nodes retrieve from the Tangle. The Tangle acts as an arbiter between users and chains and guarantees that the transaction is valid, making it the only way to transfer assets to a chain or between the chains, albeit it is the slowest way to invoke a smart contract.

### Off-Ledger

If all necessary assets are in the chain already, it is possible to send a request directly to that chain's validator nodes.
This way it is not necessary to wait for the Tangle to process the message, making the overall confirmation time much shorter.
Due to the shorter delay, off-ledger requests are preferred over on-ledger requests, unless it is required to move assets between chains or Layer 1 accounts.

---

## Gas

Gas is used to express the "cost" of running a request in a chain. Each operaton (arithmetics, write to disk, dispatch events, etc) has an associated gas cost.

In order for users to specify how much they're willing to pay for a request, they need to specify a `GasBudget` in the request. This gas buget is the "maximum of operations that this request can execute", and will be charged as a fee based on the chain's current [Fee Policy](core_contracts/governance.md#fee-policy).

The funds to cover the gas used will be charged directly from the user's on-chain account.

---

## Allowance

Any funds sent to the chain via (on-ledger) requests are credited to the sender's account.

In order for contracts to use funds owned by the *caller*, the *caller* must specify an `Allowance` in the request. Contracts can then claim any of the allowed funds by using the sandbox `TransferAllowedFunds` function.

The Allowance properly looks like the following:

```go
{
  FungibleTokens: {
    BaseToken: uint64
    NativeTokens: [{TokenID, uint256}, ...]
  }
  NFTs: [NFTID,...]
}
```
