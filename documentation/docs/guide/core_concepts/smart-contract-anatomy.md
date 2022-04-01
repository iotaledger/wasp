---
description: Each smart contract instance has a program with a collection of entry points and a state. 
image: /img/tutorial/SC-structure.png
keywords:
- smart contracts
- structure
- state
- entry points
- Wasm
- explanation
---

# Anatomy of a Smart Contract

Smart contracts are programs which are immutably stored in the chain.

The logical structure of IOTA Smart Contracts is independent of the VM type you
use, be it a _Wasm_ smart contract or any other VM type.

![Smart Contract Structure](/img/tutorial/SC-structure.png)

## Identifying a Smart Contract

Each smart contract on the chain is identified by its name hashed into 4 bytes
and interpreted as `uint32` value: the `hname`.

For example, the `hname` of the root contract is `0xcebf5908`, the unique identifier of the
`root` contract in every chain. The exception is the `_default` contract which always has `hname` `0x00000000`.

Each smart contract instance has a program with a collection of entry points and
a state. An entry point is a function of the program through which the program
can be invoked.

Depending on the type of the entry point there are several ways to invoke an entry point: a request, a call and a view call

The smart contract program can access its state and account through an interface
layer called the _Sandbox_.

## State

The smart contract state is its data, with each update stored on the chain. The
state can only be modified by the smart contract program itself. There are two
parts of the state:

- A collection of key/value pairs called the `data state`. Each key and value
  are byte arrays of arbitrary size (there are practical limits set by the
  database, of course). The value of the key/value pair is always retrieved by
  its key.
- A collection of `color: balance` pairs called the `account`. The account
  represents the balances of tokens of specific colors controlled by the smart
  contract. Receiving and spending tokens into/from the account means changing
  the account's balances.

Only the smart contract program can change its data state and spend from its
account. Tokens can be sent to the smart contract account by any other agent on
the ledger, be it a wallet with an address or another smart contract.

See [Accounts](../core_concepts/accounts/how-accounts-work.md) for more info on sending and receiving tokens.

## Entry Points

There are two types of entry points:

- _Full entry points_(or just _entry points_): These functions can modify
  (mutate) the smart contract's state.
- _View entry points_(or _views_): These are read-only functions. They are only used
  to retrieve the information from the smart contract state. They cannot
  modify the state, i.e. are read-only calls.

## Execution Results

After a request to a Smart Contract is executed (a call to a "full entry point"),
a `receipt` will be added to the [`BlockLog`](../core_concepts/core_contracts/blocklog.md)
detailing the execution results of said request: whether it was successful, the block it was
included in, and other information. Any events dispatched by the Smart Contract in context of
this execution will also be added to the BlockLog.

## Error Handling

When a smart contract execution is interrupted for some reason (exception), or it produces an
error (missing parameter, or other inconsistency), the funds will be refunded to the caller,
except the fees. Any error that resulted from the SC execution can be viewed on the contract
`receipt` (present in the [`BlockLog`](../core_concepts/core_contracts/blocklog.md)).
